---
title: YANP TUI Roadmap v0.2.0
date: 2026-03-28
tags:
  - project
  - roadmap
  - planning
status: active
source: human
project: YANP
---

# YANP TUI Roadmap v0.2.0

The next milestone after `v0.1.0-alpha.2` is `v0.2.0`.

## Release Bar

- Periodic notes are easy and reliable from the TUI.
- The dashboard works as a useful daily landing screen.
- Vault health reports more than duplicate conflicts.
- Create, rename, publish, and setup flows feel stable.
- Repo docs and project notes stay in sync with the shipped app.

## Recommended Sequence

1. Build periodic notes first.
2. Make the dashboard more useful around those notes.
3. Expand vault health beyond duplicate-target reporting.
4. Polish the core create, rename, publish, and setup workflows.
5. Harden verification and release docs before cutting `v0.2.0`.

## Work Packages

### Periodic Notes

- Add daily, weekly, and monthly note creation and current-period open flows.
- Reuse existing notes for the current period instead of creating duplicates.
- Add dashboard shortcuts for current period notes.

### Dashboard And Queries

- Add recent, inbox, draft, stale, and publish-ready widgets.
- Add saved-query support from config.
- Make health and next actions visible from the dashboard.

### Vault Health

- Add unresolved wikilink reporting.
- Add malformed frontmatter or suspicious metadata checks where feasible.
- Add concise stale-note and draft-note summaries.

### Workflow Polish

- Improve note creation validation and feedback.
- Improve rename warnings and success feedback.
- Improve publish summaries and first-run or vault-switching UX.

### Verification And Release

- Add tests for periodic-note naming and path generation.
- Add smoke coverage for the most important changed flows.
- Update README, roadmap, release notes, and mirrored notes before release.

## Suggested Release Sequence

- `v0.2.0-beta.1`: daily notes plus dashboard entry points
- `v0.2.0-beta.2`: weekly and monthly notes plus saved queries
- `v0.2.0-rc.1`: expanded health reporting and workflow polish
- `v0.2.0`: final verification, doc sync, and release packaging
