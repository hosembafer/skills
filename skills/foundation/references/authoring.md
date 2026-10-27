# Authoring a foundation

The human drives. Your job is to surface candidate lines, contradictions, and missing contracts from the evidence, and to ask one question at a time until every line is one the human can defend in a review without rereading the spec.

## Before the first question

Read the ticket, the spec, the design, the code folder, the generated API client, and the access policies. Keep for yourself, not for the human:

- candidate vocabulary: the words the user will see, the words the code uses, the words the data model uses, and where they collide;
- candidate sources: which client, endpoint, table, or service feeds each surface, and which plausible alternative an agent would pick;
- candidate access: which scope, policy, or reach gates each surface and action;
- candidate rules: invariants an agent would plausibly break;
- contradictions between the inputs (spec against design, spec against itself, ticket against code);
- missing contracts: a field, endpoint, scope, or copy string the design needs and nothing provides.

## The first message

The first message to the human is the first question, with its candidate line, and nothing else: no status report, no list of what you read, no inventory of candidate material, no queue of lines you plan to ask about. The recon stays with you and shows up only as questions, one at a time. This holds when you are a subagent and your hand-back is the message: the hand-back is the question and its candidate line, under 120 words. Nobody who continues the work needs your inventory; they need the human's answer.

## Order of questions

1. **Vocabulary first.** Ask which word the user sees for each concept. Code follows the copy. When code and copy must differ on purpose, write both in one line marked "Deliberate exception", so nobody fixes the mismatch later. When an existing concept already owns a word, say so and keep the new concept off it.
2. **Sources.** For each surface: what feeds it, and what it must never read from. Name the client and endpoint, or the service, as an agent would grep for it.
3. **Access.** For page entry and each action: the scope or policy, and what must never be borrowed.
4. **Rules.** Invariants in one sentence each, with the concrete violation after `Never:` and the evidence after `Why:`.
5. **Open.** Every contradiction and every undecided question, each with the person who answers it.

## Each question

- One question per message, a single sentence, no list of options hanging off it. A guess the human can correct beats an open question.
- Show the candidate line in the file's shape and ask whether they can defend it. When they cannot, rewrite it in their words or drop it.
- Keep each candidate line to about 40 words. `Why:` is one reference the human can open (a ticket, a spec section, a merge request, a code path) plus at most a clause. Leave out the history of how the decision was reached, the backend mechanism, and the list of call sites; the reference carries those.
- When the human does not know a word in a candidate line, define it in plain words from the evidence, then ask whether it forces anything on this codebase. When it forces nothing, drop the line. When it forces something, write only that.
- When the inputs contradict each other, do not resolve it. Put it under Open with the owner.
- When a contract is missing, mark the line `BLOCKED` with the owner. Never fill it from what the ticket wishes for.
- A relayed claim ("backend said in chat") is not evidence. Ask for something the human can open, or keep the line `BLOCKED`.

## Playback

After each section, show the section as it stands and ask what is wrong or missing. Mark your own guesses as guesses. Only what the human confirmed is written as theirs.

## Size

Count the file's lines before you stop. Over 80: ask the human which lines an agent would never drift on, and remove them. The file holds what would drift, not what is true.

## Where and how

- Default home: `docs/foundations/<feature>.foundation.md`. When the repository already keeps `*.foundation.md` files elsewhere, use that folder.
- The file is named after the initiative, as the team says it. When that name uses a word the vocabulary forbids for a concept, the vocabulary line says so in one clause ("the initiative's name only, never in the product"), so the title and the line do not contradict each other.
- Header: the ownership rule, owner, `Status: draft`, references, the `Code:` folders, and consumers.
- Write the file only while `Status: draft`, and only lines confirmed in this conversation. The human sets `Status: established`, by hand or by telling you to in so many words; that is the last write. After that, agents propose text in chat and never write to the file.
