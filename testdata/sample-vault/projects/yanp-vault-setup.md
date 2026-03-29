---
title: YANP Vault Setup
date: 2026-03-28
tags:
  - project
  - vault
  - docs
status: active
source: human
project: YANP
---

# YANP Vault Setup

`yanp-tui` uses two different vault roles.

## Sample vault

- Path: `testdata/sample-vault`
- Tracked in the `yanp-tui` repo
- Used for tests, examples, and self-documentation

## Real vault

- Path: your own private vault path outside this repo
- Private
- Lives in its own Git repo
- Intended for actual day-to-day PKM use

## Config

Use `config.example.json` as the starting point for `C:\Users\user\.yanp\config.json` so the app opens your real vault by default.

## First Run

- If no vault is configured yet, `yanp-tui` opens a setup flow.
- You can enter the path to an existing vault and press Enter.
- You can press `V` to create a brand-new vault folder with conventional YANP subfolders.

## Changing Vaults

- Press `v` in the app to switch to an existing vault.
- Press `V` in the app to create and switch to a new vault.
- The selected vault is saved into `C:\Users\user\.yanp\config.json`.
