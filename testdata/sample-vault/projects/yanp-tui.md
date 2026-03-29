---
title: YANP TUI
date: 2026-03-28
tags:
  - project
  - pkm
  - tool
  - go
status: active
source: human
aliases:
  - YANP Tool
project: YANP
---

# YANP TUI

`yanp-tui` is a terminal-first PKM tool written in Go and built to conform to the [[YANP Vault Spec]].

## Purpose

- Dogfood the YANP format in a working tool.
- Provide a fast TUI for browsing, capturing, renaming, and publishing notes.
- Keep vault source plain text and publish to standard Markdown without mutating the source notes.

## Current Features

- Index `.md` notes across the full vault.
- Read optional YAML frontmatter and preserve unknown fields during rewrites.
- Resolve wikilinks by frontmatter title, then aliases, then filename stem.
- Merge frontmatter tags with inline tags like #pkm and #tooling.
- Rename notes using lowercase kebab-case filenames and rewrite inbound wikilinks.
- Publish notes to a separate output directory with relative Markdown links.
- Offer a Bubble Tea TUI for dashboard, health reporting, filtering, browse, preview, capture, create, rename, and publish flows.

## Project Layout

- `cmd/yanp` contains the entry point.
- `internal/vault` contains indexing, parsing, resolution, rename, and publish logic.
- `internal/ui` contains the Bubble Tea terminal interface.
- `testdata/sample-vault` is the dogfood vault for examples and tests.

## TUI Flows

- The app opens on a dashboard with selectable shortcuts for browse, inbox, today's daily note, and recent notes.
- When duplicate title, alias, or filename targets exist, the dashboard exposes a vault health report.
- `enter` opens the note browser.
- `/` filters notes by title, alias, tag, path, or body text.
- `?` opens in-app help.

## Commands

```powershell
go run ./cmd/yanp -vault .\testdata\sample-vault
go run ./cmd/yanp publish -vault .\testdata\sample-vault -out .\published
go run ./cmd/yanp capture -vault .\testdata\sample-vault -text "Follow up with [[Alice]]"
go run ./cmd/yanp rename -vault .\testdata\sample-vault -note projects/sprint-review.md -title "Sprint Demo"
```

## Open Work

- Add richer note creation flows for `daily/`, `weekly/`, and `monthly/`.
- Add saved queries and dashboard widgets from config.
- Expand vault health beyond duplicate conflict reporting.
