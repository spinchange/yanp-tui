---
title: YANP Vault Spec
date: 2026-03-28
tags:
  - spec
  - pkm
  - markdown
status: active
source: human
aliases:
  - YANP Spec
  - YANP vault specification draft
project: YANP
---

# YANP Vault Spec

Reference: [spinchange.github.io/yanp](https://spinchange.github.io/yanp/)

## Core Rules

- Notes are UTF-8 Markdown files with lowercase kebab-case filenames.
- YAML frontmatter is optional and unknown fields must be preserved.
- Internal links use wikilinks like `[[Note Title]]` or `[[Note Title|Display Text]]`.
- Resolution order is title, aliases, then filename stem.
- Publish output must be valid CommonMark with wikilinks transformed to relative Markdown links.

## Tooling Implications

- Rename flows need to rewrite inbound wikilinks across the whole vault.
- Publish must never mutate source notes.
- Tags come from both frontmatter and inline body tags.
- Periodic notes conventionally live in `daily/`, `weekly/`, and `monthly/`.
