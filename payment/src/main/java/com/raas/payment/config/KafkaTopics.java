package com.raas.payment.config;

import org.springframework.boot.context.properties.ConfigurationProperties;

@ConfigurationProperties(prefix = "app.kafka.topics")
public class KafkaTopics {
    private String bookingCreated;
    private String paymentSucceeded;
    private String paymentFailed;

    public String getBookingCreated() {
        return bookingCreated;
    }

    public void setBookingCreated(String bookingCreated) {
        this.bookingCreated = bookingCreated;
    }

    public String getPaymentSucceeded() {
        return paymentSucceeded;
    }

    public void setPaymentSucceeded(String paymentSucceeded) {
        this.paymentSucceeded = paymentSucceeded;
    }

    public String getPaymentFailed() {
        return paymentFailed;
    }

    public void setPaymentFailed(String paymentFailed) {
        this.paymentFailed = paymentFailed;
    }
}
