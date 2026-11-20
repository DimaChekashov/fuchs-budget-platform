"use client";
import { useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { dictionary, type Language } from "@/shared/i18n";

type Props = {
  children?: React.ReactNode;
  language: Language;
  endpoint: string;
  label: string;
  destination?: string;
  method?: "POST" | "PATCH";
  money?: { walletID: string };
  convert?: "currency";
};

export default function ActionForm({ children, language, endpoint, label, destination, method = "POST", money, convert }: Props) {
  const t = dictionary(language);
  const router = useRouter();
  const [pending, setPending] = useState(false);
  const inFlight = useRef(false);
  const [error, setError] = useState("");
  const retry = useRef<{ body: string; key: string } | null>(null);
  async function submit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (inFlight.current) return;
    const form = event.currentTarget;
    const values = Object.fromEntries(new FormData(form));
    let body = JSON.stringify(values);
    if (money) {
      const amount = String(values.amount).trim().replace(",", ".");
      if (!/^(?:0|[1-9]\d*)(?:\.\d{1,2})?$/.test(amount)) {
        setError(t.requestFailed);
        return;
      }
      body = '{"wallet_id":' + JSON.stringify(money.walletID) + ',"description":' +
        JSON.stringify(values.description || "") + ',"amount":' + amount + "}";
    }
    if (convert === "currency") body = JSON.stringify({ ...values, precision: Number(values.precision) });
    if (money && retry.current?.body !== body) retry.current = { body, key: crypto.randomUUID() };
    setError("");
    inFlight.current = true;
    setPending(true);
    try {
      const headers = new Headers({ "content-type": "application/json" });
      if (money && retry.current) headers.set("idempotency-key", retry.current.key);
      const response = await fetch("/api/" + endpoint, { method, headers, body });
      const result = await response.json();
      if (result.code === "session") {
        router.replace("/" + language + "/login");
        return;
      }
      if (!response.ok) {
        setError(result.code === "credentials" ? t.invalidCredentials : result.code === "emailTaken" ? t.emailTaken :
          result.code === "insufficient" ? t.insufficient : t.requestFailed);
        return;
      }
      retry.current = null;
      form.reset();
      if (destination) router.replace(destination);
      router.refresh();
    } catch {
      setError(t.requestFailed);
    } finally {
      inFlight.current = false;
      setPending(false);
    }
  }
  return <form className="form" onSubmit={submit}>
    {children}
    {error && <p className="error" role="alert">{error}</p>}
    <button className="button" disabled={pending} aria-busy={pending}>{label}</button>
  </form>;
}
