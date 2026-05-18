import { useState, useEffect, useCallback } from "react";
import { useSelector } from "react-redux";
import {
  Plus,
  Search,
  CheckCircle,
  XCircle,
  FileText,
  RefreshCw,
  Eye,
  AlertTriangle,
  Package,
  Calendar,
  User,
  DollarSign,
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

const emptyRequest = {
  member_id: "",
  resource_item_id: "",
  fiscal_year_id: "",
  quantity_requested: "",
  remarks: "",
};

export default function Requests() {
  const { user } = useSelector((state) => state.auth);
  const { addToast } = useToast();
  const isMember = user?.role === "member";
  const canApprove = user?.role === "admin" || user?.role === "staff";

  const [requests, setRequests] = useState([]);
  const [members, setMembers] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState("");
  const [fiscalYearFilter, setFiscalYearFilter] = useState("");
  const [activeTab, setActiveTab] = useState("all");
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [refreshing, setRefreshing] = useState(false);
  const [stats, setStats] = useState(null);

  // Modal states
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showActionModal, setShowActionModal] = useState(false);
  const [showViewModal, setShowViewModal] = useState(false);
  const [actionType, setActionType] = useState("");
  const [selectedRequest, setSelectedRequest] = useState(null);
  const [viewingRequest, setViewingRequest] = useState(null);
  const [remark, setRemark] = useState("");
  const [approveQuantity, setApproveQuantity] = useState("");
  const [saving, setSaving] = useState(false);

  // Form state
  const [form, setForm] = useState(emptyRequest);
  const [resourceItems, setResourceItems] = useState([]);
  const [fiscalYears, setFiscalYears] = useState([]);

  const fetchMasterData = useCallback(async () => {
    try {
      const [itemsRes, fiscalYearsRes] = await Promise.all([
        api.getResourceItems(),
        api.getFiscalYears(),
      ]);
      if (itemsRes.success) setResourceItems(itemsRes.data || []);
      if (fiscalYearsRes.success) setFiscalYears(fiscalYearsRes.data || []);
    } catch (err) {
      console.error("Failed to fetch master data:", err);
    }
  }, []);

  const fetchMembers = useCallback(async () => {
    try {
      const res = await api.getMembers({ limit: 100 });
      if (res.success) setMembers(res.data || []);
    } catch (err) {
      console.error("Failed to fetch members:", err);
    }
  }, []);

  const fetchRequests = useCallback(async () => {
    setIsLoading(true);
    try {
      const fetchFn =
        activeTab === "my" && isMember
          ? api.getMyRequests.bind(api)
          : api.getRequests.bind(api);
      const res = await fetchFn({
        page,
        limit: 10,
        search: search || undefined,
        status: statusFilter || undefined,
        fiscal_year_id: fiscalYearFilter || undefined,
      });
      if (res.success) {
        setRequests(res.data || []);
        setTotalPages(res.meta?.total_pages || 1);
      }
    } catch (err) {
      addToast("Failed to load requests", "error");
    } finally {
      setIsLoading(false);
      setRefreshing(false);
    }
  }, [
    page,
    search,
    statusFilter,
    fiscalYearFilter,
    activeTab,
    isMember,
    addToast,
  ]);

  const fetchStats = useCallback(async () => {
    try {
      const res = await api.getRequestStats({
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
    if (!isMember && canApprove) {
      fetchMembers();
    }
  }, [fetchMasterData, fetchMembers, isMember, canApprove]);

  useEffect(() => {
    fetchRequests();
    if (canApprove) fetchStats();
  }, [fetchRequests, fetchStats, canApprove]);

  const handleCreate = async () => {
    setSaving(true);
    try {
      const payload = {
        member_id:
          !isMember && canApprove && form.member_id
            ? Number(form.member_id)
            : undefined,
        resource_item_id: Number(form.resource_item_id),
        fiscal_year_id: Number(form.fiscal_year_id),
        quantity_requested: Number(form.quantity_requested),
        remarks: form.remarks || null,
      };

      const res = await api.createRequest(payload);
      if (res.success) {
        addToast("Request submitted successfully", "success");
        setShowCreateModal(false);
        setForm(emptyRequest);
        fetchRequests();
        fetchStats();
      } else {
        addToast(res.message || "Failed to create request", "error");
      }
    } catch (err) {
      addToast(err.message || "Failed to create request", "error");
    } finally {
      setSaving(false);
    }
  };

  const handleApprove = async () => {
    if (!selectedRequest) return;
    setSaving(true);
    try {
      const payload = {
        remarks: remark || null,
      };
      if (approveQuantity) {
        payload.quantity_approved = Number(approveQuantity);
      }

      const res = await api.approveRequest(selectedRequest.id, payload);
      if (res.success) {
        addToast("Request approved successfully", "success");
        setShowActionModal(false);
        setSelectedRequest(null);
        setRemark("");
        setApproveQuantity("");
        fetchRequests();
        fetchStats();
      } else {
        addToast(res.message || "Failed to approve", "error");
      }
    } catch (err) {
      addToast(err.message || "Failed to approve", "error");
    } finally {
      setSaving(false);
    }
  };

  const handleReject = async () => {
    if (!selectedRequest) return;
    setSaving(true);
    try {
      const res = await api.rejectRequest(selectedRequest.id, remark);
      if (res.success) {
        addToast("Request rejected", "success");
        setShowActionModal(false);
        setSelectedRequest(null);
        setRemark("");
        fetchRequests();
        fetchStats();
      } else {
        addToast(res.message || "Failed to reject", "error");
      }
    } catch (err) {
      addToast(err.message || "Failed to reject", "error");
    } finally {
      setSaving(false);
    }
  };

  const openView = async (request) => {
    setViewingRequest(request);
    setShowViewModal(true);
  };

  const openActionModal = (request, type) => {
    setSelectedRequest(request);
    setActionType(type);
    setRemark("");
    setApproveQuantity(
      type === "approve" ? String(request.quantity_requested) : "",
    );
    setShowActionModal(true);
  };

  const tabs = [
    { id: "all", label: "All Requests" },
    ...(isMember ? [{ id: "my", label: "My Requests" }] : []),
  ];

  const getSelectedItemDetails = () => {
    const item = resourceItems.find(
      (i) => i.id === Number(form.resource_item_id),
    );
    return item;
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">
            Resource Requests
          </h1>
          <p className="text-sm text-gray-500 dark:text-gray-400">
            Submit and manage forest resource requests
          </p>
        </div>
        <div className="flex gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => {
              setRefreshing(true);
              fetchRequests();
            }}
            isLoading={refreshing}
          >
            <RefreshCw size={14} className="mr-1" /> Refresh
          </Button>
          <Button
            onClick={() => {
              setForm(emptyRequest);
              setShowCreateModal(true);
            }}
          >
            <Plus size={16} /> New Request
          </Button>
        </div>
      </div>

      {/* Stats Cards - Admin only */}
      {canApprove && stats && (
        <div className="grid grid-cols-2 sm:grid-cols-5 gap-3">
          <Card>
            <CardContent className="p-3 text-center">
              <p className="text-2xl font-bold text-gray-900 dark:text-gray-100">
                {stats.total || 0}
              </p>
              <p className="text-xs text-gray-500">Total</p>
            </CardContent>
          </Card>
          <Card className="border-amber-200 bg-amber-50 dark:bg-amber-900/20">
            <CardContent className="p-3 text-center">
              <p className="text-2xl font-bold text-amber-600">
                {stats.pending || 0}
              </p>
              <p className="text-xs text-amber-600">Pending</p>
            </CardContent>
          </Card>
          <Card className="border-emerald-200 bg-emerald-50 dark:bg-emerald-900/20">
            <CardContent className="p-3 text-center">
              <p className="text-2xl font-bold text-emerald-600">
                {stats.approved || 0}
              </p>
              <p className="text-xs text-emerald-600">Approved</p>
            </CardContent>
          </Card>
          <Card className="border-blue-200 bg-blue-50 dark:bg-blue-900/20">
            <CardContent className="p-3 text-center">
              <p className="text-2xl font-bold text-blue-600">
                {stats.completed || 0}
              </p>
              <p className="text-xs text-blue-600">Completed</p>
            </CardContent>
          </Card>
          <Card className="border-red-200 bg-red-50 dark:bg-red-900/20">
            <CardContent className="p-3 text-center">
              <p className="text-2xl font-bold text-red-600">
                {stats.rejected || 0}
              </p>
              <p className="text-xs text-red-600">Rejected</p>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Tabs & Filters */}
      <Card>
        <CardContent className="p-4">
          <div className="flex flex-col sm:flex-row gap-3">
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
            <div className="flex-1 relative">
              <Search
                size={16}
                className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400"
              />
              <input
                type="text"
                placeholder="Search by member name or membership no..."
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
                { value: "approved", label: "Approved" },
                { value: "rejected", label: "Rejected" },
                { value: "completed", label: "Completed" },
              ]}
              className="w-36"
            />
            {canApprove && (
              <Select
                value={fiscalYearFilter}
                onChange={(e) => {
                  setFiscalYearFilter(e.target.value);
                  setPage(1);
                }}
                options={[
                  { value: "", label: "All Fiscal Years" },
                  ...fiscalYears.map((fy) => ({
                    value: String(fy.id),
                    label: fy.name,
                  })),
                ]}
                className="w-40"
              />
            )}
          </div>
        </CardContent>
      </Card>

      {/* Requests Table */}
      <Card>
        <CardContent className="p-0">
          {isLoading ? (
            <LoadingSpinner text="Loading requests..." />
          ) : requests.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 text-gray-400">
              <FileText size={48} className="mb-3 opacity-30" />
              <p className="text-lg font-medium">No requests found</p>
              <Button
                variant="outline"
                className="mt-4"
                onClick={() => setShowCreateModal(true)}
              >
                <Plus size={14} /> Create your first request
              </Button>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>ID</TableHead>
                    <TableHead>Member</TableHead>
                    <TableHead>Resource</TableHead>
                    <TableHead>Requested</TableHead>
                    <TableHead>Approved</TableHead>
                    <TableHead>Total Amount</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Date</TableHead>
                    <TableHead>Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {requests.map((req) => (
                    <TableRow key={req.id}>
                      <TableCell className="font-mono text-xs">
                        #{req.id}
                      </TableCell>
                      <TableCell className="font-medium">
                        {req.member?.name || "-"}
                        <br />
                        <span className="text-xs text-gray-400">
                          {req.member?.membership_no}
                        </span>
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center gap-2">
                          <Package size={14} className="text-gray-400" />
                          {req.resource_item?.name || "-"}
                        </div>
                        <span className="text-xs text-gray-400">
                          {req.resource_item?.type?.unit}
                        </span>
                      </TableCell>
                      <TableCell>{req.quantity_requested}</TableCell>
                      <TableCell>{req.quantity_approved || "-"}</TableCell>
                      <TableCell className="font-semibold">
                        {req.total_amount
                          ? formatCurrency(req.total_amount)
                          : "-"}
                      </TableCell>
                      <TableCell>
                        <Badge status={req.status} />
                      </TableCell>
                      <TableCell>{formatDate(req.requested_at)}</TableCell>
                      <TableCell>
                        <div className="flex items-center gap-1">
                          <Button
                            variant="ghost"
                            size="icon"
                            onClick={() => openView(req)}
                            title="View Details"
                          >
                            <Eye size={15} />
                          </Button>
                          {canApprove && req.status === "pending" && (
                            <>
                              <Button
                                variant="ghost"
                                size="icon"
                                onClick={() => openActionModal(req, "approve")}
                                title="Approve"
                                className="text-emerald-600 hover:text-emerald-700"
                              >
                                <CheckCircle size={16} />
                              </Button>
                              <Button
                                variant="ghost"
                                size="icon"
                                onClick={() => openActionModal(req, "reject")}
                                title="Reject"
                                className="text-red-600 hover:text-red-700"
                              >
                                <XCircle size={16} />
                              </Button>
                            </>
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

      {/* Create Request Modal */}
      <Modal
        isOpen={showCreateModal}
        onClose={() => {
          setShowCreateModal(false);
          setForm(emptyRequest);
        }}
        title="New Resource Request"
        size="lg"
        footer={
          <>
            <Button variant="outline" onClick={() => setShowCreateModal(false)}>
              Cancel
            </Button>
            <Button onClick={handleCreate} isLoading={saving}>
              Submit Request
            </Button>
          </>
        }
      >
        <div className="space-y-4">
          {/* Member Selection - Only show for admin/staff */}
          {!isMember && canApprove && (
            <Select
              label="Select Member"
              value={form.member_id}
              onChange={(e) => setForm({ ...form, member_id: e.target.value })}
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

          <Select
            label="Fiscal Year"
            value={form.fiscal_year_id}
            onChange={(e) =>
              setForm({ ...form, fiscal_year_id: e.target.value })
            }
            options={fiscalYears.map((fy) => ({
              value: String(fy.id),
              label: fy.name,
            }))}
            placeholder="Select fiscal year"
            required
          />

          <Select
            label="Resource Item"
            value={form.resource_item_id}
            onChange={(e) =>
              setForm({ ...form, resource_item_id: e.target.value })
            }
            options={resourceItems.map((item) => ({
              value: String(item.id),
              label: `${item.name} (${item.type?.name || "Unknown"} - ${item.type?.unit || "unit"})`,
            }))}
            placeholder="Select resource item"
            required
          />

          {getSelectedItemDetails() && (
            <div className="text-xs text-gray-500 -mt-2">
              Unit: {getSelectedItemDetails()?.type?.unit}
            </div>
          )}

          <Input
            label="Quantity Requested"
            type="number"
            step="0.01"
            value={form.quantity_requested}
            onChange={(e) =>
              setForm({ ...form, quantity_requested: e.target.value })
            }
            placeholder="Enter quantity"
            required
          />

          <Textarea
            label="Remarks / Purpose"
            value={form.remarks}
            onChange={(e) => setForm({ ...form, remarks: e.target.value })}
            placeholder="Describe the purpose of this request"
            rows={3}
          />

          {/* Info message for admin/staff */}
          {!isMember && canApprove && (
            <div className="text-xs text-blue-600 bg-blue-50 dark:bg-blue-900/20 p-3 rounded-lg">
              ℹ️ As an administrator, you are creating this request on behalf of
              the selected member. The member will be notified about this
              request.
            </div>
          )}
        </div>
      </Modal>

      {/* Approve/Reject Modal */}
      <Modal
        isOpen={showActionModal}
        onClose={() => {
          setShowActionModal(false);
          setSelectedRequest(null);
          setRemark("");
          setApproveQuantity("");
        }}
        title={actionType === "approve" ? "Approve Request" : "Reject Request"}
        footer={
          <>
            <Button variant="outline" onClick={() => setShowActionModal(false)}>
              Cancel
            </Button>
            <Button
              variant={actionType === "approve" ? "success" : "danger"}
              onClick={actionType === "approve" ? handleApprove : handleReject}
              isLoading={saving}
            >
              {actionType === "approve" ? "Approve" : "Reject"}
            </Button>
          </>
        }
      >
        <div className="space-y-4">
          {selectedRequest && (
            <div className="p-3 bg-gray-50 dark:bg-gray-800/50 rounded-lg">
              <p className="text-sm font-medium">
                {selectedRequest.resource_item?.name}
              </p>
              <p className="text-sm text-gray-600 dark:text-gray-400">
                Requested by: {selectedRequest.member?.name} | Qty:{" "}
                {selectedRequest.quantity_requested}{" "}
                {selectedRequest.resource_item?.type?.unit}
              </p>
            </div>
          )}
          {actionType === "approve" && (
            <Input
              label="Approved Quantity"
              type="number"
              step="0.01"
              value={approveQuantity}
              onChange={(e) => setApproveQuantity(e.target.value)}
              placeholder={`Max: ${selectedRequest?.quantity_requested || 0}`}
            />
          )}
          <Textarea
            label="Remarks (optional)"
            value={remark}
            onChange={(e) => setRemark(e.target.value)}
            placeholder="Add any remarks..."
            rows={3}
          />
        </div>
      </Modal>

      {/* View Request Modal */}
      <Modal
        isOpen={showViewModal}
        onClose={() => {
          setShowViewModal(false);
          setViewingRequest(null);
        }}
        title="Request Details"
        size="lg"
      >
        {viewingRequest && (
          <div className="space-y-6">
            {/* Header */}
            <div className="flex items-center justify-between">
              <Badge
                status={viewingRequest.status}
                className="text-sm px-3 py-1"
              >
                {viewingRequest.status?.toUpperCase()}
              </Badge>
              <span className="text-xs font-mono text-gray-500">
                #{viewingRequest.id}
              </span>
            </div>

            {/* Member Info */}
            <div className="border-b border-gray-200 dark:border-white/10 pb-4">
              <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100">
                {viewingRequest.member?.name}
              </h3>
              <p className="text-sm text-gray-500">
                Membership: {viewingRequest.member?.membership_no}
              </p>
            </div>

            {/* Request Details */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="flex items-start gap-3">
                <Package size={18} className="text-gray-400 mt-0.5" />
                <div>
                  <p className="text-xs text-gray-400">Resource</p>
                  <p className="text-gray-900 dark:text-gray-100">
                    {viewingRequest.resource_item?.name}
                  </p>
                  <p className="text-xs text-gray-500">
                    {viewingRequest.resource_item?.type?.name} |{" "}
                    {viewingRequest.resource_item?.type?.unit}
                  </p>
                </div>
              </div>
              <div className="flex items-start gap-3">
                <Calendar size={18} className="text-gray-400 mt-0.5" />
                <div>
                  <p className="text-xs text-gray-400">Fiscal Year</p>
                  <p className="text-gray-900 dark:text-gray-100">
                    {viewingRequest.fiscal_year?.name}
                  </p>
                </div>
              </div>
            </div>

            {/* Quantities */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="bg-gray-50 dark:bg-gray-800/50 rounded-lg p-3 text-center">
                <p className="text-xs text-gray-400">Requested Quantity</p>
                <p className="text-xl font-bold text-gray-900 dark:text-gray-100">
                  {viewingRequest.quantity_requested}
                </p>
              </div>
              <div className="bg-gray-50 dark:bg-gray-800/50 rounded-lg p-3 text-center">
                <p className="text-xs text-gray-400">Approved Quantity</p>
                <p className="text-xl font-bold text-emerald-600">
                  {viewingRequest.quantity_approved || "-"}
                </p>
              </div>
            </div>

            {/* Financial Details */}
            {viewingRequest.total_amount && (
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="flex items-start gap-3">
                  <DollarSign size={18} className="text-gray-400 mt-0.5" />
                  <div>
                    <p className="text-xs text-gray-400">Rate Per Unit</p>
                    <p className="text-gray-900 dark:text-gray-100">
                      {formatCurrency(viewingRequest.rate_per_unit || 0)}
                    </p>
                  </div>
                </div>
                <div className="flex items-start gap-3">
                  <DollarSign size={18} className="text-gray-400 mt-0.5" />
                  <div>
                    <p className="text-xs text-gray-400">Total Amount</p>
                    <p className="text-xl font-bold text-emerald-600">
                      {formatCurrency(viewingRequest.total_amount)}
                    </p>
                  </div>
                </div>
              </div>
            )}

            {/* Remarks */}
            {viewingRequest.remarks && (
              <div className="bg-gray-50 dark:bg-gray-800/50 rounded-lg p-4">
                <p className="text-xs text-gray-400 mb-1">Remarks</p>
                <p className="text-gray-700 dark:text-gray-300">
                  {viewingRequest.remarks}
                </p>
              </div>
            )}

            {/* Approval Info */}
            {viewingRequest.approver && (
              <div className="border-t border-gray-200 dark:border-white/10 pt-4">
                <div className="flex items-start gap-3">
                  <User size={18} className="text-gray-400 mt-0.5" />
                  <div>
                    <p className="text-xs text-gray-400">
                      {viewingRequest.status === "approved"
                        ? "Approved"
                        : "Processed"}{" "}
                      By
                    </p>
                    <p className="text-gray-900 dark:text-gray-100">
                      {viewingRequest.approver?.name}
                    </p>
                    <p className="text-xs text-gray-500">
                      {formatDate(viewingRequest.approved_at)}
                    </p>
                  </div>
                </div>
              </div>
            )}

            {/* Payment Info */}
            {viewingRequest.payments && viewingRequest.payments.length > 0 && (
              <div className="border-t border-gray-200 dark:border-white/10 pt-4">
                <h4 className="font-medium text-gray-900 dark:text-gray-100 mb-2">
                  Payments
                </h4>
                <div className="space-y-2">
                  {viewingRequest.payments.map((payment) => (
                    <div
                      key={payment.id}
                      className="flex items-center justify-between p-2 bg-gray-50 dark:bg-gray-800/50 rounded-lg"
                    >
                      <div>
                        <p className="text-sm font-medium">
                          {formatCurrency(payment.amount)}
                        </p>
                        <p className="text-xs text-gray-500">
                          {payment.payment_method}
                        </p>
                      </div>
                      <Badge
                        status={
                          payment.status === "paid" ? "success" : "warning"
                        }
                      >
                        {payment.status}
                      </Badge>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        )}
      </Modal>
    </div>
  );
}
