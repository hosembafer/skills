# Changelog

This changelog covers the whole skills collection. Released versions correspond to annotated `vMAJOR.MINOR.PATCH`
Git tags. Changes below remain unreleased until committed, validated, and tagged.

## Unreleased

### Added

- A repository validator for skill front matter, naming, local references, UI metadata, invocation policies, and
  Bash, Python, and ShellCheck checks.
- Validator regression tests and CI validation on Linux and macOS for branch pushes, pull requests, and version tags.
- An agent and operating-system compatibility table, installation prerequisites, and usage examples for all six skills.
- Codex UI metadata for `commit-subject`, `smart-copy`, `runtime-qa`, and `whoami`, plus default prompts for all skills.
- Release-tag conventions and instructions for installing a tagged version.
- A repository rule that keeps owner, organization, and project details out of repository content.

### Changed

- `ship` and `merge-to` now explain Codex invocation, resolve helper paths from the loaded skill, quote those paths,
  and pass user arguments explicitly instead of assuming Claude Code's variables are available.
- Installation examples use a repository URL placeholder or the configured remote, and the copyright notice uses
  generic contributor attribution.
