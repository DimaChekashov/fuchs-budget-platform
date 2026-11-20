import type { components } from "@/shared/api/exchange.generated";
import { readAPI } from "@/shared/api/server";
import type { Language } from "@/shared/i18n";

export type Currency = components["schemas"]["Currency"];
export type Rate = components["schemas"]["Rate"];

export const getCurrencies = (language: Language) =>
  readAPI<Currency[]>("exchange", "/api/v1/currencies", language);
export const getRates = (language: Language, base: string) =>
  readAPI<Rate[]>("exchange", "/api/v1/rates?base=" + encodeURIComponent(base), language);
