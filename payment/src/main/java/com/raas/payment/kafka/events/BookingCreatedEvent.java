package com.raas.payment.kafka.events;

import com.fasterxml.jackson.annotation.JsonProperty;
import java.math.BigDecimal;

public record BookingCreatedEvent(
    @JsonProperty("booking_id") String bookingId,
    @JsonProperty("guest_id") String guestId,
    @JsonProperty("amount") BigDecimal amount,
    @JsonProperty("status") String status
) {}
