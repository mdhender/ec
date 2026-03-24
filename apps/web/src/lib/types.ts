// Shared request/response types for the API client.

export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  token: string;
}

export interface MeResponse {
  empire: number;
  authenticated: boolean;
  name: string;
}

export interface ReportSummary {
  turn_year: number;
  turn_quarter: number;
  link: string;
}

export interface Empire {
  id: string;
  name: string;
}

export interface Profile {
  handle: string;
  email: string;
  role: string;
  empire: Empire;
}

export interface UserSummary {
  user_id: string;
  username: string;
  email: string;
  role: string;
}

export interface CreateUserRequest {
  username: string;
  email: string;
  password: string;
  role: string;
}
