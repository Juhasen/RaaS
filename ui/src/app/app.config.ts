import { ApplicationConfig, provideBrowserGlobalErrorListeners } from '@angular/core';
import { provideRouter } from '@angular/router';
import { provideHttpClient, withInterceptors, HttpInterceptorFn } from '@angular/common/http';

import { routes } from './app.routes';
import { provideClientHydration, withEventReplay } from '@angular/platform-browser';

// jwtInterceptor appends a mock JWT authorization token to all outgoing requests.
export const jwtInterceptor: HttpInterceptorFn = (req, next) => {
  const mockJwt = 'mock-jwt-token-value-here';
  const authReq = req.clone({
    setHeaders: {
      Authorization: `Bearer ${mockJwt}`
    }
  });
  return next(authReq);
};

export const appConfig: ApplicationConfig = {
  providers: [
    provideBrowserGlobalErrorListeners(),
    provideRouter(routes),
    provideClientHydration(withEventReplay()),
    provideHttpClient(withInterceptors([jwtInterceptor]))
  ]
};
