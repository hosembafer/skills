# Releases

Version the whole collection with annotated Git tags named `vMAJOR.MINOR.PATCH`. Use a patch version for corrections
that preserve workflows, a minor version for new skills or compatible capabilities, and a major version when existing
invocations, requirements, or authorization behavior change incompatibly. During `0.x`, record breaking changes
explicitly and advance the minor version.

There is no tagged release yet. The first release can be `v0.1.0`; the commands below are a maintainer procedure to run
when the release is ready, not a record of an existing release.

## Prepare a release

1. Move the completed entries from `Unreleased` in [CHANGELOG.md](../CHANGELOG.md) into a version heading with its
   release date, for example `## 0.1.0 - 2026-10-03`. Keep an empty `Unreleased` section for subsequent work.
2. Run the [local validation commands](../README.md#validate). Review and commit the release notes and skill changes,
   then merge them into `main` through the repository's normal PR process.
3. Wait for both CI validation jobs to pass on the committed `main` revision. Confirm the working tree is clean and
   that `HEAD` is the revision being released.
4. Create an annotated tag for that revision. Substitute the chosen version in these commands:

   ```bash
   git tag -a v0.1.0 -m "Release skills v0.1.0"
   git show --no-patch v0.1.0
   ```

5. Push that tag to this private repository when ready to make the version available to repository collaborators:

   ```bash
   git push origin v0.1.0
   ```

The validation workflow also runs on `v*` tag pushes. Tags identify committed releases; they do not include
uncommitted changes. Keep existing release tags fixed and create a new version for corrections.

## Install a tagged version

Once the tag exists, install its contents using the `skills` CLI's `#ref` Git-source syntax. From a clone of this
skills repository, read the configured remote URL and append the tag. The quotes keep the source literal in your shell:

```bash
skills_repository=$(git remote get-url origin)
npx skills add "${skills_repository}#v0.1.0" -g --skill '*' -a claude-code -a codex
```

The unversioned command in the README follows the repository's default branch. To stay on a particular release,
reinstall from that explicit tag when choosing a version; do not use the general update command to select a release.
