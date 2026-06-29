# Staging smoke + UAT + beta gates (GEC-111)

Use this file as the human sign-off record for the final release gate. Automated
checks can prove the deployment is serving the happy path; only the owner/team can
accept UAT and beta.

## Release candidate
- Staging URL:
- Commit SHA:
- Release date:
- Owner/UAT lead:

## Automated smoke (every deploy)
- The **Smoke** workflow (`.github/workflows/smoke.yml`) runs `scripts/smoke.sh`
  against the supplied staging URL. Its manual default is **two consecutive runs**
  (health → ready → login → grounded Daily Brief → metrics).
- The **E2E** workflow (`.github/workflows/e2e.yml`) runs the demo-narrative
  Playwright spec (login → brief → network → ask → my-day → approvals) against a
  freshly-built API + SPA. Dispatch it manually with `repeat_count=2` before any
  stakeholder demo.
- Plus the CI gates: backend coverage (>80%), frontend tests, lint, codegen-drift,
  integration (testcontainers), CodeQL, Trivy, secret-scan, Lighthouse budgets.

## Automated gate record
| Gate | Run URL / evidence | Result | Date | Reviewer |
|---|---|---|---|---|
| CI | | | | |
| Smoke run 1 | | | | |
| Smoke run 2 | | | | |
| E2E run 1 | | | | |
| E2E run 2 | | | | |
| DAST staging scan | | | | |

## UAT scenarios (manual, per persona)
| Persona | Scenario | Expected | Result | Notes |
|---|---|---|---|---|
| Executive | Sign in (+ MFA if enrolled) | Access to cockpit; failed login is generic | | |
| Executive | Read the Daily Brief | Worst item is first; figures look right; source/freshness visible | | |
| Executive | Open a facility from Network/search | Detail shows only computed inventory, staff, alerts, and figures | | |
| Executive | Search "Kasoa" | Ranked facility match opens the right facility | | |
| Executive | Ask "Why is Tafo critical?" | Grounded answer cites real facilities/figures only | | |
| Executive | Turn a brief item into a task | New task appears in My Day with source trace | | |
| Executive | Approve/decline an approval | Confirmation is required; decided item cannot be re-decided | | |
| Executive | Update watched metrics in Settings | Preference saves and reorders watched KPIs | | |
| Executive | Toggle theme and reload | Theme is remembered; circular reveal degrades gracefully | | |
| Facility manager | Sign in | Network-wide aggregate routes return 403 unless explicitly relaxed by owner | | |
| Facility manager | Review assigned tasks/alerts | Data is limited to the assigned facility; no cross-facility access | | |

## Beta gates
- **Entry:** all CI gates green; smoke + UAT scenarios pass on staging; the pasted
  dev Anthropic key rotated; secrets set in the Render env group.
- **Exit (GA):** no Sev1/Sev2 open; SLOs (`infra/observability/slo.md`) met for the
  beta window; runbooks (`docs/runbooks.md`) reviewed; acceptance package signed
  (`docs/acceptance-handover.md`).

## Sign-off
UAT and beta sign-off are **human gates**. Record the decision here before flipping
GEC-111 to `☑`.

| Gate | Decision | Name | Date | Notes |
|---|---|---|---|---|
| UAT | | | | |
| Beta entry | | | | |
| Beta exit / GA | | | | |
