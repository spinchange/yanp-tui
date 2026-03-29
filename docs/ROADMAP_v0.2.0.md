# Roadmap

## Target: v0.2.0

### Goal

Move `yanp-tui` from a strong alpha preview toward a more complete daily-driver PKM tool.

### Release bar

Call `v0.2.0` done when:

- periodic notes are easy and reliable to create from the TUI
- the dashboard is useful as a daily landing screen instead of just a navigation menu
- vault health surfaces real problems beyond duplicate-target conflicts
- create, rename, publish, and first-run flows feel stable and low-friction
- docs, sample-vault notes, and release notes match the shipped behavior

### Recommended sequence

#### 1. Periodic notes first

Ship the biggest daily-use gap before polishing secondary surfaces.

- Add daily-note creation and open-today behavior
- Add weekly-note creation and open-current-week behavior
- Add monthly-note creation and open-current-month behavior
- Add dashboard shortcuts for current period notes
- Add tests for naming, path generation, and existing-note reuse

#### 2. Make the dashboard useful

Turn the dashboard into a real home screen after periodic-note primitives exist.

- Add recent-notes and inbox-focused widgets
- Add current-period note shortcuts and summaries
- Add saved-query support from config
- Add clearer health summaries with direct next actions

#### 3. Expand vault health

Build on the current duplicate-target checks with higher-value diagnostics.

- Add unresolved wikilink reporting
- Add malformed or suspicious frontmatter reporting where feasible
- Add draft, stale-note, or other actionable content summaries
- Keep health output deterministic and easy to scan in the TUI

#### 4. Polish core workflows

Improve the flows users hit most often once the major feature gaps are closed.

- Tighten note creation validation and feedback
- Improve rename warnings and success messaging
- Make publish results and warnings easier to understand
- Smooth first-run, switch-vault, and create-vault interactions

#### 5. Harden the release

Do not cut `v0.2.0` until the main flows are verified end to end.

- Extend unit tests around periodic-note and health logic
- Add smoke-test coverage for the main CLI and TUI paths that changed
- Update README, roadmap, release notes, and mirrored sample-vault notes
- Run full verification and rebuild release artifacts

### Work packages

#### A. Periodic note engine

- Create `daily/`, `weekly/`, and `monthly/` notes from the TUI
- Follow YANP naming conventions directly
- Reuse an existing note when the current period already has one
- Support clean note titles and default starter content where appropriate

#### B. Dashboard and queries

- Add widgets for recent notes, stale notes, draft notes, and publish-ready notes
- Add support for saved queries from config
- Make dashboard health more actionable
- Make current-period notes reachable in one or two keypresses

#### C. Vault health

- Extend health reporting beyond duplicate targets
- Add unresolved link reporting
- Add stale-note and draft-note summaries
- Prefer concise summaries with drill-down details rather than noisy walls of text

#### D. Workflow polish

- Improve note creation forms
- Improve rename feedback and warnings
- Add clearer publish options and summaries
- Tighten setup and vault-switching UX

#### E. Verification and release

- Add regression tests for new date-based path and note-generation logic
- Add smoke coverage for the most important interactive flows
- Keep release docs and project notes in sync with shipped behavior

### Suggested release sequence

- `v0.2.0-beta.1`: daily notes plus dashboard entry points
- `v0.2.0-beta.2`: weekly and monthly notes plus saved queries
- `v0.2.0-rc.1`: expanded health reporting and workflow polish
- `v0.2.0`: release after verification, doc sync, and artifact refresh
