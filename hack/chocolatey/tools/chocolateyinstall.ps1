$ErrorActionPreference = 'Stop'
$toolsDir   = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"
$url = "https://github.com/open-component-model/ocm/releases/download/v0.14.0/ocm-0.14.0-windows-386.zip"
$url64 = "https://github.com/open-component-model/ocm/releases/download/v0.14.0/ocm-0.14.0-windows-amd64.zip"

$packageArgs = @{
  packageName   = $env:ChocolateyPackageName
  unzipLocation = $toolsDir
  fileType      = 'exe'
  url           = $url
  url64bit      = $url64
  softwareName  = 'ocm-cli*'

  # checksum      = ''
  # checksumType  = 'sha256'
  # checksum64    = ''
  # checksumType64= 'sha256'
}

Install-ChocolateyZipPackage @packageArgs

