# Release Notes

## v0.1.0-alpha.1

First public alpha-style preview of `yanp-tui`, a Go PKM tool with a Bubble Tea TUI built around the draft YANP vault spec.

### Included

- YANP vault indexing for UTF-8 Markdown notes with optional YAML frontmatter
- Wikilink resolution by title, alias, then filename stem
- Publish transform to CommonMark in a separate output directory
- Rename flow that rewrites inbound wikilinks across the vault
- Inline tag parsing with code-block and inline-code exclusions
- Deterministic conflict resolution based on most recent modification time
- Conflict detection support in the vault engine
- Bubble Tea TUI with:
  - dashboard
  - note browser and preview
  - note filtering
  - quick capture to `inbox.md`
  - note creation
  - rename and publish flows
  - in-app help
- Self-documenting sample vault for dogfooding
- Windows release packaging through `build.ps1`

### Known Gaps

- Dashboard conflict surfacing is not wired yet
- Periodic note creation flows are still minimal
- Release distribution is currently focused on Windows `amd64`

### Artifacts

- `dist\yanp.exe`
- `dist\yanp-v0.1.0-alpha.1-windows-amd64.zip`
