# CLAUDE.md

## Project Overview

Hugo documentation site for "EC" (Empyrean Challenge) using the Hextra theme, published to https://epimethian.dev/.

## Quick Reference

- **Config**: `hugo.yaml`
- **Theme**: Hextra (`github.com/imfing/hextra`) via Hugo modules
- **Content**: `content/` — Markdown with YAML front matter
- **Custom layouts**: `layouts/report/`
- **Custom CSS**: `assets/css/`
- **Skills**: `skills/diataxis/` — load when writing/reviewing docs

## Content Organization

Content uses the Diátaxis framework, organized by audience (`players`, `referees`, `developers`) with sub-sections: `tutorials/`, `how-to/`, `reference/`, `explanation/`.

Other top-level sections: `history/` (preserved manuals), `reference/` (lookup tables), `blog/`, `reports/`.

## Key Conventions

- YAML front matter (`---`), not TOML (`+++`)
- Section index files: `_index.md`
- Leaf pages: kebab-case filenames
- Hextra shortcodes: `callout`, `cards`, `card`, `tabs`, `tab`
- Historical content (`history/`) must stay faithful to originals

## Commands

```sh
hugo server    # Dev server at localhost:1313
hugo           # Build to public/
```

## Don't Touch

- `go.mod` / `go.sum` (Hugo module managed)
- `public/` (build output, gitignored)
- `history/` content (don't modernize without explicit ask)
