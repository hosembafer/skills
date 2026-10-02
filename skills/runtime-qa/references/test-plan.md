# Test plan and findings

## Plan

One block per flow. TC ids run on across flows (TC001, TC002, …).

```markdown
## <Flow name>

- **TC001: <short title>**
  - **Steps:** 1. <step> → 2. <step>
  - **Expected:** <what the user sees>
  - **Checkpoint:** <e.g. `GET /api/orders` returns 200; no console errors>
  - **Status:** PENDING
```

Status is one of `PENDING`, `PASSED`, `FAILED`. A TC that cannot run keeps `PENDING`, and the findings say what blocks
it and what would unblock it.

### Where the TCs come from

- **The user named a scope** (arguments, ticket, page, behavior): that scope is the plan.
- **No scope:** the branch's changes. Run `git diff <default branch>...HEAD --stat`, map each changed area to the pages
  that render it (routes, page components, templates), then write:
  - one smoke TC per touched page: it loads, its main content renders, no console errors, its API calls succeed;
  - one TC per changed behavior: new inputs and controls, dialogs and drawers, permission-dependent UI, empty,
    loading and error states, and sorting, filtering or pagination when the diff touches them.

## Checkpoints

Every TC has one. Network and console are the default pair.

- **Network:** the page's requests (`read_network_requests` in Claude in Chrome). A 4xx or 5xx the TC did not expect
  fails it. Record it as `METHOD /path → status`.
- **Console:** the page's console, filtered to what matters (`read_console_messages` with a `pattern`). Page errors fail
  the TC. Third-party noise (analytics, chat widgets) is reported once per session and fails nothing.
- **Visual:** a screenshot for any layout or visual defect, described in plain words.
- **State:** script reads of the DOM or app state (`javascript_tool`; see `browser-gotchas.md` before writing one).

## Scoreboard

After each flow, repost the whole plan with updated statuses.

## Findings

One block per failed or blocked TC:

```markdown
- **TC004 FAILED: <title>**
  - **Steps to reproduce:** 1. … → 2. …
  - **Expected:** …
  - **Actual:** …
  - **Evidence:** `POST /api/orders/search → 500`; console: `<verbatim excerpt>`; screenshot if visual.
  - **Suspected cause:** `path/to/file.ts:42`, if found. Proposed fix, not applied.
```

A blocked TC uses `BLOCKED` in the heading and gives the blocking reason and what unblocks it.

**Off-plan observations:** anything unexpected seen along the way (an error on an unrelated widget, a failing
background request) goes in its own list. Note it; don't investigate it.
