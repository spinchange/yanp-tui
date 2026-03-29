---
title: YANP TUI Release Notes
date: 2026-03-28
tags:
  - project
  - release
  - docs
status: active
source: human
project: YANP
---

# YANP TUI Release Notes

## v0.1.0-alpha.2

Parser-fix follow-up release for [[YANP TUI]].

### Included

- Vault indexing
- Wikilink resolution and publish transform
- Rename backlink rewriting
- Conflict resolution tests and engine support
- Windows-safe slug generation for note creation and rename
- CRLF frontmatter normalization for Windows-authored notes
- Inline tag parsing that correctly skips escaped-backtick and multi-backtick code spans
- Dashboard, browser, filtering, help, capture, create, rename, and publish flows
- Windows release packaging via `build.ps1`

### Artifacts

- `dist\yanp.exe`
- `dist\yanp-v0.1.0-alpha.2-windows-amd64.zip`
