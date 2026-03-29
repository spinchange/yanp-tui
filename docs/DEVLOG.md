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

### Notes

- Go verification should be run with `-buildvcs=false` in this environment because `C:\Users\user` is a large Git worktree and default VCS stamping is slow.
- The current release artifact name is `yanp.exe`.

### Next likely work

- Surface conflict warnings in the dashboard.
- Add periodic note creation flows for `daily/`, `weekly/`, and `monthly/`.
- Surface conflicts and other vault health warnings in the dashboard.
