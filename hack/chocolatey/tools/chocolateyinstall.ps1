$ErrorActionPreference = 'Stop'
$toolsDir   = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"
$url = "run: update.ps1"
$url64 = "run: update.ps1"

$packageArgs = @{
  packageName   = $env:ChocolateyPackageName
  unzipLocation = $toolsDir
  fileType      = 'exe'
  url           = $url
  url64bit      = $url64
  softwareName  = 'ocm-cli*'

  checksum = 'run: update.ps1'
  checksumType  = 'sha256'
  checksum64 = 'run: update.ps1'
  checksumType64= 'sha256'
}

Install-ChocolateyZipPackage @packageArgs
