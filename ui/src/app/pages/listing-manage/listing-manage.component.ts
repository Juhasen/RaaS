import { Component, ChangeDetectionStrategy, OnInit, signal, inject, PLATFORM_ID } from '@angular/core';
import { isPlatformBrowser } from '@angular/common';
import { RouterLink } from '@angular/router';
import { ListingService } from '../../services/listing.service';
import { Listing } from '../../models/listing.model';

@Component({
  selector: 'app-listing-manage',
  imports: [RouterLink],
  templateUrl: './listing-manage.component.html',
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class ListingManageComponent implements OnInit {
  private listingService = inject(ListingService);
  private platformId = inject(PLATFORM_ID);

  listings = signal<Listing[]>([]);
  isLoading = signal<boolean>(true);
  errorMessage = signal<string | null>(null);

  ngOnInit(): void {
    if (isPlatformBrowser(this.platformId)) {
      this.loadListings();
    } else {
      // Set loading to false for server-side prerender
      this.isLoading.set(false);
    }
  }

  loadListings(): void {
    this.isLoading.set(true);
    this.errorMessage.set(null);

    this.listingService.getListings().subscribe({
      next: (data) => {
        this.listings.set(data || []);
        this.isLoading.set(false);
      },
      error: (err) => {
        this.isLoading.set(false);
        this.errorMessage.set('Failed to load listings. Please make sure the service is running.');
        console.error(err);
      }
    });
  }
}
