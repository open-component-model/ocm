# chocolatey

## package and publish instructions

Change directory into the git-repo-root folder, then execute the following:

```powershell
rm -r .\ocm-cli.*.nupkg
.\hack\chocolatey\update.ps1
choco push ocm-cli.*.nupkg --source https://push.chocolatey.org/ --apikey ********
```

## in case we need to manually re-submit a version

[Open the github releases rest endpoint](https://api.github.com/repos/open-component-model/ocm/releases) and find the `"id"` for the release you need to re-submit.
Then replace `latest` in [api.github.com/repos/\$owner/\$repo/releases/latest](update.ps1#L16) with that **id** and execute the script: `.\hack\chocolatey\update.ps1`.

Or manually adjust [ocm-cli.nuspec](ocm-cli.nuspec) and [chocolateyinstall.ps1](tools/chocolateyinstall.ps1) and run:

```powershell
$version = "0.14.0"
rm -r .\ocm-cli.$version.nupkg
choco pack --version $version .\hack\chocolatey\ocm-cli.nuspec
choco push ocm-cli.$version.nupkg --source https://push.chocolatey.org/ --apikey ********
```
