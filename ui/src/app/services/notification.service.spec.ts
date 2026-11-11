import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { NotificationService } from './notification.service';
import { NotificationListResponse } from '../models/notification.model';

describe('NotificationService', () => {
  let service: NotificationService;
  let httpMock: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()]
    });
    service = TestBed.inject(NotificationService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('fetches user notifications with default skip and limit params', () => {
    const mockResponse: NotificationListResponse = {
      total: 1,
      count: 1,
      items: [
        {
          id: 'notif-1',
          user_id: 'user-123',
          message: 'Your booking #b1 for listing #l1 was CONFIRMED!',
          notification_type: 'EMAIL',
          status: 'FAILED',
          created_at: '2026-11-11T00:00:00Z'
        }
      ]
    };

    let result: NotificationListResponse | undefined;
    service.getUserNotifications('user-123').subscribe((res) => {
      result = res;
    });

    const req = httpMock.expectOne(
      (r) =>
        r.url === '/api/notifications/user/user-123' &&
        r.params.get('skip') === '0' &&
        r.params.get('limit') === '10'
    );
    expect(req.request.method).toBe('GET');
    req.flush(mockResponse);

    expect(result).toEqual(mockResponse);
  });
});
