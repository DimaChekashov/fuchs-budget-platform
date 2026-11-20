"use client";
import Link from "next/link";
import { usePathname } from "next/navigation";
import type { Language } from "@/shared/i18n";

export default function LanguageSwitch({ language }: { language: Language }) {
  const pathname = usePathname();
  return (
    <nav className="language" aria-label={language === "ru" ? "Язык" : "Language"}>
      {(["ru", "en"] as const).map((locale) => (
        <Link
          key={locale}
          aria-current={language === locale ? "true" : undefined}
          href={pathname.replace(/^\/(ru|en)(?=\/|$)/, "/" + locale)}
        >
          {locale.toUpperCase()}
        </Link>
      ))}
    </nav>
  );
}
