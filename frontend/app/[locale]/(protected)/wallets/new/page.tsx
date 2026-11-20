import { notFound } from "next/navigation";
import { NewWallet } from "@/views/wallet";
import { isLanguage } from "@/shared/i18n";
export default async function NewWalletPage({ params }: PageProps<"/[locale]/wallets/new">) {
  const { locale } = await params;
  if (!isLanguage(locale)) notFound();
  return <NewWallet language={locale} />;
}
