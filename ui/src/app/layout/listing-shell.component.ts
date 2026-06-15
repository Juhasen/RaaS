import { Component, ChangeDetectionStrategy } from '@angular/core';
import { RouterLink, RouterOutlet } from '@angular/router';

@Component({
  selector: 'app-listing-shell',
  imports: [RouterOutlet, RouterLink],
  templateUrl: './listing-shell.component.html',
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class ListingShellComponent {
  mockHostId = 'host123';
}
