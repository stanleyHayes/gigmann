# Release hardening checklist

This records the in-repository controls now present for the production-readiness
push. It does not replace the formal staging penetration test or human UAT sign-off.

## Implemented controls

- API security headers: strict JSON-only CSP, HSTS in production, frame denial,
  no-sniff, no-referrer, COOP/CORP, Origin-Agent-Cluster, DNS-prefetch off,
  cross-domain policy denial, Permissions-Policy, and `Cache-Control: no-store`.
- Frontend static headers in `infra/render.yaml`: CSP with a JSON-LD hash, HSTS,
  frame denial, no-sniff, no-referrer, COOP/CORP, Origin-Agent-Cluster,
  DNS-prefetch off, cross-domain policy denial, and Permissions-Policy.
- Auth and authorization: bearer tokens, MFA + recovery codes, refresh-token
  revalidation, hash-only single-use password reset tokens, per-user session
  revocation after password/MFA changes, explicit reset-password strength checks,
  brute-force rate limits with `Retry-After` responses, executive-only aggregate
  views, facility manager scoping, and IDOR regression coverage.
- Data protection readiness: synthetic demo data, RBAC/facility scoping, audit logs,
  encrypted managed services path, Ghana Act 843 alignment notes, demo privacy
  notice, demo terms, and a security reporting channel.
- Security automation: gitleaks, govulncheck, npm audit, CodeQL, SBOM, Trivy image
  scan, OWASP ZAP baseline tooling, and a documented internal security assessment.
- SEO/crawler hygiene: static public pages, canonical links, Open Graph/Twitter tags,
  JSON-LD structured data, public-only sitemap, robots rules that disallow the
  authenticated cockpit, and `/.well-known/security.txt`.

## External evidence still required

- Formal penetration test against a deployed staging URL, with critical/high
  findings triaged and remediated before GEC-82 closes.
- Human UAT/beta sign-off, recorded in `docs/uat-checklist.md`, before GEC-111
  closes.
- Production organisation privacy policy, terms, lawful-basis record, retention
  schedule, support contacts, and any processor/transfer agreements before real
  personal data is processed.
- Optional auth-storage hardening: move the SPA/API behind a same-site BFF or
  cookie-aware API domain before switching refresh tokens to httpOnly cookies and
  CSRF tokens. Do not store only part of the auth state in cookies.
