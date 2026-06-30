# Ghana Data Protection Act, 2012 (Act 843) — Alignment (GEC-81)

This PoC processes a **synthetic** 12-facility network — no real personal or patient
data. This note records how the design aligns with Act 843 so the same controls hold
when real data is introduced, and is **not legal advice**.

## Principles → controls
- **Lawfulness, purpose specification & minimisation:** the cockpit stores only
  operational facility figures and the executive's own account/preferences; no
  patient records. Real deployments must register the controller with the Data
  Protection Commission and define a lawful basis per data category.
- **Accuracy:** figures are computed deterministically by the signal engine; the AI
  cannot alter them.
- **Storage limitation:** refresh tokens are single-use and expire; audit/log
  retention to be set per policy.
- **Security safeguards (s.28):** encryption in transit (TLS at Render) and at rest
  (Render-managed Postgres/Redis AES-256); secrets in env groups; access control
  (RBAC, facility scoping); audit logging; MFA.
- **Data subject rights:** the user record is a single row keyed by id/email — access,
  rectification (preferences), and erasure are straightforward to implement.
- **Cross-border transfer (s.18):** Claude (Anthropic) and Voyage process prompt text
  abroad. For real data, either (a) keep prompts free of personal data — the current
  design sends only computed figures and facility names — or (b) obtain consent /
  put a transfer agreement in place. The deterministic local narrator/embedder allow
  fully on-prem operation with **no third-party transfer**.

## Actions before processing real data
1. Register with the Data Protection Commission and appoint a data protection
   supervisor.
2. Complete a DPIA.
3. Sign DPAs with Anthropic/Voyage or disable them (local fallbacks).
4. Define retention and erasure runbooks for the production controller.
5. Replace the demo privacy notice at `frontend/public/privacy.html` with the
   production controller's published privacy notice and support process.
