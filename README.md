# skills

My agent skills for Claude Code, Codex and other agents that read [Agent Skills](https://agentskills.io). Install them with the [`skills`](https://github.com/vercel-labs/skills) CLI.

## Skills

| Skill            | What it does                                                                                                 | Started by                 |
| ---------------- | ------------------------------------------------------------------------------------------------------------ | -------------------------- |
| `ship`           | Commits the working tree as granular commits in the repo's style, pushes the branch, and opens its MR/PR.    | `/ship` only               |
| `commit-subject` | Writes a one-line commit subject for the current changes in the repo's style and copies it to the clipboard. | Agent or `/commit-subject` |
| `smart-copy`     | Copies what you'll paste next (a continue prompt, a drafted message, a command) to the clipboard, verbatim.  | Agent or `/smart-copy`     |
| `runtime-qa`     | Runs manual QA in Chrome via Claude in Chrome, each session on its own dev-server port and Chrome profile.   | Agent or `/runtime-qa`     |
| `whoami`         | Reports which agent CLI is running the session and which account is logged into it.                          | Agent or `/whoami`         |

`ship` never starts on its own. Running it is your approval to commit and push in that run.

`commit-subject` and `smart-copy` use `pbcopy`, so they need macOS.

`runtime-qa` needs Claude Code with Claude in Chrome, and Google Chrome on macOS. Each QA slot is a separate Chrome
instance with its own profile under `~/.runtime-qa/chrome`. The first run of a slot asks you to add the Claude
extension in that window and sign in; later runs reuse the profile, its extension and its sign-ins.

## Install

All skills, for Claude Code and Codex:

```bash
npx skills add git@github.com:hosembafer/skills.git -g --skill '*' -a claude-code -a codex
```

One skill:

```bash
npx skills add git@github.com:hosembafer/skills.git -g --skill ship
```

Leave out `-a` to choose agents interactively, and `-g` to install into the current project instead of your user directory.

## Update

```bash
npx skills update -g
```

Updates replace the installed copies, so edit skills in a clone of this repo and push, not in `~/.agents/skills`.

## Layout

Each skill is a folder under `skills/` whose name matches the skill's `name` field:

```
skills/
  ship/
    SKILL.md
    agents/openai.yaml   # Codex: manual invocation only
  commit-subject/SKILL.md
  smart-copy/SKILL.md
  runtime-qa/
    SKILL.md
    scripts/port.sh      # picks the dev-server port for the checkout
    scripts/chrome.sh    # opens the slot's isolated Chrome profile
    references/          # test plan format, Claude in Chrome gotchas
  whoami/SKILL.md
```

## License

MIT. See [LICENSE](LICENSE).
