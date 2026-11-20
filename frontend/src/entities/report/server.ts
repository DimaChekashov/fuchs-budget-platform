import type { components } from "@/shared/api/ledger.generated";
import { readAPI } from "@/shared/api/server";
import type { Language } from "@/shared/i18n";

export type Report = components["schemas"]["Report"];
export type ReportFilters = { from?: string; to?: string; wallet_id?: string; currency?: string };

export function getReport(language: Language, filters: ReportFilters) {
  const query = new URLSearchParams();
  for (const [field, value] of Object.entries(filters)) {
    if (value) query.set(field, value);
  }
  return readAPI<Report>("ledger", "/api/v1/reports/summary?" + query, language);
}
