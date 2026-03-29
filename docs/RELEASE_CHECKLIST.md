# Release Checklist

## Repo

- Confirm working tree is clean
- Confirm version string matches intended release
- Confirm README and devlog reflect current behavior
- Confirm release notes are updated

## Verification

- Run `go test -buildvcs=false ./...`
- Run `go build -buildvcs=false ./cmd/yanp`
- Run `.\build.ps1`
- Smoke-test `dist\yanp.exe`

## Packaging

- Confirm `dist\yanp.exe` exists
- Confirm `dist\yanp-v0.1.0-alpha.1-windows-amd64.zip` exists
- Confirm the zip contains:
  - `yanp.exe`
  - `README.md`
  - `docs\DEVLOG.md`

## Publish

- Create or confirm the remote repository
- Push the default branch
- Create a Git tag for `v0.1.0-alpha.1`
- Upload the zip artifact
- Paste in the release notes from `docs\RELEASE_NOTES_v0.1.0-alpha.1.md`

## After Release

- Record the release in the devlog
- Decide next milestone
- Prioritize post-alpha polish items
