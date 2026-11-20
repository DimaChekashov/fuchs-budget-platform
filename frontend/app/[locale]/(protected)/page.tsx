import { notFound } from "next/navigation";
import Dashboard from "@/views/dashboard";
import { isLanguage } from "@/shared/i18n";
export default async function DashboardPage({ params, searchParams }: PageProps<"/[locale]">) {
  const { locale } = await params;
  if (!isLanguage(locale)) notFound();
  const query = await searchParams;
  const getFilterValue = (key: string) => (typeof query[key] === "string" ? query[key] : undefined);
  return (
    <Dashboard
      language={locale}
      filters={{
        from: getFilterValue("from"),
        to: getFilterValue("to"),
        wallet_id: getFilterValue("wallet_id"),
        currency: getFilterValue("currency"),
      }}
    />
  );
}
