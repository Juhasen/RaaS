import { inject, Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { NotificationListResponse } from '../models/notification.model';

@Injectable({
  providedIn: 'root'
})
export class NotificationService {
  private readonly http = inject(HttpClient);
  private readonly apiUrl = '/api/notifications';

  public getUserNotifications(
    userId: string,
    skip = 0,
    limit = 10
  ): Observable<NotificationListResponse> {
    const params = new HttpParams()
      .set('skip', String(skip))
      .set('limit', String(limit));
    return this.http.get<NotificationListResponse>(
      `${this.apiUrl}/user/${userId}`,
      { params }
    );
  }
}
