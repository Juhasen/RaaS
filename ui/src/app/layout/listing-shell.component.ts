import { Component, ChangeDetectionStrategy, signal } from '@angular/core';
import { RouterLink, RouterLinkActive, RouterOutlet } from '@angular/router';

@Component({
  selector: 'app-listing-shell',
  imports: [RouterOutlet, RouterLink, RouterLinkActive],
  templateUrl: './listing-shell.component.html',
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class ListingShellComponent {
  mockHostId = 'host123';
}


