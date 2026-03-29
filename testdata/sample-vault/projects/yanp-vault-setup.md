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
