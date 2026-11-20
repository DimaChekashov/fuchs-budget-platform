import { notFound } from "next/navigation";
import AuthView from "@/views/auth";
import { isLanguage } from "@/shared/i18n";

export default async function LoginPage({ params, searchParams }: PageProps<"/[locale]/login">) {
  const { locale } = await params;
  if (!isLanguage(locale)) notFound();
  return <AuthView language={locale} registered={(await searchParams).registered === "1"} />;
}
