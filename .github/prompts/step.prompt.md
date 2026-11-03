---
description: "Implement the next unchecked step from the current PLAN file's Steps list, then stop for review."
agent: "Stepwise"
argument-hint: "(optional) step number or note"
---
If a step was already implemented and reported, the user has reviewed and approved it: tick it in
the `## Steps` list of the current `PLAN-*.md`. Then implement the next unchecked step (or the step
given in the argument), following the Stepwise rules exactly. Report and stop for the user's
review.
