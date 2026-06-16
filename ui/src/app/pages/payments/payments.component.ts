import { Component, ChangeDetectionStrategy, OnInit, signal, computed, inject, PLATFORM_ID } from '@angular/core';
import { isPlatformBrowser, CommonModule } from '@angular/common';
import { PaymentService } from '../../services/payment.service';
import { Transaction } from '../../models/payment.model';

@Component({
  selector: 'app-payments',
  imports: [CommonModule],
  templateUrl: './payments.component.html',
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class PaymentsComponent implements OnInit {
  private paymentService = inject(PaymentService);
  private platformId = inject(PLATFORM_ID);

  transactions = signal<Transaction[]>([]);
  isLoading = signal<boolean>(true);
  errorMessage = signal<string | null>(null);
  searchQuery = signal<string>('');

  ngOnInit(): void {
    if (isPlatformBrowser(this.platformId)) {
      this.loadTransactions();
    } else {
      this.isLoading.set(false);
    }
  }

  loadTransactions(): void {
    this.isLoading.set(true);
    this.errorMessage.set(null);

    this.paymentService.getTransactions().subscribe({
      next: (data) => {
        // Sort transactions by date descending
        const sorted = (data || []).sort((a, b) => 
          new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
        );
        this.transactions.set(sorted);
        this.isLoading.set(false);
      },
      error: (err) => {
        this.isLoading.set(false);
        this.errorMessage.set('Failed to load transaction logs. Please make sure the payment service is running.');
        console.error(err);
      }
    });
  }

  // Filtered transactions computed property
  filteredTransactions = computed(() => {
    const query = this.searchQuery().toLowerCase().trim();
    if (!query) {
      return this.transactions();
    }
    return this.transactions().filter(tx => 
      tx.id.toLowerCase().includes(query) || 
      tx.bookingId.toLowerCase().includes(query) || 
      tx.status.toLowerCase().includes(query) ||
      tx.paymentMethod.toLowerCase().includes(query)
    );
  });

  onSearch(event: Event): void {
    const target = event.target as HTMLInputElement;
    this.searchQuery.set(target.value);
  }
}
