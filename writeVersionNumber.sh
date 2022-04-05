#!/bin/bash

cat << EOF > walrss/internal/state/version.go
package state

const Version = "$(git rev-parse HEAD | head -c 7)"
EOF