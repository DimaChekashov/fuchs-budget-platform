import { notFound } from "next/navigation";
import { EditWallet } from "@/views/wallet";
import { isLanguage } from "@/shared/i18n";
export default async function EditWalletPage({ params, searchParams }: PageProps<"/[locale]/wallets/[id]/edit">) {
  const { locale, id } = await params;
  if (!isLanguage(locale) || !/^[0-9a-f]{8}(?:-[0-9a-f]{4}){3}-[0-9a-f]{12}$/i.test(id)) notFound();
  const rawOffset = (await searchParams).offset;
  const offset = typeof rawOffset === "string" && /^\d+$/.test(rawOffset) ? Number(rawOffset) : 0;
  return <EditWallet language={locale} id={id} offset={Number.isSafeInteger(offset) ? offset : 0} />;
}
