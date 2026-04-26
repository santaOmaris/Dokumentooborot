package ru.docflow.notification;

public class NotificationEvent {
    public String recipientEmail;
    public String recipientName;
    public String action;
    public String documentTitle;
    public String details;

    // Jackson требует конструктор без аргументов
    public NotificationEvent() {}
}
