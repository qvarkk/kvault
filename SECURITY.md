# Security Policy

kvault is a system for **personal self-hosting**. The backend and the frontend share one security model, so this document applies to the whole project.

> [!IMPORTANT]
> kvault was designed as a personal application deployed in a trusted environment. It has **no protection intended for public internet exposure**. Read this document before deploying.

---

## Threat model and assumptions

kvault assumes that:

- the server is operated by the user themselves or by a fully trusted person;
- the service is **not exposed** directly to the open internet;
- access to the server is restricted at the network level (VPN, firewall, private network).

If these assumptions do not hold, data security is not guaranteed.

---

## What you should know about data security

### The server administrator sees everything

The server owner has full access to **all** users' data:

- note content and text extracted from web pages;
- uploaded files;
- tags and stopwords;
- users' **API keys**.

Data is stored in the database and the file storage **without encryption**. Only entrust your data to a server operated by you or by someone you fully trust.

### API key lifetime and revocation

Each login issues a **separate per-device key** — other sessions are not invalidated. A key works on a **sliding expiration**: it expires if unused for longer than `AUTH_API_KEY_TTL` (30 days by default). Every use extends the lifetime.

If a key leaks (unlikely for a personal, non-exposed server, but still possible), access can be revoked in the account settings:

- **delete a specific key** — disconnects the corresponding device;
- **log out** — revokes the current session's key;
- **log out on other devices** — revokes all keys except the current one.

Expired keys are rejected at authentication.

### There is no encryption

kvault does not implement at-rest data encryption and has no protection against an attacker who gains access to the server or the Docker volumes. Data protection rests entirely on the **environment** the service runs in.

---

## Deployment recommendations

- **Do not publish the service to the open internet.** Use a VPN that admits only trusted devices, or restrict access with an IP firewall. A simple ready-made option is to expose kvault only inside a tailnet via Tailscale: see [Access via Tailscale](./HOSTING.md#access-via-tailscale-vpn).
- **Use HTTPS** when accessing via a domain — terminate TLS at the reverse proxy (see [HOSTING.md](./HOSTING.md)).
- **Use generated secrets.** `setup.sh` creates `.env` with random passwords and keys — there are no defaults. If you fill `.env` manually, generate secrets with `openssl rand -hex 24` and avoid dictionary passwords.
- **Protect `.env`.** The file contains database passwords and storage access keys. Do not commit it to version control; restrict its permissions on the server.
- **Make backups** of the PostgreSQL database and the Garage storage (`garage_meta`/`garage_data`, or the `./data/garage` directory with bind mounts) — step by step in [HOSTING.md](./HOSTING.md#updates-and-maintenance). Keep the copies somewhere safe.
- **Keep up to date.** Periodically pull fresh images: `docker compose pull && docker compose up -d` — see [HOSTING.md](./HOSTING.md#updates-and-maintenance).

---

## Reporting a vulnerability

If you find a vulnerability, **do not open a public issue**. Email **kvault@gmail.com**.

Describe the problem, reproduction steps, and potential impact. We will try to respond within a reasonable time.
