import { getCurrencies, getRates } from "@/entities/exchange/server";
import ActionForm from "@/features/form";
import { APIError } from "@/shared/api/server";
import { dictionary, type Language } from "@/shared/i18n";

export default async function RatesView({ language, base }: { language: Language; base: string }) {
  const labels = dictionary(language);
  let currencies: Awaited<ReturnType<typeof getCurrencies>> | "error";
  try {
    currencies = await getCurrencies(language);
  } catch (error) {
    if (!(error instanceof APIError)) throw error;
    currencies = "error";
  }
  return (
    <>
      <header className="page-heading">
        <h1>{labels.rates}</h1>
      </header>
      <div className="dashboard-grid">
        <section className="panel">
          {currencies === "error" ? (
            <p role="status">{labels.unavailable}</p>
          ) : (
            <>
              <form className="filters" method="get">
                <label>
                  {labels.base}
                  <select name="base" defaultValue={base}>
                    {currencies.map(({ code }) => (
                      <option key={code}>{code}</option>
                    ))}
                  </select>
                </label>
                <button className="button">{labels.filter}</button>
              </form>
              <QuoteTable language={language} base={base} />
            </>
          )}
        </section>
        <section className="panel">
          <h2>{labels.addCurrency}</h2>
          <ActionForm
            language={language}
            endpoint="currencies"
            label={labels.addCurrency}
            convert="currency"
          >
            <label>
              {labels.code}
              <input
                name="code"
                required
                minLength={3}
                maxLength={10}
                pattern="[A-Za-z][A-Za-z0-9]{2,9}"
              />
            </label>
            <label>
              {labels.precision}
              <select name="precision" defaultValue="2">
                {[0, 1, 2].map((precision) => (
                  <option key={precision}>{precision}</option>
                ))}
              </select>
            </label>
          </ActionForm>
        </section>
      </div>
    </>
  );
}

async function QuoteTable({ language, base }: { language: Language; base: string }) {
  const labels = dictionary(language);
  let rates: Awaited<ReturnType<typeof getRates>>;
  try {
    rates = await getRates(language, base);
  } catch (error) {
    if (error instanceof APIError) return <p role="status">{labels.unavailable}</p>;
    throw error;
  }
  if (!rates.length) return <p>{labels.noRates}</p>;
  const dateFormatter = new Intl.DateTimeFormat(language, { dateStyle: "medium", timeZone: "UTC" });
  const rateFormatter = new Intl.NumberFormat(language, { maximumFractionDigits: 12 });
  return (
    <div className="table-wrap">
      <table>
        <thead>
          <tr>
            <th scope="col">{labels.quote}</th>
            <th scope="col">{labels.lastUpdated}</th>
            <th scope="col">{labels.rate}</th>
          </tr>
        </thead>
        <tbody>
          {rates.map((rate) => (
            <tr key={rate.quote}>
              <th scope="row">{rate.quote}</th>
              <td>
                <time dateTime={rate.date}>{dateFormatter.format(new Date(rate.date))}</time>
                <br />
                <span className="muted">{rate.source}</span>
              </td>
              <td>{rateFormatter.format(rate.rate as `${number}`)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
