// Centralized API client for backend communication.
// All fetch/auth logic should live here, not in pages or components.

import { authHeaders, setToken } from "./auth";
import type {
  CreateUserRequest,
  LoginRequest,
  LoginResponse,
  MeResponse,
  Profile,
  ReportSummary,
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

export async function apiFetchText(
  path: string,
  options?: RequestInit,
): Promise<string> {
  const res = await fetch(`${BASE_URL}${path}`, {
    ...options,
    headers: {
      ...authHeaders(),
      ...options?.headers,
    },
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || `API error: ${res.status}`);
  }
  return res.text();
}

export async function login(email: string, password: string): Promise<void> {
  const data = await apiFetch<LoginResponse>("/login", {
    method: "POST",
    body: JSON.stringify({ email, password } satisfies LoginRequest),
  });
  setToken(data.token);
}

export async function loginWithMagicLink(magicLink: string): Promise<void> {
  const data = await apiFetch<{ access_token: string }>(
    `/login/${encodeURIComponent(magicLink)}`,
    { method: "POST" },
  );
  setToken(data.access_token);
}

export async function fetchMe(): Promise<MeResponse> {
  return apiFetch<MeResponse>("/me");
}

export async function fetchProfile(): Promise<Profile> {
  return apiFetch<Profile>("/profile");
}

export async function fetchOrders(empireNo: number): Promise<string> {
  return apiFetchText(`/${empireNo}/orders`);
}

export async function submitOrders(
  empireNo: number,
  orders: string,
): Promise<void> {
  await apiFetchVoid(`/${empireNo}/orders`, {
    method: "POST",
    body: orders,
    headers: { "Content-Type": "text/plain" },
  });
}

export async function fetchReports(empireNo: number): Promise<ReportSummary[]> {
  return apiFetch<ReportSummary[]>(`/${empireNo}/reports`);
}

export async function fetchReportByLink(link: string): Promise<unknown> {
  const res = await fetch(link, { headers: authHeaders() });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || `API error: ${res.status}`);
  }
  return res.json();
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
