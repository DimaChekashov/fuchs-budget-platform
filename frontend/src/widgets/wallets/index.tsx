import Link from "next/link";
import { money, type Wallet } from "@/entities/wallet";
import { dictionary, type Language } from "@/shared/i18n";

export default function Wallets({ wallets, language }: { wallets: Wallet[]; language: Language }) {
  const labels = dictionary(language);
  if (!wallets.length) {
    return (
      <section className="panel empty">
        <p>{labels.emptyWallets}</p>
      </section>
    );
  }
  return (
    <div className="wallet-grid">
      {wallets.map((wallet) => (
        <article className="wallet-card" key={wallet.ID}>
          <div className="wallet-top">
            <span className="wallet-name">{wallet.Name}</span>
            <span className="currency">{wallet.Currency}</span>
          </div>
          <span className="balance-label">{labels.balance}</span>
          <strong className="balance">
            {money(wallet.Balance, wallet.Currency, language, wallet.CurrencyPrecision)}
          </strong>
          <Link className="text-link" href={"/" + language + "/wallets/" + wallet.ID + "/edit"}>
            {labels.editWallet}
          </Link>
        </article>
      ))}
    </div>
  );
}
