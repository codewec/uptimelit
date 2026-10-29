# Getting started

## First sign-in

[Start UptimeLit](installation.md) and open `http://localhost:8080` (or your public deployment URL).

With an empty database, enter a new username and password on the login page. This creates the first administrator. A successful OIDC sign-in can also create the first administrator when a provider is configured.

Usernames accept 1–64 ASCII letters, digits, dots, underscores or dashes. Passwords must contain 8–72 bytes. Once any account exists, sign-in only authenticates existing users; there is no public registration.

## Navbar

| Control | Action |
| --- | --- |
| UptimeLit logo | Return to monitors |
| Search | Filter monitors by name/address or notifiers by name/provider |
| Plus icon | Create a monitor from the monitors page |
| Theme icon | Choose a built-in theme or edit custom colors |
| Bell icon | Open your notifiers |
| Shield icon | Open administration; administrators only |
| Profile icon | Change account settings |
| Sign-out icon | Revoke the current session and return to login |

On narrow screens, the menu contains notifier, profile, administration and theme actions. The plus icon remains available on the monitors page.

## Add your first monitor

1. Click the **+** icon in the navbar.
2. Enter a name and choose **HTTP / HTTPS** or **TCP port**.
3. Enter a URL such as `https://api.example.com/health`, or a TCP address such as `database:5432`.
4. Choose an interval of at least 10 seconds and a positive failure threshold. The form starts at 60 seconds and three failures.
5. Select your notification channels if any exist, then click **Create monitor**.

New workers start with a short stagger before the first check. Cards show the result after a check completes; the visible monitor list refreshes approximately every five seconds while the tab is visible.

For alerts, open the bell icon, create a [notifier](notifications.md), then edit the monitor to attach it. The monitor's actions menu provides history, editing, pause/resume and deletion.

## Create other users

Open the shield icon and choose **Create user**. Set a username, role and password, or pre-provision a username without a password when OIDC is configured. See [users and sessions](users.md) and [OIDC](oidc.md).
