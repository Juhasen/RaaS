import { Component } from '@angular/core';
import { RouterLink, RouterLinkActive, RouterOutlet } from '@angular/router';

@Component({
  selector: 'app-listing-shell',
  imports: [RouterOutlet, RouterLink, RouterLinkActive],
  templateUrl: './listing-shell.component.html'
})
export class ListingShellComponent {}
