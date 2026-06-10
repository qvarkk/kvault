export type StopwordSource = 'user' | 'default'

export interface Stopword {
  word: string
  source: StopwordSource
  isEnabled: boolean
  updatedAt: string
}

export interface StopwordCreatePayload {
  word: string
}