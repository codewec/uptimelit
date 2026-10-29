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

An outage notification is triggered once when consecutive failures reach the monitor's threshold. Further failures in the same outage do not send duplicates. Recovery is triggered only after that outage notification; a brief failure below the threshold produces neither an outage nor a recovery notification.

For the default threshold of three, one or two failed checks followed by success are silent. Three failed checks trigger an outage alert; the next successful check triggers recovery and resets the failure count. A threshold of one alerts on the first failed check, including at startup. Delivery still requires attached notification channels and valid provider settings.

See [monitoring](usage.md) for check semantics and [API usage](api.md) for programmatic notifier management.
