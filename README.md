# skills

My agent skills for Claude Code, Codex and other agents that read [Agent Skills](https://agentskills.io). Install them with the [`skills`](https://github.com/vercel-labs/skills) CLI.

## Skills

| Skill            | What it does                                                                                                 | Started by                 |
| ---------------- | ------------------------------------------------------------------------------------------------------------ | -------------------------- |
| `ship`           | Commits the working tree as granular commits in the repo's style, pushes the branch, and opens its MR/PR.    | `/ship` only               |
| `commit-subject` | Writes a one-line commit subject for the current changes in the repo's style and copies it to the clipboard. | Agent or `/commit-subject` |
| `smart-copy`     | Copies what you'll paste next (a continue prompt, a drafted message, a command) to the clipboard, verbatim.  | Agent or `/smart-copy`     |
| `whoami`         | Reports which agent CLI is running the session and which account is logged into it.                          | Agent or `/whoami`         |

`ship` never starts on its own. Running it is your approval to commit and push in that run.

`commit-subject` and `smart-copy` use `pbcopy`, so they need macOS.

## Install

All skills, for Claude Code and Codex:

```bash
npx skills add hosembafer/skills -g --skill '*' -a claude-code -a codex
```

One skill:

```bash
npx skills add hosembafer/skills -g --skill ship
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
  whoami/SKILL.md
```

## License

MIT. See [LICENSE](LICENSE).
