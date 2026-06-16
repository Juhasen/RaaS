import { Component, ChangeDetectionStrategy, signal, inject, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, Validators, ReactiveFormsModule } from '@angular/forms';
import { Router, ActivatedRoute, RouterLink } from '@angular/router';
import { forkJoin, of, switchMap } from 'rxjs';
import { ListingService } from '../../services/listing.service';
import { AuthService } from '../../services/auth.service';
import { Listing } from '../../models/listing.model';

@Component({
  selector: 'app-listing-create',
  imports: [ReactiveFormsModule, RouterLink],
  templateUrl: './listing-create.component.html',
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class ListingCreateComponent implements OnInit {
  private fb = inject(FormBuilder);
  private listingService = inject(ListingService);
  private authService = inject(AuthService);
  private router = inject(Router);
  private route = inject(ActivatedRoute);

  listingForm: FormGroup = this.fb.group({
    title: ['', [Validators.required, Validators.minLength(3)]],
    description: ['', [Validators.required, Validators.minLength(10)]],
    price_per_day: [null, [Validators.required, Validators.min(1)]],
    location_id: ['', [Validators.required]],
    location_label: ['', [Validators.required]]
  });

  isSubmitting = signal<boolean>(false);
  errorMessage = signal<string | null>(null);
  successMessage = signal<string | null>(null);

  selectedFiles = signal<File[]>([]);
  previewUrls = signal<string[]>([]);
  isEditMode = signal<boolean>(false);
  listingId = signal<string | null>(null);
  existingMediaUrls = signal<string[]>([]);

  ngOnInit(): void {
    const id = this.route.snapshot.paramMap.get('id');
    if (id) {
      this.isEditMode.set(true);
      this.listingId.set(id);
      this.loadListing(id);
    }
  }

  loadListing(id: string): void {
    this.isSubmitting.set(true);
    this.listingService.getListing(id).subscribe({
      next: (listing) => {
        this.isSubmitting.set(false);
        this.listingForm.patchValue({
          title: listing.title,
          description: listing.description,
          price_per_day: listing.price_per_day,
          location_id: listing.location_id,
          location_label: listing.location_label
        });
        if (listing.media_urls) {
          this.existingMediaUrls.set(listing.media_urls);
        }
      },
      error: (err) => {
        this.isSubmitting.set(false);
        this.errorMessage.set('Failed to load listing details.');
        console.error(err);
      }
    });
  }

  onFileSelected(event: Event): void {
    const input = event.target as HTMLInputElement;
    if (input.files) {
      const files = Array.from(input.files);
      this.selectedFiles.update(current => [...current, ...files]);

      files.forEach(file => {
        const reader = new FileReader();
        reader.onload = (e) => {
          if (e.target?.result) {
            this.previewUrls.update(urls => [...urls, e.target!.result as string]);
          }
        };
        reader.readAsDataURL(file);
      });
    }
  }

  removeFile(index: number): void {
    this.selectedFiles.update(current => current.filter((_, i) => i !== index));
    this.previewUrls.update(urls => urls.filter((_, i) => i !== index));
  }

  onSubmit(): void {
    if (this.listingForm.invalid) {
      this.listingForm.markAllAsTouched();
      return;
    }

    this.isSubmitting.set(true);
    this.errorMessage.set(null);
    this.successMessage.set(null);

    const hostId = this.authService.currentUser()?.id || 'host123';
    const listingData: Listing = {
      host_id: hostId,
      title: this.listingForm.value.title,
      description: this.listingForm.value.description,
      price_per_day: this.listingForm.value.price_per_day,
      location_id: this.listingForm.value.location_id,
      location_label: this.listingForm.value.location_label
    };

    const action$ = this.isEditMode()
      ? this.listingService.updateListing(this.listingId()!, listingData)
      : this.listingService.createListing(listingData);

    action$.pipe(
      switchMap((savedListing: Listing) => {
        const targetId = this.isEditMode() ? this.listingId()! : savedListing.id!;
        const files = this.selectedFiles();
        if (files.length === 0) {
          return of(savedListing);
        }
        const uploadObservables = files.map(file =>
          this.listingService.uploadPhoto(targetId, file)
        );
        return forkJoin(uploadObservables);
      })
    ).subscribe({
      next: () => {
        this.isSubmitting.set(false);
        this.successMessage.set(this.isEditMode() ? 'Listing updated successfully!' : 'Listing published and photos uploaded successfully!');
        if (!this.isEditMode()) {
          this.listingForm.reset();
          this.selectedFiles.set([]);
          this.previewUrls.set([]);
        }
        setTimeout(() => {
          this.router.navigate(['/listing/manage']);
        }, 1500);
      },
      error: (err) => {
        this.isSubmitting.set(false);
        this.errorMessage.set(err.error?.error || `Failed to ${this.isEditMode() ? 'update' : 'publish'} listing. Please try again.`);
      }
    });
  }
}
