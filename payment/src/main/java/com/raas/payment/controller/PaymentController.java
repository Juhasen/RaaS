package com.raas.payment.controller;

import com.raas.payment.model.Transaction;
import com.raas.payment.repository.TransactionRepository;
import com.raas.payment.service.PaymentIntentDetails;
import com.raas.payment.service.StripeService;
import com.raas.payment.web.CreateIntentRequest;
import com.raas.payment.web.CreateIntentResponse;
import java.util.List;
import java.util.Map;
import java.util.Collections;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/payments")
public class PaymentController {
    private final TransactionRepository transactionRepository;
    private final StripeService stripeService;

    @Value("${stripe.publishableKey:}")
    private String publishableKey;

    public PaymentController(TransactionRepository transactionRepository, StripeService stripeService) {
        this.transactionRepository = transactionRepository;
        this.stripeService = stripeService;
    }

    @GetMapping
    public List<Transaction> getAllTransactions() {
        return transactionRepository.findAll();
    }

    @GetMapping("/public-key")
    public Map<String, String> getPublishableKey() {
        return Collections.singletonMap("publicKey", publishableKey);
    }

    @PostMapping("/create-intent")
    public CreateIntentResponse createPaymentIntent(@RequestBody CreateIntentRequest request) {
        PaymentIntentDetails details = stripeService.createPaymentIntent(request.amount());
        return new CreateIntentResponse(details.clientSecret(), details.paymentIntentId());
    }
}
