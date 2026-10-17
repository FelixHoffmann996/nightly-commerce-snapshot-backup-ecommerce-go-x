#!/bin/sh
set -eu

go run ./cmd/nightly-snapshot -once -input commerce.example.json -bucket commerce-nightly
