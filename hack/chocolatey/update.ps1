# GitHub repository details
$owner = "open-component-model"
$repo = "ocm"

# Query contributors and update authors in the nuspec file
$url = "https://api.github.com/repos/$owner/$repo/contributors"
$response = Invoke-RestMethod -Uri $url -Headers @{ "User-Agent" = "PowerShell" }
$filteredContributors = $response | Where-Object { $_.login -notin @('dependabot[bot]', 'github-actions[bot]', 'gardener-robot', 'ocmbot[bot]') } | Select-Object -ExpandProperty login
$sortedContributors = $filteredContributors | Sort-Object
$authors = $sortedContributors -join ", "
$nuspecPath = Join-Path -Path $PSScriptRoot -ChildPath "ocm-cli.nuspec"
$nuspecContent = Get-Content -Path $nuspecPath -Raw
$updatedContent = $nuspecContent -replace '<authors>.*?</authors>', "<authors>$authors</authors>"

# Fetch the latest release version and asset URLs for Windows
$url = "https://api.github.com/repos/$owner/$repo/releases/latest"
$response = Invoke-RestMethod -Uri $url -Headers @{ "User-Agent" = "PowerShell" }
$latestVersion = $response.tag_name -replace '^v', ''
Write-Output "The latest released ocm-cli version is $latestVersion"
$assets = $response.assets
$url = $assets | Where-Object { $_.name -match 'windows-386.zip$' } | Select-Object -ExpandProperty browser_download_url
$url64 = $assets | Where-Object { $_.name -match 'windows-amd64.zip$' } | Select-Object -ExpandProperty browser_download_url
$sha256url = $assets | Where-Object { $_.name -match 'windows-386.zip.sha256$' } | Select-Object -ExpandProperty browser_download_url
$sha256url64 = $assets | Where-Object { $_.name -match 'windows-amd64.zip.sha256$' } | Select-Object -ExpandProperty browser_download_url
$sha256 = [System.Text.Encoding]::UTF8.GetString((Invoke-WebRequest -Uri $sha256url).Content)
$sha256_64 = [System.Text.Encoding]::UTF8.GetString((Invoke-WebRequest -Uri $sha256url64).Content)

# Update the description and release notes in the nuspec file
$description = Get-Content -Path 'docs\reference\ocm.md' -Raw
# description is too long, chocolatey has a limit of 4000 characters
$startIndex = $description.IndexOf("### Description")
$endIndex = $description.IndexOf("Every option value has the format") - $startIndex
$description = $description.Substring($startIndex, $endIndex)
$description = $description -replace '### Description', '### ocm - Open Component Model Command Line Client'
# replace all xml problematic characters and html tags
$description = $description -replace '&mdash;', '-' # TODO replace unknown entity &mdash; with - in *.go
$description = $description -replace '&bsol;', '\'  # TODO replace unknown entity &bsol; with \ in *.go
$description = $description -replace '</?code>', '´' # TODO replace inline code in *.go with ``
$description = $description -replace '\s*<pre>', "```````n" # TODO replace code blocka in *.go with ```text
$description = $description -replace '</pre>', "`n``````" # TODO replace code blocka in *.go with ```text
$description = $description -replace '</?center>', '' # TODO remove center tags in *.go
$description = $description -replace '</?b>', '' # TODO replace bold tags in *.go with **
$description = $description -replace '<br\s*/?>', '' # TODO replace line breaks in *.go with \n
$description = $description -replace '\]\(ocm_', '](https://github.com/open-component-model/ocm/blob/main/docs/reference/ocm_'
$description = $description -replace '<', '&lt;' # used in code blocks and examples
$description = $description -replace '>', '&gt;' # used in code blocks and examples
$description += "`nContinue reading on [ocm.software / cli-reference](https://ocm.software/docs/cli-reference/)"
# release notes do hopefully not contain xml tags
$releaseNotes = Get-Content -Path "docs\releasenotes\v$latestVersion.md" -Raw
$releaseNotes = $releaseNotes -replace '\(#(\d+)\)', '([$1](https://github.com/open-component-model/ocm/pull/$1))'
$updatedContent = $updatedContent -replace '(?ms)<description>.*<\/description>', "<description>$description</description>"
$updatedContent = $updatedContent -replace '(?ms)<releaseNotes>.*<\/releaseNotes>', "<releaseNotes>$releaseNotes</releaseNotes>"
Set-Content -Path $nuspecPath -Value $updatedContent
Write-Output "Updated the <authors> tag in the nuspec file with the sorted list of contributors."

# Update the install script with the new URLs
$scriptPath = Join-Path -Path $PSScriptRoot -ChildPath "tools\chocolateyinstall.ps1"
$scriptContent = Get-Content -Path $scriptPath -Raw
$updatedContent = $scriptContent -replace '\$url\s*=\s*".*"', (-join('$url = "', $url, '"'))
$updatedContent = $updatedContent -replace '\$url64\s*=\s*".*"', (-join('$url64 = "', $url64, '"'))
$updatedContent = $updatedContent -replace "checksum\s*=\s*'.*'", "checksum = '$sha256'"
$updatedContent = $updatedContent -replace "checksum64\s*=\s*'.*'", "checksum64 = '$sha256_64'"
Set-Content -Path $scriptPath -Value $updatedContent
Write-Output "Using $url ($sha256)"
Write-Output "and $url64 ($sha256_64) as package sources."

# Copy the LICENSE file to the tools directory
$licenseDest = Join-Path -Path $PSScriptRoot -ChildPath "tools\LICENSE.txt"
Copy-Item -Path LICENSE -Destination $licenseDest -Force

# Update the choco package
choco pack --version $latestVersion $nuspecPath
