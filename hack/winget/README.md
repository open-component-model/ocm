# winget

## our first contribution

[New package|Open-Component-Model.ocm-cli version 0.15.0 #175451](https://github.com/microsoft/winget-pkgs/pull/175451)

## package and publish instructions

Input             | Value
------------------|---------------------------------------------------------
PackageIdentifier | Open-Component-Model.ocm-cli
PackageVersion    | 0.15.0
alias             | ocm-cli
Publisher         | SAP SE
PackageName       | ocm-cli
License           | Apache-2.0
ShortDescription  | Open Component Model Command Line Interface (ocm-cli)

```powershell
winget install wingetcreate
wingetcreate new https://github.com/open-component-model/ocm/releases/download/v0.15.0/ocm-0.15.0-windows-386.zip https://github.com/open-component-model/ocm/releases/download/v0.15.0/ocm-0.15.0-windows-amd64.zip https://github.com/open-component-model/ocm/releases/download/v0.15.0/ocm-0.15.0-windows-arm64.zip
```

Use the table above for your inputs.

Then verify the generated manifest:

```powershell
winget validate --manifest .\manifests\o\Open-Component-Model\ocm-cli\0.15.0\
winget install  --manifest .\manifests\o\Open-Component-Model\ocm-cli\0.15.0\
```

## update package

```powershell
wingetcreate update --urls https://github.com/open-component-model/ocm/releases/download/v0.15.0/ocm-0.15.0-windows-amd64.zip https://github.com/open-component-model/ocm/releases/download/v0.15.0/ocm-0.15.0-windows-arm64.zip --version 0.15.0 --release-notes-url https://github.com/open-component-model/ocm/releases/tag/v0.15.0 ` Open-Component-Model.ocm-cli
```

## github action

[push-to-winget](../../.github/workflows/publish-to-other-than-github.yaml#L124) requires a ["Personal access tokens (classic)"](https://github.com/organizations/open-component-model/settings/secrets/actions/OCM_CI_ROBOT_0_REPO). There is an open issue on [winget-create](https://github.com/microsoft/winget-create/issues/470). We [tried it](https://github.com/open-component-model/ocm/pull/1133) already with fine grained tokens and our [ocmbot](https://github.com/organizations/open-component-model/settings/apps/ocmbot). But that doesn't work: [ERROR: Resource not accessible by integration](https://github.com/open-component-model/ocm/actions/runs/12008922878/job/33472565698).

## winget packages repository

The [pull request](https://github.com/microsoft/winget-pkgs/pull/193703) creator of has to sign the [Contributor License Agreement](https://cla.opensource.microsoft.com/microsoft/winget-pkgs). In case someone else than [ocm-ci-robot-0](https://github.com/orgs/open-component-model/people/ocm-ci-robot-0) creates such an update PR, you'll need to reply with:

```text
@microsoft-github-policy-service agree company="SAP SE"
```

on your PR at least once.
