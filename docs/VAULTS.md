# Vault Setup

`yanp-tui` now distinguishes between two vault roles:

## 1. Sample vault

Tracked in this repo at:

- `testdata/sample-vault`

Purpose:

- dogfood the YANP format
- provide stable fixtures for tests
- keep project documentation inside a real YANP vault

This vault is intentionally committed to Git.

## 2. Real vault

Expected private vault pattern:

- a private vault path that you choose yourself

Purpose:

- day-to-day notes
- private knowledge management
- your personal working vault outside this repo

This vault should stay outside this repo and live in its own Git repository.

## Recommended config

Copy `config.example.json` to:

- `C:\Users\user\.yanp\config.json`

That points `yanp-tui` at your chosen real vault while leaving the sample vault available for tests and docs.

## Typical usage

- Use `testdata/sample-vault` when developing or testing `yanp-tui`
- Use your own private real-vault path in everyday use
- Keep release artifacts and binaries out of both vaults
