import { Component } from '@angular/core';
import { RouterLink, RouterOutlet } from '@angular/router';

@Component({
  selector: 'app-listing-shell',
  standalone: true,
  imports: [RouterOutlet, RouterLink],
  templateUrl: './listing-shell.component.html',
  styleUrl: './listing-shell.component.css'
})
export class ListingShellComponent {}
