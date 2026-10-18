import { inject, Injectable, signal, computed } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Router } from '@angular/router';
import { finalize, Observable, tap } from 'rxjs';
import { User, TokenResponse } from '../models/user.model';

@Injectable({
  providedIn: 'root'
})
export class AuthService {
  private http = inject(HttpClient);
  private router = inject(Router);
  
  private currentUserSignal = signal<User | null>(null);
  
  currentUser = computed(() => this.currentUserSignal());
  isAuthenticated = computed(() => this.currentUserSignal() !== null);

  // Settles once the initial /users/me lookup is done (or skipped), so guards can wait for it.
  readonly userLoaded: Promise<void>;
  
  constructor() {
    this.userLoaded = this.loadCurrentUser();
  }
  
  private loadCurrentUser(): Promise<void> {
    if (typeof window === 'undefined' || !localStorage.getItem('access_token')) {
      return Promise.resolve();
    }
    // Fetch current user details from API
    return new Promise((resolve) => {
      this.http.get<User>('/api/users/me').subscribe({
        next: (user) => {
          this.currentUserSignal.set(user);
          resolve();
        },
        error: () => {
          this.clearAuth();
          resolve();
        }
      });
    });
  }
  
  login(credentials: { email: string; password: string }): Observable<TokenResponse> {
    return this.http.post<TokenResponse>('/api/users/login', credentials).pipe(
      tap((response) => {
        if (typeof window !== 'undefined') {
          localStorage.setItem('access_token', response.access_token);
          this.currentUserSignal.set(response.user);
        }
      })
    );
  }
  
  register(userData: { email: string; password: string; role: string }): Observable<User> {
    return this.http.post<User>('/api/users/register', userData);
  }
  
  logout(): void {
    // Clears the HttpOnly cookie the admin tools rely on. Navigation waits for the request so it is not
    // cancelled; the local session ends even if the call fails (the cookie then expires on its own).
    this.http.post('/api/users/logout', null).pipe(finalize(() => {
      this.clearAuth();
      this.router.navigate(['/login']);
    })).subscribe({ error: () => undefined });
  }
  
  private clearAuth(): void {
    if (typeof window !== 'undefined') {
      localStorage.removeItem('access_token');
    }
    this.currentUserSignal.set(null);
  }
}
