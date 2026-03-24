// Centralized API client for backend communication.
// All fetch/auth logic should live here, not in pages or components.

import { authHeaders, setToken } from "./auth";
import type {
  CreateUserRequest,
  LoginRequest,
  LoginResponse,
  Profile,
  UserSummary,
} from "./types";

const BASE_URL = "/api";

export async function apiFetch<T>(
  path: string,
  options?: RequestInit,
): Promise<T> {
  const res = await fetch(`${BASE_URL}${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...authHeaders(),
      ...options?.headers,
    },
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || `API error: ${res.status}`);
  }
  return res.json() as Promise<T>;
}

export async function apiFetchVoid(
  path: string,
  options?: RequestInit,
): Promise<void> {
  const res = await fetch(`${BASE_URL}${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...authHeaders(),
      ...options?.headers,
    },
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || `API error: ${res.status}`);
  }
}

export async function login(email: string, password: string): Promise<void> {
  const data = await apiFetch<LoginResponse>("/login", {
    method: "POST",
    body: JSON.stringify({ email, password } satisfies LoginRequest),
  });
  setToken(data.token);
}

export async function fetchProfile(): Promise<Profile> {
  return apiFetch<Profile>("/profile");
}

export async function fetchUsers(): Promise<UserSummary[]> {
  return apiFetch<UserSummary[]>("/admin/users");
}

export async function createUser(req: CreateUserRequest): Promise<void> {
  await apiFetchVoid("/admin/users", {
    method: "POST",
    body: JSON.stringify(req),
  });
}
