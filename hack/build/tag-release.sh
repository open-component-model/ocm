#!/bin/bash

export default_branch="main"

function ensure_release_notes() {
  local release_notes_file="docs/release_notes/$1.md"
  if [[ ! -f "${release_notes_file}" ]]; then
    >&2 echo "Must have release notes ${release_notes_file}"
    exit 6
  fi
  echo "$release_notes_file"
}

function create_branch_if_doesnt_exist() {
  wanted_branch="$1"
  if ! git checkout "${wanted_branch}" >/dev/null; then
      echo "Creating ${wanted_branch} from $(git branch --show-current)"
      git checkout -b "${wanted_branch}"
      git push origin "$(git branch --show-current)"
  fi
}

function tag_and_push_release() {
    local version="${1}"
    local msg="${2}"
    for tag in "${version}" "${version}"; do
      git tag --annotate --message "${msg}" "${tag}"
      git push origin "${tag}"
    done
}

# Check prerequisites
gh version
gh auth status

release_version=$(go run pkg/version/generate/release_generate.go print-version)
ensure_release_notes "${release_version}"

# Create release  branch
release_branch=$(release-"${release_version}")
create_branch_if_doesnt_exist "${release_branch}"
git checkout "${release_branch}"
git pull --ff-only origin "${release_branch}" || echo "${release_branch} not found in origin, pushing new branch upstream."

# Tag and push release
msg="Release ${release_version}"
tag_and_push_release "${release_version}" "${msg}"