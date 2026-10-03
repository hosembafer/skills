# Usage examples

Enter these prompts in your agent conversation after installing the skills. Claude Code uses `/skill-name`; Codex
uses `$skill-name`. Results below are illustrative: subjects, hashes, branch names, URLs, and accounts depend on the
repository and session. The [compatibility table](../README.md#compatibility) lists the required tools.

## Ship

You are on a feature branch with changes ready to commit, and the MR/PR should target `main`.

Claude Code:

```text
/ship main
```

Codex:

```text
$ship main
```

The agent reads the repository's rules and commit style, creates commits grouped by concern, pushes the branch, and
opens its MR/PR or reports the existing one. Its reply includes each commit's short hash and subject, the push result,
and the MR/PR link. If neither Git hosting CLI can open the MR/PR, it gives you a creation link and the drafted title
and description. Invoking the skill authorizes those actions for this run; unrelated changes and unclear targets
still need your decision.

## Merge to

You are on a feature branch whose committed changes should join the existing `staging` integration branch.

Claude Code:

```text
/merge-to staging
```

Codex:

```text
$merge-to staging
```

The agent pushes the feature branch, merges it into `staging` in a temporary worktree, and pushes the merge. Your
feature branch stays checked out, and uncommitted changes stay out of the merge. The reply identifies the merge
commit and both push results. Non-trivial conflicts are presented with both sides and a recommendation; the agent
waits for your decision before resolving them. Invoking the skill authorizes both pushes for this run.

## Commit subject

Your current diff fixes an empty search-result label, and the repository uses plain imperative commit subjects.

Claude Code:

```text
/commit-subject
```

Codex:

```text
$commit-subject
```

An example reply is:

```text
Fix the empty search-result label
```

The same line is on the macOS clipboard. The skill uses staged changes when present, otherwise unstaged changes,
and creates no commit. If neither diff contains changes, it prints and copies `No Changes Made`.

## Smart copy

The agent's previous reply contains a command you want to paste into your terminal.

Claude Code:

```text
/smart-copy the command from your previous reply
```

Codex:

```text
$smart-copy the command from your previous reply
```

The clipboard contains the command verbatim, without its code fence or surrounding explanation. An example reply is:

```text
Copied the command (1 line).
```

## Runtime QA

This branch changes a checkout form, and you want to verify its empty-field error and successful submission.

Claude Code:

```text
/runtime-qa verify the checkout form's empty-field error and successful submission
```

Codex:

```text
$runtime-qa verify the checkout form's empty-field error and successful submission
```

The agent chooses or reuses a dev-server port and drafts a test plan for your approval. Once approved, it opens the
slot's isolated Chrome profile and runs the plan using its browser extension. The first use of a slot may require
installing the extension and signing in. The final reply includes case statuses, findings and evidence, the port and
slot, and the server log path when the agent started the server. Fixes need separate approval. The server and Chrome
stay running for follow-up work.

## Who am I

Claude Code:

```text
/whoami
```

Codex:

```text
$whoami
```

A Codex session with a supported account lookup might return:

```text
Agent: OpenAI Codex
User:  developer@example.com
```

When the account cannot be established, the second line is `User:  unknown`. The local helper identifies the local
Codex CLI login, which may differ from a remote or separately authenticated app session. Claude Code and other
agents use account information explicitly supplied in session context.
