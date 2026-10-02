---
name: runtime-qa
description: Use when the user asks for runtime or manual QA of a change in the running app with Claude in Chrome or Codex's Chrome extension ("/runtime-qa", "run runtime QA", "QA this branch in Chrome", "verify it in the browser"), including when several QA sessions or worktrees run at the same time.
argument-hint: "[what to test: a page, flow, ticket or behavior]"
allowed-tools: Bash(bash ${CLAUDE_SKILL_DIR}/scripts/port.sh *) Bash(bash ${CLAUDE_SKILL_DIR}/scripts/chrome.sh *)
---

# Runtime QA

Test the running app the way a manual QA engineer would, driving Chrome through your agent's browser extension: Claude
in Chrome in Claude Code, the ChatGPT extension in Codex. Every session gets a **slot**: its own dev-server port and its
own Chrome instance with a separate profile. Sessions in different checkouts or worktrees then run side by side without
sharing a server, cookies, storage or sign-ins.

Slot `n` is the `n`-th port of the dev server's port series and the Chrome profile `slot-n`. The scripts below pick
and open both; run them exactly as written. `${CLAUDE_SKILL_DIR}` is the folder holding this file; in Codex, write that
path out.

## No fixes without approval

Change code or config to fix a finding only after the user approved that specific fix in this conversation, after you
proposed it and named the files it touches. A finding that blocks other test cases doesn't change this: mark those
cases blocked and go on with the rest.

## 1. Dev server

Find the project's dev command and its default port (package scripts, the dev tool's config, the README). Then run:

```bash
bash ${CLAUDE_SKILL_DIR}/scripts/port.sh <default port>
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

Testing a deployed URL instead: skip this step and use the lowest slot that `chrome.sh --list` shows idle.

## 2. Test plan

Draft the plan as in `references/test-plan.md`: from the arguments if the user gave a scope, otherwise from the
branch's changes. Show it and wait for the user's approval before any browser action.

## 3. Browser

1. Open the slot's Chrome on the app. In Codex, add `--agent codex` and run it outside the sandbox, which can't start
   Chrome:

   ```bash
   bash ${CLAUDE_SKILL_DIR}/scripts/chrome.sh <slot> <app origin>
   ```

   The `id` line is the id your agent selects this Chrome by; browser names don't matter. `extension missing` means
   this slot hasn't been used with your agent before: the script opened the extension's store page in the slot's
   window. Ask the user to add the extension and sign in to it there, then read the slot's id from `chrome.sh --list`
   (`claude=` or `codex=`).

2. Connect this session to that Chrome and open a tab, as below for your agent.
3. If the app asks for a sign-in, ask the user to sign in in that window. The profile keeps the session for later runs
   of this slot.

### Claude Code

1. Load the `claude-in-chrome` skill, then every browser tool you will need in one ToolSearch call, including
   `list_connected_browsers`, `select_browser` and `switch_browser`.
2. Call `list_connected_browsers`. If the slot's id is listed and already `inUse`, go on. If it is listed but not in
   use, ask with AskUserQuestion which browser to use: the slot's id first, labeled with its name and `slot <n>`, and
   marked Recommended. Call `select_browser` with the answer.
3. If the id is `none` or not listed, call `switch_browser`. The prompt appears in every Chrome that has the extension,
   so tell the user to click Connect only in the slot's window.
4. Call `tabs_context_mcp` with `createIfEmpty: true`. Use the empty tab it creates, or `tabs_create_mcp` when the group
   already has tabs.
5. Call `read_console_messages` and `read_network_requests` once on that tab. They only record from their first call,
   so they must run before the page loads. Then navigate to the app origin.

### Codex

1. Set up the browser runtime the way the Chrome plugin's `control-chrome` skill describes, up to
   `const agent = await setupBrowserRuntime()`.
2. Bind the slot's Chrome by its id. `agent.browsers.get("chrome")` can bind a different Chrome, so call
   `agent.browsers.list()` instead, take the `extension` browser whose `metadata.extensionInstanceId` equals the
   slot's id, and bind it with `agent.browsers.get(<that browser's id>)`. Read its whole `documentation()` before using
   it.
3. If no browser has that id, ask the user to check that the ChatGPT extension is signed in in the slot's window, then
   list again. If it is still missing, follow the plugin's `chrome-troubleshooting` documentation.
4. Name the session, open a new tab, start the network and console capture its documentation describes, then navigate
   to the app origin. Codex asks the user before its first visit to a site, so tell the user to allow the app's origin.
   A non-interactive run such as `codex exec` can't ask, and the visit is denied.

## 4. Run the plan

Go flow by flow. Check every case's checkpoint as described in `references/test-plan.md`. In Claude Code, read
`references/browser-gotchas.md` before your first `javascript_tool` call and whenever a click, read or measurement
behaves oddly. In Codex, the browser's `documentation()` covers the same ground.

After each flow, post the scoreboard, the findings and the off-plan observations.

## 5. Fix and re-test

For each failure, propose a fix with its root cause and the files it touches, and wait. After an approved fix, wait for
the rebuild, reload the page, re-run the affected cases and post the updated scoreboard. Repeat until every case passes
or the user stops.

## Reply

Leave the dev server and the slot's Chrome running. End with the final scoreboard, the findings, the off-plan
observations, the port and slot, and the server log path if this session started the server.
