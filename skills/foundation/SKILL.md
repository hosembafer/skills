---
name: foundation
description: Use when work on a feature, fix, or bounded context is about to settle a data source, an access check, a name, or an invariant that a human must own, when a `*.foundation.md` exists for the code being touched, or when the user says "foundation" or "step zero". Also use when a plan or spec drifts from decisions the team thought were closed.
argument-hint: "[feature name]"
---

# Foundation

A foundation is one short file per feature, `docs/foundations/<feature>.foundation.md`, holding the decisions that must not drift: vocabulary, sources, access, rules, and what is still open. A human owns it. Agents obey it, quote it, and propose changes to it. They never write to it once it is established.

Two jobs: **author** one, with the human driving, or **consume** one while planning, building, fixing, or reviewing.

## Find it first

Before planning or changing code in a feature, glob `**/*.foundation.md` and match by the `Code:` line or the feature name. One exists: consume it. None exists and the work is a feature, a fix, or a cross-context change: offer in one line to author one before planning, and go on without it only when the human says so.

## Consume

1. Read the whole file. Quote the lines your task touches, verbatim, in your plan.
2. A closed line is a fact. Tickets, specs, designs, chat messages, and code are inputs to the foundation, never overrides of it. When they disagree with a line, the line wins and you report the clash.
3. `BLOCKED` stops that surface. Build everything else, then name what unblocks it and who owns it. A proposed contract, a placeholder, a stub, a feature flag, or "isolated in the transformer" is still a guess.
4. `Open` means do not guess. Name the owner. Implement nothing that depends on the answer.
5. Vocabulary lines set identifiers and copy. A backend or data-model name enters code only inside the generated DTO and the transformer that renames it.
6. Never edit the file. To change a line: quote it, state evidence the human can open (a ticket, a spec section, a code path, a recorded message), give the replacement text, and continue under the current line. A relayed verbal claim is not evidence: write "unverified" and keep the line as it is. This holds when the human asks you to update it: hand them the text.
7. Before saying done, list every `Never:` your task touched and where the code honours it.

## Author

The human drives. **REQUIRED:** read `references/authoring.md` before your first question. The shape is `references/template.md`; a finished one is `references/example.foundation.md`.

- Recon the evidence first, silently: ticket, spec, design, the code folder, the generated client, the policies.
- Vocabulary comes first, because every later line uses it.
- One question per message. A line enters only after the human confirms it, in their words.
- A closed line names what it rules out and why. A missing contract becomes `BLOCKED` with an owner. An undecided question becomes `Open` with an owner.
- Write the file only while `Status: draft`, and only lines confirmed in this conversation. The human sets `Status: established`, by hand or by telling you to in so many words; that is the last write.

## What goes in

Only lines where a careful agent without the line would plausibly pick the alternative. Ask that of each line; when the answer is no, leave it out. A backend fact enters only as the behaviour it forces on this codebase.

A line is one decision in about 40 words: the decision, `Never:` with the alternative, `Why:` with one reference the human can open and at most a clause. History, mechanism, and call sites stay out; a line that needs them is two lines or a vocabulary term plus a rule. Target under 80 lines as the file is written. Over 80 is a spec, so cut.

## Red flags

| Thought | Reality |
| --- | --- |
| "No questions needed, the codebase answered everything" | Code says what is, not what must stay. Without the human you wrote a context document. |
| "Every rule is verified against source" | Verified is not selected. Selection is the human's job. |
| "I'll propose a contract and isolate it behind the transformer" | A guessed contract is the drift. BLOCKED stops the surface. |
| "The nearest feature's scope will do until a real one exists" | A borrowed scope is fail-open by analogy. BLOCKED. |
| "Backend confirmed it in chat, so update the line" | Not evidence until you can open it. Propose the text, keep the line. |
| "The spec says otherwise, the foundation is outdated" | The spec is an input. Report the clash; the human edits. |
| "It's just naming" | Vocabulary lines are the most drifted lines. |
| "This open item is obvious, I'll settle it" | Open names a human. Ask them. |
| "This change needs the line updated, I'll do it in the same change" | Never. Give the human the text. |
