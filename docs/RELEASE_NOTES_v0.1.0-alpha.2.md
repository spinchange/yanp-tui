# Release Notes

## v0.1.0-alpha.2

Follow-up alpha release for `yanp-tui`, focused on vault parser correctness and Windows-safe note naming.

### Included

- Windows-safe slug generation for note creation and rename
- CRLF frontmatter normalization for Windows-authored Markdown files
- Inline tag parsing that correctly skips escaped-backtick and multi-backtick code spans
- Regression tests for slash-containing slugs, CRLF frontmatter parsing, and escaped backtick tag extraction
- Existing YANP vault indexing, wikilink resolution, publish transform, rename backlink rewriting, and Bubble Tea TUI flows
- Windows release packaging through `build.ps1`

### Known Gaps

- Periodic note creation flows are still minimal
- Release distribution is currently focused on Windows `amd64`

### Artifacts

- `dist\yanp.exe`
- `dist\yanp-v0.1.0-alpha.2-windows-amd64.zip`
