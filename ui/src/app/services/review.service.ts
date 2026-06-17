import { inject, Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { Review } from '../models/review.model';

@Injectable({
  providedIn: 'root'
})
export class ReviewService {
  private http = inject(HttpClient);
  private apiUrl = '/api/reviews';

  getReviews(): Observable<Review[]> {
    return this.http.get<Review[]>(this.apiUrl);
  }

  createReview(review: Review): Observable<Review> {
    return this.http.post<Review>(this.apiUrl, review);
  }

  getReviewsByListing(listingId: string): Observable<Review[]> {
    return this.http.get<Review[]>(`${this.apiUrl}/listing/${listingId}`);
  }

  updateReview(id: string, reviewerId: string, rating: number, comment: string): Observable<Review> {
    return this.http.put<Review>(`${this.apiUrl}/${id}`, { reviewerId, rating, comment });
  }

  deleteReview(id: string, reviewerId: string): Observable<void> {
    return this.http.delete<void>(`${this.apiUrl}/${id}`, { params: { reviewerId } });
  }
}
