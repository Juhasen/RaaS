import { signal } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { ListingShellComponent } from './listing-shell.component';
import { AuthService } from '../services/auth.service';
import { User } from '../models/user.model';

describe('ListingShellComponent admin link', () => {
  async function render(user: User | null): Promise<HTMLElement> {
    const current = signal(user);
    TestBed.configureTestingModule({
      providers: [
        provideRouter([]),
        { provide: AuthService, useValue: { currentUser: current, isAuthenticated: () => user !== null } }
      ]
    });
    const fixture = TestBed.createComponent(ListingShellComponent);
    await fixture.whenStable();
    return fixture.nativeElement;
  }
  const adminLink = (el: HTMLElement) => el.querySelector('a[href="/admin/tools"]');
  const as = (role: User['role']): User => ({ id: '1', email: 'a@b.c', role, created_at: '' });

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
