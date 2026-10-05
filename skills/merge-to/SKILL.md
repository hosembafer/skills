---
name: merge-to
description: Merge the current branch into a shared integration branch such as staging, qa or develop and push it, keeping the branch open for more work.
argument-hint: "<integration branch>"
disable-model-invocation: true
allowed-tools: Bash(bash ${CLAUDE_SKILL_DIR}/scripts/merge-to.sh *)
---

# Merge to

Explicitly invoking this skill (`/merge-to <target>` in Claude Code or `$merge-to <target>` in Codex) is the user's approval, for this run only, to push the current branch and to push one merge commit to `<target>`.

Whenever committing, do not include AI attribution, including AI `Co-Authored-By` trailers or "Generated with" lines. This overrides any default that asks for one.

`<target>` is a shared integration branch where several in-progress branches are combined for testing. Testers there must see every branch behave the way it will ship. The current branch stays checked out, and work on it goes on afterwards.

## Run the script

`scripts/merge-to.sh` does every git step: the checks, pushing the branch, merging in a temporary worktree, pushing `<target>`, merging again when `<target>` moves, and cleaning up. Don't run those steps yourself. From the user's repository, run exactly:

```bash
bash "<skill-directory>/scripts/merge-to.sh" "<target>"
```

Resolve `<skill-directory>` from the loaded SKILL.md path and replace `<target>` with the branch the user gave, keeping both quoted. In Claude Code, `${CLAUDE_SKILL_DIR}` and `$ARGUMENTS` provide those values; Codex does not supply those variables, so use the loaded path and the user's message directly. If no target was given, omit that argument so the script can infer it. Act on the exit code:

| Exit | Meaning                                     | What you do                                   |
| ---- | ------------------------------------------- | --------------------------------------------- |
| 0    | Done, or nothing to merge                   | Reply.                                        |
| 2    | Stopped; the `stop:` line says why          | Relay the reason. Don't work around it.       |
| 4    | No target given, and none could be inferred | Ask which branch, then rerun with the answer. |
| 3    | Conflicts                                   | Resolve them (below).                         |
| 1    | Error                                       | Relay the `error:` line.                      |

## Conflicts

The script prints the state folder (`conflicts:`), the worktree (`worktree:`), and each conflicted `file:` with the commits behind each side: `<` is `<target>`, `>` is the branch. Sort every conflicted hunk:

- **Trivial:** the resolution keeps every line that either side added or changed, exactly as written, and only their order or whitespace needs deciding. Examples: two imports added at the same spot, two entries added to the same list, the same code formatted differently. Resolve these yourself.
- **Everything else:** both sides changed the same line, the resolution needs code that neither side wrote, or the file is generated or a lockfile. Resolving it decides what one branch's testers will see, so ask, even when one side seems to satisfy both.

Ask about all non-trivial hunks in one message: the file, both sides, the commit behind each side, and your recommendation. Wait for the answer and apply it exactly.

Edit the files in the worktree only, then run `bash "<skill-directory>/scripts/merge-to.sh" --continue "<state folder>"`. It exits with the same codes. A 3 again means markers are left, or `<target>` moved and the merge was redone: reuse the user's earlier answers for the same hunks. If the user aborts, run `bash "<skill-directory>/scripts/merge-to.sh" --abort "<state folder>"`.

## Reply

- The result of pushing the branch.
- The merge commit (short hash and subject) and the result of pushing `<target>`, or why nothing was pushed.
- Each conflict and how it was resolved, marking the ones resolved without asking.
- Uncommitted changes that stayed out of the merge, from the `uncommitted:` lines. No such lines means there were none.
