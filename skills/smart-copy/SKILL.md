---
name: smart-copy
description: Use when the user asks to copy or move something to their copy buffer or clipboard ("copy to my copy buffer", "copy the continue prompt", "move this to my clipboard", /smart-copy), usually the continue prompt after a handoff, a drafted message or MR description, or a command from the previous response that they will paste somewhere themselves. Not for copying a whole response verbatim; the built-in /copy does that.
argument-hint: "[what to copy, e.g. 'the second command']"
---

# Smart Copy

Put the payload on the macOS clipboard, character for character. The payload is the part of the conversation that the user will paste somewhere else.

## Pick the payload

Read your most recent substantive response. Skip short replies such as "Copied …". If the user passed arguments, they name the payload. Otherwise take the first rule that matches:

1. **A file that holds the deliverable.** The response names a file containing text for the user to paste or send, such as a draft, a description, a handoff or a prompt. The payload is that file.
2. **Text the user is meant to paste, send or run.** Examples: a continue or resume prompt, a drafted message or description, a command, a commit message. It usually sits in a fenced block or a blockquote, or follows a label like "Continue prompt:".
3. **Nothing like that.** The payload is the one value the answer is about: a URL, path, branch name, ID or command.

If the same user message asks you to produce something and copy it ("give me the continue prompt and copy it"), write it in your reply first, then copy what you wrote.

## Shape the clipboard text

- The clipboard holds the payload's characters and nothing else. Leave out code fences, `>` quote markers, the label line, and any surrounding prose.
- Everything inside the payload stays exactly as written, including a leading `!`, quotes, backticks, `$` and line breaks.
- Several items for the same action, such as two commands, are copied as separate blocks. Keep each one verbatim, in its original order, with one blank line between them.

## Copy and verify

For a file:

```bash
pbcopy < "/path/to/file" && cmp -s <(pbpaste) "/path/to/file" && echo identical
```

For text, use a quoted heredoc so that `$`, backticks and quotes stay literal. `perl` removes the heredoc's trailing newline, so a pasted command doesn't run on its own:

```bash
perl -pe 'chomp if eof' <<'SMART_COPY_EOF' | pbcopy
<payload, verbatim>
SMART_COPY_EOF
pbpaste | head -c 300; echo; pbpaste | wc -l
```

If the payload itself contains a line `SMART_COPY_EOF`, pick another delimiter.

## Reply

Reply with one line saying what was copied and how long it is, for example: `Copied the continue prompt (1 line).` For several items, name them: `Copied both test commands (2 blocks).`

## Common mistakes

| Mistake | Fix |
| --- | --- |
| Two commands merged into one line with `;`, or a `!` dropped | Copy each command verbatim as its own block |
| `printf '%s' '…'` breaks on an apostrophe (don't, it's) | Use the quoted heredoc |
| A file's contents retyped from memory | `pbcopy < file`, then `cmp` |
| The label or fences (`Continue prompt:`, ```` ``` ````) copied too | Copy only what is inside them |
