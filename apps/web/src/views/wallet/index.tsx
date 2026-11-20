import Link from "next/link";
import { notFound } from "next/navigation";
import { dateTime, money, type Transaction, type Wallet } from "@/entities/wallet";
import ActionForm from "@/features/form";
import { APIError, readAPI } from "@/shared/api/server";
import { dictionary, type Language } from "@/shared/i18n";

export function NewWallet({ language }: { language: Language }) {
  const t = dictionary(language);
  return <section className="form-page">
    <h1>{t.newWallet}</h1>
    <div className="panel">
      <ActionForm language={language} label={t.save} endpoint="wallets" destination={"/" + language}>
        <label>{t.name}<input name="name" required maxLength={100} /></label>
        <label>{t.currency}<select name="currency" defaultValue="RUB">
          {["RUB", "USD", "EUR", "CNY"].map((code) => <option key={code}>{code}</option>)}
        </select></label>
      </ActionForm>
    </div>
  </section>;
}

export async function EditWallet({ language, id, offset }: { language: Language; id: string; offset: number }) {
  const t = dictionary(language);
  const path = "/api/v1/wallets/" + id;
  let wallet: Wallet;
  let history: Transaction[];
  try {
    [wallet, history] = await Promise.all([
      readAPI<Wallet>("wallet", path, language),
      readAPI<Transaction[]>("wallet", path + "/transactions?offset=" + offset, language),
    ]);
  } catch (error) {
    if (error instanceof APIError && error.status === 404) notFound();
    if (error instanceof APIError) return <p role="status">{t.unavailable}</p>;
    throw error;
  }
  const self = "/" + language + "/wallets/" + id + "/edit";
  return <>
    <header className="page-heading"><div><p className="muted">{t.editWallet}</p><h1>{wallet.Name}</h1></div>
      <div className="page-total"><span>{t.balance}</span><strong data-testid="wallet-balance">{money(wallet.Balance, wallet.Currency, language)}</strong></div>
    </header>
    <div className="detail-grid">
      <section className="panel">
        <h2>{t.editWallet}</h2>
        <ActionForm key={wallet.Name} language={language} endpoint={"wallets/" + id} method="PATCH" label={t.save}>
          <label>{t.name}<input name="name" required maxLength={100} defaultValue={wallet.Name} /></label>
        </ActionForm>
      </section>
      {(["deposit", "withdraw"] as const).map((operation) => (
        <section className="panel" key={operation} aria-label={t[operation]}>
          <h2>{t[operation]}</h2>
          <ActionForm language={language} endpoint={"wallets/" + operation} label={t[operation]} money={{ walletID: id }}>
            <label>{t.amount} ({wallet.Currency})<input name="amount" required inputMode="decimal" /></label>
            <label>{t.description}<input name="description" maxLength={255} /></label>
          </ActionForm>
        </section>
      ))}
    </div>
    <section className="panel history">
      <h2>{t.history}</h2>
      {history.length ? <ul className="history-list">{history.map((transaction) => (
        <li key={transaction.ID}><div><strong>{transaction.Description ||
          (transaction.Type === "deposit" ? t.deposit : transaction.Type === "withdrawal" ? t.withdraw : t.activity)}</strong>
          <time dateTime={transaction.CreatedAt}>{dateTime(transaction.CreatedAt, language)}</time></div>
          <span className={transaction.Type === "deposit" ? "credit" : ""}>
            {transaction.Type === "deposit" ? "+" : transaction.Type === "withdrawal" ? "−" : ""}
            {money(transaction.Amount, wallet.Currency, language)}</span></li>
      ))}</ul> : <p className="muted">{t.emptyHistory}</p>}
      <div className="pagination">
        {offset > 0 && <Link href={self + "?offset=" + Math.max(offset - 50, 0)}>{t.previous}</Link>}
        {history.length === 50 && <Link href={self + "?offset=" + (offset + 50)}>{t.next}</Link>}
      </div>
    </section>
  </>;
}
