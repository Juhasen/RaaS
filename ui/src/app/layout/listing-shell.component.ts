import { Component, ChangeDetectionStrategy, inject } from '@angular/core';
import { RouterLink, RouterOutlet } from '@angular/router';
import { AuthService } from '../services/auth.service';

@Component({
  selector: 'app-listing-shell',
  imports: [RouterOutlet, RouterLink],
  templateUrl: './listing-shell.component.html',
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class ListingShellComponent {
  authService = inject(AuthService);
}
