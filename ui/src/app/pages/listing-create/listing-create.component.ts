import { Component, ChangeDetectionStrategy, signal, inject } from '@angular/core';
import { FormBuilder, FormGroup, Validators, ReactiveFormsModule } from '@angular/forms';
import { Router } from '@angular/router';
import { ListingService } from '../../services/listing.service';
import { Listing } from '../../models/listing.model';

@Component({
  selector: 'app-listing-create',
  imports: [ReactiveFormsModule],
  templateUrl: './listing-create.component.html',
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class ListingCreateComponent {
  private fb = inject(FormBuilder);
  private listingService = inject(ListingService);
  private router = inject(Router);

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

  onSubmit(): void {
    if (this.listingForm.invalid) {
      this.listingForm.markAllAsTouched();
      return;
    }

    this.isSubmitting.set(true);
    this.errorMessage.set(null);
    this.successMessage.set(null);

    const newListing: Listing = {
      host_id: 'host123', // Mock host_id
      title: this.listingForm.value.title,
      description: this.listingForm.value.description,
      price_per_day: this.listingForm.value.price_per_day,
      location_id: this.listingForm.value.location_id,
      location_label: this.listingForm.value.location_label
    };

    this.listingService.createListing(newListing).subscribe({
      next: () => {
        this.isSubmitting.set(false);
        this.successMessage.set('Listing published successfully!');
        this.listingForm.reset();
        setTimeout(() => {
          this.router.navigate(['/listing/manage']);
        }, 1500);
      },
      error: (err) => {
        this.isSubmitting.set(false);
        this.errorMessage.set(err.error?.error || 'Failed to publish listing. Please try again.');
      }
    });
  }
}
