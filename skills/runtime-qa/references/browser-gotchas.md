# Chrome DevTools MCP: QA gotchas

## Pages and interactions

- Stay on this run's MCP server namespace and page ID. Pass `pageId` on page-scoped tools when available; older servers
  use `select_page` to set the target. Page IDs and element UIDs belong to that server, not to another MCP connection.
- Take a fresh `take_snapshot` before interacting. Use its element `uid` with `click`, `fill`, `fill_form` and `hover`.
  After navigation or a DOM change, refresh the snapshot instead of guessing or reusing stale UIDs.
- Prefer browser input tools for actions. `evaluate_script` reads can help diagnosis, but changing app state or calling
  handlers directly does not verify that a user can perform the flow.
- Popovers and tooltips may open on hover. Use `hover` and verify that the panel has content. Re-hover immediately
  before measuring hover styles, and check `el.matches(':hover')`.
- Use `handle_dialog` for native `alert`, `confirm` or `prompt` dialogs when the approved flow requires them. Check the
  dialog's text before accepting an action that changes data.

## Requests and console

- MCP collects events while attached to the page. Create the blank page first, then navigate; listing requests or
  console messages is a read, not a switch that starts capture. If attachment happened after the failing load, reload
  while attached to reproduce it.
- `list_network_requests` and `list_console_messages` normally cover the current navigation. Read evidence before
  leaving it; use `includePreservedRequests` / `includePreservedMessages` for recent earlier navigations when supported.
  Follow pagination when the returned list is incomplete.
- `get_network_request` can return headers and request/response bodies directly; do not patch `fetch` or
  `XMLHttpRequest` just to capture a response. Request details can contain credentials: omit tokens, cookies and
  `Authorization` / `Set-Cookie` values from reports or saved artifacts. Verify login identity in the app's account UI.
- Use `get_console_message` for the details of a listed `msgid`; correlate requests and messages with the current TC.

## `evaluate_script`

- Read the installed schema. The `function` argument normally takes a function expression such as
  `() => ({path: location.pathname, width: innerWidth})`; use an async function when awaiting a short read. Return
  JSON-serializable values rather than DOM nodes. This differs from extension tools that evaluate a bare script.
- Keep probes read-only and bounded. Use `wait_for` for expected page text instead of long timers or waiting on
  `requestAnimationFrame` inside a script.
- Never read tokens, cookies or passwords to replay an API request. Trigger the app's UI and inspect the resulting
  request instead. Return only the state needed for the assertion.

## Layout and timing

- After `resize_page`, confirm `innerWidth` and `innerHeight` before trusting viewport measurements.
- A background page can throttle timers and animations. Use `select_page` with `bringToFront` when supported and check
  `document.visibilityState`; do not report a paused background animation as an app defect.
- Frameworks can scope CSS animation names. Read `getComputedStyle(el).animationName` on the affected element instead
  of assuming its source name remains unchanged at runtime.

## Dev server and rebuilds

- The server serves the checkout's working tree, including uncommitted changes.
- After a rebuild, reload with `navigate_page` using `type: "reload"`. For stale chunks, set `ignoreCache: true`.
  If the error persists, check the server log; ask before restarting a server the user started.
- Wait for a server by polling its port (`lsof -nP -iTCP:<port> -sTCP:LISTEN`) or `curl`, not a colored log message.
- If multiple QA pages share a dev server, a rebuild may reload all of them. When the project supports it, disable
  automatic live reload for parallel runs while retaining rebuilds, and reload the affected page explicitly.

Use the upstream [tool reference](https://github.com/ChromeDevTools/chrome-devtools-mcp/blob/main/docs/tool-reference.md)
for version-specific parameters.
