import { inject, Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

export interface HostListingStats {
  listing_id: string;
  title: string;
  price_per_day: number;
  bookings_count: number;
  pending_bookings: number;
  confirmed_bookings: number;
  rejected_bookings: number;
}

export interface HostAnalyticsResponse {
  host_id: string;
  total_listings: number;
  total_bookings: number;
  total_pending: number;
  total_confirmed: number;
  total_rejected: number;
  listings: HostListingStats[];
}

@Injectable({
  providedIn: 'root'
})
export class AnalyticsService {
  private http = inject(HttpClient);
  private apiUrl = '/api/analytics/host';

  getHostAnalytics(hostId: string): Observable<HostAnalyticsResponse> {
    return this.http.get<HostAnalyticsResponse>(`${this.apiUrl}/${hostId}`);
  }
}
