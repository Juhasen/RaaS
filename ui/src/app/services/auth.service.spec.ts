import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { provideRouter } from '@angular/router';
import { AuthService } from './auth.service';

describe('AuthService.logout', () => {
  it('clears the token and asks the server to clear the auth cookie', () => {
    TestBed.configureTestingModule({
      providers: [provideRouter([{ path: 'login', children: [] }]), provideHttpClient(), provideHttpClientTesting()]
    });
    const http = TestBed.inject(HttpTestingController);
    const auth = TestBed.inject(AuthService);
    localStorage.setItem('access_token', 't');

    auth.logout();

    http.expectOne({ method: 'POST', url: '/api/users/logout' }).flush(null);
    expect(localStorage.getItem('access_token')).toBeNull();
    http.verify();
  });
});
