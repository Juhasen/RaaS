import { signal } from '@angular/core';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { Observable, of, throwError } from 'rxjs';
import { ListingShellComponent } from './listing-shell.component';
import { AuthService } from '../services/auth.service';
import { NotificationService } from '../services/notification.service';
import { NotificationItem, NotificationListResponse } from '../models/notification.model';
import { User } from '../models/user.model';

describe('ListingShellComponent', () => {
  async function renderFixture(
    user: User | null,
    itemsOrObservable: NotificationItem[] | Observable<NotificationListResponse> = []
  ): Promise<ComponentFixture<ListingShellComponent>> {
    const current = signal(user);
    const response$ = Array.isArray(itemsOrObservable)
      ? of({ total: itemsOrObservable.length, count: itemsOrObservable.length, items: itemsOrObservable })
      : itemsOrObservable;
    TestBed.configureTestingModule({
      providers: [
        provideRouter([]),
        { provide: AuthService, useValue: { currentUser: current, isAuthenticated: () => user !== null } },
        {
          provide: NotificationService,
          useValue: {
            getUserNotifications: () => response$
          }
        }
      ]
    });
    const fixture = TestBed.createComponent(ListingShellComponent);
    await fixture.whenStable();
    return fixture;
  }

  async function render(
    user: User | null,
    items: NotificationItem[] = []
  ): Promise<HTMLElement> {
    const fixture = await renderFixture(user, items);
    return fixture.nativeElement;
  }

  const adminLink = (el: HTMLElement) => el.querySelector('a[href="/admin/tools"]');
  const bellButton = (el: HTMLElement) => el.querySelector<HTMLButtonElement>('[data-testid="notification-bell"]');
  const dropdown = (el: HTMLElement) => el.querySelector<HTMLElement>('[data-testid="notification-dropdown"]');
  const as = (role: User['role']): User => ({ id: '1', email: 'a@b.c', role, created_at: '' });

  describe('admin link', () => {
    it('shows for an admin', async () => {
      expect(adminLink(await render(as('admin')))).not.toBeNull();
    });

    it.each(['host', 'guest'] as const)('is hidden for a %s', async (role) => {
      expect(adminLink(await render(as(role)))).toBeNull();
    });

    it('is hidden when logged out', async () => {
      expect(adminLink(await render(null))).toBeNull();
    });
  });

  describe('notification bell', () => {
    it.each(['guest', 'host', 'admin'] as const)('shows bell button for authenticated %s', async (role) => {
      const el = await render(as(role));
      const bell = bellButton(el);
      expect(bell).not.toBeNull();
      expect(bell?.getAttribute('aria-label')).toBe('Notifications');
      expect(bell?.getAttribute('aria-expanded')).toBe('false');
    });

    it('hides bell button when logged out', async () => {
      const el = await render(null);
      expect(bellButton(el)).toBeNull();
    });

    it('opens dropdown on click and renders confirmed and rejected messages even when status is FAILED', async () => {
      const items: NotificationItem[] = [
        {
          id: 'n1',
          user_id: '1',
          message: 'Your booking #b1 for listing #l1 was CONFIRMED!',
          notification_type: 'EMAIL',
          status: 'FAILED',
          created_at: '2026-11-11T10:00:00Z'
        },
        {
          id: 'n2',
          user_id: '1',
          message: 'Your booking #b2 for listing #l1 was REJECTED.',
          notification_type: 'EMAIL',
          status: 'FAILED',
          created_at: '2026-11-11T11:00:00Z'
        }
      ];
      const fixture = await renderFixture(as('guest'), items);
      const el: HTMLElement = fixture.nativeElement;

      expect(dropdown(el)).toBeNull();
      bellButton(el)?.click();
      await fixture.whenStable();

      const panel = dropdown(el);
      expect(panel).not.toBeNull();
      expect(bellButton(el)?.getAttribute('aria-expanded')).toBe('true');
      expect(panel?.textContent).toContain('Your booking #b1 for listing #l1 was CONFIRMED!');
      expect(panel?.textContent).toContain('Your booking #b2 for listing #l1 was REJECTED.');
      expect(panel?.textContent).not.toContain('FAILED');
    });

    it('shows empty state when user has no notifications', async () => {
      const fixture = await renderFixture(as('guest'), []);
      const el: HTMLElement = fixture.nativeElement;

      bellButton(el)?.click();
      await fixture.whenStable();

      const panel = dropdown(el);
      expect(panel).not.toBeNull();
      expect(panel?.textContent).toContain('No notifications yet');
    });

    it('shows error message when fetching notifications fails', async () => {
      const fixture = await renderFixture(
        as('guest'),
        throwError(() => new Error('network error'))
      );
      const el: HTMLElement = fixture.nativeElement;

      bellButton(el)?.click();
      await fixture.whenStable();

      const panel = dropdown(el);
      expect(panel).not.toBeNull();
      expect(panel?.textContent).toContain('Failed to load notifications');
    });
  });
});
