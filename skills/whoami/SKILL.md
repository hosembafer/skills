---
name: whoami
description: Report which agent CLI is running this session (Claude Code, Codex, Gemini, etc.) and which user account is logged into it. Prompt-based — no file reads, no shell probes. Use when the user invokes /whoami or asks who is logged into the current agent session.
---

# Identify the current agent session and its logged-in user

Answer the question: "Which user does use this <agent> session right now?"

## Instructions

1. From your own session context, identify:
   - **Which agent CLI you are** (Claude Code, OpenAI Codex, Gemini CLI, Cursor, etc.) and, if known, its version.
   - **Which account is logged into that agent** — the user's name and/or email as it appears in your session context.
2. Do NOT read any files, run shell commands, or probe the environment. This is a pure introspection prompt — answer from what you already know about the session you are running in.
3. If you genuinely do not know one of the two pieces, say "unknown" for that field rather than guessing.
4. Output exactly two lines, nothing else — no preamble, no trailing commentary, no follow-up offers:

   ```
   Agent: <agent name and version if known>
   User:  <name and/or email, or "unknown">
   ```
