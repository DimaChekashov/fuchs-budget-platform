import { notFound } from "next/navigation";
import { isLanguage } from "@/shared/i18n";
import RatesView from "@/views/rates";

export default async function RatesPage({ params, searchParams }: PageProps<"/[locale]/rates">) {
  const { locale } = await params;
  if (!isLanguage(locale)) notFound();
  const raw = (await searchParams).base;
  const base = typeof raw === "string" && /^[A-Z][A-Z0-9]{2,9}$/.test(raw) ? raw : "RUB";
  return <RatesView language={locale} base={base} />;
}
