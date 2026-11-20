import { notFound } from "next/navigation";
import AuthView from "@/views/auth";
import { isLanguage } from "@/shared/i18n";

export default async function RegisterPage({ params }: PageProps<"/[locale]/register">) {
  const { locale } = await params;
  if (!isLanguage(locale)) notFound();
  return <AuthView language={locale} register />;
}
