export interface User {
  id: string
  username: string
  // Only present on register / login / refresh responses (issued once).
  // Absent on GET /auth/me — the key is stored hashed and never returned there.
  apiKey?: string
  updatedAt: string
  createdAt: string
}

export interface LoginPayload {
  username: string
  password: string
}

export type SignupPayload = LoginPayload