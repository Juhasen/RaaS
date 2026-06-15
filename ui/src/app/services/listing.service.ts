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

  getListings(hostId?: string): Observable<Listing[]> {
    const url = hostId ? `${this.apiUrl}?host_id=${hostId}` : this.apiUrl;
    return this.http.get<Listing[]>(url);
  }

  getListing(id: string): Observable<Listing> {
    return this.http.get<Listing>(`${this.apiUrl}/${id}`);
  }

  createListing(listing: Listing): Observable<Listing> {
    return this.http.post<Listing>(this.apiUrl, listing);
  }
}
