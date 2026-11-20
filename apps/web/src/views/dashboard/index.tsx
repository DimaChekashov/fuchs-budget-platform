import Link from "next/link";
import type { Wallet } from "@/entities/wallet";
import { APIError, readAPI } from "@/shared/api/server";
import Wallets from "@/widgets/wallets";
import { dictionary, type Language } from "@/shared/i18n";

export default async function Dashboard({ language }: { language: Language }) {
  const t = dictionary(language);
  let wallets: Wallet[] | "error";
  try {
    wallets = (await readAPI<Wallet[] | null>("wallet", "/api/v1/wallets", language)) || [];
  } catch (error) {
    if (!(error instanceof APIError)) throw error;
    wallets = "error";
  }
  return <>
    <header className="page-heading"><h1>{t.heading}</h1>
      <Link className="button" href={"/" + language + "/wallets/new"}>{t.newWallet}</Link>
    </header>
    {wallets === "error" ? <p role="status">{t.unavailable}</p> : <Wallets language={language} wallets={wallets} />}
  </>;
}
