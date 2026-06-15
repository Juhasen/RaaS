import { Component, ChangeDetectionStrategy, OnInit, signal, inject, PLATFORM_ID } from '@angular/core';
import { isPlatformBrowser } from '@angular/common';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { ListingService } from '../../services/listing.service';
import { Listing } from '../../models/listing.model';

@Component({
  selector: 'app-listing-detail',
  imports: [RouterLink],
  templateUrl: './listing-detail.component.html',
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class ListingDetailComponent implements OnInit {
  private route = inject(ActivatedRoute);
  private listingService = inject(ListingService);
  private platformId = inject(PLATFORM_ID);

  listing = signal<Listing | null>(null);
  isLoading = signal<boolean>(true);
  errorMessage = signal<string | null>(null);
  fromPage = signal<string>('catalog');
  activePhotoIndex = signal<number>(0);

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
}
