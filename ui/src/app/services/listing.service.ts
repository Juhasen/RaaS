import { inject, Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { Listing } from '../models/listing.model';

@Injectable({
  providedIn: 'root'
})
export class ListingService {
  private http = inject(HttpClient);
  private apiUrl = '/api/listings';

  getListings(filters?: {
    host_id?: string;
    checkin?: string;
    checkout?: string;
    location?: string;
    name?: string;
  }): Observable<Listing[]> {
    const params: any = {};
    if (filters) {
      if (filters.host_id) params.host_id = filters.host_id;
      if (filters.checkin) params.checkin = filters.checkin;
      if (filters.checkout) params.checkout = filters.checkout;
      if (filters.location) params.location = filters.location;
      if (filters.name) params.name = filters.name;
    }
    return this.http.get<Listing[]>(this.apiUrl, { params });
  }

  getListing(id: string): Observable<Listing> {
    return this.http.get<Listing>(`${this.apiUrl}/${id}`);
  }

  createListing(listing: Listing): Observable<Listing> {
    return this.http.post<Listing>(this.apiUrl, listing);
  }

  uploadPhoto(listingId: string, file: File): Observable<any> {
    const formData = new FormData();
    formData.append('listing_id', listingId);
    formData.append('file', file);
    return this.http.post<any>('/api/media/upload', formData);
  }
}
