import { useState, useEffect, useCallback, useRef } from "react";
import { useSelector } from "react-redux";
import {
  Plus,
  Search,
  Edit2,
  Trash2,
  AlertTriangle,
  RefreshCw,
  Eye,
  Upload,
  Image as ImageIcon,
  Download,
  CheckCircle,
  XCircle,
  Calendar,
  User,
  FileText,
  DollarSign,
  AlertCircle,
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
import { formatCurrency, formatDate, getImageUrl } from "../../utils/helpers";
import {
  buildFiscalYearOptions,
  getActiveFiscalYearId,
} from "../../utils/fiscalYears";

const emptyFine = {
  fiscal_year_id: "",
  member_id: "",
  name: "",
  violation_type: "",
  description: "",
  fine_amount: "",
  incident_date: new Date().toISOString().split("T")[0],
  photo: "",
  remarks: "",
};

export default function Fines() {
  const { user } = useSelector((state) => state.auth);
  const { addToast } = useToast();
  const canEdit = user?.role === "admin" || user?.role === "staff";
  const isAdmin = user?.role === "admin";

  const [fines, setFines] = useState([]);
  const [members, setMembers] = useState([]);
  const [fiscalYears, setFiscalYears] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState("");
  const [fiscalYearFilter, setFiscalYearFilter] = useState("");
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [refreshing, setRefreshing] = useState(false);
  const [stats, setStats] = useState(null);

  // Modal states
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showEditModal, setShowEditModal] = useState(false);
  const [showViewModal, setShowViewModal] = useState(false);
  const [showStatusModal, setShowStatusModal] = useState(false);
  const [editingFine, setEditingFine] = useState(null);
  const [viewingFine, setViewingFine] = useState(null);
  const [selectedFine, setSelectedFine] = useState(null);
  const [form, setForm] = useState(emptyFine);
  const [statusForm, setStatusForm] = useState({ status: "paid", payment_reference: "" });
  const [saving, setSaving] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState(null);
  const fileInputRef = useRef(null);

  const fetchMasterData = useCallback(async () => {
    try {
      const [membersRes, fiscalYearsRes] = await Promise.all([
        api.getMembers({ per_page: 100 }),
        api.getFiscalYears(),
      ]);
      if (membersRes.success) setMembers(membersRes.data || []);
      if (fiscalYearsRes.success) {
        const years = fiscalYearsRes.data || [];
        const activeId = getActiveFiscalYearId(years);
        setFiscalYears(years);
        setFiscalYearFilter((current) => current || activeId);
        setForm((current) => ({
          ...current,
          fiscal_year_id: current.fiscal_year_id || activeId,
        }));
      }
    } catch (err) {
      console.error("Failed to fetch master data:", err);
    }
  }, []);

  const fetchFines = useCallback(async () => {
    setIsLoading(true);
    try {
      const res = await api.getFines({
        page,
        per_page: 10,
        search: search || undefined,
        status: statusFilter || undefined,
        fiscal_year_id: fiscalYearFilter || undefined,
      });
      if (res.success) {
        setFines(res.data || []);
        setTotalPages(res.meta?.total_pages || 1);
      }
    } catch (_err) {
      addToast("Failed to load fines", "error");
    } finally {
      setIsLoading(false);
      setRefreshing(false);
    }
  }, [page, search, statusFilter, fiscalYearFilter, addToast]);

  const fetchStats = useCallback(async () => {
    try {
      const res = await api.getFineStats({ fiscal_year_id: fiscalYearFilter || undefined });
      if (res.success) setStats(res.data);
    } catch (err) {
      console.error("Failed to fetch stats:", err);
    }
  }, [fiscalYearFilter]);

  useEffect(() => {
    fetchMasterData();
  }, [fetchMasterData]);

  useEffect(() => {
    fetchFines();
    fetchStats();
  }, [fetchFines, fetchStats]);

  const newFineForm = () => ({
    ...emptyFine,
    fiscal_year_id: getActiveFiscalYearId(fiscalYears),
    incident_date: new Date().toISOString().split("T")[0],
  });

  const openCreateFine = () => {
    setForm(newFineForm());
    setShowCreateModal(true);
  };

  const handleCreate = async () => {
    setSaving(true);
    try {
      const payload = {
        fiscal_year_id: Number(form.fiscal_year_id),
        member_id: form.member_id ? Number(form.member_id) : null,
        name: !form.member_id ? form.name : "",
        violation_type: form.violation_type,
        description: form.description || null,
        fine_amount: Number(form.fine_amount),
        incident_date: form.incident_date,
        photo: form.photo || null,
        remarks: form.remarks || null,
      };

      const res = await api.createFine(payload);
      if (res.success) {
        addToast("Fine created successfully", "success");
        setShowCreateModal(false);
        setForm(newFineForm());
        fetchFines();
        fetchStats();
      } else {
        addToast(res.message || "Failed to create fine", "error");
      }
    } catch (err) {
      addToast(err.message || "Failed to create fine", "error");
    } finally {
      setSaving(false);
    }
  };

  const handleUpdate = async () => {
    if (!editingFine) return;
    setSaving(true);
    try {
      const payload = {
        violation_type: form.violation_type,
        description: form.description || null,
        fine_amount: Number(form.fine_amount),
        incident_date: form.incident_date,
        photo: form.photo || null,
        remarks: form.remarks || null,
      };

      const res = await api.updateFine(editingFine.id, payload);
      if (res.success) {
        addToast("Fine updated successfully", "success");
        setShowEditModal(false);
        setEditingFine(null);
        setForm(newFineForm());
        fetchFines();
        fetchStats();
      } else {
        addToast(res.message || "Failed to update fine", "error");
      }
    } catch (err) {
      addToast(err.message || "Failed to update fine", "error");
    } finally {
      setSaving(false);
    }
  };

  const handleUpdateStatus = async () => {
    if (!selectedFine) return;
    setSaving(true);
    try {
      const res = await api.updateFineStatus(selectedFine.id, statusForm);
      if (res.success) {
        addToast(`Fine marked as ${statusForm.status}`, "success");
        setShowStatusModal(false);
        setSelectedFine(null);
        setStatusForm({ status: "paid", payment_reference: "" });
        fetchFines();
        fetchStats();
      } else {
        addToast(res.message || "Failed to update status", "error");
      }
    } catch (err) {
      addToast(err.message || "Failed to update status", "error");
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async () => {
    if (!deleteTarget) return;
    setSaving(true);
    try {
      const res = await api.deleteFine(deleteTarget);
      if (res.success) {
        addToast("Fine deleted successfully", "success");
        setDeleteTarget(null);
        fetchFines();
        fetchStats();
      } else {
        addToast(res.message || "Failed to delete", "error");
      }
    } catch (err) {
      addToast(err.message || "Failed to delete", "error");
    } finally {
      setSaving(false);
    }
  };

  const handlePhotoUpload = async (fineId, file) => {
    const formData = new FormData();
    formData.append("photo", file);

    setUploading(true);
    try {
      const res = await api.uploadFinePhoto(fineId, formData);
      if (res.success) {
        addToast("Photo uploaded successfully", "success");
        fetchFines();
        if (viewingFine && viewingFine.id === fineId) {
          const updated = await api.getFine(fineId);
          if (updated.success) setViewingFine(updated.data);
        }
      } else {
        addToast(res.message || "Upload failed", "error");
      }
    } catch (err) {
      addToast(err.message || "Upload failed", "error");
    } finally {
      setUploading(false);
      if (fileInputRef.current) fileInputRef.current.value = "";
    }
  };

  const openView = async (fine) => {
    setViewingFine(fine);
    setShowViewModal(true);
  };

  const openEdit = (fine) => {
    setEditingFine(fine);
    setForm({
      fiscal_year_id: String(fine.fiscal_year_id || ""),
      member_id: String(fine.member_id || ""),
      name: fine.name || "",
      violation_type: fine.violation_type,
      description: fine.description || "",
      fine_amount: String(fine.fine_amount),
      incident_date: fine.incident_date?.split("T")[0] || "",
      photo: fine.photo || "",
      remarks: fine.remarks || "",
    });
    setShowEditModal(true);
  };

  const openStatusModal = (fine) => {
    setSelectedFine(fine);
    setStatusForm({ status: "paid", payment_reference: "" });
    setShowStatusModal(true);
  };

  const getStatusBadge = (status) => {
    switch (status) {
      case "pending":
        return <Badge status="warning">Pending</Badge>;
      case "paid":
        return <Badge status="success">Paid</Badge>;
      case "waived":
        return <Badge status="info">Waived</Badge>;
      default:
        return <Badge status="default">{status}</Badge>;
    }
  };

  const statusOptions = [
    { value: "paid", label: "✅ Paid" },
    { value: "waived", label: "📝 Waived" },
  ];

  const violationTypeOptions = [
    { value: "Illegal Logging", label: "🌲 Illegal Logging" },
    { value: "Unauthorized Tree Cutting", label: "🪓 Unauthorized Tree Cutting" },
    { value: "Illegal Hunting", label: "🦌 Illegal Hunting" },
    { value: "Forest Fire", label: "🔥 Forest Fire" },
    { value: "Grazing Violation", label: "🐄 Grazing Violation" },
    { value: "Waste Dumping", label: "🗑️ Waste Dumping" },
    { value: "Late Payment Penalty", label: "⏰ Late Payment Penalty" },
    { value: "Other", label: "📋 Other" },
  ];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">
            Fines Management
          </h1>
          <p className="text-sm text-gray-500 dark:text-gray-400">
            Track and manage violation fines
          </p>
        </div>
        <div className="flex gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => {
              setRefreshing(true);
              fetchFines();
            }}
            isLoading={refreshing}
          >
            <RefreshCw size={14} className="mr-1" /> Refresh
          </Button>
          {canEdit && (
            <Button onClick={openCreateFine}>
              <Plus size={16} /> Add Fine
            </Button>
          )}
        </div>
      </div>

      {/* Stats Cards */}
      {stats && (
        <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3">
          <Card>
            <CardContent className="p-3 text-center">
              <p className="text-2xl font-bold text-gray-900 dark:text-gray-100">{stats.total || 0}</p>
              <p className="text-xs text-gray-500">Total</p>
            </CardContent>
          </Card>
          <Card className="border-amber-200 bg-amber-50 dark:bg-amber-900/20">
            <CardContent className="p-3 text-center">
              <p className="text-2xl font-bold text-amber-600">{stats.pending || 0}</p>
              <p className="text-xs text-amber-600">Pending</p>
            </CardContent>
          </Card>
          <Card className="border-emerald-200 bg-emerald-50 dark:bg-emerald-900/20">
            <CardContent className="p-3 text-center">
              <p className="text-2xl font-bold text-emerald-600">{stats.paid || 0}</p>
              <p className="text-xs text-emerald-600">Paid</p>
            </CardContent>
          </Card>
          <Card className="border-blue-200 bg-blue-50 dark:bg-blue-900/20">
            <CardContent className="p-3 text-center">
              <p className="text-2xl font-bold text-blue-600">{stats.waived || 0}</p>
              <p className="text-xs text-blue-600">Waived</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-3 text-center">
              <p className="text-xl font-bold text-red-600">{formatCurrency(stats.total_value || 0)}</p>
              <p className="text-xs text-gray-500">Total Value</p>
            </CardContent>
          </Card>
          <Card className="border-emerald-200 bg-emerald-50 dark:bg-emerald-900/20">
            <CardContent className="p-3 text-center">
              <p className="text-xl font-bold text-emerald-600">{formatCurrency(stats.paid_value || 0)}</p>
              <p className="text-xs text-emerald-600">Collected</p>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Filters */}
      <Card>
        <CardContent className="p-4">
          <div className="flex flex-col sm:flex-row gap-3">
            <div className="flex-1 relative">
              <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
              <input
                type="text"
                placeholder="Search by name or violation type..."
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
                { value: "waived", label: "Waived" },
              ]}
              className="w-36"
            />
            {canEdit && (
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

      {/* Fines Table */}
      <Card>
        <CardContent className="p-0">
          {isLoading ? (
            <LoadingSpinner text="Loading fines..." />
          ) : fines.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 text-gray-400">
              <AlertTriangle size={48} className="mb-3 opacity-30" />
              <p className="text-lg font-medium">No fines found</p>
              {canEdit && (
                <Button variant="outline" className="mt-4" onClick={openCreateFine}>
                  <Plus size={14} /> Add your first fine
                </Button>
              )}
            </div>
          ) : (
            <div className="overflow-x-auto">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>ID</TableHead>
                    <TableHead>Violator</TableHead>
                    <TableHead>Violation Type</TableHead>
                    <TableHead>Amount</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Date</TableHead>
                    <TableHead>Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {fines.map((fine) => (
                    <TableRow key={fine.id}>
                      <TableCell className="font-mono text-xs">#{fine.id}</TableCell>
                      <TableCell className="font-medium">
                        {fine.member?.name || fine.name || "-"}
                        {fine.member?.membership_no && (
                          <><br /><span className="text-xs text-gray-400">{fine.member.membership_no}</span></>
                        )}
                      </TableCell>
                      <TableCell>{fine.violation_type}</TableCell>
                      <TableCell className="font-semibold text-red-600">
                        {formatCurrency(fine.fine_amount)}
                      </TableCell>
                      <TableCell>{getStatusBadge(fine.status)}</TableCell>
                      <TableCell>{formatDate(fine.incident_date)}</TableCell>
                      <TableCell>
                        <div className="flex items-center gap-1">
                          <Button
                            variant="ghost"
                            size="icon"
                            onClick={() => openView(fine)}
                            title="View Details"
                          >
                            <Eye size={15} />
                          </Button>
                          {canEdit && fine.status === "pending" && (
                            <Button
                              variant="ghost"
                              size="icon"
                              onClick={() => openEdit(fine)}
                              title="Edit"
                            >
                              <Edit2 size={15} />
                            </Button>
                          )}
                          {isAdmin && fine.status === "pending" && (
                            <>
                              <Button
                                variant="ghost"
                                size="icon"
                                onClick={() => openStatusModal(fine)}
                                title="Update Status"
                                className="text-emerald-600 hover:text-emerald-700"
                              >
                                <CheckCircle size={16} />
                              </Button>
                              <Button
                                variant="ghost"
                                size="icon"
                                onClick={() => setDeleteTarget(fine.id)}
                                title="Delete"
                              >
                                <Trash2 size={15} className="text-red-500" />
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
          <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => setPage(page - 1)}>
            Previous
          </Button>
          <span className="text-sm text-gray-500">Page {page} of {totalPages}</span>
          <Button variant="outline" size="sm" disabled={page >= totalPages} onClick={() => setPage(page + 1)}>
            Next
          </Button>
        </div>
      )}

      {/* Create Fine Modal */}
      <Modal
        isOpen={showCreateModal}
        onClose={() => {
          setShowCreateModal(false);
          setForm(newFineForm());
        }}
        title="Add Fine"
        size="lg"
        footer={
          <>
            <Button variant="outline" onClick={() => setShowCreateModal(false)}>Cancel</Button>
            <Button onClick={handleCreate} isLoading={saving}>Create Fine</Button>
          </>
        }
      >
        <div className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Select
              label="Fiscal Year"
              value={form.fiscal_year_id}
              onChange={(e) => setForm({ ...form, fiscal_year_id: e.target.value })}
              options={buildFiscalYearOptions(fiscalYears)}
              placeholder="Select fiscal year"
              required
            />
            <Select
              label="Member (Optional)"
              value={form.member_id}
              onChange={(e) => {
                setForm({ ...form, member_id: e.target.value, name: "" });
              }}
              options={[
                { value: "", label: "Non-member (enter name below)" },
                ...members.map(m => ({ value: String(m.id), label: `${m.name} (${m.membership_no})` })),
              ]}
            />
          </div>

          {!form.member_id && (
            <Input
              label="Violator Name"
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
              placeholder="Enter name for non-member violator"
              required={!form.member_id}
            />
          )}

          <Select
            label="Violation Type"
            value={form.violation_type}
            onChange={(e) => setForm({ ...form, violation_type: e.target.value })}
            options={violationTypeOptions}
            required
          />

          <Textarea
            label="Description"
            value={form.description}
            onChange={(e) => setForm({ ...form, description: e.target.value })}
            placeholder="Detailed description of the violation"
            rows={3}
          />

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Input
              label="Fine Amount (NPR)"
              type="number"
              step="0.01"
              value={form.fine_amount}
              onChange={(e) => setForm({ ...form, fine_amount: e.target.value })}
              placeholder="0.00"
              required
            />
            <Input
              label="Incident Date"
              type="date"
              value={form.incident_date}
              onChange={(e) => setForm({ ...form, incident_date: e.target.value })}
              required
            />
          </div>

          <Input
            label="Photo URL"
            value={form.photo}
            onChange={(e) => setForm({ ...form, photo: e.target.value })}
            placeholder="Image URL (or upload after creation)"
          />

          <Textarea
            label="Remarks"
            value={form.remarks}
            onChange={(e) => setForm({ ...form, remarks: e.target.value })}
            placeholder="Additional notes..."
            rows={2}
          />
        </div>
      </Modal>

      {/* Edit Fine Modal */}
      <Modal
        isOpen={showEditModal}
        onClose={() => {
          setShowEditModal(false);
          setEditingFine(null);
          setForm(newFineForm());
        }}
        title="Edit Fine"
        size="lg"
        footer={
          <>
            <Button variant="outline" onClick={() => setShowEditModal(false)}>Cancel</Button>
            <Button onClick={handleUpdate} isLoading={saving}>Update Fine</Button>
          </>
        }
      >
        <div className="space-y-4">
          <Select
            label="Violation Type"
            value={form.violation_type}
            onChange={(e) => setForm({ ...form, violation_type: e.target.value })}
            options={violationTypeOptions}
            required
          />

          <Textarea
            label="Description"
            value={form.description}
            onChange={(e) => setForm({ ...form, description: e.target.value })}
            placeholder="Detailed description of the violation"
            rows={3}
          />

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Input
              label="Fine Amount (NPR)"
              type="number"
              step="0.01"
              value={form.fine_amount}
              onChange={(e) => setForm({ ...form, fine_amount: e.target.value })}
              required
            />
            <Input
              label="Incident Date"
              type="date"
              value={form.incident_date}
              onChange={(e) => setForm({ ...form, incident_date: e.target.value })}
              required
            />
          </div>

          <Input
            label="Photo URL"
            value={form.photo}
            onChange={(e) => setForm({ ...form, photo: e.target.value })}
            placeholder="Image URL"
          />

          <Textarea
            label="Remarks"
            value={form.remarks}
            onChange={(e) => setForm({ ...form, remarks: e.target.value })}
            placeholder="Additional notes..."
            rows={2}
          />
        </div>
      </Modal>

      {/* Update Status Modal */}
      <Modal
        isOpen={showStatusModal}
        onClose={() => {
          setShowStatusModal(false);
          setSelectedFine(null);
          setStatusForm({ status: "paid", payment_reference: "" });
        }}
        title="Update Fine Status"
        size="sm"
        footer={
          <>
            <Button variant="outline" onClick={() => setShowStatusModal(false)}>Cancel</Button>
            <Button
              variant={statusForm.status === "paid" ? "success" : "primary"}
              onClick={handleUpdateStatus}
              isLoading={saving}
            >
              Update Status
            </Button>
          </>
        }
      >
        <div className="space-y-4">
          {selectedFine && (
            <div className="p-3 bg-gray-50 dark:bg-gray-800/50 rounded-lg">
              <p className="text-sm font-medium">{selectedFine.violation_type}</p>
              <p className="text-sm text-gray-600 dark:text-gray-400">
                Violator: {selectedFine.member?.name || selectedFine.name}
              </p>
              <p className="text-lg font-bold text-red-600 mt-1">
                {formatCurrency(selectedFine.fine_amount)}
              </p>
            </div>
          )}

          <Select
            label="New Status"
            value={statusForm.status}
            onChange={(e) => setStatusForm({ ...statusForm, status: e.target.value })}
            options={statusOptions}
            required
          />

          <Input
            label="Payment Reference / Receipt No"
            value={statusForm.payment_reference}
            onChange={(e) => setStatusForm({ ...statusForm, payment_reference: e.target.value })}
            placeholder="Enter receipt number or reference"
          />

          <div className="text-xs text-gray-500">
            {statusForm.status === "paid" 
              ? "This will mark the fine as paid and create a transaction record."
              : "This will waive the fine. No payment required."}
          </div>
        </div>
      </Modal>

      {/* View Fine Modal */}
      <Modal
        isOpen={showViewModal}
        onClose={() => {
          setShowViewModal(false);
          setViewingFine(null);
        }}
        title="Fine Details"
        size="lg"
      >
        {viewingFine && (
          <div className="space-y-6">
            {/* Header */}
            <div className="flex items-center justify-between">
              {getStatusBadge(viewingFine.status)}
              <span className="text-xs font-mono text-gray-500">#{viewingFine.id}</span>
            </div>

            {/* Amount */}
            <div className="text-center">
              <p className="text-3xl font-bold text-red-600">
                {formatCurrency(viewingFine.fine_amount)}
              </p>
              <p className="text-sm text-gray-500">Fine Amount</p>
            </div>

            {/* Violator Info */}
            <div className="border-t border-gray-200 dark:border-white/10 pt-4">
              <div className="flex items-start gap-3">
                <User size={18} className="text-gray-400 mt-0.5" />
                <div>
                  <p className="text-xs text-gray-400">Violator</p>
                  <p className="font-medium text-gray-900 dark:text-gray-100">
                    {viewingFine.member?.name || viewingFine.name || "-"}
                  </p>
                  {viewingFine.member?.membership_no && (
                    <p className="text-xs text-gray-500">
                      Membership: {viewingFine.member.membership_no}
                    </p>
                  )}
                </div>
              </div>
            </div>

            {/* Violation Details */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="flex items-start gap-3">
                <AlertCircle size={18} className="text-gray-400 mt-0.5" />
                <div>
                  <p className="text-xs text-gray-400">Violation Type</p>
                  <p className="text-gray-900 dark:text-gray-100">
                    {viewingFine.violation_type}
                  </p>
                </div>
              </div>
              <div className="flex items-start gap-3">
                <Calendar size={18} className="text-gray-400 mt-0.5" />
                <div>
                  <p className="text-xs text-gray-400">Incident Date</p>
                  <p className="text-gray-900 dark:text-gray-100">
                    {formatDate(viewingFine.incident_date)}
                  </p>
                </div>
              </div>
            </div>

            {/* Description */}
            {viewingFine.description && (
              <div className="bg-gray-50 dark:bg-gray-800/50 rounded-lg p-4">
                <p className="text-xs text-gray-400 mb-1">Description</p>
                <p className="text-gray-700 dark:text-gray-300">
                  {viewingFine.description}
                </p>
              </div>
            )}

            {/* Payment Info */}
            {viewingFine.status === "paid" && viewingFine.payment_reference && (
              <div className="bg-emerald-50 dark:bg-emerald-900/20 rounded-lg p-4">
                <p className="text-xs text-emerald-600 mb-1">Payment Reference</p>
                <p className="font-mono text-sm text-emerald-700 dark:text-emerald-400">
                  {viewingFine.payment_reference}
                </p>
              </div>
            )}

            {/* Remarks */}
            {viewingFine.remarks && (
              <div className="border-t border-gray-200 dark:border-white/10 pt-4">
                <p className="text-xs text-gray-400 mb-1">Remarks</p>
                <p className="text-gray-700 dark:text-gray-300">{viewingFine.remarks}</p>
              </div>
            )}

            {/* Photo Section */}
            <div className="border-t border-gray-200 dark:border-white/10 pt-4">
              <div className="flex items-center justify-between mb-3">
                <h4 className="font-medium text-gray-900 dark:text-gray-100 flex items-center gap-2">
                  <ImageIcon size={16} />
                  Evidence Photo
                </h4>
                {canEdit && (
                  <div>
                    <input
                      type="file"
                      ref={fileInputRef}
                      accept="image/*"
                      className="hidden"
                      onChange={(e) => {
                        const file = e.target.files?.[0];
                        if (file && viewingFine) {
                          handlePhotoUpload(viewingFine.id, file);
                        }
                      }}
                    />
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={() => fileInputRef.current?.click()}
                      isLoading={uploading}
                    >
                      <Upload size={14} className="mr-1" />
                      {viewingFine.photo ? "Replace" : "Upload"}
                    </Button>
                  </div>
                )}
              </div>
              {viewingFine.photo ? (
                <div className="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800/50 rounded-lg">
                  <div className="flex items-center gap-3">
                    <ImageIcon size={20} className="text-emerald-500" />
                    <span className="text-sm text-gray-700 dark:text-gray-300">
                      {viewingFine.photo.split("/").pop()}
                    </span>
                  </div>
                  <a
                    href={getImageUrl(viewingFine.photo)}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-emerald-600 hover:text-emerald-700 text-sm flex items-center gap-1"
                  >
                    <Download size={14} /> View
                  </a>
                </div>
              ) : (
                <p className="text-sm text-gray-400 text-center py-4">
                  No evidence photo attached
                </p>
              )}
            </div>

            {/* Meta Info */}
            <div className="text-xs text-gray-400 border-t border-gray-200 dark:border-white/10 pt-4">
              <p>Created by: {viewingFine.creator?.name || "Unknown"}</p>
              <p>Created at: {formatDate(viewingFine.created_at)}</p>
              {viewingFine.updated_at !== viewingFine.created_at && (
                <p>Last updated: {formatDate(viewingFine.updated_at)}</p>
              )}
            </div>
          </div>
        )}
      </Modal>

      {/* Delete Confirmation Modal */}
      <Modal
        isOpen={!!deleteTarget}
        onClose={() => setDeleteTarget(null)}
        title="Confirm Delete"
        size="sm"
        footer={
          <>
            <Button variant="outline" onClick={() => setDeleteTarget(null)}>Cancel</Button>
            <Button variant="danger" onClick={handleDelete} isLoading={saving}>Delete</Button>
          </>
        }
      >
        <div className="flex items-start gap-3">
          <AlertTriangle size={20} className="text-red-500 shrink-0 mt-0.5" />
          <div>
            <p className="text-sm text-gray-700 dark:text-gray-300">
              Are you sure you want to delete this fine?
            </p>
            <p className="text-xs text-gray-500 mt-1">
              This action cannot be undone.
            </p>
          </div>
        </div>
      </Modal>
    </div>
  );
}