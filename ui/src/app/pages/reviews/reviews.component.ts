import { Component, ChangeDetectionStrategy, OnInit, signal, inject, PLATFORM_ID } from '@angular/core';
import { isPlatformBrowser, CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, Validators, ReactiveFormsModule } from '@angular/forms';
import { ActivatedRoute } from '@angular/router';
import { ReviewService } from '../../services/review.service';
import { ListingService } from '../../services/listing.service';
import { Review } from '../../models/review.model';
import { Listing } from '../../models/listing.model';
import { forkJoin } from 'rxjs';

@Component({
  selector: 'app-reviews',
  imports: [CommonModule, ReactiveFormsModule],
  templateUrl: './reviews.component.html',
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class ReviewsComponent implements OnInit {
  private fb = inject(FormBuilder);
  private reviewService = inject(ReviewService);
  private listingService = inject(ListingService);
  private platformId = inject(PLATFORM_ID);
  private route = inject(ActivatedRoute);

  mockReviewerId = 'mock-user-123';
  reviews = signal<Review[]>([]);
  listings = signal<Listing[]>([]);
  isLoading = signal<boolean>(true);
  isSubmitting = signal<boolean>(false);
  errorMessage = signal<string | null>(null);
  successMessage = signal<string | null>(null);

  // Form setup
  reviewForm!: FormGroup;
  selectedRating = signal<number>(5);

  ngOnInit(): void {
    this.initForm();
    if (isPlatformBrowser(this.platformId)) {
      this.loadData();
      
      // Pre-select listing if listingId parameter is present in URL
      this.route.queryParams.subscribe(params => {
        if (params['listingId']) {
          this.reviewForm.patchValue({ listingId: params['listingId'] });
        }
      });
    } else {
      this.isLoading.set(false);
    }
  }

  initForm(): void {
    this.reviewForm = this.fb.group({
      listingId: ['', Validators.required],
      rating: [5, [Validators.required, Validators.min(1), Validators.max(5)]],
      comment: ['', [Validators.required, Validators.minLength(5)]],
      bookingId: [this.generateUuid(), Validators.required],
      reviewerId: [this.mockReviewerId, Validators.required]
    });
  }

  loadData(): void {
    this.isLoading.set(true);
    this.errorMessage.set(null);

    forkJoin({
      reviews: this.reviewService.getReviews(),
      listings: this.listingService.getListings()
    }).subscribe({
      next: (data) => {
        this.reviews.set(data.reviews || []);
        this.listings.set(data.listings || []);
        this.isLoading.set(false);
      },
      error: (err) => {
        this.isLoading.set(false);
        this.errorMessage.set('Failed to load reviews. Make sure backend microservices are running.');
        console.error(err);
      }
    });
  }

  setRating(rating: number): void {
    this.selectedRating.set(rating);
    this.reviewForm.patchValue({ rating });
  }

  getListingTitle(listingId: string): string {
    const list = this.listings().find(l => l.id === listingId);
    return list ? list.title : `Listing #${listingId.substring(0, 8)}`;
  }

  onSubmit(): void {
    if (this.reviewForm.invalid) {
      this.reviewForm.markAllAsTouched();
      return;
    }

    this.isSubmitting.set(true);
    this.errorMessage.set(null);
    this.successMessage.set(null);

    const newReview: Review = {
      bookingId: this.reviewForm.value.bookingId,
      reviewerId: this.reviewForm.value.reviewerId,
      listingId: this.reviewForm.value.listingId,
      rating: this.reviewForm.value.rating,
      comment: this.reviewForm.value.comment
    };

    this.reviewService.createReview(newReview).subscribe({
      next: (savedReview) => {
        this.isSubmitting.set(false);
        this.successMessage.set('Review submitted successfully!');
        
        // Add new review to feed
        this.reviews.update(items => [savedReview, ...items]);

        // Reset form but prefill new values for next submission
        this.reviewForm.reset({
          listingId: '',
          rating: 5,
          comment: '',
          bookingId: this.generateUuid(),
          reviewerId: this.mockReviewerId
        });
        this.selectedRating.set(5);

        // Clear success message after 3 seconds
        setTimeout(() => {
          this.successMessage.set(null);
        }, 3000);
      },
      error: (err) => {
        this.isSubmitting.set(false);
        this.errorMessage.set(err.error?.message || 'Failed to submit review. Please try again.');
        console.error(err);
      }
    });
  }

  private generateUuid(): string {
    return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
      const r = (Math.random() * 16) | 0;
      const v = c === 'x' ? r : (r & 0x3) | 0x8;
      return v.toString(16);
    });
  }
}
