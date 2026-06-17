import { Component, ChangeDetectionStrategy, OnInit, signal, inject, PLATFORM_ID, effect } from '@angular/core';
import { isPlatformBrowser } from '@angular/common';
import { RouterLink, ActivatedRoute } from '@angular/router';
import { forkJoin, of } from 'rxjs';
import { catchError } from 'rxjs/operators';
import { ListingService } from '../../services/listing.service';
import { BookingService } from '../../services/booking.service';
import { AuthService } from '../../services/auth.service';
import { AnalyticsService, HostAnalyticsResponse } from '../../services/analytics.service';
import { FavoritesService } from '../../services/favorites.service';
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
  private analyticsService = inject(AnalyticsService);
  private favoritesService = inject(FavoritesService);
  authService = inject(AuthService);
  private platformId = inject(PLATFORM_ID);
  private route = inject(ActivatedRoute);

  constructor() {
    effect(() => {
      const user = this.authService.currentUser();
      
      // If there's a token but the user is not loaded yet, wait.
      // Once user loads or clears, this effect will run again.
      const tokenExists = typeof window !== 'undefined' && !!localStorage.getItem('access_token');
      if (tokenExists && !user) {
        return;
      }

      if (user && user.role === 'guest') {
        this.activeTab.set('my-bookings');
      }

      if (isPlatformBrowser(this.platformId)) {
        this.loadAllData();
      }
    });
  }


  // Data signals
  listings = signal<Listing[]>([]);
  myBookings = signal<Booking[]>([]);
  guestBookings = signal<Booking[]>([]);
  allListingsMap = signal<Listing[]>([]);
  hostAnalytics = signal<HostAnalyticsResponse | null>(null);
  favoriteListings = signal<Listing[]>([]);

  activeTab = signal<'listings' | 'my-bookings' | 'guest-bookings' | 'analytics' | 'observed'>('listings');
  isLoading = signal<boolean>(true);
  errorMessage = signal<string | null>(null);

  get hostId(): string {
    return this.authService.currentUser()?.id || 'host123';
  }

  get guestId(): string {
    return this.authService.currentUser()?.id || 'guest123';
  }

  ngOnInit(): void {
    if (isPlatformBrowser(this.platformId)) {
      this.route.queryParams.subscribe(params => {
        const tab = params['tab'];
        if (tab && ['listings', 'my-bookings', 'guest-bookings', 'analytics', 'observed'].includes(tab)) {
          this.activeTab.set(tab as any);
        } else if (this.authService.currentUser()?.role === 'guest') {
          this.activeTab.set('my-bookings');
        } else {
          this.activeTab.set('listings');
        }
      });
    } else {
      if (this.authService.currentUser()?.role === 'guest') {
        this.activeTab.set('my-bookings');
      }
      this.isLoading.set(false);
    }
  }

  loadAllData(): void {
    this.isLoading.set(true);
    this.errorMessage.set(null);

    // 1. Fetch all listings to resolve titles/images for bookings
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

        // 4. Fetch host analytics if host
        const isHost = this.authService.currentUser()?.role !== 'guest';
        const hostAnalyticsObs = isHost ? this.analyticsService.getHostAnalytics(this.hostId).pipe(
          catchError((err) => {
            console.error('Error loading host analytics:', err);
            return of(null);
          })
        ) : of(null);

        // 5. Fetch favorites (observed listings)
        const favoritesObs = this.favoritesService.getFavorites(this.guestId).pipe(
          catchError((err) => {
            console.error('Error loading favorites:', err);
            return of([] as string[]);
          })
        );

        forkJoin({
          myBookings: myBookingsObs,
          allBookings: allBookingsObs,
          hostAnalytics: hostAnalyticsObs,
          favorites: favoritesObs
        }).subscribe({
          next: (res) => {
            this.myBookings.set(res.myBookings);
            
            // Filter all bookings to only those for listings I own
            const guestBookingsFiltered = res.allBookings.filter(b => myOwnedIds.has(b.listing_id));
            this.guestBookings.set(guestBookingsFiltered);

            if (res.hostAnalytics) {
              this.hostAnalytics.set(res.hostAnalytics);
            }

            // Resolve favorite listings from the allListings map
            const favIds = new Set(res.favorites);
            const favs = allListings.filter(l => l.id && favIds.has(l.id));
            this.favoriteListings.set(favs);
            
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

  switchTab(tab: 'listings' | 'my-bookings' | 'guest-bookings' | 'analytics' | 'observed'): void {
    this.activeTab.set(tab);
  }

  toggleFavorite(listingId: string): void {
    this.isLoading.set(true);
    this.favoritesService.removeFavorite(this.guestId, listingId).subscribe({
      next: () => {
        this.loadAllData();
      },
      error: (err) => {
        this.isLoading.set(false);
        this.errorMessage.set(`Failed to remove observed listing: ${err.error?.error || 'Unknown error'}`);
        console.error(err);
      }
    });
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
