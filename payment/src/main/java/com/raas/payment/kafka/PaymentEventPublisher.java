package com.raas.payment.kafka;

import com.raas.payment.config.KafkaTopics;
import com.raas.payment.kafka.events.PaymentResultEvent;
import com.raas.payment.model.Transaction;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.stereotype.Component;

@Component
public class PaymentEventPublisher {
    private final KafkaTemplate<String, Object> kafkaTemplate;
    private final KafkaTopics topics;

    public PaymentEventPublisher(KafkaTemplate<String, Object> kafkaTemplate, KafkaTopics topics) {
        this.kafkaTemplate = kafkaTemplate;
        this.topics = topics;
    }

    public void publishSucceeded(Transaction transaction, String bookingId) {
        PaymentResultEvent event = new PaymentResultEvent(
            transaction.getId().toString(),
            bookingId,
            "SUCCESS",
            null
        );
        kafkaTemplate.send(topics.getPaymentSucceeded(), event);
    }

    public void publishFailed(Transaction transaction, String bookingId, String reason) {
        String paymentId = transaction != null ? transaction.getId().toString() : null;
        PaymentResultEvent event = new PaymentResultEvent(
            paymentId,
            bookingId,
            "FAILED",
            reason
        );
        kafkaTemplate.send(topics.getPaymentFailed(), event);
    }
}
