const connectionErrors: Record<string, string> = {
  permission_denied: 'Permission denied',
  ping_error: 'Ping error',
  timeout: 'Timeout',
  connection_refused: 'Connection refused',
  dns_error: 'DNS error',
  network_unreachable: 'Network unreachable',
  connection_reset: 'Connection reset',
  connection_error: 'Connection error',
}

export function networkStatus(isUp: boolean | null | undefined, errorCode: string | undefined, protocol: 'tcp' | 'ping'): string {
  if (isUp == null) return '—'
  if (isUp) return protocol === 'ping' ? 'Reply received' : 'Connected'
  return connectionErrors[errorCode ?? ''] ?? (protocol === 'ping' ? 'No reply' : 'Failed')
}
