import { inject, Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { Booking } from '../models/booking.model';

@Injectable({
  providedIn: 'root'
})
export class BookingService {
  private http = inject(HttpClient);
  private apiUrl = '/api/bookings';

  getBookings(guestId?: string, listingId?: string): Observable<Booking[]> {
    let url = this.apiUrl;
    const params: string[] = [];
    if (guestId) params.push(`guest_id=${guestId}`);
    if (listingId) params.push(`listing_id=${listingId}`);
    if (params.length > 0) {
      url += `?${params.join('&')}`;
    }
    return this.http.get<Booking[]>(url);
  }

  getBooking(id: string): Observable<Booking> {
    return this.http.get<Booking>(`${this.apiUrl}/${id}`);
  }

  createBooking(booking: Booking): Observable<Booking> {
    return this.http.post<Booking>(this.apiUrl, booking);
  }

  updateBooking(id: string, booking: Partial<Booking>): Observable<Booking> {
    return this.http.put<Booking>(`${this.apiUrl}/${id}`, booking);
  }

  deleteBooking(id: string): Observable<void> {
    return this.http.delete<void>(`${this.apiUrl}/${id}`);
  }
}
