import { useState, useEffect, useCallback, useMemo } from "react";
import { useSelector } from "react-redux";
import {
  Search,
  ArrowUpRight,
  ArrowDownLeft,
  Wallet,
  BookOpen,
  RefreshCw,
  Eye,
  TrendingUp,
  TrendingDown,
  Calendar,
  User,
  FileText,
} from "lucide-react";
import { api } from "../../services/api";
import { Card, CardContent } from "../../components/ui/Card";
import Button from "../../components/ui/Button";
import Input from "../../components/ui/Input";
import Select from "../../components/ui/Select";
import Badge from "../../components/ui/Badge";
import Modal from "../../components/ui/Modal";
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

export default function Transactions() {
  const { user } = useSelector((state) => state.auth);
  const { addToast } = useToast();
  const isMember = user?.role === "member";

  const [transactions, setTransactions] = useState([]);
  const [summary, setSummary] = useState({
    total_revenue: 0,
    resource_sales: 0,
    membership_fees: 0,
    total_collected: 0,
    total_remaining: 0,
    total_expenses: 0,
    total_fines_collected: 0,
    net_balance: 0,
  });
  const [isLoading, setIsLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [typeFilter, setTypeFilter] = useState("");
  const [fiscalYearFilter, setFiscalYearFilter] = useState("");
  const [activeTab, setActiveTab] = useState(isMember ? "my" : "all");
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [refreshing, setRefreshing] = useState(false);
  const [showViewModal, setShowViewModal] = useState(false);
  const [viewingTransaction, setViewingTransaction] = useState(null);
  const [fiscalYears, setFiscalYears] = useState([]);

  const fetchMasterData = useCallback(async () => {
    try {
      const res = await api.getFiscalYears();
      if (res.success) {
        const years = res.data || [];
        setFiscalYears(years);
        if (!isMember) {
          setFiscalYearFilter(
            (current) => current || getActiveFiscalYearId(years),
          );
        }
      }
    } catch (err) {
      console.error("Failed to fetch fiscal years:", err);
    }
  }, [isMember]);

  const fetchTransactions = useCallback(async () => {
    setIsLoading(true);
    try {
      const fetchFn = isMember
        ? api.getMyTransactions.bind(api)
        : api.getTransactions.bind(api);
      const res = await fetchFn({
        page,
        per_page: isMember ? 100 : 10,
        search: search || undefined,
        type: typeFilter || undefined,
        fiscal_year_id: fiscalYearFilter || undefined,
      });
      if (res.success) {
        setTransactions(res.data || []);
        setTotalPages(res.meta?.total_pages || 1);
      }
    } catch (_err) {
      addToast("Failed to load transactions", "error");
    } finally {
      setIsLoading(false);
      setRefreshing(false);
    }
  }, [
    page,
    search,
    typeFilter,
    fiscalYearFilter,
    isMember,
    addToast,
  ]);

  const fetchSummary = useCallback(async () => {
    if (fiscalYearFilter) {
      try {
        const res = await api.getTransactionSummary(fiscalYearFilter);
        if (res.success && res.data) setSummary(res.data);
      } catch (err) {
        console.error("Failed to fetch summary:", err);
      }
    }
  }, [fiscalYearFilter]);

  useEffect(() => {
    fetchMasterData();
  }, [fetchMasterData]);

  useEffect(() => {
    fetchTransactions();
    if (!isMember && fiscalYearFilter) {
      fetchSummary();
    }
  }, [fetchTransactions, fetchSummary, isMember, fiscalYearFilter]);

  const openView = async (transaction) => {
    setViewingTransaction(transaction);
    setShowViewModal(true);
  };

  const visibleTransactions = useMemo(() => {
    if (!isMember) return transactions;
    const term = search.trim().toLowerCase();
    return transactions.filter((transaction) => {
      const searchable = [
        transaction.id,
        transaction.receipt_no,
        transaction.type,
        transaction.resource_item?.name,
      ]
        .filter((value) => value != null)
        .join(" ")
        .toLowerCase();
      const matchesSearch = !term || searchable.includes(term);
      const matchesType = !typeFilter || transaction.type === typeFilter;
      return matchesSearch && matchesType;
    });
  }, [isMember, transactions, search, typeFilter]);

  const tabs = isMember
    ? [{ id: "my", label: "My Transactions" }]
    : [{ id: "all", label: "All Transactions" }];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">
            {isMember ? "My Ledger" : "Transactions"}
          </h1>
          <p className="text-sm text-gray-500 dark:text-gray-400">
            {isMember ? "Review your charges, payments and receipts" : "View complete financial ledger"}
          </p>
        </div>
        <div className="flex gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => {
              setRefreshing(true);
              fetchTransactions();
            }}
            isLoading={refreshing}
          >
            <RefreshCw size={14} className="mr-1" /> Refresh
          </Button>
        </div>
      </div>

      {/* Summary Cards - Admin only */}
      {!isMember && fiscalYearFilter && (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          <Card className="border-emerald-200 bg-emerald-50 dark:bg-emerald-900/20">
            <CardContent className="p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-gray-600 dark:text-gray-400">
                    Total Revenue
                  </p>
                  <p className="text-2xl font-bold text-emerald-600">
                    {formatCurrency(summary.total_revenue || 0)}
                  </p>
                </div>
                <TrendingUp size={24} className="text-emerald-500" />
              </div>
              <div className="mt-2 text-xs text-gray-500">
                <span>
                  Sales: {formatCurrency(summary.resource_sales || 0)}
                </span>
                <span className="mx-2">|</span>
                <span>
                  Fees: {formatCurrency(summary.membership_fees || 0)}
                </span>
              </div>
            </CardContent>
          </Card>

          <Card className="border-red-200 bg-red-50 dark:bg-red-900/20">
            <CardContent className="p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-gray-600 dark:text-gray-400">
                    Total Expenses
                  </p>
                  <p className="text-2xl font-bold text-red-600">
                    {formatCurrency(summary.total_expenses || 0)}
                  </p>
                </div>
                <TrendingDown size={24} className="text-red-500" />
              </div>
            </CardContent>
          </Card>

          <Card className="border-blue-200 bg-blue-50 dark:bg-blue-900/20">
            <CardContent className="p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-gray-600 dark:text-gray-400">
                    Fines Collected
                  </p>
                  <p className="text-2xl font-bold text-blue-600">
                    {formatCurrency(summary.total_fines_collected || 0)}
                  </p>
                </div>
                <Wallet size={24} className="text-blue-500" />
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-gray-600 dark:text-gray-400">
                    Net Balance
                  </p>
                  <p
                    className={`text-2xl font-bold ${(summary.net_balance || 0) >= 0 ? "text-emerald-600" : "text-red-600"}`}
                  >
                    {formatCurrency(Math.abs(summary.net_balance || 0))}
                    {(summary.net_balance || 0) >= 0 ? " CR" : " DR"}
                  </p>
                </div>
                <Wallet size={24} className="text-gray-400" />
              </div>
              <div className="mt-2 text-xs text-gray-500">
                <span>
                  Collected: {formatCurrency(summary.total_collected || 0)}
                </span>
                <span className="mx-2">|</span>
                <span>
                  Remaining: {formatCurrency(summary.total_remaining || 0)}
                </span>
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Filters */}
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
              value={typeFilter}
              onChange={(e) => {
                setTypeFilter(e.target.value);
                setPage(1);
              }}
              options={[
                { value: "", label: "All Types" },
                { value: "resource_sale", label: "Resource Sale" },
                { value: "membership_fee", label: "Membership Fee" },
              ]}
              className="w-40"
            />
            {!isMember && (
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

      {/* Transactions Table */}
      <Card>
        <CardContent className="p-0">
          {isLoading ? (
            <LoadingSpinner text="Loading transactions..." />
          ) : visibleTransactions.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 text-gray-400">
              <BookOpen size={48} className="mb-3 opacity-30" />
              <p className="text-lg font-medium">No transactions found</p>
              {!isMember && !fiscalYearFilter && (
                <p className="text-sm mt-2">
                  Please select a fiscal year to view transactions
                </p>
              )}
            </div>
          ) : (
            <div className="overflow-x-auto">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Receipt #</TableHead>
                    <TableHead>Member</TableHead>
                    <TableHead>Type</TableHead>
                    <TableHead>Description</TableHead>
                    <TableHead>Amount</TableHead>
                    <TableHead>Paid</TableHead>
                    <TableHead>Balance</TableHead>
                    <TableHead>Date</TableHead>
                    <TableHead>Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {visibleTransactions.map((txn) => (
                    <TableRow key={txn.id}>
                      <TableCell className="font-mono text-xs font-medium">
                        {txn.receipt_no}
                      </TableCell>
                      <TableCell className="font-medium">
                        {txn.member?.name || "-"}
                        <br />
                        <span className="text-xs text-gray-400">
                          {txn.member?.membership_no}
                        </span>
                      </TableCell>
                      <TableCell>
                        <Badge
                          status={
                            txn.type === "resource_sale" ? "active" : "info"
                          }
                        >
                          {txn.type === "resource_sale"
                            ? "Resource Sale"
                            : "Membership Fee"}
                        </Badge>
                      </TableCell>
                      <TableCell className="max-w-50 truncate">
                        {txn.resource_item?.name || "Membership Fee"}
                      </TableCell>
                      <TableCell className="font-semibold text-gray-900 dark:text-gray-100">
                        {formatCurrency(txn.total_amount)}
                      </TableCell>
                      <TableCell className="font-semibold text-emerald-600">
                        {formatCurrency(txn.amount_paid)}
                      </TableCell>
                      <TableCell className="font-semibold">
                        {txn.amount_remaining > 0 ? (
                          <span className="text-amber-600">
                            {formatCurrency(txn.amount_remaining)}
                          </span>
                        ) : (
                          <span className="text-emerald-600">Paid</span>
                        )}
                      </TableCell>
                      <TableCell>{formatDate(txn.date)}</TableCell>
                      <TableCell>
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => openView(txn)}
                          title="View Details"
                        >
                          <Eye size={15} />
                        </Button>
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

      {/* View Transaction Modal */}
      <Modal
        isOpen={showViewModal}
        onClose={() => {
          setShowViewModal(false);
          setViewingTransaction(null);
        }}
        title="Transaction Details"
        size="lg"
      >
        {viewingTransaction && (
          <div className="space-y-6">
            {/* Header */}
            <div className="flex items-center justify-between">
              <Badge
                status={
                  viewingTransaction.type === "resource_sale"
                    ? "active"
                    : "info"
                }
                className="text-sm px-3 py-1"
              >
                {viewingTransaction.type === "resource_sale"
                  ? "Resource Sale"
                  : "Membership Fee"}
              </Badge>
              <span className="text-xs font-mono text-gray-500">
                Receipt: {viewingTransaction.receipt_no}
              </span>
            </div>

            {/* Amount */}
            <div className="text-center">
              <p className="text-3xl font-bold text-emerald-600">
                {formatCurrency(viewingTransaction.total_amount)}
              </p>
              <p className="text-sm text-gray-500">Total Amount</p>
            </div>

            {/* Payment Summary */}
            <div className="grid grid-cols-2 gap-4">
              <div className="bg-emerald-50 dark:bg-emerald-900/20 rounded-lg p-3 text-center">
                <p className="text-xs text-gray-500">Amount Paid</p>
                <p className="text-xl font-bold text-emerald-600">
                  {formatCurrency(viewingTransaction.amount_paid)}
                </p>
              </div>
              <div className="bg-amber-50 dark:bg-amber-900/20 rounded-lg p-3 text-center">
                <p className="text-xs text-gray-500">Remaining Balance</p>
                <p className="text-xl font-bold text-amber-600">
                  {formatCurrency(viewingTransaction.amount_remaining)}
                </p>
              </div>
            </div>

            {/* Member Info */}
            <div className="border-t border-gray-200 dark:border-white/10 pt-4">
              <div className="flex items-start gap-3">
                <User size={18} className="text-gray-400 mt-0.5" />
                <div>
                  <p className="text-xs text-gray-400">Member</p>
                  <p className="font-medium text-gray-900 dark:text-gray-100">
                    {viewingTransaction.member?.name}
                  </p>
                  <p className="text-xs text-gray-500">
                    Membership: {viewingTransaction.member?.membership_no}
                  </p>
                </div>
              </div>
            </div>

            {/* Transaction Details */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="flex items-start gap-3">
                <Calendar size={18} className="text-gray-400 mt-0.5" />
                <div>
                  <p className="text-xs text-gray-400">Transaction Date</p>
                  <p className="text-gray-900 dark:text-gray-100">
                    {formatDate(viewingTransaction.date)}
                  </p>
                </div>
              </div>
              <div className="flex items-start gap-3">
                <FileText size={18} className="text-gray-400 mt-0.5" />
                <div>
                  <p className="text-xs text-gray-400">Fiscal Year</p>
                  <p className="text-gray-900 dark:text-gray-100">
                    {viewingTransaction.fiscal_year?.name}
                  </p>
                </div>
              </div>
            </div>

            {/* Resource Details (if resource sale) */}
            {viewingTransaction.resource_item && (
              <div className="bg-gray-50 dark:bg-gray-800/50 rounded-lg p-4">
                <p className="text-xs text-gray-400 mb-2">Resource Details</p>
                <div className="grid grid-cols-3 gap-2 text-center">
                  <div>
                    <p className="text-xs text-gray-500">Item</p>
                    <p className="font-medium">
                      {viewingTransaction.resource_item.name}
                    </p>
                  </div>
                  <div>
                    <p className="text-xs text-gray-500">Quantity</p>
                    <p className="font-medium">
                      {viewingTransaction.quantity}{" "}
                      {viewingTransaction.resource_item.type?.unit}
                    </p>
                  </div>
                  <div>
                    <p className="text-xs text-gray-500">Rate</p>
                    <p className="font-medium">
                      {formatCurrency(viewingTransaction.rate_per_unit || 0)}
                    </p>
                  </div>
                </div>
              </div>
            )}

            {/* Remarks */}
            {viewingTransaction.remarks && (
              <div className="border-t border-gray-200 dark:border-white/10 pt-4">
                <p className="text-xs text-gray-400 mb-1">Remarks</p>
                <p className="text-gray-700 dark:text-gray-300">
                  {viewingTransaction.remarks}
                </p>
              </div>
            )}
          </div>
        )}
      </Modal>
    </div>
  );
}
