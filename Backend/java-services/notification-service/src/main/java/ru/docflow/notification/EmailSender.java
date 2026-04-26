package ru.docflow.notification;

import javax.mail.*;
import javax.mail.internet.InternetAddress;
import javax.mail.internet.MimeMessage;
import java.util.Properties;
import java.util.logging.Logger;

public class EmailSender {
    private static final Logger log = Logger.getLogger(EmailSender.class.getName());

    private final Session session;
    private final String from;
    private final boolean enabled;

    public EmailSender(String host, int port, String user, String password, String from) {
        this.from = from;
        this.enabled = host != null && !host.isBlank();

        if (!enabled) {
            log.warning("EmailSender: SMTP_HOST not set, email sending is disabled");
            session = null;
            return;
        }

        Properties props = new Properties();
        props.put("mail.smtp.host", host);
        props.put("mail.smtp.port", String.valueOf(port));
        props.put("mail.smtp.auth", "true");
        props.put("mail.smtp.starttls.enable", "true");

        session = Session.getInstance(props, new Authenticator() {
            @Override
            protected PasswordAuthentication getPasswordAuthentication() {
                return new PasswordAuthentication(user, password);
            }
        });
    }

    public void send(String to, String subject, String body) {
        if (!enabled) {
            log.info(String.format("[EMAIL STUB] to=%s subject=%s body=%s", to, subject, body));
            return;
        }
        try {
            MimeMessage msg = new MimeMessage(session);
            msg.setFrom(new InternetAddress(from));
            msg.setRecipients(Message.RecipientType.TO, InternetAddress.parse(to));
            msg.setSubject(subject, "UTF-8");
            msg.setText(body, "UTF-8");
            Transport.send(msg);
            log.info("Email sent to " + to);
        } catch (MessagingException e) {
            log.severe("Failed to send email to " + to + ": " + e.getMessage());
        }
    }
}
