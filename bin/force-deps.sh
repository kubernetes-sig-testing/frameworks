#!/bin/bash

set -eu

[ -z "${DEBUG:-}" ] || {
  set -x
}


##------------------------------------------------------------
BASE_DIR="$( cd "$(dirname "$0")/.." && pwd )"
PATH="${PATH}:${GOPATH}/bin"

##------------------------------------------------------------
forceVersion() {
  local package="$1"
  local version="$2"

  go get "${package}/..."
  (
    cd "${GOPATH}/src/${package}"
    git checkout "$version" >/dev/null 2>&1
    godep restore ./...
  )
}

##------------------------------------------------------------
deleteFromVendor() {
  for d in "$@"
  do
    find "${BASE_DIR}/vendor" \
      -wholename "*/${d}" \
      -type d \
      -exec rm -rf '{}' \+
  done
}

##------------------------------------------------------------
printVersion() {
  local getVerScript='
echo "found in $1:"
cd "$1"
git log -1 --no-color --format="%H, %aI, %d"
'
  for p in "$@"
  do
    find "${GOPATH}/src" \
      -wholename "*/${p}" \
      -type d \
      -exec sh -c "$getVerScript" -- '{}' \;
  done
}

##------------------------------------------------------------
main() {
  [ -z "${CLIENT_GO_VERSION:-}" ] || {
    {
      echo "Setting client-go version to ${CLIENT_GO_VERSION}"
      echo
    }>&2

    # deleteFromVendor 'k8s.io/client-go' 'k8s.io/api' 'k8s.io/apimachinery'
    deleteFromVendor 'k8s.io'

    forceVersion 'k8s.io/client-go' "$CLIENT_GO_VERSION"
  }

  printVersion 'k8s.io/client-go' 'k8s.io/api' 'k8s.io/apimachinery'
}

##------------------------------------------------------------
##------------------------------------------------------------
main "$@"
