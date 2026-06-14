import { useCallback, useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { useSelector } from "react-redux";
import {
  AlertCircle,
  ArrowRight,
  BadgeCheck,
  CalendarDays,
  CreditCard,
  FileText,
  MapPin,
  RefreshCw,
  ShieldCheck,
  UserRound,
  Wallet,
} from "lucide-react";
import { api } from "../../../services/api";
import { Card, CardContent, CardHeader } from "../../../components/ui/Card";
import Badge from "../../../components/ui/Badge";
import Button from "../../../components/ui/Button";
import LoadingSpinner from "../../../components/common/LoadingSpinner";
import { useToast } from "../../../components/common/Toast";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "../../../components/ui/Table";
import {
  formatCurrency,
  formatDate,
  getImageUrl,
} from "../../../utils/helpers";

const initialData = {
  requests: [],
  payments: [],
  transactions: [],
};

const transactionLabels = {
  legacy_gasti_fee: "Past Gasti fee",
  legacy_timber_sale: "Past timber sale",
  legacy_firewood_sale: "Past firewood sale",
  legacy_other_sale: "Past other sale",
  resource_sale: "Resource sale",
  membership_fee: "Annual Gasti / Membership fee",
  fine: "Fine",
};

function getTransactionLabel(type) {
  return (
    transactionLabels[type] ||
    String(type || "Ledger entry").replaceAll("_", " ")
  );
}

function isMembershipFee(transaction) {
  return ["membership_fee", "legacy_gasti_fee"].includes(transaction?.type);
}

export default function MemberDashboard() {
  const user = useSelector((state) => state.auth.user);
  const settings = useSelector((state) => state.appSettings.settings);
  const { addToast } = useToast();
  const [data, setData] = useState(initialData);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState(null);
  const [payingFeeId, setPayingFeeId] = useState(null);

  const loadMemberDashboard = useCallback(async () => {
    setIsLoading(true);
    setError(null);

    const results = await Promise.allSettled([
      api.getMyRequests({ page: 1, per_page: 100 }),
      api.getMyPayments({ page: 1, per_page: 100 }),
      api.getMyTransactions({ page: 1, per_page: 100 }),
    ]);

    const [requestResult, paymentResult, transactionResult] = results;
    const nextData = {
      requests:
        requestResult.status === "fulfilled" && requestResult.value.success
          ? requestResult.value.data || []
          : [],
      payments:
        paymentResult.status === "fulfilled" && paymentResult.value.success
          ? paymentResult.value.data || []
          : [],
      transactions:
        transactionResult.status === "fulfilled" &&
        transactionResult.value.success
          ? transactionResult.value.data || []
          : [],
    };

    setData(nextData);
    if (results.every((result) => result.status === "rejected")) {
      setError("Your dashboard data could not be loaded. Please try again.");
    }
    setIsLoading(false);
  }, []);

  useEffect(() => {
    loadMemberDashboard();
  }, [loadMemberDashboard]);

  const membershipFees = useMemo(
    () =>
      data.transactions
        .filter(
          (transaction) =>
            isMembershipFee(transaction) &&
            transaction.record_status !== "reversed",
        )
        .sort((a, b) => {
          const aDate = new Date(a.fiscal_year?.start_date || a.date || 0);
          const bDate = new Date(b.fiscal_year?.start_date || b.date || 0);
          return bDate - aDate;
        }),
    [data.transactions],
  );

  const hasPendingEsewa = useCallback(
    (transactionId) =>
      data.payments.some(
        (payment) =>
          Number(payment.ledger_transaction_id) === Number(transactionId) &&
          payment.payment_method === "esewa" &&
          payment.status === "pending",
      ),
    [data.payments],
  );

  const handleFeeEsewaPayment = async (fee) => {
    if (!fee?.id || Number(fee.amount_remaining || 0) <= 0) return;
    setPayingFeeId(fee.id);
    try {
      const result = await api.initiateEsewaLedgerPayment(fee.id);
      if (!result.success || !result.data?.action_url) {
        throw new Error(result.message || "Unable to initialize eSewa payment");
      }
      api.submitEsewaForm(result.data.action_url, result.data.fields);
    } catch (err) {
      addToast(err.message || "Unable to initialize eSewa payment", "error");
      setPayingFeeId(null);
    }
  };

  const summary = useMemo(() => {
    const pendingRequests = data.requests.filter(
      (request) => request.status === "pending",
    ).length;
    const approvedRequests = data.requests.filter(
      (request) => request.status === "approved",
    ).length;
    const paidAmount = data.payments
      .filter((payment) => payment.status === "paid")
      .reduce((sum, payment) => sum + Number(payment.amount || 0), 0);
    const historicalOutstanding = data.transactions
      .filter(
        (transaction) =>
          String(transaction.type || "").startsWith("legacy_") &&
          transaction.record_status !== "reversed",
      )
      .reduce(
        (sum, transaction) =>
          sum + Math.max(0, Number(transaction.amount_remaining || 0)),
        0,
      );
    const currentOutstanding = data.transactions
      .filter(
        (transaction) =>
          !String(transaction.type || "").startsWith("legacy_") &&
          transaction.record_status !== "reversed",
      )
      .reduce(
        (sum, transaction) =>
          sum + Math.max(0, Number(transaction.amount_remaining || 0)),
        0,
      );
    const membershipFeeOutstanding = membershipFees.reduce(
      (sum, fee) => sum + Math.max(0, Number(fee.amount_remaining || 0)),
      0,
    );

    return {
      pendingRequests,
      approvedRequests,
      paidAmount,
      outstandingAmount: historicalOutstanding + currentOutstanding,
      historicalOutstanding,
      currentOutstanding,
      membershipFeeOutstanding,
    };
  }, [data, membershipFees]);

  if (isLoading) return <LoadingSpinner text="Loading your dashboard..." />;

  const member = user?.member;
  const memberPhoto = getImageUrl(member?.photo);

  return (
    <div className="space-y-6">
      <section className="overflow-hidden rounded-3xl bg-linear-to-br from-emerald-700 via-teal-800 to-slate-900 p-6 text-white shadow-xl shadow-emerald-950/10 sm:p-8">
        <div className="flex flex-col justify-between gap-6 lg:flex-row lg:items-center">
          <div className="flex items-center gap-4">
            {memberPhoto ? (
              <img
                src={memberPhoto}
                alt={member?.name || user?.name}
                className="h-16 w-16 rounded-2xl border-2 border-white/30 object-cover"
              />
            ) : (
              <div className="flex h-16 w-16 items-center justify-center rounded-2xl bg-white/15 text-2xl font-extrabold backdrop-blur-sm">
                {user?.name?.charAt(0)?.toUpperCase() || "M"}
              </div>
            )}
            <div>
              <p className="text-sm font-semibold text-emerald-200">Member portal</p>
              <h1 className="mt-1 text-3xl font-extrabold">Namaste, {user?.name}</h1>
              <p className="mt-1 text-sm text-emerald-100/80">
                View your membership, yearly Gasti fee, requests and payments.
              </p>
            </div>
          </div>
          <div className="flex flex-wrap gap-2">
            <Link
              to="/profile"
              className="inline-flex items-center justify-center gap-2 rounded-xl bg-white px-4 py-2 text-sm font-bold text-emerald-800 transition hover:bg-emerald-50"
            >
              <UserRound size={16} /> View profile
            </Link>
            <Button type="button" variant="secondary" onClick={loadMemberDashboard}>
              <RefreshCw size={16} /> Refresh
            </Button>
          </div>
        </div>
      </section>

      {error && (
        <div className="rounded-2xl border border-red-200 bg-red-50 p-4 text-sm text-red-700 dark:border-red-900/50 dark:bg-red-950/30 dark:text-red-300">
          {error}
        </div>
      )}

      <section className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-3">
        <Card><CardContent className="flex items-center justify-between p-5"><div><p className="text-sm font-semibold text-slate-500">Pending requests</p><p className="mt-2 text-3xl font-extrabold">{summary.pendingRequests}</p></div><div className="rounded-2xl bg-amber-100 p-3 text-amber-700"><FileText size={22} /></div></CardContent></Card>
        <Card><CardContent className="flex items-center justify-between p-5"><div><p className="text-sm font-semibold text-slate-500">Approved requests</p><p className="mt-2 text-3xl font-extrabold">{summary.approvedRequests}</p></div><div className="rounded-2xl bg-emerald-100 p-3 text-emerald-700"><BadgeCheck size={22} /></div></CardContent></Card>
        <Card className="border-emerald-200 bg-emerald-50/60 dark:border-emerald-900/40 dark:bg-emerald-950/20"><CardContent className="flex items-center justify-between p-5"><div><p className="text-sm font-semibold text-emerald-700 dark:text-emerald-300">Gasti fee outstanding</p><p className="mt-2 text-2xl font-extrabold text-emerald-800 dark:text-emerald-200">{formatCurrency(summary.membershipFeeOutstanding)}</p></div><div className="rounded-2xl bg-emerald-100 p-3 text-emerald-700"><CalendarDays size={22} /></div></CardContent></Card>
        <Card><CardContent className="flex items-center justify-between p-5"><div><p className="text-sm font-semibold text-slate-500">Paid amount</p><p className="mt-2 text-2xl font-extrabold">{formatCurrency(summary.paidAmount)}</p></div><div className="rounded-2xl bg-blue-100 p-3 text-blue-700"><CreditCard size={22} /></div></CardContent></Card>
        <Card className="border-amber-200 bg-amber-50/60 dark:border-amber-900/40 dark:bg-amber-950/20"><CardContent className="flex items-center justify-between p-5"><div><p className="text-sm font-semibold text-amber-700 dark:text-amber-300">Past fiscal-year balance</p><p className="mt-2 text-2xl font-extrabold text-amber-800 dark:text-amber-200">{formatCurrency(summary.historicalOutstanding)}</p></div><div className="rounded-2xl bg-amber-100 p-3 text-amber-700"><AlertCircle size={22} /></div></CardContent></Card>
        <Card><CardContent className="flex items-center justify-between p-5"><div><p className="text-sm font-semibold text-slate-500">Total outstanding</p><p className="mt-2 text-2xl font-extrabold">{formatCurrency(summary.outstandingAmount)}</p></div><div className="rounded-2xl bg-slate-100 p-3 text-slate-700"><Wallet size={22} /></div></CardContent></Card>
      </section>

      <Card className="overflow-hidden border-emerald-200 dark:border-emerald-900/40">
        <CardHeader className="flex flex-col justify-between gap-3 bg-emerald-50/70 sm:flex-row sm:items-center dark:bg-emerald-950/20">
          <div>
            <h2 className="font-bold text-slate-900 dark:text-white">
              Gasti / Membership fee by fiscal year
            </h2>
            <p className="mt-1 text-xs text-slate-500">
              Each fiscal year is listed separately. Unpaid verified fees can be paid directly through eSewa.
            </p>
          </div>
          <div className="flex items-center gap-2 text-xs font-semibold text-emerald-700 dark:text-emerald-300">
            <ShieldCheck size={16} /> Verified ledger entries only
          </div>
        </CardHeader>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Fiscal year</TableHead>
                <TableHead>Fee type</TableHead>
                <TableHead>Total</TableHead>
                <TableHead>Paid</TableHead>
                <TableHead>Remaining</TableHead>
                <TableHead>Status</TableHead>
                <TableHead className="text-right">Action</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {membershipFees.map((fee) => {
                const remaining = Math.max(0, Number(fee.amount_remaining || 0));
                const isPaid = remaining <= 0;
                const pending = hasPendingEsewa(fee.id);
                return (
                  <TableRow key={fee.id}>
                    <TableCell className="font-bold">{fee.fiscal_year?.name || "-"}</TableCell>
                    <TableCell>
                      {fee.type === "legacy_gasti_fee" ? "Past Gasti fee" : "Annual Gasti fee"}
                      {fee.type === "legacy_gasti_fee" && (
                        <span className="ml-2 rounded-full bg-amber-100 px-2 py-0.5 text-[10px] font-bold uppercase tracking-wide text-amber-700">Old register</span>
                      )}
                    </TableCell>
                    <TableCell>{formatCurrency(fee.total_amount || 0)}</TableCell>
                    <TableCell>{formatCurrency(fee.amount_paid || 0)}</TableCell>
                    <TableCell className="font-extrabold text-amber-700 dark:text-amber-300">{formatCurrency(remaining)}</TableCell>
                    <TableCell>
                      <Badge status={isPaid ? "paid" : pending ? "pending" : "approved"}>
                        {isPaid ? "Paid" : pending ? "Payment pending" : "Unpaid"}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-right">
                      {isPaid ? (
                        <span className="text-xs font-bold text-emerald-600">Completed</span>
                      ) : pending ? (
                        <Link to="/payments" className="text-xs font-bold text-amber-700 hover:underline">Check payment</Link>
                      ) : (
                        <Button
                          size="sm"
                          onClick={() => handleFeeEsewaPayment(fee)}
                          isLoading={payingFeeId === fee.id}
                        >
                          Pay with eSewa
                        </Button>
                      )}
                    </TableCell>
                  </TableRow>
                );
              })}
              {!membershipFees.length && (
                <TableRow><TableCell colSpan={7} className="py-10 text-center text-slate-400">No Gasti or membership fee has been assigned yet.</TableCell></TableRow>
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      <section className="grid grid-cols-1 gap-6 xl:grid-cols-[1fr_1.6fr]">
        <Card>
          <CardHeader><h2 className="font-bold text-slate-900 dark:text-white">Membership details</h2></CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-start gap-3"><UserRound size={19} className="mt-0.5 text-emerald-600" /><div><p className="text-xs font-semibold uppercase tracking-wide text-slate-400">Membership number</p><p className="font-bold text-slate-900 dark:text-white">{member?.membership_no || "Not assigned"}</p></div></div>
            <div className="flex items-start gap-3"><MapPin size={19} className="mt-0.5 text-emerald-600" /><div><p className="text-xs font-semibold uppercase tracking-wide text-slate-400">Address</p><p className="font-semibold text-slate-700 dark:text-slate-200">{[member?.tole, member?.ward_no ? `Ward ${member.ward_no}` : null].filter(Boolean).join(", ") || "Not recorded"}</p></div></div>
            <div><p className="text-xs font-semibold uppercase tracking-wide text-slate-400">Member status</p><div className="mt-1"><Badge status={member?.status || "active"} /></div></div>
            <div><p className="text-xs font-semibold uppercase tracking-wide text-slate-400">Joined date</p><p className="mt-1 font-semibold text-slate-700 dark:text-slate-200">{formatDate(member?.joined_date)}</p></div>
            <Link to="/profile" className="inline-flex items-center gap-2 text-sm font-bold text-emerald-700 hover:underline dark:text-emerald-400"><UserRound size={16} /> View complete profile and family details</Link>
            <div className="rounded-xl bg-emerald-50 p-4 text-sm leading-6 text-emerald-900 dark:bg-emerald-950/40 dark:text-emerald-200">Your record is managed by {settings?.name}. Contact the office when any household information needs correction.</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex items-center justify-between"><h2 className="font-bold text-slate-900 dark:text-white">Recent requests</h2><Link to="/requests" className="text-sm font-semibold text-emerald-700 dark:text-emerald-400">View all</Link></CardHeader>
          <CardContent className="p-0">
            <Table>
              <TableHeader><TableRow><TableHead>Resource</TableHead><TableHead>Quantity</TableHead><TableHead>Date</TableHead><TableHead>Status</TableHead></TableRow></TableHeader>
              <TableBody>
                {data.requests.slice(0, 6).map((request) => (
                  <TableRow key={request.id}>
                    <TableCell className="font-semibold">{request.resource_item?.name || request.resource_name || "Resource request"}</TableCell>
                    <TableCell>{request.quantity_requested}</TableCell>
                    <TableCell>{formatDate(request.requested_at || request.created_at)}</TableCell>
                    <TableCell><Badge status={request.status} /></TableCell>
                  </TableRow>
                ))}
                {!data.requests.length && <TableRow><TableCell colSpan={4} className="py-10 text-center text-slate-400">You have not submitted a request yet.</TableCell></TableRow>}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      </section>

      <Card>
        <CardHeader className="flex items-center justify-between">
          <div>
            <h2 className="font-bold text-slate-900 dark:text-white">Recent ledger balances</h2>
            <p className="mt-1 text-xs text-slate-500">Verified current and historical entries from your member register.</p>
          </div>
          <Link to="/transactions" className="text-sm font-semibold text-emerald-700 dark:text-emerald-400">View all</Link>
        </CardHeader>
        <CardContent className="p-0">
          <Table>
            <TableHeader><TableRow><TableHead>Fiscal year</TableHead><TableHead>Entry</TableHead><TableHead>Total</TableHead><TableHead>Paid</TableHead><TableHead>Remaining</TableHead></TableRow></TableHeader>
            <TableBody>
              {data.transactions.slice(0, 8).map((transaction) => (
                <TableRow key={transaction.id}>
                  <TableCell>{transaction.fiscal_year?.name || "-"}</TableCell>
                  <TableCell className="font-semibold">
                    {getTransactionLabel(transaction.type)}
                    {String(transaction.type || "").startsWith("legacy_") && (
                      <span className="ml-2 rounded-full bg-amber-100 px-2 py-0.5 text-[10px] font-bold uppercase tracking-wide text-amber-700">Past record</span>
                    )}
                  </TableCell>
                  <TableCell>{formatCurrency(transaction.total_amount || 0)}</TableCell>
                  <TableCell>{formatCurrency(transaction.amount_paid || 0)}</TableCell>
                  <TableCell className="font-bold text-amber-700 dark:text-amber-300">{formatCurrency(transaction.amount_remaining || 0)}</TableCell>
                </TableRow>
              ))}
              {!data.transactions.length && <TableRow><TableCell colSpan={5} className="py-10 text-center text-slate-400">No verified ledger entries are available.</TableCell></TableRow>}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      <section className="grid grid-cols-1 gap-4 md:grid-cols-4">
        {[
          { to: "/profile", icon: UserRound, title: "My profile", text: "View household and family details." },
          { to: "/requests", icon: FileText, title: "My requests", text: "Submit and track resource requests." },
          { to: "/payments", icon: CreditCard, title: "My payments", text: "See your payment records and status." },
          { to: "/transactions", icon: Wallet, title: "My ledger", text: "Review charges, receipts and outstanding balances." },
        ].map((item) => (
          <Link key={item.to} to={item.to} className="group rounded-2xl border border-slate-200 bg-white p-5 transition hover:-translate-y-0.5 hover:border-emerald-300 hover:shadow-lg dark:border-white/10 dark:bg-slate-900">
            <item.icon className="text-emerald-600" />
            <h2 className="mt-3 font-bold text-slate-900 dark:text-white">{item.title}</h2>
            <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">{item.text}</p>
            <span className="mt-4 inline-flex items-center gap-1 text-sm font-bold text-emerald-700 dark:text-emerald-400">Open <ArrowRight size={15} className="transition-transform group-hover:translate-x-1" /></span>
          </Link>
        ))}
      </section>
    </div>
  );
}
