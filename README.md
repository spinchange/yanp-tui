# yanp-tui

`yanp-tui` is a Go personal knowledge management tool with a terminal UI, built to work with the draft [YANP vault specification](https://spinchange.github.io/yanp/).

YANP-TUI stands for `Yet Another Notes Project - Terminal User Interface`.

This project is currently shaped as an early preview release: `v0.1.0-alpha.1`.

## Current Status

- Core YANP vault indexing works.
- Wikilinks resolve by title, alias, then filename stem.
- Rename updates inbound wikilinks across the vault.
- Publish writes CommonMark output to a separate directory.
- The TUI includes a dashboard, browser, preview, filtering, help, capture, create, rename, and publish flows.
- The dashboard now exposes vault-health conflicts through a dedicated health report.
- Tests and builds are passing with `-buildvcs=false`.

## Features

- Index UTF-8 Markdown notes with optional YAML frontmatter
- Preserve unknown frontmatter fields on rewrite
- Parse and merge frontmatter tags with inline tags
- Resolve wikilinks using YANP precedence rules
- Detect and report deterministic conflict winners
- Rewrite inbound links when notes are renamed
- Publish without mutating source notes
- Browse a vault in a Bubble Tea TUI
- Jump from the dashboard into inbox, daily notes, and recent notes

## Download And Run

You can download the Windows build from the GitHub release page:

- [v0.1.0-alpha.1 release](https://github.com/spinchange/yanp-tui/releases/tag/v0.1.0-alpha.1)

You have two options:

1. Download `yanp.exe` directly and run it
2. Download `yanp-v0.1.0-alpha.1-windows-amd64.zip`, extract it, and run `yanp.exe`

For normal use, point the app at your own private real vault. See [docs/VAULTS.md](C:\Users\user\yanp-tui\docs\VAULTS.md) and [config.example.json](C:\Users\user\yanp-tui\config.example.json).

## Key TUI Controls

- `j` / `k`: move selection
- `enter`: open the selected dashboard item
- `/`: filter notes
- `g`: return to the dashboard
- from the dashboard, open the vault health report when conflicts are detected
- `n`: create a new note
- `c`: capture to `inbox.md`
- `R`: rename selected note
- `p`: publish
- `?` or `h`: help
- `q`: quit

## CLI Examples

```powershell
go run -buildvcs=false ./cmd/yanp publish -vault .\testdata\sample-vault -out .\published
go run -buildvcs=false ./cmd/yanp capture -vault .\testdata\sample-vault -text "Follow up with [[Alice]]"
go run -buildvcs=false ./cmd/yanp rename -vault .\testdata\sample-vault -note projects/sprint-review.md -title "Sprint Demo"
```

## Run From Source

From `C:\Users\user\yanp-tui`:

```powershell
go run -buildvcs=false ./cmd/yanp -vault .\testdata\sample-vault
```

## Build A Windows Release Artifact

```powershell
.\build.ps1
```

This produces:

- `dist\yanp.exe`
- `dist\yanp-v0.1.0-alpha.1-windows-amd64.zip`

## Project Layout

- `cmd/yanp`: entry point
- `internal/app`: command routing
- `internal/config`: YANP-style config loading
- `internal/ui`: Bubble Tea terminal interface
- `internal/vault`: indexing, parsing, resolution, publish, and rename logic
- `docs/DEVLOG.md`: development log
- `docs/VAULTS.md`: sample-vault vs real-vault guidance
- `config.example.json`: example config for your own real vault path
- `testdata/sample-vault`: dogfood vault and examples

## Release Notes

This repo is currently aimed at a small first release in the style of:

- version: `v0.1.0-alpha.1`
- platform artifact: Windows `amd64`
- distribution: zipped binary plus project docs

## Known Gaps

- Periodic note creation flows are still minimal.
- The release artifact name is `yanp.exe`.

## Next Milestone

The current working target after `v0.1.0-alpha.1` is `v0.2.0`.

Planned focus:

- periodic note creation for `daily/`, `weekly/`, and `monthly/`
- richer dashboard widgets and saved queries
- stronger vault-health surfacing beyond duplicate-target conflicts
- more polished create and rename workflows
