import { NextRequest, NextResponse } from "next/server";
import { SESSION, session, upstream, type Service } from "@/shared/api/server";
import type { paths } from "@/shared/api/identity.generated";

type LoginResult =
  paths["/api/v1/auth/login"]["post"]["responses"][200]["content"]["application/json"];
type Context = { params: Promise<{ action: string[] }> };
const walletIDPattern = /^[0-9a-f]{8}(?:-[0-9a-f]{4}){3}-[0-9a-f]{12}$/i;
const errorResponse = (code: string, status: number) => NextResponse.json({ code }, { status });
function resolveUpstreamRoute(
  pathSegments: string[],
): { service: Service; path: string; method: string } | undefined {
  const [resource, operationOrWalletID] = pathSegments;
  if (pathSegments.length === 1 && resource === "wallets") {
    return { service: "wallet", path: "/api/v1/wallets", method: "POST" };
  }
  if (pathSegments.length === 2 && resource === "wallets") {
    if (operationOrWalletID === "deposit" || operationOrWalletID === "withdraw") {
      return { service: "wallet", path: "/api/v1/wallets/" + operationOrWalletID, method: "POST" };
    }
    if (walletIDPattern.test(operationOrWalletID)) {
      return { service: "wallet", path: "/api/v1/wallets/" + operationOrWalletID, method: "PATCH" };
    }
  }
  if (pathSegments.length === 1 && resource === "currencies") {
    return { service: "exchange", path: "/api/v1/currencies", method: "POST" };
  }
}

async function handle(request: NextRequest, context: Context) {
  if (request.headers.get("origin") !== request.nextUrl.origin) {
    return errorResponse("forbidden", 403);
  }
  const pathSegments = (await context.params).action;
  const action = pathSegments.length === 1 ? pathSegments[0] : "";
  const secure = request.nextUrl.protocol === "https:";
  if (action === "logout" && request.method === "POST") {
    const response = NextResponse.json({ ok: true });
    response.cookies.set(SESSION, "", {
      path: "/",
      maxAge: 0,
      httpOnly: true,
      sameSite: "lax",
      secure,
    });
    return response;
  }
  if (request.headers.get("content-type")?.split(";")[0] !== "application/json") {
    return errorResponse("invalid", 400);
  }
  const body = await request.text();
  if (body.length > 16384) return errorResponse("invalid", 413);

  try {
    if ((action === "login" || action === "register") && request.method === "POST") {
      const response = await upstream("identity", "/api/v1/auth/" + action, {
        method: "POST",
        headers: { "content-type": "application/json" },
        body,
      });
      if (!response.ok) {
        return errorResponse(
          action === "login" && response.status === 401
            ? "credentials"
            : action === "register" && response.status === 409
              ? "emailTaken"
              : "failed",
          response.status,
        );
      }
      const reply = NextResponse.json({ ok: true }, { status: response.status });
      if (action === "login") {
        const { access_token: token } = (await response.json()) as LoginResult;
        const { exp } = JSON.parse(
          Buffer.from(token.split(".")[1], "base64url").toString("utf8"),
        ) as { exp: number };
        if (!Number.isFinite(exp) || exp <= Date.now() / 1000) return errorResponse("failed", 502);
        reply.cookies.set(SESSION, token, {
          httpOnly: true,
          sameSite: "lax",
          secure,
          path: "/",
          expires: new Date(exp * 1000),
        });
      }
      return reply;
    }
    const route = resolveUpstreamRoute(pathSegments);
    if (!route || route.method !== request.method) return errorResponse("missing", 404);
    const token = await session();
    if (!token) return errorResponse("session", 401);
    const headers = new Headers({ "content-type": "application/json" });
    const idempotencyKey = request.headers.get("idempotency-key");
    if (idempotencyKey) headers.set("idempotency-key", idempotencyKey);
    const response = await upstream(
      route.service,
      route.path,
      { method: route.method, headers, body },
      token,
    );
    if (response.status === 401) return errorResponse("session", 401);
    if (!response.ok) {
      const problem = (await response.json().catch(() => null)) as { error?: string } | null;
      return errorResponse(
        problem?.error === "insufficient funds" ? "insufficient" : "failed",
        response.status,
      );
    }
    return NextResponse.json({ ok: true }, { status: response.status });
  } catch {
    return errorResponse("failed", 503);
  }
}

export const POST = handle;
export const PATCH = handle;
