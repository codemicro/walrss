#!/usr/bin/env bash

cat << EOF > walrss/internal/core/version.go
package core

const Version = "$1"
EOF