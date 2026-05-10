package ru.docflow.notification;

import com.fasterxml.jackson.annotation.JsonAlias;

public class NotificationEvent {
    @JsonAlias("recipient_email")
    public String recipientEmail;

    @JsonAlias("recipient_name")
    public String recipientName;

    public String action;

    @JsonAlias("document_title")
    public String documentTitle;

    public String details;

    // Jackson требует конструктор без аргументов
    public NotificationEvent() {}
}
