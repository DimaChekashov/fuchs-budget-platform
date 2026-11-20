import type { components } from "@/shared/api/wallet.generated";
import type { Language } from "@/shared/i18n";

export type Wallet = Omit<components["schemas"]["Wallet"], "Balance"> & { Balance: string };
export type Transaction = Omit<components["schemas"]["Transaction"], "Amount"> & { Amount: string };

export function money(
  amount: number | string,
  currency: string,
  language: Language,
  precision = 2,
) {
  const hasThreeLetterCode = /^[A-Za-z]{3}$/.test(currency);
  const formattedAmount = new Intl.NumberFormat(language, {
    maximumFractionDigits: precision,
    minimumFractionDigits: precision,
    style: hasThreeLetterCode ? "currency" : "decimal",
    currency: hasThreeLetterCode ? currency.toUpperCase() : undefined,
    currencyDisplay: "code",
  }).format(amount as number | `${number}`);
  return hasThreeLetterCode ? formattedAmount : formattedAmount + " " + currency;
}

export function dateTime(time: string, language: Language) {
  return new Intl.DateTimeFormat(language, {
    dateStyle: "medium",
    timeStyle: "short",
    timeZone: "UTC",
  }).format(new Date(time));
}
