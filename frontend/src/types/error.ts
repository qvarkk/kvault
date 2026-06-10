export interface ApiError {
  title: string
  detail: string
  type: string
  instance: string
  status: number
  validation?: ValidationError[]
}

export interface ValidationError {
  field: string
  message: string
}