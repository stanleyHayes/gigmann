# Auth cookie/BFF migration

Gigmann currently uses a standard SPA bearer-token model: short-lived access
tokens, single-use rotating refresh tokens, strict CORS, and no ambient browser
credentials. This is acceptable for the demo/release candidate, and the remaining
hardening path should be done as one architecture migration.

## Why not a partial cookie switch

Moving only the refresh token to an httpOnly cookie while the API and static SPA
remain cross-origin changes the threat model without completing the controls:

- cookies become ambient credentials, so state-changing routes need CSRF tokens;
- cross-site cookies require `SameSite=None; Secure`, custom domains, and careful
  CORS `credentials` handling;
- WebSocket and API clients need coordinated cookie/session behavior;
- split Render subdomains make same-site semantics fragile until production
  domains are finalized.

## Target shape

1. Serve the SPA and API under the same site, preferably through a BFF or a
   stable `app.example` / `api.example` custom-domain pair with a deliberate
   cookie domain.
2. Store refresh/session state in httpOnly, `Secure`, `SameSite=Lax` cookies
   where same-site deployment permits it. If cross-site is unavoidable, require
   `SameSite=None; Secure` and a stricter CORS allow-list with credentials.
3. Add CSRF protection to every cookie-authenticated mutation using synchronizer
   tokens or double-submit tokens bound to the session.
4. Keep access tokens short-lived and server-issued; the frontend should not
   persist bearer credentials once cookie sessions are active.
5. Add e2e tests for login, refresh, logout, CSRF rejection, password reset, MFA
   enrollment/disable, and WebSocket connection under the deployed domain model.

This should be treated as a release-hardening follow-up after production domains
are final, not a component-level UI change.
