# Claude in Chrome: what trips QA runs

## Reading requests and logs

- `read_network_requests` and `read_console_messages` only start tracking on their first call. Call each once before
  the action you want to observe, or its requests and logs are lost.
- `read_network_requests` gives URL, method and status, not bodies. To see a body, patch `XMLHttpRequest` (and
  `fetch`) in the page and push what you need into an array such as `window.__qaLog`, then read the array back with
  `javascript_tool`. A full page load removes the patch; client-side route changes keep it. Re-apply after every
  reload.
- `console.log` lines from callbacks started by `javascript_tool` (a promise's `.then`) are often missing from
  `read_console_messages`. Push them into `window.__qaLog` instead.

## `javascript_tool`

- End the script with the value you want back. Top-level `await` works; an `async () => {}` wrapper comes back as `{}`.
- Calls time out after about 45 seconds. Don't wait on `requestAnimationFrame` or long timers.
- A result is blocked when the script or the value it returns contains a query string or cookie data. Return
  `location.pathname`, never `location.href`.
- Never read tokens, `Authorization` headers or cookies, even to replay a request. To call an app API, trigger the UI
  action that sends it, or get the app's own HTTP client from its framework's dev tools, so the app adds the
  credentials itself.

## Clicking and typing

- The first click after a page load is sometimes swallowed. Take a screenshot, then click again if nothing changed.
- A click by element reference can miss framework handlers on buttons and router links. Click by coordinates, or call
  `el.click()` through `javascript_tool`, or `navigate` straight to the URL.
- `Escape` doesn't always close a menu or popover. Take a screenshot before the next click near it, so a stale menu
  item isn't clicked by mistake.
- Popovers and tooltips may open on hover, not click. Hover, then check that the panel has content.
- Don't press controls that open native `alert`, `confirm` or `prompt` dialogs; they block the extension until the
  user closes them.

## Layout and timing

- `resize_window` can take effect a few calls late. Read `innerWidth` again before trusting a width measurement.
- The automated tab can report `document.visibilityState === 'hidden'`. Then CSS animations and
  `requestAnimationFrame` pause, timers slow down, and Chrome skips View Transitions. Don't judge animation or
  transition behavior there; ask the user to look in a visible tab.
- To measure layout at a width without waiting a frame, set the width inline and read `getBoundingClientRect()` in
  the same script.

## Tabs and windows

- Work only in the tab group that `tabs_context_mcp` returns for this session. A tab the user opened themselves isn't
  reachable; open the same URL in a new tab.
- Closing the last tab of the group can remove the group. `tabs_context_mcp` with `createIfEmpty: true` makes a new one.

## Dev server

- The dev server serves the checkout's working tree, uncommitted changes included.
- After a rebuild, a stale chunk error (`ChunkLoadError`, `Failed to fetch dynamically imported module`) needs a hard
  reload. If it persists, the server needs a restart; ask first when the user started it.
- Wait for a server by polling its port (`lsof -nP -iTCP:<port> -sTCP:LISTEN`) or `curl`, not by grepping its log;
  logs are often colored and the match never comes.
