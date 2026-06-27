# Staging smoke + UAT + beta gates (GEC-111)

## Automated smoke (every deploy)
- The **E2E** workflow (`.github/workflows/e2e.yml`) runs the demo-narrative
  Playwright spec (login → brief → network → ask → my-day → approvals) against a
  freshly-built API + SPA. A green run is the staging smoke gate.
- Plus the CI gates: backend coverage (>80%), frontend tests, lint, codegen-drift,
  integration (testcontainers), CodeQL, Trivy, secret-scan, Lighthouse budgets.

## UAT scenarios (manual, per persona)
**Executive (Sammy):** sign in (+ MFA if enrolled) → read the Daily Brief, confirm
the worst item is first and figures look right → drill into a facility → quick-search
"Kasoa" → ask "Why is Tafo critical?" and check the answer cites facilities → turn a
brief item into a task → approve/decline an approval → tune watched metrics in
Settings → toggle theme and reload (remembered).
**Facility manager:** sign in → confirm scope is limited to the assigned facility
(no cross-facility data) → review tasks/alerts for that facility.

## Beta gates
- **Entry:** all CI gates green; smoke + UAT scenarios pass on staging; the pasted
  dev Anthropic key rotated; secrets set in the Render env group.
- **Exit (GA):** no Sev1/Sev2 open; SLOs (`infra/observability/slo.md`) met for the
  beta window; runbooks (`docs/runbooks.md`) reviewed; acceptance package signed
  (`docs/acceptance-handover.md`).

## Sign-off
UAT and beta sign-off are **human gates** — the owner runs the scenarios on the
deployed staging URL and records pass/fail + sign-off here.
