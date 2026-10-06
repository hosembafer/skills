# skills

A collection of agent skills for Claude Code and Codex, using the [Agent Skills](https://agentskills.io) format. Install them with the [`skills`](https://github.com/vercel-labs/skills) CLI.

## Skills

| Skill            | What it does                                                                                                 | Started by                 |
| ---------------- | ------------------------------------------------------------------------------------------------------------ | -------------------------- |
| [ship](skills/ship/SKILL.md) | Commits the working tree as granular commits in the repo's style, pushes the branch, and opens its MR/PR. | Explicit invocation only |
| [merge-to](skills/merge-to/SKILL.md) | Pushes the branch, merges it into a shared integration branch such as `staging`, and pushes the merge. | Explicit invocation only |
| [commit-subject](skills/commit-subject/SKILL.md) | Writes a one-line commit subject for the current changes in the repo's style and copies it to the clipboard. | Agent or user |
| [smart-copy](skills/smart-copy/SKILL.md) | Copies what you'll paste next (a continue prompt, a drafted message, a command) to the clipboard, verbatim. | Agent or user |
| [runtime-qa](skills/runtime-qa/SKILL.md) | Runs manual QA in Chrome through the agent's browser extension, each session on its own port and profile. | Agent or user |
| [whoami](skills/whoami/SKILL.md) | Reports which agent CLI is running the session and which account is logged into it. | Agent or user |

In Claude Code, invoke a skill with `/skill-name`. In Codex, mention `$skill-name` in your prompt or select it through
`/skills`. Arguments follow the skill name: `/merge-to staging` in Claude Code, `$merge-to staging` in Codex.
See [usage examples and expected results](docs/usage.md) for every skill.

`ship` and `merge-to` never start on their own. Running one is your approval for that run only: `ship` commits and
pushes, `merge-to` pushes the branch and one merge commit.

`merge-to` resolves a conflict on its own only when it can keep both sides' lines as written. It asks about every other
conflict, so the integration branch tests each branch as it will ship.

## Compatibility

| Skill | Claude Code | Codex | Operating system | Requirements |
| --- | --- | --- | --- | --- |
| `ship` | Yes | Yes | macOS or Linux | Bash, Git, an `origin` remote, and push access. Authenticated `gh` or `glab` opens the MR/PR; otherwise the skill gives you a creation link and draft. |
| `merge-to` | Yes | Yes | macOS or Linux | Bash, Git with worktree support, an `origin` remote, and push access to both branches. |
| `commit-subject` | Yes | Yes | macOS | Git and `pbcopy`. |
| `smart-copy` | Yes | Yes | macOS | `pbcopy`, `pbpaste`, and Perl. |
| `runtime-qa` | Yes, with Claude in Chrome | Yes, with the Chrome plugin | macOS | Bash, Git, `lsof`, Google Chrome, and the agent's installed, signed-in browser extension. |
| `whoami` | Agent identity; account from context | Agent identity; account lookup when available | Any OS for context or connected account lookup; macOS or Linux for the local helper | Local Codex account lookup needs Go and `codex` on `PATH`. The helper uses only the standard library and runs from an installed bundle. Other agents use explicit session context and may report `unknown`. |

The table describes each workflow's requirements. CI validates structure and syntax on macOS and Linux; it does not
exercise Git hosting, clipboard access, account lookup, or browser workflows. Other agents that read Agent Skills may
use these instructions, but their invocation policies and tool integrations have not been verified.

Script paths come from the loaded skill's directory. Claude Code exposes `${CLAUDE_SKILL_DIR}`; in Codex, the agent
resolves that directory from the loaded `SKILL.md` path and passes arguments from your prompt explicitly.

`runtime-qa` needs Google Chrome on macOS, and Claude Code with Claude in Chrome or Codex with its Chrome plugin. Each
QA slot is a separate Chrome instance with its own profile under `~/.runtime-qa/chrome`. The first run of a slot with
an agent asks you to add that agent's extension in the slot's window and sign in; later runs reuse the profile, its
extensions and its sign-ins.

## Install

You need Node.js with `npx` and Git. Use the repository URL below to check access without installing anything.
Use the same `skills_repository` variable for the installation commands in this shell:

```bash
skills_repository='https://github.com/hosembafer/skills.git'
git ls-remote "$skills_repository" HEAD
```

It should print a commit hash and `HEAD`. If access fails while the repository is private, authenticate with a
GitHub account that has access or use `git@github.com:hosembafer/skills.git` with a registered SSH key.

All skills, for Claude Code and Codex:

```bash
npx skills add "$skills_repository" -g --skill '*' -a claude-code -a codex
```

One skill:

```bash
npx skills add "$skills_repository" -g --skill ship
```

Leave out `-a` to choose agents interactively, and `-g` to install into the current project instead of your user directory.

## Update

```bash
npx skills update -g
```

Updates replace the installed copies, so edit skills in a clone of this repo and push, not in `~/.agents/skills`.

## Validate

From a clone, with Go 1.26 or later, Bash, and ShellCheck available. Install ShellCheck directly through your system
package manager, such as `brew install shellcheck` on macOS or `sudo apt install shellcheck` on Ubuntu.

```bash
go run ./cmd/validate-skills
go test ./...
go vet ./...
```

Go downloads the single YAML library automatically; `go.mod` and `go.sum` record its version and checksums.
The validator checks front matter, matching and unique skill names, local references, UI metadata, and consistent
explicit-invocation policies. It checks Bash and Go syntax without executing helpers, and fails on ShellCheck errors
and warnings. Tests cover the validator's acceptance and rejection of broken bundles, plus the `whoami` helper's
account protocol, safe output, timeouts, and process cleanup using a fake local server.

To build a standalone validator, run `go build -o /tmp/validate-skills ./cmd/validate-skills`, then
`/tmp/validate-skills /path/to/checkout`. The binary needs Bash and ShellCheck on `PATH` and defaults to validating
the current directory when no checkout path is supplied.

`.github/workflows/validate.yml` runs validation, tests, formatting, and vet checks on macOS and Linux for branch
pushes, pull requests, and version tags. CI installs a pinned ShellCheck binary with a verified checksum and checks
Go formatting in both the validator and skill helpers.

Every skill in this collection includes `agents/openai.yaml` with a display name, a 25–64 character description, and a
default prompt naming the skill. This is a repository convention; the Agent Skills format itself does not require
that file. Explicit-only skills also set `policy.allow_implicit_invocation: false` for Codex and
`disable-model-invocation: true` for Claude Code.

## Releases

[CHANGELOG.md](CHANGELOG.md) records changes to the collection. Releases use annotated Git tags named `vMAJOR.MINOR.PATCH`.
See [release and tagged-install instructions](docs/releases.md). Changes under `Unreleased` have no release tag yet.

## Layout

Each skill is a folder under `skills/` whose name matches the skill's `name` field:

```
skills/
  ship/
    SKILL.md
    scripts/ship.sh      # the git steps the skill runs
    agents/openai.yaml   # UI metadata; Codex: explicit invocation only
  merge-to/
    SKILL.md
    scripts/merge-to.sh  # the git steps the skill runs
    agents/openai.yaml   # UI metadata; Codex: explicit invocation only
  commit-subject/
    SKILL.md
    agents/openai.yaml
  smart-copy/
    SKILL.md
    agents/openai.yaml
  runtime-qa/
    SKILL.md
    scripts/port.sh      # picks the dev-server port for the checkout
    scripts/chrome.sh    # opens the slot's isolated Chrome profile
    references/          # test plan format, Claude in Chrome gotchas
    agents/openai.yaml
  whoami/
    SKILL.md
    scripts/codex_account.go
    scripts/codex_account_test.go
    agents/openai.yaml
```

## License

MIT. See [LICENSE](LICENSE).
