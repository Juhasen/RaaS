import { Component, ChangeDetectionStrategy, OnInit, signal, inject, PLATFORM_ID } from '@angular/core';
import { isPlatformBrowser } from '@angular/common';
import { RouterLink } from '@angular/router';
import { ListingService } from '../../services/listing.service';
import { Listing } from '../../models/listing.model';

@Component({
  selector: 'app-listing-catalog',
  imports: [RouterLink],
  templateUrl: './listing-catalog.component.html',
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class ListingCatalogComponent implements OnInit {

  private listingService = inject(ListingService);
  private platformId = inject(PLATFORM_ID);

  listings = signal<Listing[]>([]);
  isLoading = signal<boolean>(true);
  errorMessage = signal<string | null>(null);

  currentPage = signal<number>(1);
  totalListings = signal<number>(0);
  pageSize = 20; // ponytail: bump to configurable if needed later

  // Filter signals
  checkinFilter = signal<string>('');
  checkoutFilter = signal<string>('');
  locationFilter = signal<string>('');
  nameFilter = signal<string>('');

  private isFirstLoad = true;

  ngOnInit(): void {
    if (isPlatformBrowser(this.platformId)) {
      this.loadListings();
    } else {
      this.isLoading.set(false);
    }
  }

  loadListings(): void {
    if (this.isFirstLoad) {
      this.isLoading.set(true);
      this.isFirstLoad = false;
    }
    this.errorMessage.set(null);

    const filters: any = {
      page: this.currentPage(),
      limit: this.pageSize,
    };
    if (this.checkinFilter() && this.checkoutFilter()) {
      filters.checkin = this.checkinFilter();
      filters.checkout = this.checkoutFilter();
    }
    if (this.locationFilter()) {
      filters.location = this.locationFilter();
    }
    if (this.nameFilter()) {
      filters.name = this.nameFilter();
    }

    this.listingService.getListings(filters).subscribe({
      next: (data) => {
        this.listings.set(data.data || []);
        this.totalListings.set(data.total);
        this.isLoading.set(false);
      },
      error: (err) => {
        this.isLoading.set(false);
        this.errorMessage.set('Failed to load listings. Please make sure the service is running.');
        console.error(err);
      }
    });
  }

  onLocationInput(event: Event): void {
    const val = (event.target as HTMLInputElement).value;
    this.locationFilter.set(val);
    this.currentPage.set(1);
    this.loadListings();
  }

  onNameInput(event: Event): void {
    const val = (event.target as HTMLInputElement).value;
    this.nameFilter.set(val);
    this.currentPage.set(1);
    this.loadListings();
  }

  onCheckinChange(event: Event): void {
    const val = (event.target as HTMLInputElement).value;
    this.checkinFilter.set(val);
    if (this.checkoutFilter()) {
      this.currentPage.set(1);
      this.loadListings();
    }
  }

  onCheckoutChange(event: Event): void {
    const val = (event.target as HTMLInputElement).value;
    this.checkoutFilter.set(val);
    if (this.checkinFilter()) {
      this.currentPage.set(1);
      this.loadListings();
    }
  }

  clearFilters(): void {
    this.checkinFilter.set('');
    this.checkoutFilter.set('');
    this.locationFilter.set('');
    this.nameFilter.set('');
    this.currentPage.set(1);
    this.loadListings();
  }

  totalPages(): number {
    return Math.ceil(this.totalListings() / this.pageSize) || 1;
  }

  goToPage(page: number): void {
    if (page < 1 || page > this.totalPages()) return;
    this.currentPage.set(page);
    this.loadListings();
  }
}
