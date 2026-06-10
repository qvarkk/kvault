# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and the project adheres to [Semantic Versioning](https://semver.org/).

## [0.3.1] - 2026-06-09

Deployment configuration restored.

### Fixed

- `docker-compose.yml` reverted to the previous version: accidentally
  committed local changes removed.

## [0.3.0] - 2026-06-03

Multi-device API keys with expiration, configurable worker retries, and documentation for HTTPS access to the storage.

### Security

- Per-device keys: every login (`login`, `register`) issues a separate key and does **not invalidate** other sessions. Stored in the new `api_keys` table (hash, label, last-used timestamp, expiration).
- Key lifetime is set by `AUTH_API_KEY_TTL` (30 days by default). The window is sliding: each use extends the key; expired keys are rejected at authentication.
- Key management: `GET /auth/keys`, `PATCH /auth/keys/:id` (label), `DELETE /auth/keys/:id`, `POST /auth/logout` (current session), `POST /auth/logout-others` (all but the current one).

### Added

- Configurable background task retries: `WORKER_MAX_RETRIES` and `WORKER_RETRY_TIMEOUT`.
- Extracted file text (`text_content`) in the `GET /files/:id/info` response.
- Automatic device label for a key based on the `User-Agent` header (renamable).
- Documentation: HTTPS access to S3 on the same domain via a path prefix + example config `deploy/nginx.example.conf`; section separators in `.env.example`.

### Changed

- `url_status` migrated from `TEXT` to an enum type (`ENUM`).
- Removed the `POST /auth/refresh` endpoint — single-key rotation replaced by per-device key management.

### Fixed

- Auto-tagging: disabled stopwords can become tags again — the `active_stopwords` filter now respects the `is_enabled` flag.

## [0.2.0] - 2026-05-30

Security hardening and worker task processing fixes.

### Security

- API keys are stored as SHA-256 hashes, generated with a CSPRNG, never logged, and not returned by `GET /auth/me`.
- Key rotation on login: `login`, `register`, and `refresh` issue a new key once.
- SSRF protection for URL fetching: private, loopback, link-local, and CGNAT addresses blocked at connection time, scheme allowlist, redirect and response size limits.
- Rate limiting on `/auth/login` and `/auth/register`.
- Parameterized `active_stopwords` call, `ORDER BY` column allowlist, switch to `ILIKE` with wildcard escaping.
- Eliminated a timing oracle in credential verification.

### Added

- `url_status` field (`pending`/`ready`/`error`) on URL notes.
- `GET /files/:id/info` endpoint and `updated_at` field in the file response.
- The `Authorization` header accepts values both with and without the `Bearer` prefix.

### Changed

- The worker logs task errors (asynq error handler + root PDF processing error).
- Swagger UI moved under `/api/v1`.

### Fixed

- NUL bytes (`0x00`) stripped from text extracted from PDFs and URLs — PostgreSQL rejected such values.
- Whitespace between words when extracting text from web pages.
- Auto-tagging of URL notes (extracted content taken into account) and tag list cache invalidation.
- A failed task enqueue during file upload no longer leaves orphaned DB records and storage objects.
- Logging of S3 object deletion errors.

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

[0.3.1]: https://github.com/qvarkk/kvault/releases/tag/v0.3.1
[0.3.0]: https://github.com/qvarkk/kvault/releases/tag/v0.3.0
[0.2.0]: https://github.com/qvarkk/kvault/releases/tag/v0.2.0
[0.1.0]: https://github.com/qvarkk/kvault/releases/tag/v0.1.0
