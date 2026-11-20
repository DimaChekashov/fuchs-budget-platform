import Link from "next/link";
import Brand from "@/shared/ui/brand";
import { dictionary, type Language } from "@/shared/i18n";
import ActionForm from "@/features/form";

export default function AuthView({ language, register = false, registered = false }: {
  language: Language; register?: boolean; registered?: boolean;
}) {
  const t = dictionary(language);
  const root = "/" + language;
  const heading = register ? t.register : t.login;
  return <main id="content" className="auth-layout">
    <div className="auth-aside"><Brand /><div className="auth-art" aria-hidden="true"><i /><i /><i /></div></div>
    <section className="auth-card">
      <h1>{heading}</h1>
      <p className="muted">{register ? t.registerHelp : t.signedInHelp}</p>
      {registered && <p role="status">{t.registerDone}</p>}
      <ActionForm language={language} label={heading} endpoint={register ? "register" : "login"}
        destination={register ? root + "/login?registered=1" : root}>
        <label>{t.email}<input required name="email" type="email" autoComplete="email" /></label>
        <label>{t.password}<input required name="password" type="password"
          autoComplete={register ? "new-password" : "current-password"} /></label>
      </ActionForm>
      <Link className="text-link" href={root + (register ? "/login" : "/register")}>{register ? t.login : t.register}</Link>
    </section>
  </main>;
}
