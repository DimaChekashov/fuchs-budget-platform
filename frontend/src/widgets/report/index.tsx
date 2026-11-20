import { getReport, type ReportFilters } from "@/entities/report/server";
import { dateTime, money, type Wallet } from "@/entities/wallet";
import { APIError } from "@/shared/api/server";
import { dictionary, type Language } from "@/shared/i18n";

export default async function ReportSummary({
  language,
  wallets,
  filters,
}: {
  language: Language;
  wallets: Wallet[];
  filters: ReportFilters;
}) {
  const labels = dictionary(language);
  let report;
  let errorMessage: string | undefined;
  try {
    report = await getReport(language, filters);
  } catch (error) {
    if (!(error instanceof APIError)) throw error;
    errorMessage = error.status === 400 ? labels.invalidRange : labels.unavailable;
  }
  const currencies = [...new Set(wallets.map((wallet) => wallet.Currency))].sort();
  return (
    <section className="panel" aria-label={labels.summary}>
      <h2>{labels.summary}</h2>
      <form className="filters report-filters" method="get">
        <label>
          {labels.from}
          <input type="date" name="from" defaultValue={filters.from || report?.from} />
        </label>
        <label>
          {labels.to}
          <input type="date" name="to" defaultValue={filters.to || report?.to} />
        </label>
        <label>
          {labels.wallets}
          <select name="wallet_id" defaultValue={filters.wallet_id || ""}>
            <option value="">{labels.allWallets}</option>
            {wallets.map((wallet) => (
              <option key={wallet.ID} value={wallet.ID}>
                {wallet.Name}
              </option>
            ))}
          </select>
        </label>
        <label>
          {labels.currency}
          <select name="currency" defaultValue={filters.currency || ""}>
            <option value="">{labels.allCurrencies}</option>
            {currencies.map((code) => (
              <option key={code}>{code}</option>
            ))}
          </select>
        </label>
        <button className="button">{labels.filter}</button>
      </form>
      {errorMessage && <p role="status">{errorMessage}</p>}
      {report &&
        (report.totals.length ? (
          <div className="table-wrap">
            <table>
              <thead>
                <tr>
                  <th scope="col">{labels.currency}</th>
                  <th scope="col">{labels.income}</th>
                  <th scope="col">{labels.expenses}</th>
                </tr>
              </thead>
              <tbody>
                {report.totals.map((total) => (
                  <tr key={total.currency}>
                    <th scope="row">{total.currency}</th>
                    <td>{money(total.income, total.currency, language)}</td>
                    <td>{money(total.expenses, total.currency, language)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <p>{labels.noReport}</p>
        ))}
      {report?.updated_at && (
        <p className="muted report-date">
          {labels.lastUpdated}{" "}
          <time dateTime={report.updated_at}>{dateTime(report.updated_at, language)} UTC</time>
        </p>
      )}
    </section>
  );
}
