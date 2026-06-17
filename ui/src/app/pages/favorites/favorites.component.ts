import { Component, ChangeDetectionStrategy, OnInit, signal, inject, PLATFORM_ID, effect } from '@angular/core';
import { isPlatformBrowser, CommonModule } from '@angular/common';
import { RouterLink } from '@angular/router';
import { FavoritesService } from '../../services/favorites.service';
import { ListingService } from '../../services/listing.service';
import { AuthService } from '../../services/auth.service';
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
  authService = inject(AuthService);
  private platformId = inject(PLATFORM_ID);

  favoriteListings = signal<Listing[]>([]);
  isLoading = signal<boolean>(true);
  errorMessage = signal<string | null>(null);

  constructor() {
    effect(() => {
      const user = this.authService.currentUser();
      if (isPlatformBrowser(this.platformId)) {
        const tokenExists = typeof window !== 'undefined' && !!localStorage.getItem('access_token');
        if (tokenExists && !user) {
          return;
        }
        this.loadFavorites();
      }
    });
  }

  ngOnInit(): void {
    if (!isPlatformBrowser(this.platformId)) {
      this.isLoading.set(false);
    }
  }

  loadFavorites(): void {
    const userId = this.authService.currentUser()?.id;
    if (!userId) {
      this.isLoading.set(false);
      this.errorMessage.set('Please log in to view observed listings.');
      return;
    }

    this.isLoading.set(true);
    this.errorMessage.set(null);

    forkJoin({
      favoriteIds: this.favoritesService.getFavorites(userId),
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
    const userId = this.authService.currentUser()?.id;
    if (!userId) return;

    this.favoritesService.removeFavorite(userId, listingId).subscribe({
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
