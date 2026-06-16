import { Component, ChangeDetectionStrategy, OnInit, signal, inject, PLATFORM_ID } from '@angular/core';
import { isPlatformBrowser, CommonModule } from '@angular/common';
import { RouterLink } from '@angular/router';
import { FavoritesService } from '../../services/favorites.service';
import { ListingService } from '../../services/listing.service';
import { Listing } from '../../models/listing.model';
import { forkJoin } from 'rxjs';

@Component({
  selector: 'app-favorites',
  imports: [CommonModule, RouterLink],
  templateUrl: './favorites.component.html',
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class FavoritesComponent implements OnInit {
  private favoritesService = inject(FavoritesService);
  private listingService = inject(ListingService);
  private platformId = inject(PLATFORM_ID);

  mockUserId = 'mock-user-123';
  favoriteListings = signal<Listing[]>([]);
  isLoading = signal<boolean>(true);
  errorMessage = signal<string | null>(null);

  ngOnInit(): void {
    if (isPlatformBrowser(this.platformId)) {
      this.loadFavorites();
    } else {
      this.isLoading.set(false);
    }
  }

  loadFavorites(): void {
    this.isLoading.set(true);
    this.errorMessage.set(null);

    forkJoin({
      favoriteIds: this.favoritesService.getFavorites(this.mockUserId),
      allListings: this.listingService.getListings()
    }).subscribe({
      next: ({ favoriteIds, allListings }) => {
        const filtered = (allListings?.data || []).filter(listing => 
          listing.id && favoriteIds.includes(listing.id)
        );
        this.favoriteListings.set(filtered);
        this.isLoading.set(false);
      },
      error: (err) => {
        this.isLoading.set(false);
        this.errorMessage.set('Failed to load favorites. Please make sure services are running.');
        console.error(err);
      }
    });
  }

  toggleFavorite(listingId: string | undefined): void {
    if (!listingId) return;

    this.favoritesService.removeFavorite(this.mockUserId, listingId).subscribe({
      next: () => {
        // Filter out from the current view
        this.favoriteListings.update(listings => 
          listings.filter(listing => listing.id !== listingId)
        );
      },
      error: (err) => {
        console.error('Failed to remove favorite', err);
      }
    });
  }
}
