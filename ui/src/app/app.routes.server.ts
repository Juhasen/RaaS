import { RenderMode, ServerRoute } from '@angular/ssr';

export const serverRoutes: ServerRoute[] = [
  {
    path: 'listing/:id',
    renderMode: RenderMode.Server
  },
  {
    path: 'listing/:id/edit',
    renderMode: RenderMode.Server
  },
  {
    path: 'admin/tools',
    renderMode: RenderMode.Client
  },
  {
    path: '**',
    renderMode: RenderMode.Prerender
  }
];
