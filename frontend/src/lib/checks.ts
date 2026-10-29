const connectionErrors: Record<string, string> = {
  timeout: 'Timeout',
  connection_refused: 'Connection refused',
  dns_error: 'DNS error',
  network_unreachable: 'Network unreachable',
  connection_reset: 'Connection reset',
  connection_error: 'Connection error',
}

export function connectionStatus(isUp: boolean | null | undefined, errorCode?: string): string {
  if (isUp == null) return '—'
  if (isUp) return 'Connected'
  return connectionErrors[errorCode ?? ''] ?? 'Failed'
}
