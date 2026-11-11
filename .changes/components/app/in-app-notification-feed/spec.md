# In-App Notification Feed

Issue: #57

## Problem Statement

When a host confirms or rejects a booking, `notification-service` consumes the `booking.confirmed` or `booking.rejected` Kafka event from `system.events` and persists a `NotificationLog` record in PostgreSQL, even when SMTP credentials are not configured (`status="FAILED"`). Although `notification/main.py` already exposes `GET /notifications/user/{user_id}` (`list_user_notifications()`) and `GET /notifications/{log_id}`, two pieces are missing:

1. `gateway/default.conf` and `k8s/apps/gateway/gateway-configmap.yaml` have no `/api/notifications` route proxying requests to `notification-service`.
2. The Angular frontend (`ui/`) has no notification service or UI component (`ListingShellComponent` bell dropdown) for authenticated users to view their notifications in-app.

As a result, guests have no visibility into booking confirmation or rejection notifications inside the app, especially in local and development environments where SMTP is unconfigured.

## Solution

Expose `notification-service` through the API gateway under `/api/notifications` (protected by `auth_request /api/auth/verify`) and add an in-app notification bell dropdown in `ListingShellComponent` backed by a new `NotificationService`. When an authenticated user logs in or opens the bell dropdown in the top navigation bar, the UI fetches their notifications from `GET /api/notifications/user/{userId}` and renders the notification list (message, booking outcome badge, and timestamp), or an empty state when none exist.

## User Stories

1. As a logged-in user (guest or host), I want to see a notification bell in the header bar with my recent booking notifications (`booking.confirmed` and `booking.rejected`), so that I know when my booking requests are approved or rejected without relying on email delivery.
2. As a logged-in user with no notifications, I want to see a clear "No notifications yet" empty state when I open the notification dropdown, so that I know the feed is working and empty.
3. As a logged-out visitor, I should not see the notification bell in the header and unauthenticated requests to `/api/notifications/...` must be rejected with `401 Unauthorized`, so that notification logs are only accessible to authenticated users.

## Implementation Decisions

### 1. API Gateway (`gateway/default.conf` & `k8s/apps/gateway/gateway-configmap.yaml`)

- In both `gateway/default.conf` and the `default.conf` payload of `k8s/apps/gateway/gateway-configmap.yaml` (which `k8s/apps/gateway/gateway.yaml` mounts over `/etc/nginx/conf.d`), add section `# 10. Notification Service` (before `# 10. UI Service Fallback`, renumbering fallback to `# 11`):
  - Route `location ~ ^/api/notifications/user/([a-zA-Z0-9\-]+)$`:
    - Gate with `auth_request /api/auth/verify;`.
    - Use runtime DNS variable `set $upstream_notification "http://notification-service.raas.svc.cluster.local:80";` (relying on the existing `resolver kube-dns.kube-system.svc.cluster.local valid=10s;`) so nginx starts even if `notification-service` is not yet resolvable.
    - Proxy with `proxy_pass $upstream_notification/notifications/user/$1$is_args$args;` so query parameters (`?skip=0&limit=10`) are preserved across the variable `proxy_pass`.
    - Forward standard proxy headers (`Host`, `X-Real-IP`, `X-Forwarded-For`, `X-Forwarded-Proto`).
  - Route `location ~ ^/api/notifications/([a-zA-Z0-9\-]+)$`:
    - Gate with `auth_request /api/auth/verify;`.
    - Set `$upstream_notification` and proxy to `$upstream_notification/notifications/$1$is_args$args` with the same headers.

### 2. Frontend Model & Service (`ui/src/app/`)

- **`ui/src/app/models/notification.model.ts`**:
  - Define `NotificationItem` interface matching `NotificationLogResponse` from `notification/schemas.py`:
    - `id: string`
    - `user_id: string`
    - `message: string`
    - `notification_type: string`
    - `status: string`
    - `created_at: string`
  - Define `NotificationListResponse` interface matching `list_user_notifications()` in `notification/main.py`:
    - `total: number`
    - `count: number`
    - `items: NotificationItem[]`
