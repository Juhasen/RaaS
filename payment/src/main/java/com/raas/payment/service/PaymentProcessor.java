package com.raas.payment.service;

import com.raas.payment.kafka.PaymentEventPublisher;
import com.raas.payment.kafka.events.BookingCreatedEvent;
import com.raas.payment.model.ProcessedBooking;
import com.raas.payment.model.Transaction;
import com.raas.payment.repository.ProcessedBookingRepository;
import com.raas.payment.repository.TransactionRepository;
import java.math.BigDecimal;
import java.util.UUID;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

@Service
public class PaymentProcessor {
    private final ProcessedBookingRepository processedBookingRepository;
    private final TransactionRepository transactionRepository;
    private final StripeService stripeService;
    private final PaymentEventPublisher paymentEventPublisher;

    public PaymentProcessor(
        ProcessedBookingRepository processedBookingRepository,
        TransactionRepository transactionRepository,
        StripeService stripeService,
        PaymentEventPublisher paymentEventPublisher
    ) {
        this.processedBookingRepository = processedBookingRepository;
        this.transactionRepository = transactionRepository;
        this.stripeService = stripeService;
        this.paymentEventPublisher = paymentEventPublisher;
    }

    @Transactional
    public void handleBooking(BookingCreatedEvent event) {
        UUID bookingId = parseBookingId(event.bookingId());
        if (bookingId == null) {
            return;
        }
        if (!markProcessed(bookingId)) {
            return;
        }

        BigDecimal amount = event.amount();
        if (amount == null) {
            paymentEventPublisher.publishFailed(null, bookingId.toString(), "Missing amount");
            return;
        }
        PaymentChargeResult result = stripeService.charge(bookingId, amount);
        String status = result.success() ? "SUCCESS" : "FAILED";
        Transaction transaction = new Transaction(bookingId, amount, status, "stripe");
        transactionRepository.save(transaction);

        if (result.success()) {
            paymentEventPublisher.publishSucceeded(transaction, bookingId.toString());
        } else {
            paymentEventPublisher.publishFailed(transaction, bookingId.toString(), result.failureReason());
        }
    }

    private UUID parseBookingId(String rawBookingId) {
        if (rawBookingId == null || rawBookingId.isBlank()) {
            return null;
        }
        try {
            return UUID.fromString(rawBookingId);
        } catch (IllegalArgumentException ex) {
            return null;
        }
    }

    private boolean markProcessed(UUID bookingId) {
        if (processedBookingRepository.existsById(bookingId)) {
            return false;
        }
        processedBookingRepository.save(new ProcessedBooking(bookingId));
        return true;
    }
}
