export type FileStatus = "uploading" | "processing" | "ready" | "error"

export interface File {
  id: string
  userId: string
  originalName: string
  s3Key: string
  size: number
  mimeType: string
  status: FileStatus
  textContent?: string
  createdAt: string
  updatedAt: string
}

export interface PresignedUrl {
  url: string
  filename: string
  mimeType: string
  size: number
  expiresAt: string
}

export interface UploadItem {
  id: string
  name: string
  progress: number
  status: "uploading" | "done" | "error" | "cancelled"
  error?: string
}
