import { Link } from "react-router-dom";
import { useSelector } from "react-redux";
import { motion as Motion } from "framer-motion";
import {
  ArrowRight,
  BadgeCheck,
  BookOpenCheck,
  Building2,
  CalendarDays,
  CreditCard,
  FileText,
  Leaf,
  Mail,
  MapPin,
  Phone,
  ShieldCheck,
  TreePine,
  Users,
} from "lucide-react";
import {
  formatDate,
  getImageUrl,
  getOrganizationLocation,
} from "../../utils/helpers";

const serviceCards = [
  {
    icon: Users,
    title: "Digital member register",
    description:
      "Maintain household heads, assistant members, family details, photos and membership information in one place.",
  },
  {
    icon: FileText,
    title: "Resource request tracking",
    description:
      "Submit and follow requests for timber, firewood, grass and other forest resources with clear status tracking.",
  },
  {
    icon: CreditCard,
    title: "Transparent financial records",
    description:
      "Keep membership fees, sales, payments, fines, expenses and receipts organized by fiscal year.",
  },
];

const postLabels = {
  chairperson: "Chairperson",
  secretary: "Secretary",
  treasurer: "Treasurer",
  member: "Committee Member",
};

export default function Landing() {
  const { settings, committee, status } = useSelector(
    (state) => state.appSettings,
  );
  const user = useSelector((state) => state.auth.user);
  const location = getOrganizationLocation(settings);
  const logoUrl = getImageUrl(settings?.logo);
  const activeCommittee = committee
    .filter((person) => person.is_active)
    .slice(0, 4);

  return (
    <div className="min-h-screen overflow-x-hidden bg-[#f6faf7] text-slate-900">
      <header className="sticky top-0 z-40 border-b border-emerald-950/5 bg-[#f6faf7]/90 backdrop-blur-xl">
        <div className="mx-auto flex max-w-7xl items-center justify-between gap-4 px-5 py-4 lg:px-8">
          <Link to="/" className="flex min-w-0 items-center gap-3">
            {logoUrl ? (
              <img
                src={logoUrl}
                alt="Organization logo"
                className="h-11 w-11 shrink-0 rounded-xl border border-emerald-100 bg-white object-contain p-1"
              />
            ) : (
              <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-emerald-700 text-white shadow-lg shadow-emerald-700/20">
                <TreePine size={23} />
              </div>
            )}
            <div className="min-w-0">
              <p className="line-clamp-2 text-sm font-extrabold leading-5 text-emerald-950 sm:text-base">
                {settings?.name}
              </p>
              <p className="hidden text-xs font-medium text-emerald-700 sm:block">
                Community Forest User Group
              </p>
            </div>
          </Link>

          <div className="flex shrink-0 items-center gap-2">
            <a
              href="#about"
              className="hidden rounded-xl px-3 py-2 text-sm font-semibold text-slate-600 hover:bg-white hover:text-emerald-800 md:block"
            >
              About
            </a>
            <a
              href="#committee"
              className="hidden rounded-xl px-3 py-2 text-sm font-semibold text-slate-600 hover:bg-white hover:text-emerald-800 md:block"
            >
              Committee
            </a>
            <Link
              to={user ? "/dashboard" : "/login"}
              className="inline-flex items-center gap-2 rounded-xl bg-emerald-700 px-4 py-2.5 text-sm font-bold text-white shadow-lg shadow-emerald-700/20 transition hover:-translate-y-0.5 hover:bg-emerald-800"
            >
              {user ? "Open dashboard" : "Sign in"}
              <ArrowRight size={16} />
            </Link>
          </div>
        </div>
      </header>

      <main>
        <section className="relative isolate overflow-hidden">
          <div className="absolute inset-0 -z-20 bg-linear-to-br from-emerald-950 via-emerald-900 to-slate-950" />
          <div className="absolute inset-0 -z-10 opacity-30 bg-[radial-gradient(circle_at_20%_20%,rgba(52,211,153,0.45),transparent_30%),radial-gradient(circle_at_80%_70%,rgba(16,185,129,0.3),transparent_34%)]" />
          <div className="mx-auto grid min-h-170 max-w-7xl items-center gap-12 px-5 py-20 lg:grid-cols-[1.15fr_0.85fr] lg:px-8 lg:py-24">
            <Motion.div
              initial={{ opacity: 0, y: 24 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.65 }}
            >
              <div className="mb-5 inline-flex items-center gap-2 rounded-full border border-emerald-300/20 bg-emerald-300/10 px-4 py-2 text-sm font-semibold text-emerald-100 backdrop-blur-sm">
                <Leaf size={16} />
                Digital service portal
              </div>
              <h1 className="max-w-4xl text-4xl font-black leading-[1.08] tracking-tight text-white sm:text-5xl lg:text-6xl">
                {settings?.name}
              </h1>
              <p className="mt-6 max-w-2xl text-lg leading-8 text-emerald-50/80">
                A unified system for member registration, forest-resource services,
                committee operations and transparent financial record keeping.
              </p>

              <div className="mt-8 flex flex-wrap gap-3">
                <Link
                  to={user ? "/dashboard" : "/login"}
                  className="inline-flex items-center gap-2 rounded-xl bg-white px-5 py-3 font-bold text-emerald-900 shadow-xl transition hover:-translate-y-0.5 hover:bg-emerald-50"
                >
                  {user ? "Go to dashboard" : "Access member portal"}
                  <ArrowRight size={18} />
                </Link>
                <a
                  href="#about"
                  className="inline-flex items-center gap-2 rounded-xl border border-white/20 bg-white/10 px-5 py-3 font-bold text-white backdrop-blur-sm transition hover:bg-white/15"
                >
                  Learn about the group
                </a>
              </div>

              <div className="mt-10 flex flex-wrap gap-x-6 gap-y-3 text-sm text-emerald-100/80">
                <span className="inline-flex items-center gap-2">
                  <ShieldCheck size={17} className="text-emerald-300" />
                  Role-based secure access
                </span>
                <span className="inline-flex items-center gap-2">
                  <BookOpenCheck size={17} className="text-emerald-300" />
                  Organized member records
                </span>
                <span className="inline-flex items-center gap-2">
                  <BadgeCheck size={17} className="text-emerald-300" />
                  Clear service history
                </span>
              </div>
            </Motion.div>

            <Motion.div
              initial={{ opacity: 0, scale: 0.96, x: 24 }}
              animate={{ opacity: 1, scale: 1, x: 0 }}
              transition={{ duration: 0.7, delay: 0.15 }}
              className="rounded-3xl border border-white/15 bg-white/10 p-6 text-white shadow-2xl backdrop-blur-xl sm:p-8"
            >
              <div className="flex items-center gap-4 border-b border-white/10 pb-6">
                {logoUrl ? (
                  <img
                    src={logoUrl}
                    alt="Organization logo"
                    className="h-20 w-20 rounded-2xl bg-white object-contain p-2"
                  />
                ) : (
                  <div className="flex h-20 w-20 items-center justify-center rounded-2xl bg-emerald-400/20 text-emerald-100">
                    <Building2 size={36} />
                  </div>
                )}
                <div>
                  <p className="text-xs font-bold uppercase tracking-[0.18em] text-emerald-300">
                    Organization profile
                  </p>
                  <p className="mt-1 text-lg font-bold leading-6">
                    {settings?.name}
                  </p>
                </div>
              </div>

              <dl className="mt-6 space-y-5">
                {settings?.registration_no && (
                  <div className="flex gap-3">
                    <BadgeCheck className="mt-0.5 shrink-0 text-emerald-300" size={19} />
                    <div>
                      <dt className="text-xs uppercase tracking-wide text-emerald-100/60">
                        Registration number
                      </dt>
                      <dd className="mt-1 font-semibold">
                        {settings.registration_no}
                      </dd>
                    </div>
                  </div>
                )}
                {location && (
                  <div className="flex gap-3">
                    <MapPin className="mt-0.5 shrink-0 text-emerald-300" size={19} />
                    <div>
                      <dt className="text-xs uppercase tracking-wide text-emerald-100/60">
                        Office location
                      </dt>
                      <dd className="mt-1 font-semibold leading-6">{location}</dd>
                    </div>
                  </div>
                )}
                {settings?.established_date && (
                  <div className="flex gap-3">
                    <CalendarDays className="mt-0.5 shrink-0 text-emerald-300" size={19} />
                    <div>
                      <dt className="text-xs uppercase tracking-wide text-emerald-100/60">
                        Established
                      </dt>
                      <dd className="mt-1 font-semibold">
                        {formatDate(settings.established_date)}
                      </dd>
                    </div>
                  </div>
                )}
              </dl>

              {status === "failed" && (
                <p className="mt-6 rounded-xl bg-amber-300/10 p-3 text-sm text-amber-100">
                  Live organization settings are temporarily unavailable.
                </p>
              )}
            </Motion.div>
          </div>
        </section>

        <section id="about" className="mx-auto max-w-7xl px-5 py-20 lg:px-8">
          <div className="mx-auto max-w-3xl text-center">
            <p className="text-sm font-black uppercase tracking-[0.2em] text-emerald-700">
              Digital community service
            </p>
            <h2 className="mt-3 text-3xl font-black tracking-tight text-emerald-950 sm:text-4xl">
              Built around the real work of the user group
            </h2>
            <p className="mt-4 text-base leading-7 text-slate-600">
              The portal replaces scattered paper records with a structured,
              searchable system while keeping member services simple for staff
              and households.
            </p>
          </div>

          <div className="mt-12 grid gap-5 md:grid-cols-3">
            {serviceCards.map((service, index) => (
              <Motion.article
                key={service.title}
                initial={{ opacity: 0, y: 18 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true, amount: 0.2 }}
                transition={{ delay: index * 0.08 }}
                className="rounded-3xl border border-emerald-950/8 bg-white p-7 shadow-sm"
              >
                <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-emerald-100 text-emerald-800">
                  <service.icon size={23} />
                </div>
                <h3 className="mt-5 text-xl font-extrabold text-emerald-950">
                  {service.title}
                </h3>
                <p className="mt-3 text-sm leading-6 text-slate-600">
                  {service.description}
                </p>
              </Motion.article>
            ))}
          </div>
        </section>

        {activeCommittee.length > 0 && (
          <section id="committee" className="bg-emerald-950 py-20 text-white">
            <div className="mx-auto max-w-7xl px-5 lg:px-8">
              <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-end">
                <div>
                  <p className="text-sm font-black uppercase tracking-[0.2em] text-emerald-300">
                    Current leadership
                  </p>
                  <h2 className="mt-3 text-3xl font-black tracking-tight sm:text-4xl">
                    Committee representatives
                  </h2>
                </div>
                <p className="max-w-xl text-sm leading-6 text-emerald-100/70">
                  Public contact information is shown only when it has been added
                  by the organization administrator.
                </p>
              </div>

              <div className="mt-10 grid gap-5 sm:grid-cols-2 xl:grid-cols-4">
                {activeCommittee.map((person) => {
                  const photoUrl = getImageUrl(person.photo);
                  return (
                    <article
                      key={person.id}
                      className="rounded-3xl border border-white/10 bg-white/8 p-5 backdrop-blur-sm"
                    >
                      {photoUrl ? (
                        <img
                          src={photoUrl}
                          alt={person.name}
                          className="h-24 w-24 rounded-2xl object-cover"
                        />
                      ) : (
                        <div className="flex h-24 w-24 items-center justify-center rounded-2xl bg-emerald-300/15 text-3xl font-black text-emerald-200">
                          {person.name?.charAt(0)?.toUpperCase() || "?"}
                        </div>
                      )}
                      <h3 className="mt-5 text-lg font-extrabold">{person.name}</h3>
                      <p className="mt-1 text-sm font-semibold text-emerald-300">
                        {postLabels[person.post] || person.post}
                      </p>
                      {person.phone && (
                        <a
                          href={`tel:${person.phone}`}
                          className="mt-4 inline-flex items-center gap-2 text-sm text-emerald-100/75 hover:text-white"
                        >
                          <Phone size={15} /> {person.phone}
                        </a>
                      )}
                    </article>
                  );
                })}
              </div>
            </div>
          </section>
        )}

        <section className="mx-auto max-w-7xl px-5 py-20 lg:px-8">
          <div className="grid gap-6 rounded-3xl border border-emerald-950/8 bg-white p-7 shadow-sm md:grid-cols-[1fr_auto] md:items-center lg:p-10">
            <div>
              <h2 className="text-2xl font-black text-emerald-950 sm:text-3xl">
                Contact the user-group office
              </h2>
              <p className="mt-2 text-sm leading-6 text-slate-600">
                For membership corrections, service questions or account access,
                contact the office using the details maintained in organization settings.
              </p>
              <div className="mt-5 flex flex-wrap gap-4 text-sm font-semibold text-slate-700">
                {settings?.contact_phone && (
                  <a href={`tel:${settings.contact_phone}`} className="inline-flex items-center gap-2 hover:text-emerald-700">
                    <Phone size={17} className="text-emerald-700" />
                    {settings.contact_phone}
                  </a>
                )}
                {settings?.contact_email && (
                  <a href={`mailto:${settings.contact_email}`} className="inline-flex items-center gap-2 hover:text-emerald-700">
                    <Mail size={17} className="text-emerald-700" />
                    {settings.contact_email}
                  </a>
                )}
                {location && (
                  <span className="inline-flex items-center gap-2">
                    <MapPin size={17} className="text-emerald-700" />
                    {location}
                  </span>
                )}
              </div>
            </div>
            <Link
              to={user ? "/dashboard" : "/login"}
              className="inline-flex items-center justify-center gap-2 rounded-xl bg-emerald-700 px-5 py-3 font-bold text-white transition hover:bg-emerald-800"
            >
              {user ? "Open dashboard" : "Sign in to portal"}
              <ArrowRight size={17} />
            </Link>
          </div>
        </section>
      </main>

      <footer className="border-t border-emerald-950/8 bg-white">
        <div className="mx-auto flex max-w-7xl flex-col justify-between gap-3 px-5 py-7 text-sm text-slate-500 sm:flex-row lg:px-8">
          <p>
            © {new Date().getFullYear()} {settings?.name}
          </p>
          <p>Community Forestry Management System</p>
        </div>
      </footer>
    </div>
  );
}
