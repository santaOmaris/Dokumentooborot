package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	db "collaboration-service/db/generated"
	"collaboration-service/service"
)

type AuditEvent struct {
	DepartmentID int32  `json:"department_id"`
	DocumentID   int32  `json:"document_id"`
	ActorLogin   string `json:"actor_login"`
	Action       string `json:"action"`
	Details      string `json:"details"`
}

func Run(ctx context.Context, amqpURL string, q *db.Queries) error {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return fmt.Errorf("amqp dial: %w", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("amqp channel: %w", err)
	}
	defer ch.Close()

	if err = ch.ExchangeDeclare("audit", "fanout", true, false, false, false, nil); err != nil {
		return fmt.Errorf("exchange declare: %w", err)
	}

	queue, err := ch.QueueDeclare("collaboration.audit", true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("queue declare: %w", err)
	}

	if err = ch.QueueBind(queue.Name, "", "audit", false, nil); err != nil {
		return fmt.Errorf("queue bind: %w", err)
	}

	msgs, err := ch.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume: %w", err)
	}

	log.Printf("audit consumer: listening on queue %q", queue.Name)

	for {
		select {
		case <-ctx.Done():
			return nil
		case d, ok := <-msgs:
			if !ok {
				return fmt.Errorf("amqp channel closed")
			}
			handleEvent(ctx, d, q)
		}
	}
}

func handleEvent(ctx context.Context, d amqp.Delivery, q *db.Queries) {
	var evt AuditEvent
	if err := json.Unmarshal(d.Body, &evt); err != nil {
		log.Printf("audit consumer: bad json: %v", err)
		d.Nack(false, false) // dead-letter, не переставляем в очередь
		return
	}

	if err := service.WriteAuditLog(ctx, q, evt.DepartmentID, evt.DocumentID, evt.ActorLogin, evt.Action, evt.Details); err != nil {
		log.Printf("audit consumer: db error: %v", err)
		d.Nack(false, true) // requeue
		return
	}

	d.Ack(false)
}
