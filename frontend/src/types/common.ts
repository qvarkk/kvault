export interface Listed<T> {
  data: T[]
}

export interface Paginated<T> extends Listed<T> {
  page: number
  pageSize: number
  total: number
}
