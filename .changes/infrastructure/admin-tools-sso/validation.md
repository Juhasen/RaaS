# Validation: admin-tools-sso

| Level | What was run | Result |
|---|---|---|
| Unit / integration (user) | `pytest` in `user/` (venv, Python 3.13, unpinned deps): cookie on login, `admin-check` via cookie and Bearer (200 + `X-Auth-User`), 403 for guest/host, 403 after demotion, 401 without/with bad token, `role=admin` register gives 422, logout clears cookie, seed (demote others, create, idempotent, password sync, no promotion of existing non-admin, short password), import fails without `JWT_SECRET` | pass, 15 tests |
| Unit (ui) | `npm run test -- --watch=false`, `npx ng build` | pass, 13 tests; build ok |
| Manifests | `kubectl apply -k k8s/infra/admin-tools`, `kubectl apply -k k8s/infra/monitoring` | applied; no NodePort left for the four tools |
| Deployment | `scripts/admin-secrets.ps1`, applied manifests, `redeploy.ps1` for user, gateway, ui; curl against `localhost:30000` | see below |
| Browser (Playwright, Edge headless) | log in as the admin, open each `/admin/<tool>/` | Grafana dashboard home, MinIO login page, Adminer login page, Console overview all render with no redirect loop |
| Manual | Click through the four buttons and log in where asked | pending the user |

Deployment checks (all against `localhost:30000`):

- `/admin/{grafana,minio,adminer,kafka}/`: no cookie 401; host cookie 403; admin cookie 200.
- Assets under the prefix: MinIO `styles/root-styles.css` 200 text/css, MinIO `api/v1/session` reaches MinIO (403 invalid session JSON), Adminer `?file=default.css` 200, Adminer login POST 302 to the prefixed URL, Console `static/js/index.*.js` 200 and `api/topics` JSON, Grafana `public/img/fav32.png` 200.
- Grafana as admin: `/admin/grafana/api/user` shows auth label "Auth Proxy" and the RaaS admin email. A different `X-WEBAUTH-USER` sent by the admin is ignored; sent by a host it gets 403.
- `POST /api/users/register` with `role=admin`: 422.
- Token signed with the old default secret: 401 on `admin-check` and on the admin path (header and cookie).
- Logout clears the cookie; the next admin request is 401.
- Old NodePorts 30901, 30090, 30808, 30300: connection refused. Old `redpanda-console-auth` objects deleted.
- Gateway restarted with the `adminer` and `redpanda-console` Services deleted: gateway came up, `/api/listings` 200, adminer path 502; services restored afterwards.
- `raas_token` stripping from the cookie forwarded to tools: tested the `map` in a stock nginx container (6 cookie shapes, including a lookalike name); not observed inside a real tool.
- All pods Running, 0 restarts (checked right after the final apply).

## Not verified

- Memory before/after: `kubectl top` has no Metrics API here and Prometheus has no container memory series. No new pod was added; `redpanda-console-auth` (16Mi request, 32Mi limit) was removed.
- Browser check covers landing pages only; logins inside MinIO and Adminer and deeper navigation were not driven. A first browser test by the user found a Grafana redirect loop (root_url without port); fixed and re-verified in the browser.
- Pre-existing admins demoted at startup: the cluster had only the account that is now the seeded admin, so the demote path ran on the unit test only.
- The 2-minute stability check was not repeated after the last restart cycle.
