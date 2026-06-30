# User Guide — Ahenfie Executive Cockpit (GEC-115)

For Sammy Adjei (CEO) and facility managers. Ahenfie is your "chief of staff": it
reads the whole network every morning and tells you, in plain English, what needs
you — worst first.

## Signing in
Open the cockpit, enter your email + password. If you've enabled an authenticator
(Settings → MFA), you'll also enter a 6-digit code. Sessions refresh silently; you
stay signed in across reloads.

Use **Forgot password?** to reset access. The demo returns a short-lived reset
token on screen because no email/SMS delivery adapter is configured; production
deployments should deliver that token out-of-band. MFA remains enabled after a
password reset.

## Today — the Daily Brief (home)
The hero screen. A short narrated summary, then **prioritised items, worst first**.
Every item shows a status dot (🟢 good · 🟠 watch · 🔴 critical) and the figures
behind it — **every number is computed, never guessed by the AI**. Use the inline
actions to dig in or turn an item into a task. **Copy** or **Download** the brief to
share. If the AI is briefly unavailable you still get the same figures in plain prose.

## Network
All 12 facilities with health, region, and payer mix. Tap a facility for its
drill-down: inventory (stock-out projections), staff (roles, licence expiry,
attrition risk), and open alerts.

## Quick find (🔍 in the top bar)
Type a name or a natural phrase — "how is the Kasoa polyclinic doing" — and jump
straight to the facility.

## KPIs
Network revenue, patients, occupancy, and NHIS denial trends, week over week.

## Ask
Ask a plain-English question about the network. Answers are **grounded** in the
computed figures and cite the facilities they're based on — no invented numbers.

## My Day
Your task list. Tasks can come from a brief item or an alert (so you can trace why).
Move them todo → in progress → done.

## Approvals
Capital / hire / reorder requests routed to you, with the context to decide. Approve
or decline — the decision is logged. **Only you act**; the AI never approves anything.

## Settings
Switch light/dark (your choice is remembered), enrol MFA, and tune what you watch.

## Good to know
- Worst-first ordering means the top of the brief is where your attention belongs.
- Figures are end-of-day from the operational data; the brief refreshes through the day.
