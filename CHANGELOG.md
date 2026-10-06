# Changelog

This changelog covers the whole skills collection. Released versions correspond to annotated `vMAJOR.MINOR.PATCH`
Git tags. Changes below remain unreleased until committed, validated, and tagged.

## Unreleased

### Added

- A Go repository validator for skill front matter, naming, local references, UI metadata, invocation policies,
  Bash and Go syntax, and ShellCheck checks.
- Validator regression tests and CI validation on Linux and macOS for branch pushes, pull requests, and version tags.
- An agent and operating-system compatibility table, installation prerequisites, and usage examples for all six skills.
- Codex UI metadata for `commit-subject`, `smart-copy`, `runtime-qa`, and `whoami`, plus default prompts for all skills.
- Release-tag conventions and instructions for installing a tagged version.
- A repository rule that permits owner attribution while keeping organization, project, and machine details private.

### Changed

- Contributor validation uses Go and a single YAML library. ShellCheck is installed directly, with a pinned version
  and verified download checksums in CI.
- The local `whoami` account helper uses the Go standard library and runs from an installed skill bundle. Tests cover
  its initialization and account lookup protocol, safe output, timeouts, and server process cleanup.
- `ship` and `merge-to` now explain Codex invocation, resolve helper paths from the loaded skill, quote those paths,
  and pass user arguments explicitly instead of assuming Claude Code's variables are available.
- Installation examples use the repository's GitHub URL, and the license credits the owner by name and email.
