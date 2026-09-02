export interface User {
  id: number
  username: string
  email: string
  avatar_url: string
  role: 'admin' | 'user'
  created_at: string
}

export interface AuthResponse {
  token: string
  id: number
  username: string
  email: string
  role: 'admin' | 'user'
}

export interface LoginForm {
  username: string
  password: string
}

export interface RegisterForm {
  username: string
  email: string
  password: string
}
