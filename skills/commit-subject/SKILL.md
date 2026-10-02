---
name: commit-subject
description: Generate a commit subject from current git changes that follows the repository's commit conventions. Prints the subject AND copies it to the macOS clipboard (pbcopy). Use when the user wants a one-line commit subject without committing.
---

# Generate a commit subject based on current changes

Analyze the current git changes and generate a concise commit subject that follows the repository's commit conventions.

## Instructions

1. First, get the 5 latest commit subjects using `git log -5 --pretty=format:"%s"` to understand the commit style patterns
2. Check for staged changes using `git diff --staged`
3. If no staged changes, check unstaged changes using `git diff`
4. If no changes at all, output `No Changes Made` and stop (still copy that string to the clipboard so the wrapper sees a deterministic value)
5. Analyze:
   - What was changed (files, nature of changes)
   - The commit message patterns from recent commits (prefixes, scopes, casing, style)
6. Generate a commit subject that:
   - Follows the repository's existing commit style/patterns
   - Uses conventional commit format if that's the pattern: `type(scope): description`
   - Types: feat, fix, refactor, docs, style, test, chore
   - Scope: optional, the module/feature affected
   - Description: imperative mood, max 50-72 chars
   - If the user passed extra arguments (e.g. `use chore`, `feat(monitoring)`, ticket id), treat them as a constraint or hint and apply them to the generated subject
7. Copy the generated subject to the macOS clipboard by piping it through `pbcopy`. Run: `printf '%s' '<subject>' | pbcopy` (escape any single quote in the subject as `'\''`).
8. Output the single line commit subject as plain text so the user can see it. No backticks, no prefix like "Based on the analysis:", no surrounding commentary — just the subject on its own line. The subject MUST be both printed AND copied to the clipboard.
