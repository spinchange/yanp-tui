---
title: YANP TUI README
date: 2026-03-28
tags:
  - project
  - docs
  - release
status: active
source: human
project: YANP
---

# YANP TUI README

This note mirrors the application README for dogfooding inside the vault.

## Summary

`yanp-tui` is a Go PKM tool with a TUI, built around [[YANP Vault Spec]].

## Release Target

- Version: `v0.1.0-alpha.1`
- Windows artifact: `dist\yanp.exe`
- Release package: `dist\yanp-v0.1.0-alpha.1-windows-amd64.zip`

## Core Features

- YANP vault indexing
- Wikilink resolution and publish transform
- Rename with inbound link rewriting
- Dashboard, filtering, help, capture, create, rename, and publish flows

## Commands

```powershell
go run -buildvcs=false ./cmd/yanp -vault .\testdata\sample-vault
.\build.ps1
```

## Notes

- The environment currently benefits from `-buildvcs=false` during Go verification.
- Future polish includes surfacing conflict warnings directly in the dashboard.
