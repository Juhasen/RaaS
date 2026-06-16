package com.raas.review.controller;

import com.raas.review.model.Review;
import com.raas.review.repository.ReviewRepository;
import com.raas.review.web.ReviewRequest;
import jakarta.validation.Valid;
import java.util.List;
import java.util.UUID;
import org.springframework.http.HttpStatus;
import org.springframework.web.server.ResponseStatusException;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.ResponseStatus;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/api/reviews")
public class ReviewController {
    private final ReviewRepository reviewRepository;

    public ReviewController(ReviewRepository reviewRepository) {
        this.reviewRepository = reviewRepository;
    }

    @PostMapping
    @ResponseStatus(HttpStatus.CREATED)
    public Review createReview(@Valid @RequestBody ReviewRequest request) {
        Review review = new Review(
            parseUuid(request.getBookingId(), "bookingId"),
            parseUuid(request.getReviewerId(), "reviewerId"),
            request.getListingId(),
            request.getRating(),
            request.getComment()
        );
        return reviewRepository.save(review);
    }

    @GetMapping("/listing/{listingId}")
    public List<Review> getByListing(@PathVariable String listingId) {
        return reviewRepository.findByListingId(listingId);
    }

    @GetMapping
    public List<Review> getAllReviews() {
        return reviewRepository.findAll();
    }


    private UUID parseUuid(String rawValue, String fieldName) {
        try {
            return UUID.fromString(rawValue);
        } catch (IllegalArgumentException ex) {
            throw new ResponseStatusException(HttpStatus.BAD_REQUEST, "Invalid " + fieldName);
        }
    }
}
