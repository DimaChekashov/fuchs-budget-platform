import Link from "next/link";
import { money, type Wallet } from "@/entities/wallet";
import { dictionary, type Language } from "@/shared/i18n";

export default function Wallets({ wallets, language }: { wallets: Wallet[]; language: Language }) {
  const t = dictionary(language);
  if (!wallets.length) return <section className="panel empty"><p>{t.emptyWallets}</p></section>;
  return <div className="wallet-grid">
    {wallets.map((wallet) => (
      <article className="wallet-card" key={wallet.ID}>
        <div className="wallet-top"><span className="wallet-name">{wallet.Name}</span><span className="currency">{wallet.Currency}</span></div>
        <span className="balance-label">{t.balance}</span>
        <strong className="balance">{money(wallet.Balance, wallet.Currency, language)}</strong>
        <Link className="text-link" href={"/" + language + "/wallets/" + wallet.ID + "/edit"}>{t.editWallet}</Link>
      </article>
    ))}
  </div>;
}
