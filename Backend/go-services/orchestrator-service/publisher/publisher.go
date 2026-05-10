package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"docflow.local/pkg/mq"
)

type Publisher struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func New(amqpURL string) (*Publisher, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("amqp dial: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("amqp channel: %w", err)
	}

	for _, ex := range []string{"audit", "notifications"} {
		if err = ch.ExchangeDeclare(ex, "fanout", true, false, false, false, nil); err != nil {
			conn.Close()
			return nil, fmt.Errorf("exchange declare %q: %w", ex, err)
		}
	}
	return &Publisher{conn: conn, ch: ch}, nil
}

func (p *Publisher) Close() {
	if p == nil {
		return
	}
	p.ch.Close()
	p.conn.Close()
}

func (p *Publisher) PublishAudit(evt mq.AuditEvent) {
	if p == nil {
		log.Printf("[AUDIT STUB] action=%s doc=%d actor=%s", evt.Action, evt.DocumentID, evt.ActorLogin)
		return
	}
	p.publish("audit", evt)
}

func (p *Publisher) PublishNotification(evt mq.NotificationEvent) {
	if p == nil {
		log.Printf("[NOTIFY STUB] to=%s action=%s", evt.RecipientEmail, evt.Action)
		return
	}
	p.publish("notifications", evt)
}

func (p *Publisher) publish(exchange string, payload interface{}) {
	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("publisher: marshal error: %v", err)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = p.ch.PublishWithContext(ctx, exchange, "", false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         body,
	}); err != nil {
		log.Printf("publisher: publish to %q error: %v", exchange, err)
	}
}
