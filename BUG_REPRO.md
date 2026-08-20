# Bug Reproduction

## What is wrong

Approval application methods do not consistently honor a caller's cancelled context. A cancelled request can still reach approval storage, write a decision, or run policy/state checks.

## How to trigger

Use a cancelled context with the approval application test cases covering decision reads/writes, plan approval listing, policy evaluation, and state validation. The baseline records storage calls and returns nil instead of stopping at cancellation.

## Observed error

The targeted tests report `cancelled decision reached storage`, `decision lost caller context`, `cancelled approval check reached storage`, `policy evaluation ignored cancellation`, and `approval validation ignored cancellation`.
