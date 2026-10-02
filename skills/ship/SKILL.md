---
name: ship
description: Commit the working tree as granular commits in the repo's style, push the branch, and open its MR/PR with a description.
argument-hint: "[target branch or notes for the MR/PR]"
disable-model-invocation: true
allowed-tools: Bash(bash ${CLAUDE_SKILL_DIR}/scripts/ship.sh *)
---

# Ship

Running `/ship` is the user's approval to commit, push and open an MR/PR in this run. It does not carry over to later turns.

Run the script calls exactly as written; `${CLAUDE_SKILL_DIR}` is the folder holding this SKILL.md.

## 1. Read the repo

Run `bash ${CLAUDE_SKILL_DIR}/scripts/ship.sh context`, adding the argument if it names a branch. It fetches `origin` and prints what the steps below need.

- Exit 3, on the default or a release branch: ask for a branch name in the repo's naming style, `git switch -c <name>`, and run it again.
- Exit 4, behind its remote: stop and ask. Never force-push.

The repo's agent instructions (CLAUDE.md, AGENTS.md) for commit, hook and branch rules win over this skill. Read the diff. If some changes look unrelated to the branch's purpose, for example local config, env files, CI or debug tweaks, scratch files, or code on another topic, list them once and ask which to include.

## 2. Commit

- One commit per concern. Each one reads on its own and builds on its own; code and its tests go together.
- Messages copy the style of the last 30 commits exactly: type, scope, casing, tense, body shape and wrap width. Add a body only when the why isn't clear from the subject.
- No AI attribution anywhere, in commits or in the MR/PR: no `Co-Authored-By` trailer, no "Generated with" line. This overrides any default that asks for one.
- Stage exact paths with `git add <paths>`, then run a plain `git commit`, never `git commit -- <paths>`. When one file mixes two concerns, put one concern in the index and keep the working tree whole.
- If a hook fails on something the commit caused, fix it and commit again. Use `--no-verify` only when the repo's instructions allow it for that failure, and say so in the reply.

## 3. Push

Run `bash ${CLAUDE_SKILL_DIR}/scripts/ship.sh push`. Exit 4: stop and ask, as in step 1.

## 4. MR/PR

If step 1 or 3 printed an `mr:` link, the push updated it. Otherwise:

- **Target:** `target:` from step 1. Ask when it says `unclear` or `not on origin`.
- **Title and description:** follow the template from step 1 and recent MRs/PRs if you can read them. The description tells a reviewer the problem, what changed (grouped by commit when that helps) and how it was verified.
- Open it:

  ```bash
  bash ${CLAUDE_SKILL_DIR}/scripts/ship.sh open <target> <<'EOF'
  <title>

  <description>
  EOF
  ```

  Exit 5 means nothing could open it. Give the user the title, the description and a link to create it, taken from `create_url:` or built from `web:`.

## Reply

List the commits (short hash and subject), the push result and the MR/PR link, plus anything the user still has to do.
