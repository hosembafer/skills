---
name: merge-to
description: Merge the current branch into a shared integration branch such as staging, qa or develop and push it, keeping the branch open for more work.
argument-hint: "<integration branch>"
disable-model-invocation: true
---

# Merge to

Running `/merge-to <target>` is the user's approval, for this run only, to push the current branch and to push one merge commit to `<target>`.

`<target>` is a shared integration branch where several in-progress branches are combined for testing. Testers there must see every branch behave the way it will ship. The current branch stays checked out, and work on it goes on afterwards.

## 1. Check

- `git fetch origin` first.
- **Target:** the argument. Without one, read the subjects of earlier merges of this branch on the remote (`git log --remotes --merges --format=%s`). If they all went into one branch, use it and say so. Otherwise ask.
- Stop and say why when HEAD is detached, the current branch is the target, the target is the default branch or a release branch (those go through an MR/PR), or `origin/<target>` doesn't exist.

## 2. Push the branch

The merge takes the pushed branch, so push it first:

- No upstream: `git push -u origin HEAD`. Ahead of its upstream: `git push`.
- Behind or diverged from its upstream: stop and ask. Never force-push.
- Uncommitted changes are not part of the merge. Leave them as they are (no commit, no stash) and mention them in the reply.

## 3. Merge in a temporary worktree

Never switch the user's checkout to the target, and never merge in another worktree that has the target checked out: both hold someone's work. Use a temporary detached worktree. Pass `-c core.hooksPath=/dev/null` to every git command that touches it, because its hooks would install dependencies or run checks in a tree that has none.

```bash
tmp="$(mktemp -d)"
git -c core.hooksPath=/dev/null worktree add --detach "$tmp" "origin/<target>"
git -C "$tmp" -c core.hooksPath=/dev/null merge --no-ff -m "<subject>" "origin/<branch>"
```

`<subject>` copies the style of the recent merges on the target (`git log origin/<target> --merges -5 --format=%s`), for example `Merge branch '<branch>' into <target>`. If git says "Already up to date", nothing is pushed; go to step 6.

## 4. Conflicts

Sort every conflicted hunk:

- **Trivial:** the resolution keeps every line that either side added or changed, exactly as written, and only their order or whitespace needs deciding. Examples: two imports added at the same spot, two entries added to the same list, the same code formatted differently. Resolve these yourself.
- **Everything else:** both sides changed the same line, the resolution needs code that neither side wrote, or the file is generated or a lockfile. Resolving it decides what one branch's testers will see, so ask, even when one side seems to satisfy both.

Ask about all non-trivial hunks in one message: the file, both sides, the commit behind each side, and your recommendation. Wait for the answer and apply it exactly. If the user aborts, run `git -C "$tmp" merge --abort` and go to step 6.

Once everything is resolved, check that no conflict markers are left (`git -C "$tmp" diff --cached --check`), then commit with the same subject: `git -C "$tmp" -c core.hooksPath=/dev/null commit -m "<subject>"`.

## 5. Push

`git -C "$tmp" -c core.hooksPath=/dev/null push origin HEAD:<target>`. Never force-push.

If the push is rejected because the target moved, fetch, run `git -C "$tmp" checkout --detach origin/<target>`, and merge again under the same rules. Reuse the user's earlier answers for the same hunks. After two rejections, stop and report.

## 6. Clean up

Always, including after an abort or an error: `git worktree remove --force "$tmp"`.

## Reply

- The result of pushing the branch.
- The merge commit (short hash and subject) and the result of pushing `<target>`, or why nothing was pushed.
- Each conflict and how it was resolved, marking the ones resolved without asking.
- Uncommitted changes that stayed out of the merge.
