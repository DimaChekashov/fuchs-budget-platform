import Link from "next/link";
import { notFound } from "next/navigation";
import { dateTime, money, type Transaction, type Wallet } from "@/entities/wallet";
import ActionForm from "@/features/form";
import { APIError, readAPI } from "@/shared/api/server";
import { dictionary, type Language } from "@/shared/i18n";
import { getCurrencies, type Currency } from "@/entities/exchange/server";

export async function NewWallet({ language }: { language: Language }) {
  const labels = dictionary(language);
  let currencies: Currency[];
  try {
    currencies = await getCurrencies(language);
  } catch (error) {
    if (error instanceof APIError) return <p role="status">{labels.unavailable}</p>;
    throw error;
  }
  return (
    <section className="form-page">
      <h1>{labels.newWallet}</h1>
      <div className="panel">
        <ActionForm
          language={language}
          label={labels.save}
          endpoint="wallets"
          destination={"/" + language}
        >
          <label>
            {labels.name}
            <input name="name" required maxLength={100} />
          </label>
          <label>
            {labels.currency}
            <select name="currency" defaultValue="RUB">
              {currencies.map(({ code }) => (
                <option key={code}>{code}</option>
              ))}
            </select>
          </label>
        </ActionForm>
      </div>
    </section>
  );
}

export async function EditWallet({
  language,
  id,
  offset,
}: {
  language: Language;
  id: string;
  offset: number;
}) {
  const labels = dictionary(language);
  const path = "/api/v1/wallets/" + id;
  let wallet: Wallet;
  let transactions: Transaction[];
  try {
    [wallet, transactions] = await Promise.all([
      readAPI<Wallet>("wallet", path, language),
      readAPI<Transaction[]>("wallet", path + "/transactions?offset=" + offset, language),
    ]);
  } catch (error) {
    if (error instanceof APIError && error.status === 404) notFound();
    if (error instanceof APIError) return <p role="status">{labels.unavailable}</p>;
    throw error;
  }
  const editWalletPath = "/" + language + "/wallets/" + id + "/edit";
  return (
    <>
      <header className="page-heading">
        <div>
          <p className="muted">{labels.editWallet}</p>
          <h1>{wallet.Name}</h1>
        </div>
        <div className="page-total">
          <span>{labels.balance}</span>
          <strong data-testid="wallet-balance">
            {money(wallet.Balance, wallet.Currency, language, wallet.CurrencyPrecision)}
          </strong>
        </div>
      </header>
      <div className="detail-grid">
        <section className="panel">
          <h2>{labels.editWallet}</h2>
          <ActionForm
            key={wallet.Name}
            language={language}
            endpoint={"wallets/" + id}
            method="PATCH"
            label={labels.save}
          >
            <label>
              {labels.name}
              <input name="name" required maxLength={100} defaultValue={wallet.Name} />
            </label>
          </ActionForm>
        </section>
        {(["deposit", "withdraw"] as const).map((operation) => (
          <section className="panel" key={operation} aria-label={labels[operation]}>
            <h2>{labels[operation]}</h2>
            <ActionForm
              language={language}
              endpoint={"wallets/" + operation}
              label={labels[operation]}
              money={{ walletID: id }}
            >
              <label>
                {labels.amount} ({wallet.Currency})
                <input name="amount" required inputMode="decimal" />
              </label>
              <label>
                {labels.description}
                <input name="description" maxLength={255} />
              </label>
            </ActionForm>
          </section>
        ))}
      </div>
      <section className="panel history">
        <h2>{labels.history}</h2>
        {transactions.length ? (
          <ul className="history-list">
            {transactions.map((transaction) => (
              <li key={transaction.ID}>
                <div>
                  <strong>
                    {transaction.Description ||
                      (transaction.Type === "deposit"
                        ? labels.deposit
                        : transaction.Type === "withdrawal"
                          ? labels.withdraw
                          : labels.activity)}
                  </strong>
                  <time dateTime={transaction.CreatedAt}>
                    {dateTime(transaction.CreatedAt, language)}
                  </time>
                </div>
                <span className={transaction.Type === "deposit" ? "credit" : ""}>
                  {transaction.Type === "deposit"
                    ? "+"
                    : transaction.Type === "withdrawal"
                      ? "−"
                      : ""}
                  {money(transaction.Amount, wallet.Currency, language, wallet.CurrencyPrecision)}
                </span>
              </li>
            ))}
          </ul>
        ) : (
          <p className="muted">{labels.emptyHistory}</p>
        )}
        <div className="pagination">
          {offset > 0 && (
            <Link href={editWalletPath + "?offset=" + Math.max(offset - 50, 0)}>
              {labels.previous}
            </Link>
          )}
          {transactions.length === 50 && (
            <Link href={editWalletPath + "?offset=" + (offset + 50)}>{labels.next}</Link>
          )}
        </div>
      </section>
    </>
  );
}
