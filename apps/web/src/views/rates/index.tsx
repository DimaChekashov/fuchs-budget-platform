import { getCurrencies, getRates } from "@/entities/exchange/server";
import ActionForm from "@/features/form";
import { APIError } from "@/shared/api/server";
import { dictionary, type Language } from "@/shared/i18n";

export default async function RatesView({ language, base }: { language: Language; base: string }) {
  const t = dictionary(language);
  let data: Awaited<ReturnType<typeof getCurrencies>> | "error";
  try {
    data = await getCurrencies(language);
  } catch (error) {
    if (!(error instanceof APIError)) throw error;
    data = "error";
  }
  return <>
    <header className="page-heading"><h1>{t.rates}</h1></header>
    <div className="dashboard-grid">
      <section className="panel">
        {data === "error" ? <p role="status">{t.unavailable}</p> : <>
          <form className="filters" method="get">
            <label>{t.base}<select name="base" defaultValue={base}>
              {data.map(({ code }) => <option key={code}>{code}</option>)}
            </select></label>
            <button className="button">{t.filter}</button>
          </form>
          <QuoteTable language={language} base={base} />
        </>}
      </section>
      <section className="panel">
        <h2>{t.addCurrency}</h2>
        <ActionForm language={language} endpoint="currencies" label={t.addCurrency} convert="currency">
          <label>{t.code}<input name="code" required minLength={3} maxLength={10} pattern="[A-Za-z][A-Za-z0-9]{2,9}" /></label>
          <label>{t.precision}<select name="precision" defaultValue="2">
            {[0, 1, 2].map((precision) => <option key={precision}>{precision}</option>)}
          </select></label>
        </ActionForm>
      </section>
    </div>
  </>;
}

async function QuoteTable({ language, base }: { language: Language; base: string }) {
  const t = dictionary(language);
  let rates: Awaited<ReturnType<typeof getRates>>;
  try {
    rates = await getRates(language, base);
  } catch (error) {
    if (error instanceof APIError) return <p role="status">{t.unavailable}</p>;
    throw error;
  }
  if (!rates.length) return <p>{t.noRates}</p>;
  const date = new Intl.DateTimeFormat(language, { dateStyle: "medium", timeZone: "UTC" });
  const decimal = new Intl.NumberFormat(language, { maximumFractionDigits: 12 });
  return <div className="table-wrap"><table>
    <thead><tr><th scope="col">{t.quote}</th><th scope="col">{t.lastUpdated}</th><th scope="col">{t.rate}</th></tr></thead>
    <tbody>{rates.map((rate) => <tr key={rate.quote}>
      <th scope="row">{rate.quote}</th>
      <td><time dateTime={rate.date}>{date.format(new Date(rate.date))}</time><br /><span className="muted">{rate.source}</span></td>
      <td>{decimal.format(rate.rate as `${number}`)}</td>
    </tr>)}</tbody>
  </table></div>;
}
