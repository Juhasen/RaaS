import { Component, ChangeDetectionStrategy, OnInit, signal, inject, PLATFORM_ID } from '@angular/core';
import { isPlatformBrowser } from '@angular/common';
import { RouterLink } from '@angular/router';
import { forkJoin, of } from 'rxjs';
import { catchError } from 'rxjs/operators';
import { ListingService } from '../../services/listing.service';
import { BookingService } from '../../services/booking.service';
import { AuthService } from '../../services/auth.service';
import { Listing } from '../../models/listing.model';
import { Booking } from '../../models/booking.model';

@Component({
  selector: 'app-listing-manage',
  imports: [RouterLink],
  templateUrl: './listing-manage.component.html',
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class ListingManageComponent implements OnInit {
  private listingService = inject(ListingService);
  private bookingService = inject(BookingService);
  authService = inject(AuthService);
  private platformId = inject(PLATFORM_ID);

  // Data signals
  listings = signal<Listing[]>([]);
  myBookings = signal<Booking[]>([]);
  guestBookings = signal<Booking[]>([]);
  allListingsMap = signal<Listing[]>([]);

  activeTab = signal<'listings' | 'my-bookings' | 'guest-bookings'>('listings');
  isLoading = signal<boolean>(true);
  errorMessage = signal<string | null>(null);

  get hostId(): string {
    return this.authService.currentUser()?.id || 'host123';
  }

  get guestId(): string {
    return this.authService.currentUser()?.id || 'guest123';
  }

  ngOnInit(): void {
    if (this.authService.currentUser()?.role === 'guest') {
      this.activeTab.set('my-bookings');
    }
    if (isPlatformBrowser(this.platformId)) {
      this.loadAllData();
    } else {
      this.isLoading.set(false);
    }
  }

  loadAllData(): void {
    this.isLoading.set(true);
    this.errorMessage.set(null);

    // 1. Fetch all listings to resolve titles/images for bookings
    // ponytail: high limit to get all listings for cross-referencing bookings
    this.listingService.getListings({ limit: 100 }).subscribe({
      next: (res) => {
        const allListings = res.data || [];
        this.allListingsMap.set(allListings);

        // Filter listings owned by this host
        const myOwnedListings = allListings.filter(l => l.host_id === this.hostId);
        this.listings.set(myOwnedListings);

        const myOwnedIds = new Set(myOwnedListings.map(l => l.id));

        // 2. Fetch my bookings (as guest)
        const myBookingsObs = this.bookingService.getBookings(this.guestId).pipe(
          catchError((err) => {
            console.error('Error loading my bookings:', err);
            return of([] as Booking[]);
          })
        );

        // 3. Fetch all bookings (to extract other guests' bookings on my listings)
        const allBookingsObs = this.bookingService.getBookings().pipe(
          catchError((err) => {
            console.error('Error loading all bookings:', err);
            return of([] as Booking[]);
          })
        );

        forkJoin({
          myBookings: myBookingsObs,
          allBookings: allBookingsObs
        }).subscribe({
          next: (res) => {
            this.myBookings.set(res.myBookings);
            
            // Filter all bookings to only those for listings I own
            const guestBookingsFiltered = res.allBookings.filter(b => myOwnedIds.has(b.listing_id));
            this.guestBookings.set(guestBookingsFiltered);
            
            this.isLoading.set(false);
          },
          error: (err) => {
            this.isLoading.set(false);
            this.errorMessage.set('Failed to load bookings data. Please try again.');
            console.error(err);
          }
        });
      },
      error: (err) => {
        this.isLoading.set(false);
        this.errorMessage.set('Failed to load listings data. Please make sure services are running.');
        console.error(err);
      }
    });
  }

  switchTab(tab: 'listings' | 'my-bookings' | 'guest-bookings'): void {
    this.activeTab.set(tab);
  }

  getListing(listingId: string): Listing | undefined {
    return this.allListingsMap().find(l => l.id === listingId);
  }

  updateBookingStatus(bookingId: string, status: 'CONFIRMED' | 'REJECTED'): void {
    this.isLoading.set(true);
    this.bookingService.updateBooking(bookingId, { status }).subscribe({
      next: () => {
        this.loadAllData();
      },
      error: (err) => {
        this.isLoading.set(false);
        this.errorMessage.set(`Failed to update booking status: ${err.error?.error || 'Unknown error'}`);
        console.error(err);
      }
    });
  }

  deleteBooking(bookingId: string): void {
    if (!confirm('Are you sure you want to cancel this booking?')) return;
    this.isLoading.set(true);
    this.bookingService.deleteBooking(bookingId).subscribe({
      next: () => {
        this.loadAllData();
      },
      error: (err) => {
        this.isLoading.set(false);
        this.errorMessage.set(`Failed to cancel booking: ${err.error?.error || 'Unknown error'}`);
        console.error(err);
      }
    });
  }

  deleteListing(listingId: string): void {
    if (!confirm('Are you sure you want to delete this listing? This action cannot be undone.')) return;
    this.isLoading.set(true);
    this.listingService.deleteListing(listingId).subscribe({
      next: () => {
        this.loadAllData();
      },
      error: (err) => {
        this.isLoading.set(false);
        this.errorMessage.set(`Failed to delete listing: ${err.error?.error || 'Unknown error'}`);
        console.error(err);
      }
    });
  }
}
