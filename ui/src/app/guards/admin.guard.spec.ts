import { TestBed } from '@angular/core/testing';
import { Router, UrlTree, provideRouter } from '@angular/router';
import { signal } from '@angular/core';
import { adminGuard } from './admin.guard';
import { AuthService } from '../services/auth.service';
import { User } from '../models/user.model';

function user(role: User['role']): User {
  return { id: '1', email: 'a@b.c', role, created_at: '' };
}

describe('adminGuard', () => {
  const current = signal<User | null>(null);
  let resolveLoaded: () => void;

  function run(): Promise<boolean | UrlTree> {
    return TestBed.runInInjectionContext(() => adminGuard({} as never, {} as never)) as Promise<boolean | UrlTree>;
  }

  beforeEach(() => {
    current.set(null);
    const loaded = new Promise<void>((resolve) => (resolveLoaded = resolve));
    TestBed.configureTestingModule({
      providers: [
        provideRouter([]),
        { provide: AuthService, useValue: { currentUser: current, userLoaded: loaded } }
      ]
    });
  });

  it('allows an admin', async () => {
    current.set(user('admin'));
    resolveLoaded();
    expect(await run()).toBe(true);
  });

  it.each(['host', 'guest'] as const)('redirects a %s to /', async (role) => {
    current.set(user(role));
    resolveLoaded();
    const result = await run();
    expect(result).toEqual(TestBed.inject(Router).parseUrl('/'));
  });

  it('redirects a logged-out visitor to /', async () => {
    resolveLoaded();
    expect(await run()).toEqual(TestBed.inject(Router).parseUrl('/'));
  });

  it('waits for the user to load before deciding', async () => {
    const pending = run();
    current.set(user('admin'));
    resolveLoaded();
    expect(await pending).toBe(true);
  });
});
