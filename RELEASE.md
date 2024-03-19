# Releasing the Conformance Suite

The conformance repository has a release workflow which publishes the test runner
binary, `connectconformance`, as an artifact that is part of the GitHub release.
The process was designed to be as friction-free as possible — the result is that
you just create/tag a release in GitHub, and the creation and upload of artifacts
is automated after that.

Using the Github UI, [create a new release](https://github.com/connectrpc/conformance/releases/new)
like so:
* Under “Choose a tag”, type in “vX.Y.Z” to create a new tag for the release upon publish.
  Note: The release job does infer the version from the release string and expects versions
  to start with `v` (e.g. `v0.1.0`).
* Target the main branch.
* Title the Release the same as the tag: “vX.Y.Z”.
* Click “Set as latest release”.
* If this is a release candidate, or any other kind of pre-release, click "Set as a pre-release".
* Set the last version as the “Previous tag”.
* Click “Generate release notes” to autogenerate release notes.
* Edit the release notes.
   * Tweak the change description for each if necessary so it summarizes the salient
     aspect(s) of the change in a single sentence. Detail is not needed as a user could
     follow the link to the relevant PR. (Potentially take a pass at PR descriptions
     and revise to increase clarity for users that visit the PRs from the release notes.)
   * Related commits can be grouped together with a single entry that has links to all
     relevant PRs (and attributes all relevant contributors).
   * A summary may be added if warranted.
   * The items in the list should be broken up into sub-categories. The typical
     sub-categories to use follow:
      * **Bugfixes**: Self-explanatory -- fixes to defects. These can be bugs in the
        test runner or bugs in the reference implementations.
      * **Enhancements**: New features or additions/improvements to existing features.
        This can include new command-line flags to enable new functionality and also
        includes addition of test cases for enhanced test coverage.
      * **Other Changes**: Other noteworthy changes in the codebase or tests. Use your
        best judgement when deciding if something warrants appearing here. Things like
        dependency updates and the like do _not_ warrant appearing here and should be
        removed from the auto-generated release notes.
* Click "Save Draft" and then share the link to have the notes reviewed by at least one
  other [maintainer](https://github.com/connectrpc/conformance/blob/main/MAINTAINERS.md).
* After the notes are approved (after some potential iteration and revision), you can
  finally create the release by clicking "Publish Release".

After the GitHub release has been created, you can verify that the rest of the process
completes successfully by following the corresponding run of the "release" GitHub
workflow: https://github.com/connectrpc/conformance/actions/workflows/release.yml

Once the workflow is complete, the artifacts will be attached to the
[GitHub release](https://github.com/connectrpc/conformance/releases).
