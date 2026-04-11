# YANP TUI Devlog

## 2026-04-11

### Milestone

Released `v0.2.0`.

### What shipped

**Periodic notes**
- Daily, weekly, and monthly notes now use date-specific titles (`2026-04-11`, `2026-W15`, `2026-04`) instead of the generic "Daily Note" — all notes are now distinct in the browser.
- Extracted `PeriodicRelPath` from vault to eliminate duplicated path logic in the UI layer.
- Dashboard periodic shortcuts no longer appear twice when the note already exists; each period shows one entry (direct link if present, create shortcut if not).

**Dashboard widgets**
- Recent notes count now respects `cfg.Defaults.DashboardLimit`.
- `StaleNotes(days, asOf)` added to vault; stale count shown in the overview panel.
- `DraftNotes()` added to vault; draft count in the overview panel and a "Browse drafts" shortcut on the dashboard (bypasses text filter, uses the method directly).
- Saved queries: `Config.Queries` changed from an unused `string` to `[]SavedQuery{Name, Filter}`. Queries appear as selectable dashboard items and activate the existing filter machinery.

**Vault health**
- Health view is now scrollable via `j/k` and `ctrl+d/u` — content no longer silently overflows on large vaults.
- Malformed frontmatter is now recorded rather than fatal: `Vault.ParseErrors []ParseError` is populated during `Load`, and notes with bad YAML are still indexed with their filename stem as title. Parse errors are surfaced in the health view and counted in the overview.

**Workflow polish**
- Note creation (`n`) now accepts a `path/Title` prefix to target a subdirectory (e.g. `projects/My Note` → `projects/my-note.md`). Status message shows the created path.
- `RenameNote` now rejects destination conflicts before writing any file ("a note already exists at …") and same-title renames ("new title produces the same filename").
- Rename and publish warnings are now shown in the browser viewport after the operation completes, not just counted in the status bar.
- Switch-vault error messages distinguish "path does not exist" (with a hint to use Shift+V) from "path is a file".
- Vault creation status message lists the directories and files actually scaffolded.

**Verification**
- `internal/app` now has smoke tests covering `capture`, `rename`, and `publish` subcommands (12 tests).
- New vault tests: `TestPeriodicNoteHasDateSpecificTitle`, `TestPeriodicRelPath`, `TestStaleNotes`, `TestDraftNotes`, `TestMalformedFrontmatterIsRecordedNotFatal`, `TestRenameRejectsConflictingDestination`, `TestRenameRejectsSameTitle`.
- New UI tests: `TestParseNoteInput`, `TestDashboardItemsIncludesSavedQueries`.
- New config tests: `TestSavedQueryRoundTrip`, `TestEmptyQueriesOmittedFromJSON`.

### Notes

- `PeriodicRelPath` is now the canonical exported function for computing periodic note paths. The internal `periodicSpec` was simplified to three return values (relPath, metadata, body) — the separate `title` return was redundant since title is embedded in metadata.
- The `vaultPeriodicSpec` helper in `model.go` was removed entirely; all callers use `vault.PeriodicRelPath`.
- `parseNote` now returns `(*Note, *ParseError, error)`. Callers that write files themselves (`CreateNote`, `EnsurePeriodicNote`) discard the parse error with `_`.

## 2026-03-28

### Milestone

Prepared `yanp-tui` for a first alpha-style release as `v0.1.0-alpha.1`.

### Follow-up release

Prepared `v0.1.0-alpha.2` as a parser and packaging follow-up release.

### What changed in alpha.2

- Fixed slug generation so `/` and `\` no longer leak into note filenames on Windows.
- Normalized CRLF frontmatter parsing so metadata is read cleanly from Windows-authored notes.
- Reworked inline-code stripping so escaped backticks and multi-backtick spans do not suppress real tag extraction.
- Added regression tests for slash-containing slugs, CRLF frontmatter, and escaped backtick tag extraction.
- Updated release metadata and rebuilt the Windows artifact set for `v0.1.0-alpha.2`.

### What shipped into this milestone

- Built a Go PKM tool around the YANP vault spec.
- Added a Bubble Tea TUI with:
  - dashboard home screen
  - note browser and preview
  - filtering
  - quick capture to `inbox.md`
  - note creation
  - rename flow with inbound wikilink rewriting
  - publish flow to CommonMark output
  - in-app help
- Added self-documenting notes inside the sample YANP vault.
- Added conflict-resolution tests and fixed title-vs-filename precedence behavior.
- Added `build.ps1` to produce a Windows binary and release zip.
- Surfaced vault conflict health in the dashboard through a dedicated health report view.
- Documented the distinction between the tracked `sample-vault` and a user-chosen real private vault.
- Added a first-run setup flow so the app prompts for a vault instead of guessing the current folder.
- Added in-app vault management for switching to an existing vault or creating a new one.
- Added daily, weekly, and monthly note creation/open flows with dashboard and keyboard entry points.
- Added dashboard summaries for inbox state and current daily, weekly, and monthly notes.
- Expanded vault health to report unresolved wikilinks alongside duplicate targets.

### Notes

- Go verification should be run with `-buildvcs=false` in this environment because `C:\Users\user` is a large Git worktree and default VCS stamping is slow.
- The current release artifact name is `yanp.exe`.
- `YANP` now uses `C:\Users\user\.yanp\config.json` as its own config home.

### Next likely work

- Add richer dashboard widgets and saved queries.
- Expand health reporting beyond duplicate targets and unresolved wikilinks.
- Polish the periodic-note experience with stronger summaries and current-period widgets.
