package ru.docflow.notification;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.rabbitmq.client.*;

import java.io.IOException;
import java.util.logging.Logger;

public class NotificationConsumer {
    private static final Logger log    = Logger.getLogger(NotificationConsumer.class.getName());
    private static final String EXCHANGE = "notifications";
    private static final String QUEUE    = "notifications.queue";

    private final Connection  connection;
    private final Channel     channel;
    private final EmailSender emailSender;
    private final ObjectMapper mapper = new ObjectMapper();

    public NotificationConsumer(String amqpUrl, EmailSender emailSender) throws Exception {
        ConnectionFactory factory = new ConnectionFactory();
        factory.setUri(amqpUrl);
        this.connection  = factory.newConnection();
        this.channel     = connection.createChannel();
        this.emailSender = emailSender;

        channel.exchangeDeclare(EXCHANGE, BuiltinExchangeType.FANOUT, true);
        channel.queueDeclare(QUEUE, true, false, false, null);
        channel.queueBind(QUEUE, EXCHANGE, "");
    }

    public void start() throws IOException {
        log.info("notification-service: listening on queue " + QUEUE);
        channel.basicConsume(QUEUE, false, (tag, delivery) -> {
            try {
                NotificationEvent evt = mapper.readValue(delivery.getBody(), NotificationEvent.class);
                String subject = buildSubject(evt);
                String body    = buildBody(evt);
                emailSender.send(evt.recipientEmail, subject, body);
                channel.basicAck(delivery.getEnvelope().getDeliveryTag(), false);
            } catch (Exception e) {
                log.severe("notification consumer error: " + e.getMessage());
                try {
                    channel.basicNack(delivery.getEnvelope().getDeliveryTag(), false, false);
                } catch (IOException ignored) {}
            }
        }, tag -> log.warning("consumer cancelled: " + tag));
    }

    public void close() {
        try {
            channel.close();
            connection.close();
        } catch (Exception ignored) {}
    }

    // ─── message builders ────────────────────────────────────────────────────

    private String buildSubject(NotificationEvent evt) {
        return switch (evt.action) {
            case "DOC_APPROVED"          -> "DocFlow: документ завизирован";
            case "DOC_REJECTED"          -> "DocFlow: документ не принят";
            case "DOC_DELEGATED"         -> "DocFlow: вам делегирован документ";
            case "DOC_APPROVAL_REQUESTED"-> "DocFlow: запрос на согласование";
            case "DOC_SENT_FOR_APPROVAL" -> "DocFlow: документ отправлен на визирование";
            default                      -> "DocFlow: уведомление";
        };
    }

    private String buildBody(NotificationEvent evt) {
        StringBuilder sb = new StringBuilder();
        sb.append("Здравствуйте, ").append(evt.recipientName).append("!\n\n");

        switch (evt.action) {
            case "DOC_APPROVED" ->
                sb.append("Ваш документ «").append(evt.documentTitle).append("» успешно завизирован.");
            case "DOC_REJECTED" ->
                sb.append("Ваш документ «").append(evt.documentTitle)
                  .append("» не принят.\nКомментарий: ").append(evt.details);
            case "DOC_DELEGATED" ->
                sb.append("Вам делегирован документ «").append(evt.documentTitle)
                  .append("».\nЗадача: ").append(evt.details);
            case "DOC_APPROVAL_REQUESTED" ->
                sb.append("В ваш отдел поступил запрос на согласование документа «")
                  .append(evt.documentTitle).append("».\nВопрос: ").append(evt.details);
            case "DOC_SENT_FOR_APPROVAL" ->
                sb.append("Документ «").append(evt.documentTitle)
                  .append("» отправлен на визирование начальнику отдела.");
            default ->
                sb.append(evt.details);
        }

        sb.append("\n\n-- DocFlow");
        return sb.toString();
    }
}
