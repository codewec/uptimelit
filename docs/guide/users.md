# Users and sessions

## First administrator

When the database has no users, the first password sign-in creates an administrator with the entered credentials. The first successful OIDC sign-in can do the same when a provider is configured. After the first account exists, neither login method creates arbitrary new accounts.

There is no public registration. Usernames accept 1–64 ASCII letters, digits, dots, underscores or dashes. Passwords must contain 8–72 bytes.

## Roles and private data

| Capability | User | Administrator |
| --- | --- | --- |
| Manage own monitors, history and notifiers | Yes | Yes |
| Edit own profile | Yes | Yes |
| Create and manage user accounts | No | Yes |
| Access another user's monitors or notifier settings | No | No |

The shield icon opens **Administration** for administrators. The page lists accounts and their sign-in methods, not their private monitoring data.

## Manage accounts

Administrators can create users, change usernames and roles, reset passwords, reset OIDC bindings and delete accounts. Creation requires a password unless OIDC is configured; an account without a password is pre-provisioned for [OIDC linking](oidc.md).

Resetting an OIDC binding requires an available password login method. The last administrator cannot be demoted or deleted. You cannot delete your own account. Deleting another account removes its monitors, check history, notifiers and sessions and stops its monitoring workers.

## Profile

Open the profile icon to change your username or password. Username changes are unavailable in an OIDC session. A password change requires the current password when the account already has one. An OIDC-only user can set a first password from a confirmed OIDC session while password login is enabled.

Saving profile changes revokes your sessions and returns you to login. Use the updated credentials to sign in again.

## Sessions

Sessions expire after seven days. The browser uses an HttpOnly cookie; the [API](api.md) also accepts Bearer tokens. Signing out revokes the current session.

Profile changes, administrative account updates and initial linking to an OIDC identity revoke the affected user's existing sessions. Account deletion revokes all of that account's sessions. Cookies and Bearer tokens refer to the same server-side session records.
