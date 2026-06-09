package com.raas.review.repository;

import com.raas.review.model.Review;
import java.util.List;
import java.util.UUID;
import org.springframework.data.jpa.repository.JpaRepository;

public interface ReviewRepository extends JpaRepository<Review, UUID> {
    List<Review> findByListingId(String listingId);
}
