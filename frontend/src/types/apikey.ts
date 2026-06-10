export interface ApiKey {
  id: string
  label: string
  // True for the key backing the current session (server-computed).
  current: boolean
  lastUsedAt: string
  createdAt: string
  expiresAt: string
}
