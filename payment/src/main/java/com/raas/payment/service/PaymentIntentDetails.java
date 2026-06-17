package com.raas.payment.service;

public record PaymentIntentDetails(String clientSecret, String paymentIntentId) {}
