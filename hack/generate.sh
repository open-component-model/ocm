#!/bin/bash

set -e

echo "> Generate"

GO111MODULE=on go generate -mod=mod $@
