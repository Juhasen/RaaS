import { Routes } from '@angular/router';
import { ListingShellComponent } from './layout/listing-shell.component';
import { ListingCatalogComponent } from './pages/listing-catalog/listing-catalog.component';
import { ListingCreateComponent } from './pages/listing-create/listing-create.component';
import { ListingManageComponent } from './pages/listing-manage/listing-manage.component';
import { ListingDetailComponent } from './pages/listing-detail/listing-detail.component';
import { LoginComponent } from './pages/login/login.component';
import { RegisterComponent } from './pages/register/register.component';
import { FavoritesComponent } from './pages/favorites/favorites.component';
import { NotFoundComponent } from './pages/not-found/not-found.component';

export const routes: Routes = [
	{
		path: '',
		component: ListingShellComponent,
		children: [
			{ path: '', component: ListingCatalogComponent },
			{ path: 'listing', redirectTo: '', pathMatch: 'full' },
			{ path: 'listing/create', component: ListingCreateComponent },
			{ path: 'listing/manage', component: ListingManageComponent },
			{ path: 'listing/:id', component: ListingDetailComponent },
			{ path: 'favorites', component: FavoritesComponent },
			{ path: 'login', component: LoginComponent },
			{ path: 'register', component: RegisterComponent }
		]
	},
	{ path: '**', component: NotFoundComponent }
];

