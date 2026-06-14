import { useEffect } from "react";
import { Link } from "react-router-dom";
import { useDispatch, useSelector } from "react-redux";
import { motion as Motion } from "framer-motion";
import {
  AlertCircle,
  ArrowRight,
  CreditCard,
  FileText,
  History,
  RefreshCw,
  Users,
  UserCheck,
  Wallet,
} from "lucide-react";
import {
  Bar,
  BarChart,
  CartesianGrid,
  Cell,
  Legend,
  Line,
  LineChart,
  Pie,
  PieChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import {
  fetchDashboard,
  fetchDashboardCharts,
} from "../../../redux/slices/dashboardSlice";
import { Card, CardContent, CardHeader } from "../../../components/ui/Card";
import Button from "../../../components/ui/Button";
import Badge from "../../../components/ui/Badge";
import LoadingSpinner from "../../../components/common/LoadingSpinner";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "../../../components/ui/Table";
import { formatCurrency, formatNumber } from "../../../utils/helpers";

const chartColors = ["#059669", "#2563eb", "#d97706", "#7c3aed", "#dc2626"];

function EmptyChart({ text }) {
  return (
    <div className="flex h-full items-center justify-center rounded-xl border border-dashed border-slate-200 text-sm text-slate-400 dark:border-white/10">
      {text}
    </div>
  );
}

export default function AdminStaffDashboard() {
  const dispatch = useDispatch();
  const user = useSelector((state) => state.auth.user);
  const settings = useSelector((state) => state.appSettings.settings);
  const { data, charts, dataStatus, chartStatus, error } = useSelector(
    (state) => state.dashboard,
  );
  const isAdmin = user?.role === "admin";

  const refresh = () => {
    dispatch(fetchDashboard());
    dispatch(fetchDashboardCharts());
  };

  useEffect(() => {
    if (dataStatus === "idle") dispatch(fetchDashboard());
    if (chartStatus === "idle") dispatch(fetchDashboardCharts());
  }, [dispatch, dataStatus, chartStatus]);

  if (dataStatus === "loading" && !data) {
    return <LoadingSpinner text="Loading dashboard..." />;
  }

  const adminCards = [
    {
      key: "total_members",
      label: "Registered members",
      icon: Users,
      format: "number",
      tone: "emerald",
    },
    {
      key: "pending_requests",
      label: "Pending requests",
      icon: FileText,
      format: "number",
      tone: "amber",
    },
    {
      key: "members_with_historical_due",
      label: "Members with past balance",
      icon: History,
      format: "number",
      tone: "rose",
    },
    {
      key: "historical_outstanding",
      label: "Past balance outstanding",
      icon: History,
      format: "currency",
      tone: "rose",
    },
    {
      key: "unverified_historical_entries",
      label: "Past records to verify",
      icon: AlertCircle,
      format: "number",
      tone: "amber",
    },
    {
      key: "total_revenue",
      label: "Revenue this fiscal year",
      icon: CreditCard,
      format: "currency",
      tone: "blue",
    },
    {
      key: "balance",
      label: "Current balance",
      icon: Wallet,
      format: "currency",
      tone: "violet",
    },
  ];

  const staffCards = [
    {
      key: "total_members",
      label: "Registered members",
      icon: Users,
      format: "number",
      tone: "emerald",
    },
    {
      key: "active_members",
      label: "Active members",
      icon: UserCheck,
      format: "number",
      tone: "blue",
    },
    {
      key: "members_with_historical_due",
      label: "Members with past balance",
      icon: History,
      format: "number",
      tone: "rose",
    },
    {
      key: "pending_requests",
      label: "Requests to review",
      icon: FileText,
      format: "number",
      tone: "violet",
    },
    {
      key: "unverified_historical_entries",
      label: "Past records awaiting admin",
      icon: AlertCircle,
      format: "number",
      tone: "amber",
    },
  ];

  const toneClasses = {
    emerald: "bg-emerald-100 text-emerald-700 dark:bg-emerald-950 dark:text-emerald-300",
    amber: "bg-amber-100 text-amber-700 dark:bg-amber-950 dark:text-amber-300",
    blue: "bg-blue-100 text-blue-700 dark:bg-blue-950 dark:text-blue-300",
    violet: "bg-violet-100 text-violet-700 dark:bg-violet-950 dark:text-violet-300",
    rose: "bg-rose-100 text-rose-700 dark:bg-rose-950 dark:text-rose-300",
  };

  const statCards = isAdmin ? adminCards : staffCards;
  const monthlyFinancials = charts?.monthly_financials || [];
  const memberGrowth = charts?.member_growth || [];
  const requestDistribution = charts?.request_status_distribution || [];

  return (
    <div className="space-y-6">
      <section className="overflow-hidden rounded-3xl bg-linear-to-br from-emerald-700 via-emerald-800 to-slate-900 p-6 text-white shadow-xl shadow-emerald-950/10 sm:p-8">
        <div className="flex flex-col justify-between gap-5 lg:flex-row lg:items-end">
          <div>
            <p className="text-sm font-semibold text-emerald-200">
              {isAdmin ? "Administration overview" : "Operations workspace"}
            </p>
            <h1 className="mt-2 text-3xl font-extrabold tracking-tight sm:text-4xl">
              Namaste, {user?.name}
            </h1>
            <p className="mt-2 max-w-2xl text-sm leading-6 text-emerald-100/80 sm:text-base">
              Monitor the member register, service requests and daily work of {settings?.name}.
            </p>
          </div>
          <div className="flex flex-wrap items-center gap-2">
            {data?.active_fiscal_year && (
              <div className="rounded-xl border border-white/15 bg-white/10 px-4 py-2 backdrop-blur-sm">
                <p className="text-[11px] uppercase tracking-wide text-emerald-200">
                  Active fiscal year
                </p>
                <p className="font-bold">{data.active_fiscal_year}</p>
              </div>
            )}
            <Button
              type="button"
              variant="secondary"
              onClick={refresh}
              isLoading={dataStatus === "loading" || chartStatus === "loading"}
            >
              <RefreshCw size={16} /> Refresh
            </Button>
          </div>
        </div>
      </section>

      {error && (
        <div className="flex flex-col gap-3 rounded-2xl border border-red-200 bg-red-50 p-4 text-sm text-red-700 sm:flex-row sm:items-center sm:justify-between dark:border-red-900/50 dark:bg-red-950/30 dark:text-red-300">
          <span>{error}</span>
          <Button type="button" size="sm" variant="outline" onClick={refresh}>
            Try again
          </Button>
        </div>
      )}

      <section className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4">
        {statCards.map((card, index) => (
          <Motion.div
            key={card.key}
            initial={{ opacity: 0, y: 16 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: index * 0.06 }}
          >
            <Card className="h-full">
              <CardContent className="flex items-start justify-between gap-4 p-5">
                <div>
                  <p className="text-sm font-semibold text-slate-500 dark:text-slate-400">
                    {card.label}
                  </p>
                  <p className="mt-2 text-3xl font-extrabold tracking-tight text-slate-950 dark:text-white">
                    {card.format === "currency"
                      ? formatCurrency(data?.[card.key] || 0)
                      : formatNumber(data?.[card.key] || 0)}
                  </p>
                </div>
                <div className={`rounded-2xl p-3 ${toneClasses[card.tone]}`}>
                  <card.icon size={22} />
                </div>
              </CardContent>
            </Card>
          </Motion.div>
        ))}
      </section>

      <section className="grid grid-cols-1 gap-4 md:grid-cols-3">
        <Link to="/members" className="group rounded-2xl border border-slate-200 bg-white p-5 transition hover:-translate-y-0.5 hover:border-emerald-300 hover:shadow-lg dark:border-white/10 dark:bg-slate-900">
          <Users className="text-emerald-600" />
          <h2 className="mt-3 font-bold text-slate-900 dark:text-white">Member register</h2>
          <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">Add, search and update household membership records.</p>
          <span className="mt-4 inline-flex items-center gap-1 text-sm font-bold text-emerald-700 dark:text-emerald-400">Open register <ArrowRight size={15} className="transition-transform group-hover:translate-x-1" /></span>
        </Link>
        <Link to="/requests" className="group rounded-2xl border border-slate-200 bg-white p-5 transition hover:-translate-y-0.5 hover:border-emerald-300 hover:shadow-lg dark:border-white/10 dark:bg-slate-900">
          <FileText className="text-emerald-600" />
          <h2 className="mt-3 font-bold text-slate-900 dark:text-white">Review requests</h2>
          <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">Process timber, firewood and other resource requests.</p>
          <span className="mt-4 inline-flex items-center gap-1 text-sm font-bold text-emerald-700 dark:text-emerald-400">View requests <ArrowRight size={15} className="transition-transform group-hover:translate-x-1" /></span>
        </Link>
        <Link to={isAdmin ? "/reports" : "/payments"} className="group rounded-2xl border border-slate-200 bg-white p-5 transition hover:-translate-y-0.5 hover:border-emerald-300 hover:shadow-lg dark:border-white/10 dark:bg-slate-900">
          {isAdmin ? <Wallet className="text-emerald-600" /> : <CreditCard className="text-emerald-600" />}
          <h2 className="mt-3 font-bold text-slate-900 dark:text-white">{isAdmin ? "Financial reports" : "Payment records"}</h2>
          <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">{isAdmin ? "Review income, expenses and fiscal-year summaries." : "Review member payment records."}</p>
          <span className="mt-4 inline-flex items-center gap-1 text-sm font-bold text-emerald-700 dark:text-emerald-400">Open section <ArrowRight size={15} className="transition-transform group-hover:translate-x-1" /></span>
        </Link>
      </section>

      <section className="grid grid-cols-1 gap-6 xl:grid-cols-2">
        {isAdmin ? (
          <Card>
            <CardHeader>
              <h2 className="font-bold text-slate-900 dark:text-white">Monthly income and expenses</h2>
            </CardHeader>
            <CardContent className="h-80">
              {monthlyFinancials.length ? (
                <ResponsiveContainer width="100%" height="100%">
                  <BarChart data={monthlyFinancials}>
                    <CartesianGrid strokeDasharray="3 3" vertical={false} />
                    <XAxis dataKey="month" tick={{ fontSize: 11 }} />
                    <YAxis tick={{ fontSize: 11 }} />
                    <Tooltip formatter={(value) => formatCurrency(value)} />
                    <Legend />
                    <Bar dataKey="income" name="Income" fill="#059669" radius={[5, 5, 0, 0]} />
                    <Bar dataKey="expense" name="Expense" fill="#dc2626" radius={[5, 5, 0, 0]} />
                  </BarChart>
                </ResponsiveContainer>
              ) : (
                <EmptyChart text="No financial chart data yet" />
              )}
            </CardContent>
          </Card>
        ) : (
          <Card>
            <CardHeader>
              <h2 className="font-bold text-slate-900 dark:text-white">Member growth</h2>
            </CardHeader>
            <CardContent className="h-80">
              {memberGrowth.length ? (
                <ResponsiveContainer width="100%" height="100%">
                  <LineChart data={memberGrowth}>
                    <CartesianGrid strokeDasharray="3 3" vertical={false} />
                    <XAxis dataKey="month" tick={{ fontSize: 11 }} />
                    <YAxis tick={{ fontSize: 11 }} allowDecimals={false} />
                    <Tooltip />
                    <Line type="monotone" dataKey="count" name="New members" stroke="#059669" strokeWidth={3} />
                  </LineChart>
                </ResponsiveContainer>
              ) : (
                <EmptyChart text="No member growth data yet" />
              )}
            </CardContent>
          </Card>
        )}

        <Card>
          <CardHeader>
            <h2 className="font-bold text-slate-900 dark:text-white">Request status</h2>
          </CardHeader>
          <CardContent className="h-80">
            {requestDistribution.length ? (
              <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie
                    data={requestDistribution}
                    dataKey="count"
                    nameKey="status"
                    innerRadius={62}
                    outerRadius={96}
                    paddingAngle={4}
                  >
                    {requestDistribution.map((item, index) => (
                      <Cell key={`${item.status}-${index}`} fill={chartColors[index % chartColors.length]} />
                    ))}
                  </Pie>
                  <Tooltip />
                  <Legend />
                </PieChart>
              </ResponsiveContainer>
            ) : (
              <EmptyChart text="No request data yet" />
            )}
          </CardContent>
        </Card>
      </section>

      <section className="grid grid-cols-1 gap-6 xl:grid-cols-2">
        <Card>
          <CardHeader className="flex items-center justify-between">
            <h2 className="font-bold text-slate-900 dark:text-white">Recent requests</h2>
            <Link to="/requests" className="text-sm font-semibold text-emerald-700 dark:text-emerald-400">View all</Link>
          </CardHeader>
          <CardContent className="p-0">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Member</TableHead>
                  <TableHead>Resource</TableHead>
                  <TableHead>Status</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {(data?.recent_requests || []).map((request) => (
                  <TableRow key={request.id}>
                    <TableCell className="font-semibold">{request.member_name}</TableCell>
                    <TableCell>{request.resource_name}</TableCell>
                    <TableCell><Badge status={request.status} /></TableCell>
                  </TableRow>
                ))}
                {!data?.recent_requests?.length && (
                  <TableRow><TableCell colSpan={3} className="py-10 text-center text-slate-400">No recent requests</TableCell></TableRow>
                )}
              </TableBody>
            </Table>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex items-center justify-between">
            <h2 className="font-bold text-slate-900 dark:text-white">Recent payments</h2>
            <Link to="/payments" className="text-sm font-semibold text-emerald-700 dark:text-emerald-400">View all</Link>
          </CardHeader>
          <CardContent className="p-0">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Member</TableHead>
                  <TableHead>Amount</TableHead>
                  <TableHead>Status</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {(data?.recent_payments || []).map((payment) => (
                  <TableRow key={payment.id}>
                    <TableCell className="font-semibold">{payment.member_name}</TableCell>
                    <TableCell>{formatCurrency(payment.amount || 0)}</TableCell>
                    <TableCell><Badge status={payment.status} /></TableCell>
                  </TableRow>
                ))}
                {!data?.recent_payments?.length && (
                  <TableRow><TableCell colSpan={3} className="py-10 text-center text-slate-400">No recent payments</TableCell></TableRow>
                )}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      </section>
    </div>
  );
}
