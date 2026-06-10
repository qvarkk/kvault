import type { TagRef } from "./tag"

export type NoteType = "text" | "url"

export type UrlStatus = "pending" | "ready" | "error"

export interface UrlMetadata {
  title: string
  description: string
  image: string
  siteName: string
}

export interface Note {
  id: string
  userId: string
  title: string
  content: string
  type: NoteType
  sourceUrl?: string
  urlStatus?: UrlStatus
  urlMetadata?: UrlMetadata
  extractedContent?: string
  tags: TagRef[]
  updatedAt: string
  createdAt: string
}

export interface NoteCreatePayload {
  title: string
  content?: string
  type: NoteType
  sourceUrl?: string
}

export interface NoteUpdatePayload {
  title?: string
  content?: string
}

export interface BindTagPayload {
  tag_id: string
}
