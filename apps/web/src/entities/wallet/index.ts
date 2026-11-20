import type { components } from "@/shared/api/wallet.generated";
import type { Language } from "@/shared/i18n";

export type Wallet = components["schemas"]["Wallet"];
export type Transaction = components["schemas"]["Transaction"];

export function money(amount: number | string, currency: string, language: Language) {
  return new Intl.NumberFormat(language, {
    maximumFractionDigits: 2,
    minimumFractionDigits: 2,
    style: /^[A-Za-z]{3}$/.test(currency) ? "currency" : "decimal",
    currency: /^[A-Za-z]{3}$/.test(currency) ? currency.toUpperCase() : undefined,
    currencyDisplay: "code",
  }).format(amount as number) + (/^[A-Za-z]{3}$/.test(currency) ? "" : " " + currency);
}

export function dateTime(time: string, language: Language) {
  return new Intl.DateTimeFormat(language, { dateStyle: "medium", timeStyle: "short", timeZone: "UTC" }).format(new Date(time));
}
