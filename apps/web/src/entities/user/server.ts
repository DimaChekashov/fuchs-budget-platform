import type { paths } from "@/shared/api/identity.generated";
import { readAPI } from "@/shared/api/server";
import type { Language } from "@/shared/i18n";

type User = paths["/api/v1/users/me"]["get"]["responses"][200]["content"]["application/json"];
export function apiProfile(language: Language) {
  return readAPI<User>("identity", "/api/v1/users/me", language);
}
