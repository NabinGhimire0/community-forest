import { useCallback, useEffect, useState } from "react";
import { useSelector } from "react-redux";
import {
  CalendarDays,
  CircleUserRound,
  Hash,
  Home,
  Mail,
  MapPin,
  Phone,
  ShieldCheck,
  UserRound,
  UsersRound,
} from "lucide-react";
import { api } from "../../services/api";
import { Card, CardContent, CardHeader } from "../../components/ui/Card";
import Badge from "../../components/ui/Badge";
import LoadingSpinner from "../../components/common/LoadingSpinner";
import AccountSecurity from "../../components/security/AccountSecurity";
import {
  formatDate,
  getImageUrl,
  getRoleLabel,
} from "../../utils/helpers";

function Detail({ icon: _Icon, label, value }) {
  return (
    <div className="flex items-start gap-3 rounded-xl border border-slate-200/80 p-4 dark:border-white/10">
      <div className="rounded-xl bg-emerald-50 p-2 text-emerald-700 dark:bg-emerald-950/50 dark:text-emerald-300">
        <_Icon size={18} />
      </div>
      <div className="min-w-0">
        <p className="text-xs font-bold uppercase tracking-wide text-slate-400">
          {label}
        </p>
        <p className="mt-1 break-words font-semibold text-slate-800 dark:text-slate-100">
          {value || "Not recorded"}
        </p>
      </div>
    </div>
  );
}

