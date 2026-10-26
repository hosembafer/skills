# Runtime QA regression plan

Run this plan after changing `runtime-qa` to check that browser control, evidence collection and storage isolation still
work. It tests the skill against public pages, without a project dev server or login. Use the skill version being
changed, including its local references, rather than assuming an installed copy has updated.

Present these three cases for approval before browser actions. An approval already given for this run still applies.
Start each new run with all statuses `PENDING`; the recorded baseline below is historical evidence, not a new result.

## Setup

- **Target:** [OpenAI's public homepage](https://openai.com/).
- **Connection:** Chrome DevTools MCP with network, console, snapshot and script tools available.
- **Isolation:** read the installed `new_page` schema. When it exposes `isolatedContext`, create two blank pages in
  different, uniquely named contexts, then navigate both to the target:

  ```json
  {"url": "about:blank", "isolatedContext": "runtime-qa-<run-id>-a"}
  {"url": "about:blank", "isolatedContext": "runtime-qa-<run-id>-b"}
  ```

  Record each returned page ID and context name. Reusing a context name shares its storage. Two ordinary tabs in the
  same context do not establish storage isolation. The server's `--isolated` flag creates a profile per server process;
  it does not isolate each tab. If named contexts are unavailable, use two dedicated MCP servers with separate profiles
  as described in [chrome-devtools.md](chrome-devtools.md).
- **Routing:** keep every tool call on the recorded server and page ID. Use `pageId` when exposed; otherwise select the
  correct page before acting. Run both pages at the same time, and check their origins match before testing storage.
- **Evidence:** capture title, pathname, visible main content, relevant request statuses and console messages per page.
  Read evidence before navigating away. Save any screenshots outside the repository and omit credentials or query
  parameters that are irrelevant to the assertion.

## Public-page smoke test

- **TC001: Both isolated pages load the homepage.**
  - **Steps:** Navigate A and B to the target. Take a snapshot of each. Check each page's title, visible main content,
    document request and console messages.
  - **Expected:** Both pages render OpenAI's homepage at the same origin. Each page has its own ID and isolation context.
  - **Checkpoint:** Both main document requests return 200; both snapshots show meaningful page content. Report page
    errors and unexpected failed requests. Classify third-party noise separately under the normal test-plan rules.
  - **Status:** PENDING

## Storage isolation

- **TC002: Local storage stays separate in both directions.**
  - **Steps:** Choose a unique key such as `__runtime_qa_isolation_<run-id>`. Confirm it is absent in both contexts. Set
    its value to `context-A` in A; read it in B. Set it to `context-B` in B; read it again in A. Remove it from A and
    confirm B still holds `context-B`. Remove it from B and confirm it is absent in both contexts.
  - **Expected:** B cannot see A's value, and writing or removing B's value does not affect A. Each context retains only
    its own value until cleanup.
  - **Checkpoint:** Capture the sequence below using `evaluate_script`, returning only the QA marker and page origin.
    The origins must match; different origins would make this test inconclusive.
  - **Status:** PENDING

  | Operation | A's value | B's value |
  | --- | --- | --- |
  | Before writing | `null` | `null` |
  | Write in A | `context-A` | `null` |
  | Write in B | `context-A` | `context-B` |
  | Remove from A | `null` | `context-B` |
  | Remove from B | `null` | `null` |

This case explicitly allows writing and removing the test's own local-storage key. Use `localStorage`, since
`sessionStorage` is already separated by tab and would not prove context isolation. Keep ordinary app-state probes
read-only. Do not enumerate unrelated storage or clear all storage. If the case fails partway through, still remove
the test key from every reachable context.

Example functions for `evaluate_script`; pass one function per call and replace `<run-id>` consistently before executing:

```javascript
// Write in A. Use "context-B" for the write in B.
() => {
  const key = "__runtime_qa_isolation_<run-id>";
  localStorage.setItem(key, "context-A");
  return { origin: location.origin, value: localStorage.getItem(key) };
}

// Read in either context.
() => ({
  origin: location.origin,
  value: localStorage.getItem("__runtime_qa_isolation_<run-id>")
})

// Cleanup in each context.
() => {
  const key = "__runtime_qa_isolation_<run-id>";
  localStorage.removeItem(key);
  return { value: localStorage.getItem(key) };
}
```

## Independent navigation

- **TC003: Navigating A leaves B on the homepage.**
  - **Steps:** Record B's pathname and title. Take a fresh snapshot in A and click its Company/About link using the
    current UID. Wait for the destination's visible content. Read the pathname, title and main content in both pages.
  - **Expected:** A displays the About page while B still displays the homepage.
  - **Checkpoint:** A reaches `/about/`, its About content is visible, and the relevant navigation or client-route
    request succeeds. B remains at `/` with its original homepage title. Check network and console evidence per page.
  - **Status:** PENDING

If the site's layout changes, locate the current About link from its snapshot rather than reusing a saved UID. If the
site is unavailable or presents a challenge that prevents execution, report affected cases as blocked, without treating
that as proof of a skill regression. A navigation pass alone also does not prove storage isolation; TC002 must pass.

## Results to keep

Report each case's status, findings and off-plan observations, along with the tested skill path, MCP server, context
names, page IDs and final URLs. Confirm marker cleanup. Leave the pages available for follow-up work while the MCP
connection remains open. A deployed-site run has no local dev-server port, slot or server log.

## Recorded baseline: 2026-10-26

The migrated repository copy passed all three cases using one `chrome-devtools` server and two distinct named browser
contexts. Both homepage documents returned 200 and visible main content rendered. The local-storage values followed
the table above, and both markers were removed. A then reached [About](https://openai.com/about/) through the Company
link while B remained on the [homepage](https://openai.com/).

| Case | Status |
| --- | --- |
| TC001: Homepage in both contexts | PASSED |
| TC002: Local-storage isolation and cleanup | PASSED |
| TC003: Independent navigation | PASSED |

The browser identity for this run was:

| Server | Context | Page ID | Final URL |
| --- | --- | --- | --- |
| `chrome-devtools` | `runtime-qa-mvpej0z3-a` | 2 | `https://openai.com/about/` |
| `chrome-devtools` | `runtime-qa-mvpej0z3-b` | 3 | `https://openai.com/` |

These IDs identify the recorded run only; use fresh context names and returned page IDs on later runs. No local server
was used, so there was no dev-server port, slot or log.

Both contexts logged a Statsig cache warning and `POST https://statsigapi.net/v1/sdk_exception` returning 401. This was
reported as third-party telemetry noise; it did not prevent the tested page loads or navigation. Reassess failures on
future runs instead of treating this observation as a blanket exception.

The installed skill copy still used the extension workflow during this baseline; it was not the version tested.
