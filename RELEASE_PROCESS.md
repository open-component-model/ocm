# Release Process

## General Information

The content of a branch can be released by GitHub `release` action. The name of the release is based on the content of
the file [`VERSION`](./VERSION). During development, the content of this file is the complete release name of the
release currently under development and the suffix `-dev` (e.g. `0.1.0-dev`). The content of this file is used for
generating the version information compiled into the ocm executables.

If a release is created, the `-dev`-suffix is removed and an optional pre-release suffix is appended to generate the
name of the release (e.g. `0.1.0` or `0.1.0-alpha1`) and to prepare a commit for the release which is used to create a
release tag.

Additionally, this commit will also add a new release note file under [./docs/releasenotes](./docs/releasenotes). It is
generated from the appropriate draft release with the basic release name (e.g. `0.1.0`).
After the release is done, for final releases, a new commit is created to prepare the development of the next release
by adapting the [`VERSION`](./VERSION) file again. Thereby, if the patch level is 0 (e.g. `0.1.0`), the minor
version number is increased (e.g. `0.2.0`). If the patch level is not zero (e.g. `0.1.1`), the patch level is increased
(e.g. `0.1.2`). This commit is pushed to the branch for which you created the release, and it therefore contains the
release commit in its history.

When creating a minor release, it is possible to optionally created a patch branch. In this case, a new branch with the
name `releases/<release-name>` is created. This branch is prepared with a commit which adjusts the patch level in the
[`VERSION`](./VERSION) file to 1 (e.g. for release `0.1.0`, the patch branch is prepared with `0.1.1`).


## Creating a Release

A release is created for a branch - typically, the main branch or a patch branch - by executing the GitHub action
`release`. Therefore, you have to specify the branch to release, and you can optionally indicate to create a patch
branch or to create a pre-release by specifying a pre-release name.

## Preparing a Patch Release

There are 2 possibilities to create and release patches.
1) If during the creation of a minor release the option to create patch branch has been selected, there is already a
patch branch `releases/<minor-release>` which can be used to prepare commits to be released.
2) If no patch branch has been created in advance for any existing minor release, a patch branch can be created using
the GitHub action `release-branch`. Therefore, you have to select the tag of the intended release. As a result, the
patch branch `releases/<minor-release>` is prepared with the appropriate version file
(containg `x.<minor-release>.1-dev`).

   > **NOTE**:
   > If this is not possible because the release is older than the latest version of the release action, then you have to
   > manually specify the intended tag in the input field of the action.

On the patch branch (like on the main branch), new commits can be added using pull requests. Once a patch should be
released, the release action is executed on the patch branch (theoretically, for patches, pre-releases are also
possible).

## What is part of a release?

During the build of a release, a OCM CTF (Common Transport Format Archive) is created (through `make ctf`), containing
all the provided component versions described by the actual git snapshot. This archive is then published to
ghcr.io/open-component-model/ocm. Additionally, a GitHub release is created, exposing the OCM CTF and the ocm-cli
executables for various platforms. These executables are build using the go releaser plugin. Furthermore, packages for
debian and brew are created and uploaded to respective package repositories.