export default function Profile() {
  const account = useSelector((state) => state.auth.user);
  const [member, setMember] = useState(account?.member || null);
  const [loading, setLoading] = useState(account?.role === "member");
  const [error, setError] = useState("");

  const loadMember = useCallback(async () => {
    if (account?.role !== "member") return;
    setLoading(true);
    setError("");
    try {
      const response = await api.getMyMemberProfile();
      if (response.success) setMember(response.data);
    } catch (err) {
      setError(err.message || "Could not load member details.");
    } finally {
      setLoading(false);
    }
  }, [account?.role]);

  useEffect(() => {
    loadMember();
  }, [loadMember]);

  if (loading) return <LoadingSpinner text="Loading your profile..." />;

  const photo = getImageUrl(member?.photo);
  const assistantPhoto = getImageUrl(member?.assistant_photo);
  const familyMembers = member?.family_members || [];

  return (
    <div className="space-y-6">
      <section className="overflow-hidden rounded-3xl bg-linear-to-br from-emerald-700 via-teal-800 to-slate-900 p-6 text-white shadow-xl sm:p-8">
        <div className="flex flex-col gap-5 sm:flex-row sm:items-center">
          {photo ? (
            <img
              src={photo}
              alt={member?.name || account?.name}
              className="h-28 w-28 rounded-3xl border-4 border-white/25 object-cover"
            />
          ) : (
            <div className="flex h-28 w-28 items-center justify-center rounded-3xl border-4 border-white/20 bg-white/10">
              <CircleUserRound size={54} />
            </div>
          )}
          <div>
            <p className="text-sm font-bold uppercase tracking-[0.18em] text-emerald-200">
              {account?.role === "member" ? "Member profile" : "Account profile"}
            </p>
            <h1 className="mt-2 text-3xl font-extrabold">
              {member?.name || account?.name}
            </h1>
            <div className="mt-3 flex flex-wrap items-center gap-2">
              <Badge status={member?.status || account?.status || "active"} />
              <span className="rounded-full bg-white/15 px-3 py-1 text-xs font-semibold">
                {getRoleLabel(account?.role)}
              </span>
              {member?.membership_no && (
                <span className="rounded-full bg-white/15 px-3 py-1 text-xs font-semibold">
                  Member #{member.membership_no}
                </span>
              )}
            </div>
          </div>
        </div>
      </section>

      {error && (
        <div className="rounded-2xl border border-red-200 bg-red-50 p-4 text-sm text-red-700 dark:border-red-900/50 dark:bg-red-950/30 dark:text-red-300">
          {error}
        </div>
      )}

      <AccountSecurity />

      <Card>
        <CardHeader>
          <h2 className="font-bold text-slate-900 dark:text-white">
            Personal and membership details
          </h2>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
            <Detail icon={UserRound} label="Full name" value={member?.name || account?.name} />
            <Detail icon={Hash} label="Membership number" value={member?.membership_no} />
            <Detail icon={Phone} label="Phone" value={member?.phone || account?.phone} />
            <Detail icon={Mail} label="Email" value={account?.email} />
            <Detail icon={UserRound} label="Assistant household head" value={member?.assistant_name} />
            <Detail icon={UserRound} label="Father / Husband name" value={member?.father_name} />
            <Detail icon={MapPin} label="Tole / Road" value={member?.tole} />
            <Detail icon={Home} label="Ward" value={member?.ward_no ? `Ward ${member.ward_no}` : null} />
            <Detail icon={CalendarDays} label="Joined date" value={formatDate(member?.joined_date)} />
            <Detail icon={ShieldCheck} label="System role" value={getRoleLabel(account?.role)} />
          </div>

          {member?.remarks && (
            <div className="mt-4 rounded-xl bg-slate-50 p-4 text-sm text-slate-700 dark:bg-white/5 dark:text-slate-300">
              <p className="mb-1 text-xs font-bold uppercase tracking-wide text-slate-400">Remarks</p>
              {member.remarks}
            </div>
          )}
        </CardContent>
      </Card>

      {account?.role === "member" && (
        <>
          <Card>
            <CardHeader>
              <h2 className="font-bold text-slate-900 dark:text-white">
                Household photographs
              </h2>
            </CardHeader>
            <CardContent className="grid grid-cols-1 gap-5 sm:grid-cols-2">
              <div className="rounded-2xl border border-slate-200 p-4 text-center dark:border-white/10">
                {photo ? (
                  <img src={photo} alt="Household head" className="mx-auto h-44 w-44 rounded-2xl object-cover" />
                ) : (
                  <div className="mx-auto flex h-44 w-44 items-center justify-center rounded-2xl bg-slate-100 text-slate-400 dark:bg-white/5"><UserRound size={44} /></div>
                )}
                <p className="mt-3 font-bold text-slate-800 dark:text-slate-100">Household head</p>
              </div>
              <div className="rounded-2xl border border-slate-200 p-4 text-center dark:border-white/10">
                {assistantPhoto ? (
                  <img src={assistantPhoto} alt="Assistant household head" className="mx-auto h-44 w-44 rounded-2xl object-cover" />
                ) : (
                  <div className="mx-auto flex h-44 w-44 items-center justify-center rounded-2xl bg-slate-100 text-slate-400 dark:bg-white/5"><UserRound size={44} /></div>
                )}
                <p className="mt-3 font-bold text-slate-800 dark:text-slate-100">Assistant household head</p>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex items-center gap-2">
              <UsersRound size={20} className="text-emerald-600" />
              <div>
                <h2 className="font-bold text-slate-900 dark:text-white">Family members</h2>
                <p className="text-xs text-slate-500">Members recorded under this household.</p>
              </div>
            </CardHeader>
            <CardContent className="overflow-x-auto p-0">
              <table className="w-full min-w-[720px] text-left text-sm">
                <thead className="border-b border-slate-200 bg-slate-50 text-xs uppercase text-slate-500 dark:border-white/10 dark:bg-white/5">
                  <tr>
                    <th className="px-6 py-3">Name</th>
                    <th className="px-6 py-3">Relation</th>
                    <th className="px-6 py-3">Age</th>
                    <th className="px-6 py-3">Gender</th>
                    <th className="px-6 py-3">Citizenship no.</th>
                    <th className="px-6 py-3">Remarks</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-200 dark:divide-white/10">
                  {familyMembers.map((family) => (
                    <tr key={family.id}>
                      <td className="px-6 py-4 font-semibold text-slate-900 dark:text-white">{family.name}</td>
                      <td className="px-6 py-4 capitalize">{family.relation || "-"}</td>
                      <td className="px-6 py-4">{family.age ?? "-"}</td>
                      <td className="px-6 py-4 capitalize">{family.gender || "-"}</td>
                      <td className="px-6 py-4">{family.citizenship_no || "-"}</td>
                      <td className="px-6 py-4">{family.remarks || "-"}</td>
                    </tr>
                  ))}
                  {!familyMembers.length && (
                    <tr><td colSpan={6} className="px-6 py-10 text-center text-slate-400">No family-member details have been recorded.</td></tr>
                  )}
                </tbody>
              </table>
            </CardContent>
          </Card>
        </>
      )}
    </div>
  );
}
