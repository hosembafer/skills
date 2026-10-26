---
name: runtime-qa
description: Use when the user asks for runtime or manual QA of a change in the running app through Chrome DevTools MCP ("/runtime-qa", "run runtime QA", "QA this branch in Chrome", "verify it in the browser"), including when several QA sessions or worktrees run at the same time.
argument-hint: "[what to test: a page, flow, ticket or behavior]"
allowed-tools: Bash(bash ${CLAUDE_SKILL_DIR}/scripts/port.sh *)
---

# Runtime QA

Test the running app the way a manual QA engineer would, driving Chrome through **Chrome DevTools MCP** in Claude Code
or Codex. Each session uses a dev-server **slot** and a dedicated MCP browser profile, so different checkouts or worktrees
can run side by side without sharing a server, cookies, storage or sign-ins.

Slot `n` is the `n`-th port of the dev server's port series. The MCP server owns and launches Chrome; its profile is
configured separately from the app port. `${CLAUDE_SKILL_DIR}` is the folder holding this file; in Codex, resolve that
folder from the loaded `SKILL.md` and use its quoted path.

Discover the available Chrome DevTools MCP tools without opening the browser. Server prefixes vary by client and
configuration; identify them by tools such as `list_pages`, `take_snapshot`, and `list_network_requests`. Read
[references/chrome-devtools.md](references/chrome-devtools.md) when setup or session isolation needs attention. If the
tools are unavailable, report the required setup instead of falling back to a browser extension.

## No fixes without approval

Change code or config to fix a finding only after the user approved that specific fix in this conversation, after you
proposed it and named the files it touches. A finding that blocks other test cases doesn't change this: mark those
cases blocked and go on with the rest.

## 1. Dev server

Find the project's dev command and its default port (package scripts, the dev tool's config, the README). Then run:

```bash
bash "${CLAUDE_SKILL_DIR}/scripts/port.sh" <default port>
```

It scans the series `<default port>`, `+100`, `+200`, … and prints a decision line plus a `taken` line for each port
that another checkout or program holds:

- `reuse <port> slot <n>`: a server from this checkout already runs there. Use it as it is; don't restart it.
- `start <port> slot <n>`: run the dev command with its port option set to `<port>`, as
  `nohup <command> > <scratchpad>/dev-<port>.log 2>&1 &` (a background task would be stopped at its time limit).
  Wait until the port listens (a background until-loop on `lsof -nP -iTCP:<port> -sTCP:LISTEN`), then run `port.sh`
  again; it must now print `reuse <port>`. If the server exits or fails to compile, show the end of the log and stop.
- Exit 2: every port in the series is taken. Show the `taken` lines and ask which server to stop.

Never stop or restart a server from a `taken` line. If the app pins its own origin in config (an auth redirect URL, an
allowed-origins list), point it at the new port through the project's local, untracked settings before starting.

Testing a deployed URL instead: skip this step; report the URL and MCP browser identity instead of a local port and slot.

## 2. Test plan

Draft the plan as in [references/test-plan.md](references/test-plan.md): from the arguments if the user gave a scope,
otherwise from the branch's changes. Show it and get the user's approval before any browser action. An approval already
given for the same plan still applies.

## 3. Browser

1. Use one MCP server namespace and its dedicated profile for this run. Confirm isolation from its launch config:
   `--isolated` gives each server process a temporary profile; `--userDataDir` gives it an explicitly assigned persistent
   profile. Neither a different namespace nor a new tab alone proves isolation. Do not attach to the user's personal
   Chrome or a browser used by another QA session.
2. Read [references/browser-gotchas.md](references/browser-gotchas.md), then call `list_pages` on that server. Create a
   fresh `about:blank` page with `new_page`, record its page ID and the server namespace, and keep all browser calls on
   that pair. Pass `pageId` when the installed tools expose it; otherwise use `select_page` before operating on the page.
3. Navigate that page to the app with `navigate_page`. The MCP attaches before navigation and collects network and
   console events; use `list_network_requests` and `list_console_messages` to inspect the initial load.
4. If the app asks for a sign-in, use the project's approved QA login recipe or ask the user to sign in in this window.
   Verify an existing session's identity through the app's account UI before reusing it. Persistent profiles retain
   sign-ins; temporary profiles do not retain them after the browser closes.
5. Take a snapshot, wait for the expected app content, then execute the approved plan. If browser setup, login or the
   environment blocks a case, report the blocker and continue with unaffected cases.

## 4. Run the plan

Go flow by flow. Check every case's checkpoint as described in `references/test-plan.md`. Use snapshot element UIDs for
interactions, `evaluate_script` for read-only DOM or state probes, and MCP's request and console detail tools for evidence.
Follow the installed tool schemas; parameter names and page routing can differ between versions.

After each flow, post the scoreboard, the findings and the off-plan observations.

## 5. Fix and re-test

For each failure, propose a fix with its root cause and the files it touches, and wait. After an approved fix, wait for
the rebuild, reload the same page with `navigate_page` (`type: "reload"`; use `ignoreCache: true` for stale chunks),
re-run the affected cases and post the updated scoreboard. Repeat until every case passes or the user stops.

## Reply

Leave the dev server and MCP browser available for follow-up work; the client may close Chrome when its MCP server
stops. End with the final scoreboard, findings, off-plan observations, app URL, local port and slot when applicable,
MCP server namespace and page ID, and the server log path if this session started the server.

## Verifying changes to this skill

Use [references/regression-plan.md](references/regression-plan.md) for the reusable public-site regression plan covering
two isolated browser contexts, network and console evidence, local-storage separation, cleanup and independent
navigation. This is a maintenance check for the skill; ordinary app QA uses the user's own scope.
