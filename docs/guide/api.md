# REST API

All monitoring, notifier and user-management routes require authentication. Sign in with a password:

```bash
curl -H 'Content-Type: application/json' -H 'X-UptimeLit-Request: 1' \
  -d '{"username":"alice","password":"your-password"}' \
  http://localhost:8080/api/auth/login
```

The response contains `token` and `user`. Pass the token as `Authorization: Bearer <token>` for subsequent requests. The browser uses an HttpOnly session cookie. All writes without a Bearer token, including login, require `X-UptimeLit-Request: 1` for CSRF protection.

## Authentication and accounts

| Method | Route | Purpose |
| --- | --- | --- |
| GET | `/api/auth/config` | Public login-method and bootstrap configuration |
| POST | `/api/auth/login` | Password sign-in; creates the first administrator on an empty database |
| POST | `/api/auth/logout` | Revoke the current session |
| GET | `/api/auth/me` | Current user |
| PUT | `/api/auth/profile` | Update `username`, optionally `newPassword` and `currentPassword` |
| GET | `/api/auth/oidc/login` | Start OIDC login |
| GET | `/api/auth/oidc/callback` | OIDC provider callback |
| GET / POST | `/api/admin/users` | List or create users; administrators only |
| PUT / DELETE | `/api/admin/users/{id}` | Update or delete users; administrators only |

User creation accepts `username`, `role` (`admin` or `user`, default `user`) and `password` (optional with OIDC). Updates accept `username`, `role`, optional `newPassword` and `resetOIDC`. Resetting an OIDC binding requires an available password login method.

## Monitors and notifiers

Every resource route is scoped to the signed-in user. Accessing another user's resource returns `404`, including for administrators.

| Method | Route | Purpose |
| --- | --- | --- |
| GET / POST | `/api/endpoints` | List own monitors with current statistics, or create a monitor |
| GET / PUT / DELETE | `/api/endpoints/{id}` | Read, update or delete a monitor and its history |
| POST | `/api/endpoints/{id}/pause` | Toggle monitoring pause |
| GET | `/api/endpoints/{id}/history?limit=60` | Recent history for the timeline |
| GET | `/api/endpoints/{id}/checks?limit=100` | Check details with status codes and response times |
| GET / POST | `/api/notifiers` | List own notifiers or create a notifier |
| GET / PUT / DELETE | `/api/notifiers/{id}` | Read, update or delete a notifier |
| POST | `/api/notifiers/{id}/test` | Send a test notification |
| GET | `/api/status` | Application version |

## Create a monitor

Set `TOKEN` to the token returned by login, then create an HTTP monitor:

```bash
curl -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"name":"Production API","type":"http","url":"https://api.example.com/health","interval_seconds":60,"fail_threshold":3,"notifier_ids":[]}' \
  http://localhost:8080/api/endpoints
```

For TCP, use `"type":"tcp"` and a `host:port` value in `url`. Intervals must be at least 10 seconds. Use a positive `fail_threshold`. Supply the same configuration fields when updating a monitor with PUT. Notifier IDs must belong to the signed-in user; monitor changes and their notifier links save together in a transaction.

Notifier creation/update accepts `name`, `type` (`telegram`, `pushover` or `redis`) and `config_json`, which is a **JSON-encoded string**, not a nested object. For example, a Telegram request body is:

```json
{
  "name": "Operations chat",
  "type": "telegram",
  "config_json": "{\"bot_token\":\"provider-issued-token\",\"chat_id\":\"12345\"}"
}
```

See [notification providers](notifications.md) for the settings supported by each channel.

## Errors and sessions

Responses use JSON; errors normally have an `error` message. `401` means authentication is missing, expired or revoked. `403` indicates insufficient role privileges, a disabled login method or a rejected CSRF check. Resource routes return `404` for another user's objects. Password login is limited to 20 attempts per minute per remote address and returns `429` with `Retry-After` when exceeded.

Session Bearer tokens are revoked by logout and account changes just like browser cookie sessions. An administrator token cannot bypass resource ownership checks.

See [users and sessions](users.md) for account rules and [monitoring](usage.md) for check behavior.
