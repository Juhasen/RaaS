import {
  Component,
  ChangeDetectionStrategy,
  effect,
  inject,
  PLATFORM_ID,
  signal,
  untracked
} from '@angular/core';
import { DatePipe, isPlatformBrowser } from '@angular/common';
import { RouterLink, RouterOutlet } from '@angular/router';
import { AuthService } from '../services/auth.service';
import { NotificationService } from '../services/notification.service';
import { NotificationItem } from '../models/notification.model';

@Component({
  selector: 'app-listing-shell',
  imports: [RouterOutlet, RouterLink, DatePipe],
  templateUrl: './listing-shell.component.html',
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class ListingShellComponent {
  public readonly authService = inject(AuthService);
  private readonly notificationService = inject(NotificationService);
  private readonly platformId = inject(PLATFORM_ID);

  public readonly notifications = signal<NotificationItem[]>([]);
  public readonly notificationsOpen = signal<boolean>(false);
  public readonly notificationsLoading = signal<boolean>(false);
  public readonly notificationsError = signal<string | null>(null);

  constructor() {
    effect(() => {
      const userId = this.authService.currentUser()?.id;
      if (!isPlatformBrowser(this.platformId)) {
        return;
      }
      untracked(() => {
        if (userId) {
          this.loadNotifications(userId);
        } else {
          this.notifications.set([]);
          this.notificationsLoading.set(false);
          this.notificationsError.set(null);
          this.notificationsOpen.set(false);
        }
      });
    });
  }

  public toggleNotifications(): void {
    const nextOpen = !this.notificationsOpen();
    this.notificationsOpen.set(nextOpen);
    const userId = this.authService.currentUser()?.id;
    if (nextOpen && userId) {
      this.loadNotifications(userId);
    }
  }

  public loadNotifications(userId: string): void {
    this.notificationsLoading.set(true);
    this.notificationsError.set(null);
    this.notificationService.getUserNotifications(userId).subscribe({
      next: (res) => {
        if (this.authService.currentUser()?.id !== userId) {
          return;
        }
        this.notifications.set(res.items);
        this.notificationsLoading.set(false);
      },
      error: () => {
        if (this.authService.currentUser()?.id !== userId) {
          return;
        }
        this.notificationsError.set('Failed to load notifications');
        this.notificationsLoading.set(false);
      }
    });
  }
}
