import { notFound } from "next/navigation";
import Dashboard from "@/views/dashboard";
import { isLanguage } from "@/shared/i18n";
export default async function DashboardPage({ params }: PageProps<"/[locale]">) {
  const { locale } = await params;
  if (!isLanguage(locale)) notFound();
  return <Dashboard language={locale} />;
}
