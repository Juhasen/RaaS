package com.raas.payment.repository;

import com.raas.payment.model.ProcessedBooking;
import java.util.UUID;
import org.springframework.data.jpa.repository.JpaRepository;

public interface ProcessedBookingRepository extends JpaRepository<ProcessedBooking, UUID> {
}
