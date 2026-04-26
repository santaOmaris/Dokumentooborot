package ru.docflow.notification;

import java.util.logging.Logger;

public class Main {
    private static final Logger log = Logger.getLogger(Main.class.getName());

    public static void main(String[] args) throws Exception {
        String amqpUrl   = getEnv("AMQP_URL",      "amqp://guest:guest@localhost:5672/");
        String smtpHost  = getEnv("SMTP_HOST",      "");
        int    smtpPort  = Integer.parseInt(getEnv("SMTP_PORT",  "587"));
        String smtpUser  = getEnv("SMTP_USER",      "");
        String smtpPass  = getEnv("SMTP_PASSWORD",  "");
        String smtpFrom  = getEnv("SMTP_FROM",      "noreply@docflow.local");

        EmailSender emailSender = new EmailSender(smtpHost, smtpPort, smtpUser, smtpPass, smtpFrom);
        NotificationConsumer consumer = new NotificationConsumer(amqpUrl, emailSender);
        consumer.start();

        log.info("notification-service started");

        Runtime.getRuntime().addShutdownHook(new Thread(() -> {
            log.info("notification-service: shutting down");
            consumer.close();
        }));

        // Держим поток живым — consumer работает в I/O потоке RabbitMQ
        Thread.currentThread().join();
    }

    private static String getEnv(String key, String fallback) {
        String v = System.getenv(key);
        return (v != null && !v.isBlank()) ? v : fallback;
    }
}
