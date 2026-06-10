export interface Tag {
  id: string
  userId: string
  name: string
  source: "manual" | "auto"
  createdAt: string
  updatedAt: string
  itemCount: number
}

export interface TagRef {
  id: string
  name: string
  source: "manual" | "auto"
}

export interface TagCreatePayload {
  name: string
}

export type TagUpdatePayload = TagCreatePayload
