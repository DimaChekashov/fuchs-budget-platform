import { notFound } from "next/navigation";
import { isLanguage } from "@/shared/i18n";
import LanguageSwitch from "@/features/language";
import "../styles.css";

export default async function LocaleLayout({ children, params }: LayoutProps<"/[locale]">) {
  const { locale } = await params;
  if (!isLanguage(locale)) notFound();
  return <html lang={locale}><body>
    <a className="skip" href="#content">{locale === "ru" ? "К содержанию" : "Skip to content"}</a>
    <LanguageSwitch language={locale} />
    {children}
  </body></html>;
}
