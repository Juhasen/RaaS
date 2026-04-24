package com.raas.payment.kafka.events;

import com.fasterxml.jackson.annotation.JsonProperty;

public record PaymentResultEvent(
    @JsonProperty("payment_id") String paymentId,
    @JsonProperty("booking_id") String bookingId,
    @JsonProperty("status") String status,
    @JsonProperty("reason") String reason
) {}
