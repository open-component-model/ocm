$ErrorActionPreference = 'Stop'
$toolsDir   = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"
$url64 = "run: update.ps1"

$packageArgs = @{
  packageName   = $env:ChocolateyPackageName
  unzipLocation = $toolsDir
  url64bit      = $url64
  softwareName  = 'ocm-cli*'

  checksum64 = 'run: update.ps1'
  checksumType64= 'sha256'
}

Install-ChocolateyZipPackage @packageArgs
