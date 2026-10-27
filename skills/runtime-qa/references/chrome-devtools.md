# Chrome DevTools MCP setup and isolation

Use an available Chrome DevTools MCP server in the current client. The npm package is `chrome-devtools-mcp`; the MCP
server name and tool prefix are configurable. Both Claude Code and Codex use the same browser tools.

## Setup

The server needs a current Node.js LTS release with npm/npx and Google Chrome. This stdio server definition uses a
temporary profile per MCP server process:

```json
{
  "mcpServers": {
    "chrome-devtools": {
      "command": "npx",
      "args": ["-y", "chrome-devtools-mcp@latest", "--isolated"]
    }
  }
}
```

Use the client's MCP configuration format to register that command and those arguments. Merge with existing server
entries; do not replace unrelated configuration. Reconnect or restart the client if necessary to expose the tools.
Chrome starts on the first browser tool call, so defer that call until the test plan is approved. Do not use `--slim`
for this workflow: network and console evidence tools must be available.

See the upstream [client setup](https://github.com/ChromeDevTools/chrome-devtools-mcp/blob/main/docs/client-configurations.md)
and [configuration options](https://github.com/ChromeDevTools/chrome-devtools-mcp/blob/main/docs/configuration.md) for the
installed client and server version.

## Persistent profiles and concurrent runs

To keep app sign-ins across runs, replace `--isolated` with `--userDataDir=<absolute QA profile path>`. Resolve the path
before configuring it; do not assume the client expands `~` or shell variables in arguments. Use a dedicated profile
such as a new directory under `~/.runtime-qa/devtools/slot-0`, separate from the old extension profiles and personal
Chrome. Let MCP launch it consistently; opening it manually can invalidate stored sign-ins on macOS.

For concurrent runs, each MCP server process needs its own profile. Clients that start a server per conversation can
use the temporary-profile definition above. If the client shares servers between conversations, configure dedicated
entries such as `chrome-qa-0`, `chrome-qa-1`, each with a distinct `--userDataDir`, and assign one to each active run.
`--isolated` does not create a profile per tool call or per tab, and a new server name alone does not isolate storage.
Do not use the same persistent profile from two live MCP processes.

The dev-server slot from `port.sh` and the MCP profile are separate settings. Record both; do not assume selecting app
port slot 1 also selects a browser profile. With a deployed URL, only the dedicated MCP browser is needed.

If the profile is locked, check which process owns that exact profile. Reuse the current run's existing connection
when possible. Terminate a process only when it belongs to this run and is confirmed stale, then retry once. For an
active or uncertain owner, use another dedicated profile or report the blocker. Do not delete lock files or kill all
Chrome processes.

See upstream [concurrent sessions and profile behavior](https://github.com/ChromeDevTools/chrome-devtools-mcp/blob/main/docs/advanced-usage.md).
