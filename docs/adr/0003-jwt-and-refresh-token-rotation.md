# ADR-0003: HS256 access tokens with single-use rotating refresh tokens

- **Status:** accepted
- **Date:** 2026-06-26
- **Deciders:** Owner (Stanley) + engineering

## Context
ADR-0001 D-002 chose a custom JWT auth owned in Go. We had to settle the concrete token scheme:
signing algorithm, access/refresh split, refresh-token storage, and reuse handling. An early
`infra/render.yaml` draft referenced an RS256 key pair, which implies key-pair management and a
JWKS surface we do not need for a single first-party API verifying its own tokens.

## Decision
- **Access tokens: JWT signed with HS256** (`golang-jwt/jwt` v5), short-lived, carrying the subject,
  role, and facility scope used for authorization at the app boundary. A single shared secret
  (`JWT_SECRET`, from the Render env group) signs and verifies — no key distribution, no JWKS.
- **Refresh tokens: opaque, random, single-use, and rotating.** Only the **SHA-256 hash** is stored
  (`RefreshTokenStore`), never the raw token. On refresh the presented token is looked up by hash,
  consumed (deleted), and a brand-new refresh token is issued alongside the new access token.
- **Reuse detection:** a refresh token that does not resolve to a live stored hash is rejected. Logout
  consumes the stored token.

## Consequences
- Simple, self-contained verification: the API needs only its own secret, which suits the stateless
  Render deployment. Rotating the secret invalidates all live sessions (acceptable; documented).
- Refresh-token theft is contained — a stolen token is single-use, so the legitimate client's next
  refresh (or the attacker's) invalidates the other; raw tokens are never at rest.
- HS256 is symmetric: any service that must *verify* tokens would need the signing secret. We have one
  verifier (this API), so that is a non-issue today. **If a second independent verifier ever appears,
  revisit with RS256/EdDSA + JWKS** (supersede this ADR).
- The store is in-memory for now (see ADR-0002), so rotation state is per-instance and resets on
  restart; it moves to Postgres/Redis with the persistence work.

## Alternatives considered
- **RS256/EdDSA asymmetric signing** — rejected for now: key-pair + JWKS management with no
  multi-verifier requirement. Documented as the upgrade path if that requirement emerges.
- **Stateless (non-stored) refresh tokens** — rejected: no server-side revocation or reuse detection.
- **Opaque access tokens with a server-side session lookup on every request** — rejected: adds a
  store round-trip to the hot path; stateless JWT verification is cheaper and sufficient.
