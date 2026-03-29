# YANP TUI Devlog

## 2026-03-28

### Milestone

Prepared `yanp-tui` for a first alpha-style release as `v0.1.0-alpha.1`.

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

### Notes

- Go verification should be run with `-buildvcs=false` in this environment because `C:\Users\user` is a large Git worktree and default VCS stamping is slow.
- The current release artifact name is `yanp.exe`.
- `YANP` now uses `C:\Users\user\.yanp\config.json` as its own config home.

### Next likely work

- Add periodic note creation flows for `daily/`, `weekly/`, and `monthly/`.
- Add richer dashboard widgets and saved queries.
- Expand health reporting beyond duplicate title, alias, and filename targets.
