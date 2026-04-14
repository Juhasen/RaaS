package com.raas.payment.service;

import com.stripe.Stripe;
import com.stripe.exception.StripeException;
import com.stripe.model.PaymentIntent;
import com.stripe.param.PaymentIntentCreateParams;
import java.math.BigDecimal;
import java.math.RoundingMode;
import java.util.UUID;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

@Service
public class StripeService {
    private final String apiKey;

    public StripeService(@Value("${stripe.apiKey:}") String apiKey) {
        this.apiKey = apiKey;
    }

    public PaymentChargeResult charge(UUID bookingId, BigDecimal amount) {
        if (amount == null) {
            return PaymentChargeResult.failure("Missing amount");
        }
        if (apiKey == null || apiKey.isBlank()) {
            return PaymentChargeResult.success("simulated-" + bookingId);
        }
        Stripe.apiKey = apiKey;
        long minorUnits = amount.movePointRight(2)
            .setScale(0, RoundingMode.HALF_UP)
            .longValueExact();

        PaymentIntentCreateParams params = PaymentIntentCreateParams.builder()
            .setAmount(minorUnits)
            .setCurrency("usd")
            .putMetadata("booking_id", bookingId.toString())
            .build();

        try {
            PaymentIntent intent = PaymentIntent.create(params);
            return PaymentChargeResult.success(intent.getId());
        } catch (StripeException ex) {
            return PaymentChargeResult.failure(ex.getMessage());
        }
    }
}
