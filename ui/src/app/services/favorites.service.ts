import { inject, Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';

@Injectable({
  providedIn: 'root'
})
export class FavoritesService {
  private http = inject(HttpClient);
  private apiUrl = '/api/favorites';

  getFavorites(userId: string): Observable<string[]> {
    const params = new HttpParams().set('userId', userId);
    return this.http.get<string[]>(this.apiUrl, { params });
  }

  addFavorite(userId: string, listingId: string): Observable<void> {
    return this.http.post<void>(this.apiUrl, { userId, listingId });
  }

  removeFavorite(userId: string, listingId: string): Observable<void> {
    const params = new HttpParams().set('userId', userId);
    return this.http.delete<void>(`${this.apiUrl}/${listingId}`, { params });
  }
}