- **`ui/src/app/services/notification.service.ts`**:
  - Singleton `@Injectable({ providedIn: 'root' })` using `inject(HttpClient)`.
  - `getUserNotifications(userId: string, skip = 0, limit = 10): Observable<NotificationListResponse>` calling `GET /api/notifications/user/${userId}` with `HttpParams` (`skip`, `limit`). The existing `jwtInterceptor` in `app.config.ts` automatically attaches `Authorization: Bearer <token>`.

### 3. Header Bell Dropdown (`ui/src/app/layout/listing-shell.component.*`)

- Implement the notification feed as a bell icon button and dropdown panel in `ListingShellComponent` (chosen over a tab in `ListingManageComponent` so notifications are visible from every page immediately after login and `ListingManageComponent`'s 5-tab bar stays uncluttered).
- **State, error handling & SSR safety**:
  - Inject `NotificationService` and `PLATFORM_ID` with `readonly` and explicit access modifiers per `guidelines/TYPESCRIPT.md`.
  - Signals: `readonly notifications = signal<NotificationItem[]>([])`, `readonly notificationsOpen = signal(false)`, `readonly notificationsLoading = signal(false)`, `readonly notificationsError = signal<string | undefined>(undefined)`.
  - Use an `effect()` guarded by `isPlatformBrowser(this.platformId)` that reads `const userId = this.authService.currentUser()?.id` and wraps state mutations in `untracked(() => ...)` so only `currentUser()` is tracked: when `userId` is present, call `this.loadNotifications(userId)`; when `userId` is absent (logout), clear `notifications`, reset `notificationsLoading` and `notificationsError`, and close the dropdown (`notificationsOpen.set(false)`).
  - `loadNotifications(userId: string)` sets `notificationsLoading.set(true)` and `notificationsError.set(undefined)`, subscribes to `notificationService.getUserNotifications(userId)` with both `next` and `error` handlers, checks `this.authService.currentUser()?.id === userId` before applying results (ignoring stale in-flight responses after logout/user switch), populates `notifications.set(res.items)` on success, or sets `notificationsError.set('Failed to load notifications')` on error, and clears `notificationsLoading.set(false)`.
  - Toggling the bell button via `toggleNotifications()` flips `notificationsOpen` and, when opening with an authenticated user, calls `loadNotifications(user.id)` to refresh the list.
- **Presentation & SMTP status handling**:
  - Note: `NotificationLog.status` in `notification/main.py` stores the *email delivery status* (`"SENT"` or `"FAILED"` when SMTP is unconfigured), NOT the booking status. The UI must **never** display `status: "FAILED"` as an error badge on the notification item. Instead, display the notification `message`, a formatted or readable `created_at` timestamp, and an optional semantic indicator derived from `message` (`CONFIRMED` vs `REJECTED`).
  - Inside the open dropdown, render loading state (`"Loading notifications..."`) when `notificationsLoading()` is true and `notifications()` is empty, render `notificationsError()` when an error occurred, render `"No notifications yet"` when not loading, no error, and `notifications()` is empty, or render the `@for (item of notifications(); track item.id)` list.
- **Accessibility & Guidelines (`guidelines/ANGULAR.md`, `guidelines/TYPESCRIPT.md`)**:
  - `ChangeDetectionStrategy.OnPush`, standalone by default (do not set `standalone: true`), signals and native control flow (`@if`, `@for` tracking `item.id`), `DatePipe` for `created_at`.
  - Bell `<button type="button">` includes `aria-label="Notifications"`, `[attr.aria-expanded]="notificationsOpen()"`, and `data-testid="notification-bell"`.
  - Dropdown panel includes `role="region"`, `aria-label="Notifications"`, and `data-testid="notification-dropdown"`.
  - No `@HostListener` decorator; if closing on Escape or outside click is needed, bind via the `@Component` `host` metadata object or button toggle.

## Testing Decisions

### Test Seams

1. **`NotificationService` (`ui/src/app/services/notification.service.spec.ts`)**: unit test with `provideHttpClient()` and `provideHttpClientTesting()` (`HttpTestingController`) verifying `GET /api/notifications/user/{userId}?skip=0&limit=10` request URL, query params, and response mapping.
2. **`ListingShellComponent` (`ui/src/app/layout/listing-shell.component.spec.ts`)**:
   - Keep existing admin-link tests passing by providing `NotificationService` stub (or `provideHttpClient()`).
   - Add unit tests verifying:
     - Bell button is rendered when authenticated (`guest`, `host`, `admin`) and hidden when logged out.
     - Clicking the bell button opens the dropdown and displays notification messages (both confirmed and rejected messages, even when `status === 'FAILED'`), or displays `"No notifications yet"` when the list is empty.

### Verification Levels

- **Unit / Integration (`ui`)**:
  - Run `npm run test -- --watch=false` inside `ui/`.
  - Run `npm run build` inside `ui/` to verify SSR compilation and bundle checks.
- **Deployment & HTTP Gateway (`kubectl` + `curl` / `playwright-cli`)**:
  - Redeploy `gateway` and `ui` via `powershell -File scripts/redeploy.ps1 -Service gateway` and `powershell -File scripts/redeploy.ps1 -Service ui`.
  - Verify `GET http://localhost:30080/api/notifications/user/<id>` without a token returns `401 Unauthorized`.
  - Verify `GET http://localhost:30080/api/notifications/user/<id>` with a valid `Authorization: Bearer <token>` returns `200 OK` with `{ total, count, items }`.
- **Browser E2E (`playwright-cli`)**:
  - Log in as a guest/host via `http://localhost:30080/login`, verify `[data-testid="notification-bell"]` is visible in the header, click it to open `[data-testid="notification-dropdown"]`, trigger or verify a booking confirmation/rejection notification, and verify the notification message renders in the dropdown.
- **Mobile Flow (`maestro-cli` reference YAML)**:
  - RaaS is a web application without a native Android/iOS binary (`android/` or `ios/`), so live Maestro execution is `n/a` (covered by `playwright-cli` on web/mobile viewport). Reference Maestro flow for mobile web wrapper if added later:
    ```yaml
    appId: com.raas.app
    ---
    - launchApp
    - tapOn:
        id: "notification-bell"
    - assertVisible: "Notifications"
    ```

## Out of Scope

- Changing `notification/main.py` database schema (`notification_logs`), adding read/unread persistence (`is_read` column), or adding `PATCH`/`DELETE` notification endpoints.
- Real-time push via WebSockets or Server-Sent Events (SSE).
- Adding a separate notifications tab to `ListingManageComponent` (the header bell dropdown in `ListingShellComponent` satisfies the requirement across all routes).
- Configuring an external SMTP server.

## Further Notes

### Adversarial review outcome

**Accepted and applied to spec:**
- **Gateway ConfigMap mount overrides image `default.conf` (`k8s/apps/gateway/gateway-configmap.yaml`):** `k8s/apps/gateway/gateway.yaml` mounts ConfigMap `gateway-config` at `/etc/nginx/conf.d`, masking `gateway/default.conf` from the Docker image, and `scripts/redeploy.ps1 -Service gateway` does not run `kubectl apply` on the ConfigMap. Updated the spec to require modifying both `gateway/default.conf` and `k8s/apps/gateway/gateway-configmap.yaml` and running `kubectl apply -f k8s/apps/gateway/gateway-configmap.yaml` before restarting the `gateway` deployment.
- **Gateway NodePort (`30000` vs `30080`):** `k8s/apps/gateway/gateway.yaml` exposes the nginx API gateway on NodePort `30000` (while `30080` in `k8s/apps/ui/ui.yaml` hits the Angular SSR Express container directly, bypassing nginx `/api/*` routing). Fixed Verification Levels to use `http://localhost:30000`.
- **Gateway query string preservation on variable `proxy_pass`:** In `gateway/default.conf`, using `set $upstream_notification ...` with regex `location ~ ^/api/notifications/user/([a-zA-Z0-9\-]+)$` drops query strings (`?skip=0&limit=10`) unless `$is_args$args` is appended. Added `$is_args$args` explicitly to `proxy_pass`.
- **Gateway auth gate (`auth_request /api/auth/verify`):** `notification/main.py` does not authenticate requests itself; the gateway location block must enforce `auth_request /api/auth/verify` so unauthenticated calls receive `401`.
- **Misleading `status: "FAILED"` field from `NotificationLog`:** When SMTP is unconfigured, `notification/main.py` saves every notification with `status="FAILED"`. Rendering `item.status` directly in the UI would mislead users into thinking their confirmed booking failed. Specified that the UI must display the message and derive any badge from the message content (`CONFIRMED` / `REJECTED`), never surfacing email `FAILED` status as an error.
- **`effect()` signal tracking & in-flight request error/staleness handling:** Specified `untracked()` inside the `effect()` so only `authService.currentUser()` is tracked, guarded subscription writes by matching `currentUser()?.id === userId`, and added explicit error handling (`notificationsError`) and test coverage so failed HTTP requests do not surface as unhandled errors.
- **SSR safety & existing `listing-shell.component.spec.ts` compatibility:** `ListingShellComponent` renders under Angular SSR (`RenderMode.Server`) and already has unit tests in `listing-shell.component.spec.ts`. Because Vitest runs in `jsdom` (`isPlatformBrowser === true`), using bare `provideHttpClient()` in existing tests would trigger live XHR calls when the `effect()` runs. Specified a `NotificationService` stub returning `of({ total: 0, count: 0, items: [] })`.
- **Accessibility & Angular/TypeScript guidelines:** Added explicit `aria-label`, `aria-expanded`, `ChangeDetectionStrategy.OnPush`, `readonly` + access modifiers, and prohibition on `standalone: true` / `@HostListener` per `guidelines/ANGULAR.md` and `guidelines/TYPESCRIPT.md`.

**Rejected:**
- **Enforcing cross-user ID matching (`jwt.sub == user_id`) inside `notification-service` or nginx:** `gateway/default.conf`'s `auth_request /api/auth/verify` pattern is used uniformly across `/api/bookings`, `/api/analytics/host/{id}`, and `/api/favorites` without forwarding identity headers or parsing JWTs in downstream services. Changing the cross-service auth architecture or modifying `notification/main.py` is out of scope for issue #57.
- **Adding a `/api/notifications/health` gateway route or replacing regex locations with a broad prefix `location /api/notifications/`:** Other services in `gateway/default.conf` use explicit regex locations and do not expose internal `/health` or `/metrics` endpoints externally through the gateway (Kubernetes probes query container port `8003` directly).
- **Adding both a bell dropdown in `ListingShellComponent` AND a tab in `ListingManageComponent`:** Issue #57 specifies "bell dropdown in `ListingShellComponent` or tab in `ListingManageComponent`". Implementing both duplicates UI and state for no user benefit (YAGNI).
- **Adding read/unread tracking (`localStorage` or DB migration) and pagination controls in the dropdown:** Not requested by issue #57; showing the most recent notifications (`limit=10`) on bell click is the smallest complete change.
- **Polling on an interval (`setInterval` / RxJS `timer`):** Background polling in `ListingShellComponent` complicates Angular SSR hydration stability (`whenStable()`) and unit testing; fetching on login and refreshing whenever the user opens the bell dropdown is deterministic and sufficient.
