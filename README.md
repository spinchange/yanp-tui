# yanp-tui

`yanp-tui` is a Go personal knowledge management tool with a terminal UI, built to work with the draft [YANP vault specification](https://spinchange.github.io/yanp/).

YANP-TUI stands for `Yet Another Notes Project - Terminal User Interface`.

Current release: `v0.2.0`

## Current Status

- Core YANP vault indexing works.
- Wikilinks resolve by title, alias, then filename stem.
- Rename updates inbound wikilinks across the vault, with conflict detection before any file is written.
- Publish writes CommonMark output to a separate directory.
- The TUI includes a dashboard, browser, preview, filtering, help, capture, create, rename, and publish flows.
- Daily, weekly, and monthly notes use date-specific titles and can be created or opened from the dashboard or keyboard.
- The dashboard surfaces inbox state, current-period notes, recent notes, draft notes, stale note counts, and saved queries from config.
- Vault health reports duplicate targets, unresolved wikilinks, and malformed frontmatter, with a scrollable detail view.
- Rename and publish warnings are shown in the browser viewport after the operation completes.
- Tests and builds are passing with `-buildvcs=false`.

## Features

- Index UTF-8 Markdown notes with optional YAML frontmatter
- Preserve unknown frontmatter fields on rewrite
- Parse and merge frontmatter tags with inline tags
- Resolve wikilinks using YANP precedence rules
- Detect and report deterministic conflict winners
- Rewrite inbound links when notes are renamed
- Publish without mutating source notes; skip drafts with `-skip-drafts`
- Browse a vault in a Bubble Tea TUI
- Dashboard with inbox, current-period, recent, draft, and saved-query shortcuts
- Open or create the current daily, weekly, and monthly notes from the dashboard or keyboard
- Create notes in any vault subdirectory using `path/Title` syntax
- Scrollable vault health view: duplicate targets, unresolved wikilinks, malformed frontmatter
- Rename and publish warnings surfaced in the viewport after completion
- Saved queries pinned to the dashboard via `~/.yanp/config.json`
- Stale-note and draft-note counts in the dashboard overview

## Download And Run

You can download the Windows build from the GitHub release page:

- [v0.2.0 release](https://github.com/spinchange/yanp-tui/releases/tag/v0.2.0)

You have two options:

1. Download `yanp.exe` directly and run it
2. Download `yanp-v0.2.0-windows-amd64.zip`, extract it, and run `yanp.exe`

For normal use, point the app at your own private real vault. See [docs/VAULTS.md](docs/VAULTS.md) and [config.example.json](config.example.json).

On first run, if no vault is configured yet, YANP-TUI opens a setup flow. You can:

- enter the path to an existing vault
- press `V` to create a new vault at a path you choose

The selected vault is saved to `~/.yanp/config.json`.

## Config

`~/.yanp/config.json` supports the following fields:

```json
{
  "vault": "/path/to/your/vault",
  "editor": "",
  "noOpen": false,
  "defaults": {
    "staleDays": 30,
    "dashboardLimit": 5
  },
  "templates": "/path/to/templates",
  "queries": [
    { "name": "Drafts", "filter": "status:draft" },
    { "name": "Project Alpha", "filter": "#project/alpha" }
  ]
}
```

Saved queries appear as selectable dashboard shortcuts and run the same filter as the `/` prompt.

## Key TUI Controls

- `j` / `k`: move selection (also scrolls the health view)
- `ctrl+d` / `ctrl+u`: half-page scroll in the health view
- `enter`: open the selected dashboard item
- `/`: filter notes
- `g`: return to the dashboard
- `v`: switch to a different existing vault
- `V`: create and switch to a new vault
- `d`: open or create today's daily note
- `w`: open or create this week's note
- `m`: open or create this month's note
- `n`: create a new note (prefix with `path/` to target a subfolder, e.g. `projects/My Note`)
- `c`: capture to `inbox.md`
- `R`: rename selected note
- `p`: publish vault to a CommonMark output directory
- `r`: refresh the vault index
- `?` or `h`: help
- `q`: quit

## CLI Examples

```powershell
go run -buildvcs=false ./cmd/yanp publish -vault .\testdata\sample-vault -out .\published
go run -buildvcs=false ./cmd/yanp publish -vault .\testdata\sample-vault -out .\published -skip-drafts
go run -buildvcs=false ./cmd/yanp capture -vault .\testdata\sample-vault -text "Follow up with [[Alice]]"
go run -buildvcs=false ./cmd/yanp rename -vault .\testdata\sample-vault -note projects/sprint-review.md -title "Sprint Demo"
```

## Run From Source

```powershell
go run -buildvcs=false ./cmd/yanp -vault .\testdata\sample-vault
```

## Build A Windows Release Artifact

```powershell
.\build.ps1
```

This produces:

- `dist\yanp.exe`
- `dist\yanp-v0.2.0-windows-amd64.zip`

## Project Layout

- `cmd/yanp`: entry point
- `internal/app`: command routing and CLI flags
- `internal/config`: config loading and `SavedQuery` type
- `internal/ui`: Bubble Tea terminal interface
- `internal/vault`: indexing, parsing, resolution, publish, rename, and health logic
- `docs/DEVLOG.md`: development log
- `docs/VAULTS.md`: sample-vault vs real-vault guidance
- `config.example.json`: example config for your own real vault path
- `testdata/sample-vault`: dogfood vault and examples
