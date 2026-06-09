import { Routes } from '@angular/router';
import { ListingShellComponent } from './layout/listing-shell.component';
import { ListingCreateComponent } from './pages/listing-create/listing-create.component';
import { ListingManageComponent } from './pages/listing-manage/listing-manage.component';
import { NotFoundComponent } from './pages/not-found/not-found.component';

export const routes: Routes = [
	{
		path: 'listing',
		component: ListingShellComponent,
		children: [
			{ path: 'create', component: ListingCreateComponent },
			{ path: 'manage', component: ListingManageComponent },
			{ path: '', redirectTo: 'create', pathMatch: 'full' }
		]
	},
	{ path: '', redirectTo: 'listing/create', pathMatch: 'full' },
	{ path: '**', component: NotFoundComponent }
];
