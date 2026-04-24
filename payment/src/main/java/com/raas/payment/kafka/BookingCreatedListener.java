package com.raas.payment.kafka;

import com.raas.payment.kafka.events.BookingCreatedEvent;
import com.raas.payment.service.PaymentProcessor;
import org.springframework.kafka.annotation.KafkaListener;
import org.springframework.stereotype.Component;

@Component
public class BookingCreatedListener {
    private final PaymentProcessor paymentProcessor;

    public BookingCreatedListener(PaymentProcessor paymentProcessor) {
        this.paymentProcessor = paymentProcessor;
    }

    @KafkaListener(topics = "${app.kafka.topics.bookingCreated}")
    public void onMessage(BookingCreatedEvent event) {
        paymentProcessor.handleBooking(event);
    }
}
