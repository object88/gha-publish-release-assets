#!/usr/bin/env bash
set -e

cd $(dirname "$0")

# Use this to ensure that we have all the tools required to do a build.
export CGO_ENABLED=0
export GO111MODULE=on
export GOFLAGS="-mod=vendor"

# default to mostly true, set env val to override
DO_TEST=${DO_TEST:-"true"}
DO_VERIFY=${DO_VERIFY:-"true"}
DO_VET=${DO_VET:-"true"}

while [[ $# -gt 0 ]]; do
  key="$1"
  case $key in
    --fast)
        DO_VET="false"
        DO_VERIFY="false"
        DO_TEST="false"
        shift
        ;;
    --no-test)
        DO_TEST="false"
        shift
        ;;
    --no-verify)
        DO_VERIFY="false"
        shift
        ;;
    --no-vet)
        DO_VET="false"
        shift
        ;;
    *)
      shift
      ;;
  esac
done

if [[ $DO_TEST == "true" ]]; then
  if ! [ -x ./bin/mockgen ]; then
    echo "Building mockgen"
    go build -o ./bin/mockgen ./vendor/github.com/golang/mock/mockgen
    echo ""
  fi
  echo "mockgen tool checked"

  echo "Generating mock for RoundTripper"
  ./bin/mockgen -destination=./mocks/mock_httproundtripper.go -package=mocks net/http RoundTripper
  echo ""
fi

if [[ $DO_VERIFY == "true" ]]; then
  echo "Verifying modules"
  go mod verify
  echo ""
fi

if [[ $DO_VET == "true" ]]; then
  echo "Running vet"
  go vet $(go list ./...)
  echo ""
fi

# build executable(s)
if [ $(uname) == "Darwin" ]; then
  # Cannot do a static compilation on Darwin.
  go build -o ./bin/publish -ldflags "-s -w" ./main/main.go
else
  go build -o ./bin/publish -tags "netgo" -ldflags "-extldflags \"-static\" -s -w" ./main/main.go
fi

# test executables and binaries
if [[ $DO_TEST == "true" ]]; then
  go test ./... -count=1
fi
