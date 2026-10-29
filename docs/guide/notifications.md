# Notifications

## Create and attach a channel

Open the bell icon in the navbar, then choose **New notifier**. On a narrow screen, open **Menu → Notifiers**. Name the channel and select its provider.

After saving, use **Send test** to verify delivery. The test sends a real notification to the configured destination. Edit a monitor and select the channel under **Notification channels** to receive its outage and recovery alerts. Creating a notifier alone does not attach it to any monitors.

You can attach multiple notifiers to one monitor. Notifiers and their settings are private to their owner; cross-user links are rejected.

## Providers

| Provider | Required settings | Optional settings |
| --- | --- | --- |
| Telegram | Bot token, chat ID | — |
| Pushover | Application token, user key | Priority: Low, Normal or High |
| Redis | Server address (`host:port`) | Password, database index and key prefix |

Telegram sends messages through the configured bot to the selected chat. Pushover sends push notifications using its application token and user key.

Redis stores alert data in keys under the prefix (default `uptimelit`). For endpoint ID `42`, these include `uptimelit:endpoint:42:count`, `:name`, `:last_title` and `:last_message`. This provider writes keys; it does not publish Pub/Sub messages.

## Editing and deletion

**Edit** updates the provider settings. **Delete** removes the notifier and unlinks it from monitors after confirmation; it does not delete those monitors.

## Alert behavior

Outage notifications depend on the monitor's failure threshold. Recovery notifications follow a down-to-up transition. The first check after a worker starts establishes its baseline; a failed test/check or an empty channel list can explain why a notification was not delivered.

See [monitoring](usage.md) for check semantics and [API usage](api.md) for programmatic notifier management.
