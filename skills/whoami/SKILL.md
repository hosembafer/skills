---
name: whoami
description: Use when the user invokes /whoami or asks which agent is running and which account is signed into the current agent session, including OpenAI Codex, Claude Code, and Gemini CLI.
---

# Identify the current agent session and its logged-in user

Identify the current agent and its signed-in account. Codex supports an account lookup; other agents use session context.

## Instructions

1. Identify the agent (and its version, if known) from session context. The presence of a `codex` executable does not establish which agent is running.
2. Identify the signed-in account:
   - **Codex:** If an account method connected to the current session is available, call `account/read` with `{"refreshToken": false}` and use `result.account.email` for a ChatGPT account.
   - **Local Codex without that method:** Run the bundled helper, resolving the path relative to this skill's directory:

     ```bash
     python3 "<skill-directory>/scripts/codex_account.py"
     ```

     It performs the app-server initialization handshake, then calls only `account/read` with token refresh disabled. It prints the email or `unknown`, without exposing credentials. This reads the local CLI login; it does not establish the account of a remote or separately authenticated app session. For those sessions, use a connected account method or explicit account information from session context.
   - **Other agents:** Use the name or email explicitly supplied in session context, without shell probes.
3. Report `unknown` when a Codex lookup is unavailable, fails, or returns no account or email (including API-key authentication). On a context-only path, report `unknown` if no name or email is explicitly supplied. Never infer the signed-in account from filesystem paths, OS usernames, Git settings, or a connected third-party account. Do not read credential files, decode tokens, refresh credentials, or start a login flow.
4. Output exactly two lines, nothing else — no preamble, progress messages, trailing commentary, or follow-up offers:

   ```
   Agent: <agent name and version if known>
   User:  <name and/or email, or "unknown">
   ```

The Codex lookup follows the [official OpenAI app-server account documentation](https://learn.chatgpt.com/docs/app-server#1-check-auth-state).
