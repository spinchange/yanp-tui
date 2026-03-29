---
title: YANP TUI Release Checklist
date: 2026-03-28
tags:
  - project
  - release
  - checklist
status: active
source: human
project: YANP
---

# YANP TUI Release Checklist

- Keep the repo clean before tagging a release.
- Run `go test -buildvcs=false ./...`.
- Run `go build -buildvcs=false ./cmd/yanp`.
- Run `.\build.ps1`.
- Smoke-test `dist\yanp.exe`.
- Push the repo and tag `v0.1.0-alpha.1`.
- Publish the zip artifact and release notes.
