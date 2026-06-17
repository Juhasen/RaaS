package com.raas.payment.web;

import java.math.BigDecimal;

public record CreateIntentRequest(BigDecimal amount) {}
