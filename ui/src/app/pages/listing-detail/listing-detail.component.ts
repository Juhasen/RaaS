import { Component, ChangeDetectionStrategy, OnInit, signal, inject, PLATFORM_ID, computed } from '@angular/core';
import { isPlatformBrowser } from '@angular/common';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { ListingService } from '../../services/listing.service';
import { BookingService } from '../../services/booking.service';
import { Listing } from '../../models/listing.model';
import { Booking } from '../../models/booking.model';

@Component({
  selector: 'app-listing-detail',
  imports: [RouterLink, FormsModule],
  templateUrl: './listing-detail.component.html',
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class ListingDetailComponent implements OnInit {
  private route = inject(ActivatedRoute);
  private listingService = inject(ListingService);
  private bookingService = inject(BookingService);
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

  totalPrice = computed(() => {
    const start = this.startDate();
    const end = this.endDate();
    const price = this.listing()?.price_per_day;
    if (!start || !end || !price) return 0;
    const days = Math.ceil((new Date(end).getTime() - new Date(start).getTime()) / (1000 * 60 * 60 * 24));
    return days > 0 ? days * price : 0;
  });

  ngOnInit(): void {
    if (isPlatformBrowser(this.platformId)) {
      const from = this.route.snapshot.queryParamMap.get('from');
      if (from) {
        this.fromPage.set(from);
      }

      const id = this.route.snapshot.paramMap.get('id');

      if (id) {
        this.loadListing(id);
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

  setActivePhoto(index: number): void {
    this.activePhotoIndex.set(index);
  }

  requestBooking(): void {
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

    const booking: Booking = {
      listing_id: list.id,
      guest_id: 'guest123', // ponytail: mock guest_id until proper authentication system is introduced
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
}
