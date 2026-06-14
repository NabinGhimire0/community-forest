import { useState, useEffect, useCallback, useMemo } from "react";
import { useSelector } from "react-redux";
import {
  Plus,
  Search,
  CreditCard,
  Receipt,
  RefreshCw,
  Eye,
  AlertTriangle,
  Wallet,
  Building2,
  User as UserIcon,
} from "lucide-react";
import { api } from "../../services/api";
import { Card, CardContent } from "../../components/ui/Card";
import Button from "../../components/ui/Button";
import Input from "../../components/ui/Input";
import Select from "../../components/ui/Select";
import Textarea from "../../components/ui/Textarea";
import Modal from "../../components/ui/Modal";
import Badge from "../../components/ui/Badge";
import {
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
} from "../../components/ui/Table";
import LoadingSpinner from "../../components/common/LoadingSpinner";
import { useToast } from "../../components/common/Toast";
import { formatCurrency, formatDate } from "../../utils/helpers";
import {
  buildFiscalYearOptions,
  getActiveFiscalYearId,
} from "../../utils/fiscalYears";

const emptyPayment = {
  member_id: "",
  request_id: "",
  amount: "",
  remarks: "",
};

export default function Payments() {
  const { user } = useSelector((state) => state.auth);
  const { addToast } = useToast();
  const isMember = user?.role === "member";
  const canRecordCash = user?.role === "admin";
  const canViewStats = user?.role === "admin" || user?.role === "staff";

  const [payments, setPayments] = useState([]);
  const [members, setMembers] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState("");
  const [fiscalYearFilter, setFiscalYearFilter] = useState("");
  const [activeTab, setActiveTab] = useState(isMember ? "my" : "all");
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [refreshing, setRefreshing] = useState(false);
  const [stats, setStats] = useState(null);

  // Modal states
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showViewModal, setShowViewModal] = useState(false);
  const [viewingPayment, setViewingPayment] = useState(null);
  const [saving, setSaving] = useState(false);
  const [checkingPaymentId, setCheckingPaymentId] = useState(null);
  const [form, setForm] = useState(emptyPayment);
  const [pendingRequests, setPendingRequests] = useState([]);
  const [filteredRequests, setFilteredRequests] = useState([]);
  const [fiscalYears, setFiscalYears] = useState([]);

  const fetchMasterData = useCallback(async () => {
    try {
      const [requestsRes, fiscalYearsRes] = await Promise.all([
        isMember
          ? api.getMyRequests({ status: "approved", per_page: 100 })
          : api.getRequests({ status: "approved", per_page: 100 }),
        api.getFiscalYears(),
      ]);
      if (requestsRes.success) {
        const approvedRequests = (requestsRes.data || []).filter(
          (r) => r.status === "approved",
        );
        setPendingRequests(approvedRequests);
      }
      if (fiscalYearsRes.success) {
        const years = fiscalYearsRes.data || [];
        setFiscalYears(years);
        if (canViewStats) {
          setFiscalYearFilter(
            (current) => current || getActiveFiscalYearId(years),
          );
        }
      }
    } catch (err) {
      console.error("Failed to fetch master data:", err);
    }
  }, [canViewStats, isMember]);

  const fetchMembers = useCallback(async () => {
    try {
      const res = await api.getMembers({ per_page: 100 });
      if (res.success) setMembers(res.data || []);
    } catch (err) {
      console.error("Failed to fetch members:", err);
    }
  }, []);

  const fetchPayments = useCallback(async () => {
    setIsLoading(true);
    try {
      const fetchFn = isMember
        ? api.getMyPayments.bind(api)
        : api.getPayments.bind(api);
      const res = await fetchFn({
        page,
        per_page: isMember ? 100 : 10,
        search: search || undefined,
        status: statusFilter || undefined,
        fiscal_year_id: fiscalYearFilter || undefined,
      });
      if (res.success) {
        setPayments(res.data || []);
        setTotalPages(res.meta?.total_pages || 1);
      }
    } catch (_err) {
      addToast("Failed to load payments", "error");
    } finally {
      setIsLoading(false);
      setRefreshing(false);
    }
  }, [
    page,
    search,
    statusFilter,
    fiscalYearFilter,
    isMember,
    addToast,
  ]);

  const fetchStats = useCallback(async () => {
    try {
      const res = await api.getPaymentStats({
        fiscal_year_id: fiscalYearFilter || undefined,
      });
      if (res.success) setStats(res.data);
    } catch (err) {
      console.error("Failed to fetch stats:", err);
    }
  }, [fiscalYearFilter]);

  useEffect(() => {
    fetchMasterData();
    // Load members for admin/staff
    if (canRecordCash) {
      fetchMembers();
    }
  }, [fetchMasterData, fetchMembers, canRecordCash]);

  useEffect(() => {
    fetchPayments();
    if (canViewStats) fetchStats();
  }, [fetchPayments, fetchStats, canViewStats]);

  // Filter requests based on selected member
  useEffect(() => {
    if (isMember) {
      setFilteredRequests(pendingRequests);
    } else if (form.member_id && pendingRequests.length > 0) {
      const filtered = pendingRequests.filter(
        (req) => req.member_id === Number(form.member_id),
      );
      setFilteredRequests(filtered);
    } else {
      setFilteredRequests([]);
    }
  }, [form.member_id, pendingRequests, isMember]);

  const handleCreate = async () => {
    setSaving(true);
    try {
      if (!form.member_id || !form.request_id || Number(form.amount) <= 0) {
        throw new Error("Select a member, approved request and valid amount");
      }
      const payload = {
        member_id: Number(form.member_id),
        request_id: Number(form.request_id),
        amount: Number(form.amount),
        remarks: form.remarks.trim() || undefined,
      };

      const res = await api.createCashPayment(payload);
      if (res.success) {
        addToast("Payment recorded successfully", "success");
        setShowCreateModal(false);
        setForm(emptyPayment);
        fetchPayments();
        fetchStats();
      } else {
        addToast(res.message || "Failed to record payment", "error");
      }
    } catch (err) {
      addToast(err.message || "Failed to record payment", "error");
    } finally {
      setSaving(false);
    }
  };

  const handleCheckEsewaStatus = async (paymentId) => {
    setCheckingPaymentId(paymentId);
    try {
      const res = await api.checkEsewaPaymentStatus(paymentId);
      const status = res.data?.status || "pending";
      addToast(
        status === "paid"
          ? "eSewa payment verified successfully"
          : `Current eSewa status: ${res.data?.gateway_status || status}`,
        status === "paid" ? "success" : "info",
      );
      await fetchPayments();
      if (canViewStats) await fetchStats();
    } catch (err) {
      addToast(err.message || "Failed to check eSewa status", "error");
    } finally {
      setCheckingPaymentId(null);
    }
  };

  const handleDownloadReceipt = async (paymentId) => {
    try {
      await api.downloadPaymentReceipt(paymentId);
    } catch (err) {
      addToast(err.message || "Failed to download receipt", "error");
    }
  };

  const openView = async (payment) => {
    setViewingPayment(payment);
    setShowViewModal(true);
  };

  const getSelectedRequestAmount = () => {
    const request = pendingRequests.find(
      (r) => r.id === Number(form.request_id),
    );
    if (request && request.total_amount) {
      return request.total_amount;
    }
    return null;
  };

  const getSelectedMemberName = () => {
    const member = members.find((m) => m.id === Number(form.member_id));
    return member?.name || "";
  };

  const visiblePayments = useMemo(() => {
    if (!isMember) return payments;
    const term = search.trim().toLowerCase();
    return payments.filter((payment) => {
      const searchable = [
        payment.id,
        payment.request?.id,
        payment.transaction_id,
        payment.payment_method,
      ]
        .filter((value) => value != null)
        .join(" ")
        .toLowerCase();
      const matchesSearch = !term || searchable.includes(term);
      const matchesStatus = !statusFilter || payment.status === statusFilter;
      return matchesSearch && matchesStatus;
    });
  }, [isMember, payments, search, statusFilter]);

  const tabs = isMember
    ? [{ id: "my", label: "My Payments" }]
    : [{ id: "all", label: "All Payments" }];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">
            {isMember ? "My Payments" : "Payments"}
          </h1>
          <p className="text-sm text-gray-500 dark:text-gray-400">
            {isMember ? "Review your payment history" : "Track and manage payments for resource requests"}
          </p>
        </div>
        <div className="flex gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => {
              setRefreshing(true);
              fetchPayments();
            }}
            isLoading={refreshing}
          >
            <RefreshCw size={14} className="mr-1" /> Refresh
          </Button>
          {canRecordCash && (
            <Button
              onClick={() => {
                setForm(emptyPayment);
                setShowCreateModal(true);
              }}
            >
              <Plus size={16} /> Record Payment
            </Button>
          )}
        </div>
      </div>

      {/* Stats Cards - Admin only */}
      {canViewStats && stats && (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-3">
          <Card>
            <CardContent className="p-3 text-center">
              <p className="text-xl font-bold text-gray-900 dark:text-gray-100">
                {formatCurrency(stats.total_amount || 0)}
              </p>
              <p className="text-xs text-gray-500">Total Collected</p>
            </CardContent>
          </Card>
          <Card className="border-emerald-200 bg-emerald-50 dark:bg-emerald-900/20">
            <CardContent className="p-3 text-center">
              <p className="text-xl font-bold text-emerald-600">
                {formatCurrency(stats.cash_amount || 0)}
              </p>
              <p className="text-xs text-emerald-600">Cash</p>
            </CardContent>
          </Card>
          <Card className="border-blue-200 bg-blue-50 dark:bg-blue-900/20">
            <CardContent className="p-3 text-center">
              <p className="text-xl font-bold text-blue-600">
                {formatCurrency(stats.online_amount || 0)}
              </p>
              <p className="text-xs text-blue-600">Online</p>
            </CardContent>
          </Card>
          <Card className="border-amber-200 bg-amber-50 dark:bg-amber-900/20">
            <CardContent className="p-3 text-center">
              <p className="text-xl font-bold text-amber-600">
                {formatCurrency(stats.pending_amount || 0)}
              </p>
              <p className="text-xs text-amber-600">Pending</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-3 text-center">
              <p className="text-xl font-bold text-gray-900 dark:text-gray-100">
                {stats.total_count || 0}
              </p>
              <p className="text-xs text-gray-500">Transactions</p>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Tabs & Filters */}
      <Card>
        <CardContent className="p-4">
          <div className="flex flex-col sm:flex-row gap-3">
            {tabs.length > 1 && (
              <div className="flex gap-1 bg-gray-100 dark:bg-gray-800/50 rounded-lg p-1">
              {tabs.map((tab) => (
                <button
                  key={tab.id}
                  onClick={() => {
                    setActiveTab(tab.id);
                    setPage(1);
                  }}
                  className={`px-3 py-1.5 rounded-md text-sm font-medium transition-colors ${
                    activeTab === tab.id
                      ? "bg-white dark:bg-gray-900 text-emerald-600 dark:text-emerald-400 shadow-sm"
                      : "text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200"
                  }`}
                >
                  {tab.label}
                </button>
              ))}
              </div>
            )}
            <div className="flex-1 relative">
              <Search
                size={16}
                className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400"
              />
              <input
                type="text"
                placeholder="Search by member name, receipt no..."
                value={search}
                onChange={(e) => {
                  setSearch(e.target.value);
                  setPage(1);
                }}
                className="w-full pl-10 pr-3 py-2 border border-gray-200 dark:border-white/10 rounded-lg bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 text-sm focus:outline-none focus:border-emerald-500 focus:ring-2 focus:ring-emerald-500/20"
              />
            </div>
            <Select
              value={statusFilter}
              onChange={(e) => {
                setStatusFilter(e.target.value);
                setPage(1);
              }}
              options={[
                { value: "", label: "All Status" },
                { value: "pending", label: "Pending" },
                { value: "paid", label: "Paid" },
                { value: "failed", label: "Failed" },
              ]}
              className="w-36"
            />
            {canViewStats && (
              <Select
                value={fiscalYearFilter}
                onChange={(e) => {
                  setFiscalYearFilter(e.target.value);
                  setPage(1);
                }}
                options={buildFiscalYearOptions(fiscalYears, {
                  includeAll: true,
                })}
                className="w-40"
              />
            )}
          </div>
        </CardContent>
      </Card>

      {/* Payments Table */}
      <Card>
        <CardContent className="p-0">
          {isLoading ? (
            <LoadingSpinner text="Loading payments..." />
          ) : visiblePayments.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 text-gray-400">
              <Receipt size={48} className="mb-3 opacity-30" />
              <p className="text-lg font-medium">No payments found</p>
              {canRecordCash && (
                <Button
                  variant="outline"
                  className="mt-4"
                  onClick={() => setShowCreateModal(true)}
                >
                  <Plus size={14} /> Record your first payment
                </Button>
              )}
            </div>
          ) : (
            <div className="overflow-x-auto">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>ID</TableHead>
                    <TableHead>Member</TableHead>
                    <TableHead>Payment For</TableHead>
                    <TableHead>Amount</TableHead>
                    <TableHead>Method</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Date</TableHead>
                    <TableHead>Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {visiblePayments.map((payment) => (
                    <TableRow key={payment.id}>
                      <TableCell className="font-mono text-xs">
                        #{payment.id}
                      </TableCell>
                      <TableCell className="font-medium">
                        {payment.member?.name || "-"}
                        <br />
                        <span className="text-xs text-gray-400">
                          {payment.member?.membership_no}
                        </span>
                      </TableCell>
                      <TableCell className="text-xs">
                        {payment.request ? (
                          <span className="font-mono">Request #{payment.request.id}</span>
                        ) : payment.ledger_transaction ? (
                          <div>
                            <p className="font-semibold text-slate-700 dark:text-slate-200">
                              {payment.ledger_transaction.type === "legacy_gasti_fee"
                                ? "Past Gasti fee"
                                : "Annual Gasti fee"}
                            </p>
                            <p className="text-[11px] text-slate-400">
                              FY {payment.ledger_transaction.fiscal_year?.name || "-"}
                            </p>
                          </div>
                        ) : (
                          "-"
                        )}
                      </TableCell>
                      <TableCell className="font-semibold text-emerald-600">
                        {formatCurrency(payment.amount)}
                      </TableCell>
                      <TableCell>
                        <Badge status={payment.payment_method}>
                          {payment.payment_method}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        <Badge
                          status={
                            payment.status === "paid" ? "success" : "warning"
                          }
                        >
                          {payment.status}
                        </Badge>
                      </TableCell>
                      <TableCell>{formatDate(payment.created_at)}</TableCell>
                      <TableCell>
                        <div className="flex items-center gap-1">
                          <Button
                            variant="ghost"
                            size="icon"
                            onClick={() => openView(payment)}
                            title="View Details"
                          >
                            <Eye size={15} />
                          </Button>
                          {payment.status === "paid" && (
                            <Button
                              variant="ghost"
                              size="icon"
                              onClick={() => handleDownloadReceipt(payment.id)}
                              title="Download receipt"
                              className="text-blue-600 hover:text-blue-700"
                            >
                              <Receipt size={16} />
                            </Button>
                          )}
                          {payment.payment_method === "esewa" && payment.status === "pending" && (
                            <Button
                              variant="ghost"
                              size="icon"
                              onClick={() => handleCheckEsewaStatus(payment.id)}
                              disabled={checkingPaymentId === payment.id}
                              title="Check eSewa status"
                              className="text-violet-600 hover:text-violet-700"
                            >
                              <RefreshCw
                                size={16}
                                className={checkingPaymentId === payment.id ? "animate-spin" : ""}
                              />
                            </Button>
                          )}
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-2">
          <Button
            variant="outline"
            size="sm"
            disabled={page <= 1}
            onClick={() => setPage(page - 1)}
          >
            Previous
          </Button>
          <span className="text-sm text-gray-500">
            Page {page} of {totalPages}
          </span>
          <Button
            variant="outline"
            size="sm"
            disabled={page >= totalPages}
            onClick={() => setPage(page + 1)}
          >
            Next
          </Button>
        </div>
      )}

      {/* Create Payment Modal */}
      <Modal
        isOpen={showCreateModal}
        onClose={() => {
          setShowCreateModal(false);
          setForm(emptyPayment);
        }}
        title="Record Cash Payment"
        size="lg"
        footer={
          <>
            <Button variant="outline" onClick={() => setShowCreateModal(false)}>
              Cancel
            </Button>
            <Button onClick={handleCreate} isLoading={saving}>
              Record Cash Payment
            </Button>
          </>
        }
      >
        <div className="space-y-4">
          {/* Member Selection - Only show for admin/staff */}
          {canRecordCash && (
            <Select
              label="Select Member"
              value={form.member_id}
              onChange={(e) =>
                setForm({ ...form, member_id: e.target.value, request_id: "" })
              }
              options={[
                { value: "", label: "Select a member..." },
                ...members.map((m) => ({
                  value: String(m.id),
                  label: `${m.name} (${m.membership_no}) - ${m.phone || "No phone"}`,
                })),
              ]}
              placeholder="Select member"
              required
            />
          )}

          {/* Show selected member name for context */}
          {form.member_id && (
            <div className="text-sm text-emerald-600 bg-emerald-50 dark:bg-emerald-900/20 p-2 rounded-lg">
              Recording payment for: <strong>{getSelectedMemberName()}</strong>
            </div>
          )}

          <Select
            label="Approved Request"
            value={form.request_id}
            onChange={(e) => setForm({ ...form, request_id: e.target.value })}
            options={[
              { value: "", label: "Select an approved request" },
              ...(form.member_id ? filteredRequests : []).map((req) => ({
                value: String(req.id),
                label: `#${req.id} - ${req.resource_item?.name} (${formatCurrency(req.total_amount || 0)})`,
              })),
            ]}
            placeholder={
              form.member_id
                ? "Select a request"
                : "Select a member first to see requests"
            }
            disabled={!form.member_id && !isMember}
          />

          {getSelectedRequestAmount() && (
            <div className="text-sm text-emerald-600 -mt-2">
              Request Total: {formatCurrency(getSelectedRequestAmount())}
            </div>
          )}

          <Input
            label="Amount"
            type="number"
            step="0.01"
            value={form.amount}
            onChange={(e) => setForm({ ...form, amount: e.target.value })}
            placeholder="Enter amount"
            required
          />

          <div className="rounded-lg border border-emerald-200 bg-emerald-50 p-3 text-sm text-emerald-700 dark:border-emerald-900/50 dark:bg-emerald-900/20 dark:text-emerald-300">
            Payment method: <strong>Cash</strong>. Only an administrator can
            record cash payments. Members make online payments from the approved
            request screen or yearly Gasti-fee table through eSewa.
          </div>

          <Textarea
            label="Remarks (optional)"
            value={form.remarks}
            onChange={(e) => setForm({ ...form, remarks: e.target.value })}
            placeholder="Cash receipt or payment note"
            rows={3}
          />
        </div>
      </Modal>

      {/* View Payment Modal */}
      <Modal
        isOpen={showViewModal}
        onClose={() => {
          setShowViewModal(false);
          setViewingPayment(null);
        }}
        title="Payment Details"
        size="lg"
      >
        {viewingPayment && (
          <div className="space-y-6">
            {/* Header */}
            <div className="flex items-center justify-between">
              <Badge
                status={
                  viewingPayment.status === "paid" ? "success" : "warning"
                }
                className="text-sm px-3 py-1"
              >
                {viewingPayment.status?.toUpperCase()}
              </Badge>
              <span className="text-xs font-mono text-gray-500">
                #{viewingPayment.id}
              </span>
            </div>

            {/* Amount */}
            <div className="text-center">
              <p className="text-3xl font-bold text-emerald-600">
                {formatCurrency(viewingPayment.amount)}
              </p>
              <p className="text-sm text-gray-500">Payment Amount</p>
            </div>

            {/* Member Info */}
            <div className="border-t border-gray-200 dark:border-white/10 pt-4">
              <div className="flex items-start gap-3">
                <UserIcon size={18} className="text-gray-400 mt-0.5" />
                <div>
                  <p className="text-xs text-gray-400">Member</p>
                  <p className="font-medium text-gray-900 dark:text-gray-100">
                    {viewingPayment.member?.name}
                  </p>
                  <p className="text-xs text-gray-500">
                    Membership: {viewingPayment.member?.membership_no}
                  </p>
                </div>
              </div>
            </div>

            {/* Payment Details */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="flex items-start gap-3">
                <CreditCard size={18} className="text-gray-400 mt-0.5" />
                <div>
                  <p className="text-xs text-gray-400">Payment Method</p>
                  <p className="text-gray-900 dark:text-gray-100 capitalize">
                    {viewingPayment.payment_method}
                  </p>
                </div>
              </div>
              {viewingPayment.transaction_id && (
                <div className="flex items-start gap-3">
                  <Building2 size={18} className="text-gray-400 mt-0.5" />
                  <div>
                    <p className="text-xs text-gray-400">Transaction ID</p>
                    <p className="text-gray-900 dark:text-gray-100 font-mono text-sm">
                      {viewingPayment.transaction_id}
                    </p>
                  </div>
                </div>
              )}
            </div>

            {viewingPayment.ledger_transaction && (
              <div className="rounded-lg bg-emerald-50 p-4 dark:bg-emerald-950/30">
                <p className="text-xs text-emerald-600 dark:text-emerald-400">Related Gasti / Membership Fee</p>
                <p className="mt-1 font-bold text-slate-900 dark:text-white">
                  Fiscal Year {viewingPayment.ledger_transaction.fiscal_year?.name || "-"}
                </p>
                <p className="mt-1 text-sm text-slate-600 dark:text-slate-300">
                  {viewingPayment.ledger_transaction.type === "legacy_gasti_fee"
                    ? "Past Gasti fee from physical register"
                    : "Automatically assigned annual Gasti / Membership fee"}
                </p>
              </div>
            )}

            {/* Related Request */}
            {viewingPayment.request && (
              <div className="bg-gray-50 dark:bg-gray-800/50 rounded-lg p-4">
                <p className="text-xs text-gray-400 mb-1">Related Request</p>
                <p className="font-medium">
                  #{viewingPayment.request.id} -{" "}
                  {viewingPayment.request.resource_item?.name}
                </p>
                <p className="text-sm text-gray-600 dark:text-gray-400">
                  Quantity: {viewingPayment.request.quantity_approved}{" "}
                  {viewingPayment.request.resource_item?.type?.unit}
                </p>
                {viewingPayment.request.total_amount && (
                  <p className="text-sm font-medium text-emerald-600 mt-1">
                    Total: {formatCurrency(viewingPayment.request.total_amount)}
                  </p>
                )}
              </div>
            )}

            {viewingPayment.status === "paid" && (
              <Button
                variant="outline"
                onClick={() => handleDownloadReceipt(viewingPayment.id)}
              >
                <Receipt size={16} /> Download Payment Receipt
              </Button>
            )}

            {/* Dates */}
            <div className="border-t border-gray-200 dark:border-white/10 pt-4">
              <div className="flex items-start gap-3">
                <Receipt size={18} className="text-gray-400 mt-0.5" />
                <div>
                  <p className="text-xs text-gray-400">Payment Date</p>
                  <p className="text-gray-900 dark:text-gray-100">
                    {formatDate(
                      viewingPayment.paid_at || viewingPayment.created_at,
                    )}
                  </p>
                </div>
              </div>
            </div>
          </div>
        )}
      </Modal>
    </div>
  );
}
