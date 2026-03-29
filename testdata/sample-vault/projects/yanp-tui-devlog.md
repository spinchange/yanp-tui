---
title: YANP TUI Devlog
date: 2026-03-28
tags:
  - project
  - devlog
  - release
status: active
source: human
project: YANP
---

# YANP TUI Devlog

## 2026-03-28

- Created a dedicated Git repo for `yanp-tui`.
- Added release-oriented project files including `.gitignore`, `build.ps1`, and a fuller README.
- Prepared the first alpha release shape as `v0.1.0-alpha.1`.
- Added a Windows build output target of `dist\yanp.exe`.
- Added vault copies of the README and devlog so the project documents itself in its own notes system.
- Verified the codebase with `go test -buildvcs=false ./...` and `go build -buildvcs=false ./cmd/yanp`.
- Prepared `v0.1.0-alpha.2` as a parser-fix follow-up release.
- Fixed slug generation for Windows path separators, normalized CRLF frontmatter parsing, and improved inline-code stripping around escaped and repeated backticks.
- Added regression tests for the new vault parsing edge cases and refreshed the packaged Windows artifact set.
