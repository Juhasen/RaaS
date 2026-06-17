package com.raas.payment.web;

public record CreateIntentResponse(String clientSecret, String paymentIntentId) {}
