# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and the project adheres to [Semantic Versioning](https://semver.org/).

## [0.3.1] - 2026-06-09

UI stabilization.

### Changed

- Added favicons for all devices.

### Fixed

- Extracted text view: very long text no longer hangs or crashes the page;
  content renders in sections as you scroll.
- Language switching: sidebar items, tooltips, and sort dropdowns (notes,
  files, search) update immediately instead of after the first interaction
  (labels no longer froze at initialization/selection).

## [0.3.0] - 2026-06-03

Sync with the updated API, global search, and session management.

### Added

- Global search: a quick-search palette (Obsidian-style) with query history
  in localStorage and per-entry deletion; a dedicated search page with "Notes"
  and "Files" sections, sorting, and tag filtering for notes. Opens from the
  sidebar or via `Ctrl/Cmd+K`.
- Session management in the "Account" section: a list of API keys with the
  current device marked, renaming, logging out a single device or all others.
- Extracted text view for files (the "Details" modal).

### Changed

- URL note card: removed the duplicated description (it stays in the source),
  a content preview is shown instead. The open-link button got an external
  link icon.
- Version in the "About" section bumped to 0.3.0.

### Fixed

- The file card no longer breaks on long file names.
- File uploads work on plain-HTTP hosts (fallback for `crypto.randomUUID`).

## [0.2.0] - 2026-05-30

Sync with API 0.2.0, UX and security improvements.

### Security

- The API key is no longer displayed in the UI ("Account" tab, sidebar).

### Added

- A "Details" menu with metadata modals for notes and files.
- For URL notes: a source modal (metadata as front-matter + image) and an extracted text modal.
- URL fetch status indicator (`pending`/`ready`/`error`) on note cards.

### Changed

- "Account" tab: the key is hidden; key refresh happens without revealing it.
- File card: the status badge removed and moved into "Details" with clear processing state descriptions.
- Repository links in the "About" section point to gitverse.
- Sync with API 0.2.0: key rotation on login, `GET /auth/me` without the key, the `url_status` field, the `/files/:id/info` endpoint, the `Bearer` header.

### Fixed

- Error notifications are no longer empty: robust API and network error handling with clear messages.

## [0.1.0] - 2026-05-28

First public release.

### Added

- Markdown text notes, URL notes, and PDF file uploads.
- Full-text search across all content with prefix search and ranking.
- Automatic and manual tagging; tag and stopword management (RU/EN).
- Tag filtering, sorting, pagination.
- Trash bin with restore and permanent deletion.
- File storage in S3-compatible storage (Garage), access via presigned URLs.
- API key authentication.
- Web UI: light/dark theme, localization (RU/EN/JA), responsive layout.
- Deployment via Docker Compose.

[0.3.1]: https://github.com/qvarkk/kvault-frontend/releases/tag/v0.3.1
[0.3.0]: https://github.com/qvarkk/kvault-frontend/releases/tag/v0.3.0
[0.2.0]: https://github.com/qvarkk/kvault-frontend/releases/tag/v0.2.0
[0.1.0]: https://github.com/qvarkk/kvault-frontend/releases/tag/v0.1.0
