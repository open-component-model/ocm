# GitHub repository details
$owner = "open-component-model"
$repo = "ocm"

# Query contributors and update authors in the nuspec file
$url = "https://api.github.com/repos/$owner/$repo/contributors"
$response = Invoke-RestMethod -Uri $url -Headers @{ "User-Agent" = "PowerShell" }
$filteredContributors = $response | Where-Object { $_.login -notin @('dependabot[bot]', 'github-actions[bot]', 'gardener-robot', 'ocmbot[bot]') } | Select-Object -ExpandProperty login
$sortedContributors = $filteredContributors | Sort-Object
$authors = $sortedContributors -join ", "
$nuspecPath = "ocm-cli.nuspec"
$nuspecContent = Get-Content -Path $nuspecPath -Raw
$updatedContent = $nuspecContent -replace '<authors>.*?</authors>', "<authors>$authors</authors>"
Set-Content -Path $nuspecPath -Value $updatedContent
Write-Output "Updated the <authors> tag in the nuspec file with the sorted list of contributors."

# Fetch the latest release version and URLs for the Windows artifacts
$url = "https://api.github.com/repos/$owner/$repo/releases/latest"
$response = Invoke-RestMethod -Uri $url -Headers @{ "User-Agent" = "PowerShell" }
$latestVersion = $response.tag_name -replace '^v', ''
Write-Output "The latest released ocm-cli version is $latestVersion"
$assets = $response.assets
$url = $assets | Where-Object { $_.name -match 'windows-386.zip' } | Select-Object -ExpandProperty browser_download_url
$url64 = $assets | Where-Object { $_.name -match 'windows-amd64.zip' } | Select-Object -ExpandProperty browser_download_url
# # SHA256 - Download the artifacts
# $artifactPath = "windows-386.zip"
# $artifactPath64 = "windows-amd64.zip"
# Invoke-WebRequest -Uri $url -OutFile $artifactPath
# Invoke-WebRequest -Uri $url64 -OutFile $artifactPath64
# # SHA256 - Compute the checksums
# $checksum = Get-FileHash -Path $artifactPath -Algorithm SHA256 | Select-Object -ExpandProperty Hash
# $checksum64 = Get-FileHash -Path $artifactPath64 -Algorithm SHA256 | Select-Object -ExpandProperty Hash

# Update the install script with the new URLs
$scriptPath = "tools\chocolateyinstall.ps1"
$scriptContent = Get-Content -Path $scriptPath -Raw
$updatedContent = $scriptContent -replace '\$url\s*=\s*".*"', (-join('$url = "', $url, '"'))
$updatedContent = $updatedContent -replace '\$url64\s*=\s*".*"', (-join('$url64 = "', $url64, '"'))
# # SHA256
# $updatedContent = $updatedContent -replace "\$checksum\s*=\s*['\"].*?['\"]", "\$checksum = '$checksum'"
# $updatedContent = $updatedContent -replace "\$checksum64\s*=\s*['\"].*?['\"]", "\$checksum64 = '$checksum64'"
Set-Content -Path $scriptPath -Value $updatedContent
Write-Output "Updated the $url and $url64 variables in the script."

# Copy the LICENSE file to the tools directory
Copy-Item -Path ..\..\LICENSE -Destination tools\LICENSE.txt -Force

# Update the choco package
choco pack --version $latestVersion

# Push the updated package to the Chocolatey repository
#choco push ocm-cli.$latestVersion.nupkg --source "'https://push.chocolatey.org/'" --api-key="'$env:CHOCO_API_KEY'"
