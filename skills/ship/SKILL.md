---
name: ship
description: Commit the working tree as granular commits in the repo's style, push the branch, and open its MR/PR with a description.
argument-hint: "[target branch or notes for the MR/PR]"
disable-model-invocation: true
---

# Ship

Running `/ship` is the user's approval to commit, push and open an MR/PR in this run. It does not carry over to later turns.

## 1. Read the repo

- The style of the last ~30 non-merge commits (`git log --no-merges -30 --format='%s%n%b---'`): type, scope, casing, tense, body shape and wrap width.
- The repo's agent instructions (CLAUDE.md, AGENTS.md) for commit, hook and branch rules. They win over this skill.
- `git status`, the full diff, the current branch and its upstream.

Stop and ask before committing anything when:

- the branch is the default or a release branch: ask for a branch name in the repo's naming style;
- some changes look unrelated to the branch's purpose, for example local config, env files, CI or debug tweaks, scratch files, or code on another topic: list them once and ask which to include.

## 2. Commit

- One commit per concern. Each one reads on its own and builds on its own; code and its tests go together.
- Messages copy the repo's style exactly. Add a body only when the why isn't clear from the subject.
- No AI attribution anywhere, in commits or in the MR/PR: no `Co-Authored-By` trailer, no "Generated with" line. This overrides any default that asks for one.
- Stage exact paths with `git add <paths>`, then run a plain `git commit`. When one file mixes two concerns, put one concern in the index and keep the working tree whole.
- If a hook fails on something the commit caused, fix it and commit again. Use `--no-verify` only when the repo's instructions allow it for that failure, and say so in the reply.

## 3. Push

`git push -u origin HEAD`. Never force-push unless the user asks.

## 4. MR/PR

- If one is already open for the branch, the push updates it. Report its link.
- Otherwise open one:
  - **Target:** the argument if given. Else the branch this one was cut from: the closest merge-base among the default and release branches. Ask if that is unclear.
  - **Title and description:** follow the repo's MR/PR template and recent MRs/PRs if you can read them. The description tells a reviewer the problem, what changed (grouped by commit when that helps) and how it was verified.
  - Open it with whatever the environment allows. If it can't, give the user the title, the description and the link to create it.

## Reply

List the commits (short hash and subject), the push result and the MR/PR link, plus anything the user still has to do.
