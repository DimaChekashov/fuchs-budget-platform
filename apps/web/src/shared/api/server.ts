import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import type { Language } from "@/shared/i18n";

const upstreams = {
  identity: process.env.IDENTITY_API_URL || "http://localhost:8080",
  wallet: process.env.WALLET_API_URL || "http://localhost:8081",
  exchange: process.env.EXCHANGE_API_URL || "http://localhost:8083",
  ledger: process.env.LEDGER_API_URL || "http://localhost:8082",
};
export type Service = keyof typeof upstreams;
export const SESSION = "fuchs_session";
export async function upstream(service: Service, path: string, options?: RequestInit, token?: string) {
  const headers = new Headers(options?.headers);
  if (token) headers.set("Authorization", "Bearer " + token);
  return fetch(new URL(path, upstreams[service]), {
    ...options, headers, cache: "no-store", signal: AbortSignal.timeout(5000),
  });
}

export async function session() {
  return (await cookies()).get(SESSION)?.value;
}

export class APIError extends Error {
  constructor(readonly status: number) {
    super("API response: " + status);
  }
}

export async function readAPI<T>(service: Service, path: string, language: Language): Promise<T> {
  const token = await session();
  if (!token) redirect("/" + language + "/login");
  let response: Response;
  try {
    response = await upstream(service, path, undefined, token);
  } catch {
    throw new APIError(503);
  }
  if (response.status === 401) redirect("/" + language + "/login");
  if (!response.ok) throw new APIError(response.status);
  return response.json() as Promise<T>;
}
