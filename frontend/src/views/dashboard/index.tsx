import Link from "next/link";
import type { Wallet } from "@/entities/wallet";
import { APIError, readAPI } from "@/shared/api/server";
import Wallets from "@/widgets/wallets";
import ReportSummary from "@/widgets/report";
import type { ReportFilters } from "@/entities/report/server";
import { dictionary, type Language } from "@/shared/i18n";

export default async function Dashboard({
  language,
  filters,
}: {
  language: Language;
  filters: ReportFilters;
}) {
  const labels = dictionary(language);
  let wallets: Wallet[] | "error";
  try {
    wallets = (await readAPI<Wallet[] | null>("wallet", "/api/v1/wallets", language)) || [];
  } catch (error) {
    if (!(error instanceof APIError)) throw error;
    wallets = "error";
  }
  return (
    <>
      <header className="page-heading">
        <h1>{labels.heading}</h1>
        <Link className="button" href={"/" + language + "/wallets/new"}>
          {labels.newWallet}
        </Link>
      </header>
      <div className="dashboard-grid">
        {wallets === "error" ? (
          <p role="status">{labels.unavailable}</p>
        ) : (
          <Wallets language={language} wallets={wallets} />
        )}
        <ReportSummary
          language={language}
          filters={filters}
          wallets={wallets === "error" ? [] : wallets}
        />
      </div>
    </>
  );
}
