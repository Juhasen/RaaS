import { Component, ChangeDetectionStrategy, OnInit, signal, inject, PLATFORM_ID, computed } from '@angular/core';
import { isPlatformBrowser, CommonModule, DatePipe } from '@angular/common';
import { HttpClient } from '@angular/common/http';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { ListingService } from '../../services/listing.service';
import { BookingService } from '../../services/booking.service';
import { ReviewService } from '../../services/review.service';
import { AuthService } from '../../services/auth.service';
import { Listing } from '../../models/listing.model';
import { Booking } from '../../models/booking.model';
import { Review } from '../../models/review.model';
import { User } from '../../models/user.model';
import { ConfirmModalComponent } from '../../components/confirm-modal/confirm-modal.component';

@Component({
  selector: 'app-listing-detail',
  imports: [RouterLink, FormsModule, CommonModule, DatePipe, ConfirmModalComponent],
  templateUrl: './listing-detail.component.html',
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class ListingDetailComponent implements OnInit {
  private route = inject(ActivatedRoute);
  private listingService = inject(ListingService);
  private bookingService = inject(BookingService);
  private reviewService = inject(ReviewService);
  private http = inject(HttpClient);
  authService = inject(AuthService);
  private platformId = inject(PLATFORM_ID);

  listing = signal<Listing | null>(null);
  isLoading = signal<boolean>(true);
  errorMessage = signal<string | null>(null);
  fromPage = signal<string>('catalog');
  activePhotoIndex = signal<number>(0);

  // Booking signals
  startDate = signal<string>('');
  endDate = signal<string>('');
  bookingSubmitting = signal<boolean>(false);
  bookingSuccess = signal<string | null>(null);
  bookingError = signal<string | null>(null);

  // Review signals
  reviews = signal<Review[]>([]);
  reviewsLoading = signal<boolean>(false);
  reviewComment = signal<string>('');
  reviewRating = signal<number>(5);
  reviewSubmitting = signal<boolean>(false);
  reviewSuccess = signal<string | null>(null);
  reviewError = signal<string | null>(null);
  reviewerEmails = signal<Record<string, string>>({});
  editingReviewId = signal<string | null>(null);
  editingComment = signal<string>('');
  editingRating = signal<number>(5);

  // Confirmation Modal signals
  confirmOpen = signal<boolean>(false);
  confirmTitle = signal<string>('');
  confirmMessage = signal<string>('');
  confirmText = signal<string>('Confirm');
  confirmType = signal<'danger' | 'primary' | 'warning'>('primary');
  private onConfirmCallback: (() => void) | null = null;

  averageRating = computed(() => {
    const r = this.reviews();
    if (r.length === 0) return 0;
    const sum = r.reduce((acc, rev) => acc + rev.rating, 0);
    return Math.round((sum / r.length) * 10) / 10;
  });

  totalPrice = computed(() => {
    const start = this.startDate();
    const end = this.endDate();
    const price = this.listing()?.price_per_day;
    if (!start || !end || !price) return 0;
    const days = Math.ceil((new Date(end).getTime() - new Date(start).getTime()) / (1000 * 60 * 60 * 24));
    return days > 0 ? days * price : 0;
  });

  isAuthenticated = computed(() => this.authService.isAuthenticated());

  ngOnInit(): void {
    if (isPlatformBrowser(this.platformId)) {
      const from = this.route.snapshot.queryParamMap.get('from');
      if (from) {
        this.fromPage.set(from);
      }

      const id = this.route.snapshot.paramMap.get('id');

      if (id) {
        this.loadListing(id);
        this.loadReviews(id);
      } else {
        this.isLoading.set(false);
        this.errorMessage.set('Invalid Listing ID.');
      }
    } else {
      this.isLoading.set(false);
    }
  }

  loadListing(id: string): void {
    this.isLoading.set(true);
    this.errorMessage.set(null);

    this.listingService.getListing(id).subscribe({
      next: (data) => {
        this.listing.set(data);
        this.activePhotoIndex.set(0);
        this.isLoading.set(false);
      },
      error: (err) => {
        this.isLoading.set(false);
        this.errorMessage.set('Failed to load listing details. Please try again.');
        console.error(err);
      }
    });
  }

  loadReviews(listingId: string): void {
    this.reviewsLoading.set(true);
    this.reviewService.getReviewsByListing(listingId).subscribe({
      next: (data) => {
        this.reviews.set(data || []);
        this.reviewsLoading.set(false);
        this.fetchReviewerEmails(data || []);
      },
      error: (err) => {
        console.error('Failed to load reviews:', err);
        this.reviewsLoading.set(false);
      }
    });
  }

  private fetchReviewerEmails(reviews: Review[]): void {
    const uniqueIds = [...new Set(reviews.map(r => r.reviewerId))];
    for (const userId of uniqueIds) {
      if (this.reviewerEmails()[userId]) continue;
      this.http.get<User>(`/api/users/${userId}`).subscribe({
        next: (user) => {
          this.reviewerEmails.update(map => ({ ...map, [userId]: user.email }));
        },
        error: () => { /* user not found, keep showing ID */ }
      });
    }
  }

  getReviewerDisplay(reviewerId: string): string {
    return this.reviewerEmails()[reviewerId] || reviewerId.substring(0, 8) + '...';
  }

  setActivePhoto(index: number): void {
    this.activePhotoIndex.set(index);
  }

  setReviewRating(rating: number): void {
    this.reviewRating.set(rating);
  }

  submitReview(): void {
    const listingId = this.listing()?.id;
    const comment = this.reviewComment();
    if (!listingId || !comment || comment.trim().length < 5) {
      this.reviewError.set('Please write at least 5 characters.');
      return;
    }

    this.reviewSubmitting.set(true);
    this.reviewError.set(null);
    this.reviewSuccess.set(null);

    const userId = this.authService.currentUser()?.id || 'anonymous-user';

    const newReview: Review = {
      bookingId: this.generateUuid(),
      reviewerId: userId,
      listingId: listingId,
      rating: this.reviewRating(),
      comment: comment.trim()
    };

    this.reviewService.createReview(newReview).subscribe({
      next: (saved) => {
        this.reviewSubmitting.set(false);
        this.reviewSuccess.set('Review posted!');
        this.reviews.update(items => [saved, ...items]);
        this.reviewComment.set('');
        this.reviewRating.set(5);

        // Cache current user's email for immediate display
        const email = this.authService.currentUser()?.email;
        if (email) {
          this.reviewerEmails.update(map => ({ ...map, [userId]: email }));
        }

        setTimeout(() => this.reviewSuccess.set(null), 3000);
      },
      error: (err) => {
        this.reviewSubmitting.set(false);
        this.reviewError.set(err.error?.message || 'Failed to post review. Please try again.');
        console.error(err);
      }
    });
  }

  startEdit(review: Review): void {
    if (!review.id) return;
    this.editingReviewId.set(review.id);
    this.editingComment.set(review.comment);
    this.editingRating.set(review.rating);
  }

  cancelEdit(): void {
    this.editingReviewId.set(null);
    this.editingComment.set('');
  }

  saveEdit(review: Review): void {
    if (!review.id) return;
    const comment = this.editingComment();
    if (!comment || comment.trim().length < 5) {
      alert('Comment must be at least 5 characters long.');
      return;
    }
    const currentUserId = this.authService.currentUser()?.id;
    if (!currentUserId || currentUserId !== review.reviewerId) {
      alert('You are not authorized to edit this review.');
      return;
    }

    this.openConfirmation(
      'Modify Review',
      'Are you sure you want to save the changes to your review?',
      'primary',
      'Save Changes',
      () => {
        this.reviewService.updateReview(review.id!, currentUserId, this.editingRating(), comment.trim()).subscribe({
          next: (updated) => {
            this.reviews.update(items => items.map(item => item.id === review.id ? updated : item));
            this.cancelEdit();
          },
          error: (err) => {
            alert(err.error?.message || 'Failed to update review.');
            console.error(err);
          }
        });
      }
    );
  }

  deleteReview(reviewId: string): void {
    const currentUserId = this.authService.currentUser()?.id;
    if (!currentUserId) return;

    this.openConfirmation(
      'Delete Review',
      'Are you sure you want to delete this review? This action cannot be undone.',
      'danger',
      'Delete',
      () => {
        this.reviewService.deleteReview(reviewId, currentUserId).subscribe({
          next: () => {
            this.reviews.update(items => items.filter(item => item.id !== reviewId));
          },
          error: (err) => {
            alert(err.error?.message || 'Failed to delete review.');
            console.error(err);
          }
        });
      }
    );
  }

  openConfirmation(title: string, message: string, type: 'danger' | 'primary' | 'warning', confirmBtnText: string, callback: () => void): void {
    this.confirmTitle.set(title);
    this.confirmMessage.set(message);
    this.confirmType.set(type);
    this.confirmText.set(confirmBtnText);
    this.onConfirmCallback = callback;
    this.confirmOpen.set(true);
  }

  handleConfirm(): void {
    const cb = this.onConfirmCallback;
    if (cb) cb();
    this.confirmOpen.set(false);
  }

  handleCancel(): void {
    this.confirmOpen.set(false);
  }

  requestBooking(): void {
    if (!this.authService.currentUser()) {
      this.bookingError.set('Please log in to request a booking.');
      return;
    }

    const start = this.startDate();
    const end = this.endDate();
    const list = this.listing();
    if (!start || !end || !list?.id) {
      this.bookingError.set('Please select both start and end dates.');
      return;
    }

    if (new Date(start) >= new Date(end)) {
      this.bookingError.set('End date must be after start date.');
      return;
    }

    this.bookingSubmitting.set(true);
    this.bookingError.set(null);
    this.bookingSuccess.set(null);

    const guestId = this.authService.currentUser()?.id;
    if (!guestId) {
      this.bookingSubmitting.set(false);
      this.bookingError.set('Please log in to request a booking.');
      return;
    }

    const booking: Booking = {
      listing_id: list.id,
      guest_id: guestId,
      start_date: start,
      end_date: end,
      total_price: this.totalPrice()
    };

    this.bookingService.createBooking(booking).subscribe({
      next: (res) => {
        this.bookingSubmitting.set(false);
        this.bookingSuccess.set(`Booking requested successfully! Booking ID: ${res.id}`);
        // Reset form
        this.startDate.set('');
        this.endDate.set('');
      },
      error: (err) => {
        this.bookingSubmitting.set(false);
        this.bookingError.set(err.error?.error || 'Failed to request booking. The selected dates might be unavailable.');
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
