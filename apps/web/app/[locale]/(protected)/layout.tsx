import Link from "next/link";
import { notFound, redirect } from "next/navigation";
import { apiProfile } from "@/entities/user/server";
import Brand from "@/shared/ui/brand";
import ActionForm from "@/features/form";
import { dictionary, isLanguage } from "@/shared/i18n";

export default async function PrivateLayout({ children, params }: LayoutProps<"/[locale]">) {
  const { locale } = await params;
  if (!isLanguage(locale)) notFound();
  const profile = await apiProfile(locale);
  if (!profile) redirect("/" + locale + "/login");
  const t = dictionary(locale);
  const root = "/" + locale;
  return <div className="shell">
    <aside className="sidebar">
      <Link className="home-link" href={root}><Brand /></Link>
      <nav className="menu" aria-label={t.dashboard}><Link href={root}>{t.dashboard}</Link></nav>
      <div className="account"><span title={profile.email}>{profile.email}</span>
        <ActionForm language={locale} label={t.logout} endpoint="logout" destination={root + "/login"} />
      </div>
    </aside>
    <main className="workspace" id="content">{children}</main>
  </div>;
}
