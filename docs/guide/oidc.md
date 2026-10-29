# OIDC authentication

UptimeLit supports OpenID Connect providers such as Pocket ID alongside local password authentication.

## Provider setup

Register a client with your provider and set its callback to the public UptimeLit URL:

```text
https://monitor.example.com/api/auth/oidc/callback
```

For the included Compose service, set these values in the project `.env` file:

```dotenv
OIDC_ISSUER=https://id.example.com
OIDC_CLIENT_ID=uptimelit
OIDC_CLIENT_SECRET=provider-issued-secret
OIDC_REDIRECT_URL=https://monitor.example.com/api/auth/oidc/callback
OIDC_BUTTON_TITLE=Sign in with Pocket ID
```

Recreate the container with `docker compose up -d`. For `docker run`, pass the values with `-e` or `--env-file`. A native process reads them from its environment, not from `.env` automatically.

Issuer and client ID must both be set. The callback must be an absolute HTTP(S) URL and match the provider configuration. UptimeLit uses discovery, authorization code exchange, PKCE, state and nonce checks, and validates the ID token's signature, issuer, audience and expiry.

## First sign-in and account linking

If there are no users, the first successful OIDC sign-in creates an administrator.

After bootstrap, an identity first matches an existing binding by issuer and subject. Otherwise, it can link to an unbound, administrator-created account whose username matches the provider's normalized username. The local password and role are preserved; a conflicting existing OIDC binding is not overwritten.

The username is derived from `preferred_username`, then the email local part, then display name. Unsupported characters become dashes, surrounding punctuation is stripped and names are limited to 64 characters. If no usable name is available, a stable identity-based fallback is used.

Use a trusted provider whose usernames you control. Administrators should pre-create the matching username before the user attempts OIDC login. Unknown identities are denied after bootstrap; there is no automatic public registration through the provider.

## Password login policy

Password login stays enabled by default. After testing the complete OIDC flow, set:

```dotenv
DISABLE_PASSWORD_LOGIN=true
```

This switch disables password login only with OIDC configured. With no issuer/client ID configured, password login remains available. Partial configuration is rejected at startup rather than enabling a partially configured provider.

## Private certificates

Prefer a provider certificate issued by a CA trusted by the host/container. If necessary, `OIDC_INSECURE_SKIP_TLS_VERIFY=true` explicitly skips certificate verification for OIDC discovery, code exchange and key retrieval. It emits a startup warning and applies only to OIDC, not monitored HTTP endpoints.

## Troubleshooting

- Ensure the issuer is reachable from the UptimeLit container and discovery succeeds.
- Check that the provider callback exactly matches `OIDC_REDIRECT_URL`.
- Forward the public host and `X-Forwarded-Proto` through the reverse proxy.
- Check that the account is pre-provisioned with the expected normalized username.
- Retry login if the temporary OIDC cookies expired; they last 10 minutes.
- Keep server clocks synchronized for token expiry checks.
