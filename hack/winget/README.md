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
wingetcreate update --urls https://github.com/open-component-model/ocm/releases/download/v0.15.0/ocm-0.15.0-windows-386.zip https://github.com/open-component-model/ocm/releases/download/v0.15.0/ocm-0.15.0-windows-amd64.zip https://github.com/open-component-model/ocm/releases/download/v0.15.0/ocm-0.15.0-windows-arm64.zip --version 0.15.0 --release-notes-url https://github.com/open-component-model/ocm/blob/main/docs/releasenotes/v0.15.0.md ` Open-Component-Model.ocm-cli
```
