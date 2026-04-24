package com.raas.payment.service;

public record PaymentChargeResult(boolean success, String paymentId, String failureReason) {
    public static PaymentChargeResult success(String paymentId) {
        return new PaymentChargeResult(true, paymentId, null);
    }

    public static PaymentChargeResult failure(String reason) {
        return new PaymentChargeResult(false, null, reason);
    }
}
