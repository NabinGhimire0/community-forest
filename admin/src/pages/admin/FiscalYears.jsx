import { useState, useEffect, useCallback } from "react";
import { useSelector } from "react-redux";
import {
  Plus,
  Check,
  Edit2,
  Trash2,
  Calendar,
  ChevronDown,
  ChevronUp,
  AlertTriangle,
  RefreshCw,
  X,
  Save,
} from "lucide-react";
import { api } from "../../services/api";
import { Card, CardContent, CardHeader } from "../../components/ui/Card";
import Button from "../../components/ui/Button";
import Input from "../../components/ui/Input";
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
import { formatDate } from "../../utils/helpers";

export default function FiscalYears() {
  const { user } = useSelector((state) => state.auth);
  const { addToast } = useToast();
  const canEdit = user?.role === "admin" || user?.role === "staff";
  const isAdmin = user?.role === "admin";

  const [fiscalYears, setFiscalYears] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [expandedId, setExpandedId] = useState(null);
  const [feeSettings, setFeeSettings] = useState({});

  // Modal states
  const [showFyModal, setShowFyModal] = useState(false);
  const [editingFy, setEditingFy] = useState(null);
  const [fyForm, setFyForm] = useState({
    name: "",
    start_date: "",
    end_date: "",
  });

  const [showFeeModal, setShowFeeModal] = useState(false);
  const [editingFee, setEditingFee] = useState(null);
  const [currentFiscalYearId, setCurrentFiscalYearId] = useState(null);
  const [feeForm, setFeeForm] = useState({ membership_fee: "" });

  const [deleteTarget, setDeleteTarget] = useState(null);
  const [saving, setSaving] = useState(false);

  const fetchFiscalYears = useCallback(async () => {
    setIsLoading(true);
    try {
      const res = await api.getFiscalYears();
      if (res.success) {
        setFiscalYears(res.data || []);
      } else {
        addToast(res.message || "Failed to load fiscal years", "error");
      }
    } catch (err) {
      addToast(err.message || "Failed to load fiscal years", "error");
    } finally {
      setIsLoading(false);
      setRefreshing(false);
    }
  }, [addToast]);

  const fetchFeeSettings = async (fiscalYearId) => {
    try {
      const res = await api.getFeeSettings(fiscalYearId);
      if (res.success) {
        setFeeSettings((prev) => ({ ...prev, [fiscalYearId]: res.data || [] }));
      }
    } catch (err) {
      console.error("Failed to fetch fee settings:", err);
      setFeeSettings((prev) => ({ ...prev, [fiscalYearId]: [] }));
    }
  };

  useEffect(() => {
    fetchFiscalYears();
  }, [fetchFiscalYears]);

  useEffect(() => {
    if (expandedId) {
      fetchFeeSettings(expandedId);
    }
  }, [expandedId]);

  const handleSaveFiscalYear = async () => {
    setSaving(true);
    try {
      let res;
      if (editingFy) {
        res = await api.updateFiscalYear(editingFy.id, fyForm);
      } else {
        res = await api.createFiscalYear(fyForm);
      }
      if (res.success) {
        addToast(
          `Fiscal year ${editingFy ? "updated" : "created"} successfully`,
          "success",
        );
        setShowFyModal(false);
        setEditingFy(null);
        setFyForm({ name: "", start_date: "", end_date: "" });
        fetchFiscalYears();
      } else {
        addToast(res.message || "Failed to save", "error");
      }
    } catch (err) {
      addToast(err.message || "Failed to save", "error");
    } finally {
      setSaving(false);
    }
  };

  const handleActivate = async (id) => {
    try {
      const res = await api.activateFiscalYear(id);
      if (res.success) {
        addToast("Fiscal year activated successfully", "success");
        fetchFiscalYears();
      } else {
        addToast(res.message || "Failed to activate", "error");
      }
    } catch (err) {
      addToast(err.message || "Failed to activate", "error");
    }
  };

  const handleSaveFee = async () => {
    if (!currentFiscalYearId) return;
    setSaving(true);
    try {
      let res;
      if (editingFee) {
        res = await api.updateFeeSetting(editingFee.id, feeForm);
      } else {
        res = await api.createFeeSetting(currentFiscalYearId, feeForm);
      }
      if (res.success) {
        addToast(
          `Fee setting ${editingFee ? "updated" : "added"} successfully`,
          "success",
        );
        setShowFeeModal(false);
        setEditingFee(null);
        setFeeForm({ membership_fee: "" });
        fetchFeeSettings(currentFiscalYearId);
      } else {
        addToast(res.message || "Failed to save fee", "error");
      }
    } catch (err) {
      addToast(err.message || "Failed to save fee", "error");
    } finally {
      setSaving(false);
    }
  };

  const handleDeleteFee = async (feeId) => {
    try {
      const res = await api.deleteFeeSetting(feeId);
      if (res.success) {
        addToast("Fee setting deleted successfully", "success");
        if (expandedId) {
          fetchFeeSettings(expandedId);
        }
      } else {
        addToast(res.message || "Failed to delete", "error");
      }
    } catch (err) {
      addToast(err.message || "Failed to delete", "error");
    }
  };

  const handleDeleteFiscalYear = async () => {
    if (!deleteTarget) return;
    setSaving(true);
    try {
      const res = await api.deleteFiscalYear(deleteTarget);
      if (res.success) {
        addToast("Fiscal year deleted successfully", "success");
        setDeleteTarget(null);
        fetchFiscalYears();
      } else {
        addToast(res.message || "Failed to delete", "error");
      }
    } catch (err) {
      addToast(err.message || "Failed to delete", "error");
    } finally {
      setSaving(false);
    }
  };

  const openEditFy = (fy) => {
    setEditingFy(fy);
    setFyForm({
      name: fy.name,
      start_date: fy.start_date?.split("T")[0] || "",
      end_date: fy.end_date?.split("T")[0] || "",
    });
    setShowFyModal(true);
  };

  const openAddFee = (fiscalYearId) => {
    setCurrentFiscalYearId(fiscalYearId);
    setEditingFee(null);
    setFeeForm({ membership_fee: "" });
    setShowFeeModal(true);
  };

  const openEditFee = (fee, fiscalYearId) => {
    setCurrentFiscalYearId(fiscalYearId);
    setEditingFee(fee);
    setFeeForm({ membership_fee: fee.membership_fee.toString() });
    setShowFeeModal(true);
  };

  const toggleExpand = async (fy) => {
    if (expandedId === fy.id) {
      setExpandedId(null);
    } else {
      setExpandedId(fy.id);
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">
            Fiscal Years
          </h1>
          <p className="text-sm text-gray-500 dark:text-gray-400">
            Manage fiscal years and membership fee settings
          </p>
        </div>
        <div className="flex gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => {
              setRefreshing(true);
              fetchFiscalYears();
            }}
            isLoading={refreshing}
          >
            <RefreshCw size={14} className="mr-1" /> Refresh
          </Button>
          {isAdmin && (
            <Button
              onClick={() => {
                setEditingFy(null);
                setFyForm({ name: "", start_date: "", end_date: "" });
                setShowFyModal(true);
              }}
            >
              <Plus size={16} /> Add Fiscal Year
            </Button>
          )}
        </div>
      </div>

      {/* Fiscal Years List */}
      {isLoading ? (
        <LoadingSpinner text="Loading fiscal years..." />
      ) : fiscalYears.length === 0 ? (
        <Card>
          <CardContent className="py-12 text-center">
            <Calendar
              size={48}
              className="mx-auto mb-3 text-gray-400 opacity-30"
            />
            <p className="text-lg font-medium text-gray-500">
              No fiscal years found
            </p>
            {isAdmin && (
              <Button
                variant="outline"
                className="mt-4"
                onClick={() => setShowFyModal(true)}
              >
                <Plus size={14} className="mr-1" /> Add your first fiscal year
              </Button>
            )}
          </CardContent>
        </Card>
      ) : (
        <div className="space-y-3">
          {fiscalYears.map((fy) => (
            <Card key={fy.id} className="overflow-hidden">
              <CardContent className="p-0">
                {/* Fiscal Year Row */}
                <div
                  className="flex items-center justify-between p-4 cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors"
                  onClick={() => toggleExpand(fy)}
                >
                  <div className="flex items-center gap-4">
                    <button className="text-gray-400">
                      {expandedId === fy.id ? (
                        <ChevronUp size={18} />
                      ) : (
                        <ChevronDown size={18} />
                      )}
                    </button>
                    <div>
                      <h4 className="font-semibold text-gray-900 dark:text-gray-100">
                        {fy.name}
                      </h4>
                      <p className="text-sm text-gray-500">
                        {formatDate(fy.start_date)} — {formatDate(fy.end_date)}
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center gap-3">
                    <Badge status={fy.is_active ? "active" : "inactive"}>
                      {fy.is_active ? "Active" : "Inactive"}
                    </Badge>
                    {isAdmin && !fy.is_active && (
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={(e) => {
                          e.stopPropagation();
                          handleActivate(fy.id);
                        }}
                      >
                        <Check size={14} className="mr-1" /> Activate
                      </Button>
                    )}
                    {isAdmin && (
                      <div
                        className="flex gap-1"
                        onClick={(e) => e.stopPropagation()}
                      >
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => openEditFy(fy)}
                          title="Edit"
                        >
                          <Edit2 size={15} />
                        </Button>
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => setDeleteTarget(fy.id)}
                          title="Delete"
                        >
                          <Trash2 size={15} className="text-red-500" />
                        </Button>
                      </div>
                    )}
                  </div>
                </div>

                {/* Expanded Fee Settings */}
                {expandedId === fy.id && (
                  <div className="border-t border-gray-200 dark:border-white/10 p-4 bg-gray-50/50 dark:bg-gray-800/30">
                    <div className="flex items-center justify-between mb-4">
                      <h5 className="font-medium text-gray-900 dark:text-gray-100">
                        Membership Fee Settings
                      </h5>
                      {isAdmin && (
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => openAddFee(fy.id)}
                        >
                          <Plus size={14} className="mr-1" /> Set Fee
                        </Button>
                      )}
                    </div>

                    {!feeSettings[fy.id] ? (
                      <LoadingSpinner size={20} text="Loading fees..." />
                    ) : feeSettings[fy.id].length === 0 ? (
                      <p className="text-sm text-gray-500 py-4 text-center">
                        No membership fee configured for this fiscal year.
                        {isAdmin && " Click 'Set Fee' to add one."}
                      </p>
                    ) : (
                      <Table>
                        <TableHeader>
                          <TableRow>
                            <TableHead>Membership Fee</TableHead>
                            <TableHead>Created At</TableHead>
                            {isAdmin && <TableHead>Actions</TableHead>}
                          </TableRow>
                        </TableHeader>
                        <TableBody>
                          {feeSettings[fy.id].map((fee) => (
                            <TableRow key={fee.id}>
                              <TableCell className="font-semibold text-emerald-600">
                                रु {fee.membership_fee.toLocaleString()}
                              </TableCell>
                              <TableCell className="text-sm text-gray-500">
                                {formatDate(fee.created_at)}
                              </TableCell>
                              {isAdmin && (
                                <TableCell>
                                  <div className="flex gap-1">
                                    <Button
                                      variant="ghost"
                                      size="icon"
                                      onClick={() => openEditFee(fee, fy.id)}
                                      title="Edit"
                                    >
                                      <Edit2 size={14} />
                                    </Button>
                                    <Button
                                      variant="ghost"
                                      size="icon"
                                      onClick={() => handleDeleteFee(fee.id)}
                                      title="Delete"
                                    >
                                      <Trash2
                                        size={14}
                                        className="text-red-500"
                                      />
                                    </Button>
                                  </div>
                                </TableCell>
                              )}
                            </TableRow>
                          ))}
                        </TableBody>
                      </Table>
                    )}
                  </div>
                )}
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {/* Add/Edit Fiscal Year Modal */}
      <Modal
        isOpen={showFyModal}
        onClose={() => {
          setShowFyModal(false);
          setEditingFy(null);
          setFyForm({ name: "", start_date: "", end_date: "" });
        }}
        title={editingFy ? "Edit Fiscal Year" : "Add Fiscal Year"}
        footer={
          <>
            <Button variant="outline" onClick={() => setShowFyModal(false)}>
              Cancel
            </Button>
            <Button onClick={handleSaveFiscalYear} isLoading={saving}>
              {editingFy ? "Update" : "Create"}
            </Button>
          </>
        }
      >
        <div className="space-y-4">
          <Input
            label="Fiscal Year Name"
            value={fyForm.name}
            onChange={(e) => setFyForm({ ...fyForm, name: e.target.value })}
            placeholder="e.g., 2080/81"
            required
          />
          <div className="grid grid-cols-2 gap-4">
            <Input
              label="Start Date"
              type="date"
              value={fyForm.start_date}
              onChange={(e) =>
                setFyForm({ ...fyForm, start_date: e.target.value })
              }
              required
            />
            <Input
              label="End Date"
              type="date"
              value={fyForm.end_date}
              onChange={(e) =>
                setFyForm({ ...fyForm, end_date: e.target.value })
              }
              required
            />
          </div>
          <p className="text-xs text-gray-500">
            Note: Nepali fiscal year typically runs from Shrawan 1 (mid-July) to
            Ashadh end (mid-July)
          </p>
        </div>
      </Modal>

      {/* Add/Edit Fee Modal */}
      <Modal
        isOpen={showFeeModal}
        onClose={() => {
          setShowFeeModal(false);
          setEditingFee(null);
          setFeeForm({ membership_fee: "" });
        }}
        title={editingFee ? "Edit Membership Fee" : "Set Membership Fee"}
        footer={
          <>
            <Button variant="outline" onClick={() => setShowFeeModal(false)}>
              Cancel
            </Button>
            <Button onClick={handleSaveFee} isLoading={saving}>
              {editingFee ? "Update" : "Save"}
            </Button>
          </>
        }
      >
        <div className="space-y-4">
          <Input
            label="Membership Fee (NPR)"
            type="number"
            value={feeForm.membership_fee}
            onChange={(e) => setFeeForm({ membership_fee: e.target.value })}
            placeholder="e.g., 500"
            required
          />
          <p className="text-xs text-gray-500">
            This fee will be charged to each member for the selected fiscal
            year.
          </p>
        </div>
      </Modal>

      {/* Delete Confirmation Modal */}
      <Modal
        isOpen={!!deleteTarget}
        onClose={() => setDeleteTarget(null)}
        title="Confirm Delete"
        size="sm"
        footer={
          <>
            <Button variant="outline" onClick={() => setDeleteTarget(null)}>
              Cancel
            </Button>
            <Button
              variant="danger"
              onClick={handleDeleteFiscalYear}
              isLoading={saving}
            >
              Delete
            </Button>
          </>
        }
      >
        <div className="flex items-start gap-3">
          <AlertTriangle size={20} className="text-red-500 shrink-0 mt-0.5" />
          <div>
            <p className="text-sm text-gray-700 dark:text-gray-300">
              Are you sure you want to delete this fiscal year?
            </p>
            <p className="text-xs text-gray-500 mt-1">
              This action cannot be undone. All associated fee settings will be
              deleted.
            </p>
          </div>
        </div>
      </Modal>
    </div>
  );
}
