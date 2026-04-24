package com.raas.payment.model;

import java.time.Instant;
import java.util.UUID;
import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.Id;
import jakarta.persistence.PrePersist;
import jakarta.persistence.Table;

@Entity
@Table(name = "processed_bookings")
public class ProcessedBooking {
    @Id
    @Column(name = "booking_id", nullable = false, updatable = false)
    private UUID bookingId;

    @Column(name = "processed_at", nullable = false)
    private Instant processedAt;

    protected ProcessedBooking() {
    }

    public ProcessedBooking(UUID bookingId) {
        this.bookingId = bookingId;
    }

    @PrePersist
    void onCreate() {
        if (processedAt == null) {
            processedAt = Instant.now();
        }
    }

    public UUID getBookingId() {
        return bookingId;
    }

    public Instant getProcessedAt() {
        return processedAt;
    }
}
