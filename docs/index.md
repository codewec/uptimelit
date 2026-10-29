# UptimeLit documentation

UptimeLit monitors websites, APIs, TCP services and hosts via ICMP Ping from a Go server, stores their history in SQLite and presents it in a Svelte web interface. Each user has a private workspace; administrators manage accounts without gaining access to other users' monitoring data.

## Start here

1. [Install UptimeLit](guide/installation.md) with Docker, Compose or a source build.
2. [Create the first administrator](guide/getting-started.md) and add a monitor.
3. [Configure notification channels](guide/notifications.md) for outages and recoveries.

## Guides

| Guide | Covers |
| --- | --- |
| [Getting started](guide/getting-started.md) | First sign-in, navbar and the first monitor |
| [Installation and updates](guide/installation.md) | GHCR, Compose, persistent storage and updates |
| [Configuration](guide/configuration.md) | Environment variables, CLI flags and reverse proxies |
| [Monitoring and themes](guide/usage.md) | Status badges, search, history, pause and appearance |
| [Notifications](guide/notifications.md) | Telegram, Pushover and Redis |
| [Users and sessions](guide/users.md) | Roles, private data, profiles and account management |
| [OIDC](guide/oidc.md) | Provider setup, account linking and password login policy |
| [REST API](guide/api.md) | Authentication, resource routes and request examples |
| [Backups and restore](guide/backups.md) | SQLite data, volumes and signing keys |
| [Development](guide/development.md) | Backend/frontend structure, builds, tests and migrations |
| [Screenshots](guide/screenshots.md) | Light and dark interface previews |

The guides describe the current UptimeLit implementation. The [README](../README.md) provides a short project overview and quick-start commands.
