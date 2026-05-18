import { useState, useEffect, useCallback } from "react";
import { useSelector } from "react-redux";
import {
  Plus,
  Edit2,
  Trash2,
  Package,
  AlertTriangle,
  RefreshCw,
  DollarSign,
  Archive,
  Layers,
  Tag,
} from "lucide-react";
import { api } from "../../services/api";
import { Card, CardContent } from "../../components/ui/Card";
import Button from "../../components/ui/Button";
import Input from "../../components/ui/Input";
import Select from "../../components/ui/Select";
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
import { formatCurrency } from "../../utils/helpers";

export default function Resources() {
  const { user } = useSelector((state) => state.auth);
  const { addToast } = useToast();
  const canEdit = user?.role === "admin" || user?.role === "staff";
  const isAdmin = user?.role === "admin";

  const [activeTab, setActiveTab] = useState("types");
  const [types, setTypes] = useState([]);
  const [items, setItems] = useState([]);
  const [rates, setRates] = useState([]);
  const [stock, setStock] = useState([]);
  const [fiscalYears, setFiscalYears] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);

  // Modal states
  const [showModal, setShowModal] = useState(false);
  const [editingItem, setEditingItem] = useState(null);
  const [form, setForm] = useState({});
  const [saving, setSaving] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState(null);

  const tabs = [
    { id: "types", label: "Resource Types", icon: Layers },
    { id: "items", label: "Resource Items", icon: Package },
    { id: "rates", label: "Rates", icon: DollarSign },
    { id: "stock", label: "Stock", icon: Archive },
  ];

  // Fetch all master data (types, items, fiscalYears) once on component mount
  const fetchMasterData = useCallback(async () => {
    try {
      const [typesRes, itemsRes, fiscalYearsRes] = await Promise.all([
        api.getResourceTypes(),
        api.getResourceItems(),
        api.getFiscalYears(),
      ]);
      if (typesRes.success) setTypes(typesRes.data || []);
      if (itemsRes.success) setItems(itemsRes.data || []);
      if (fiscalYearsRes.success) setFiscalYears(fiscalYearsRes.data || []);
    } catch (err) {
      console.error("Failed to fetch master data:", err);
    }
  }, []);

  const fetchData = useCallback(async () => {
    setIsLoading(true);
    try {
      // Always ensure master data is loaded first
      if (
        types.length === 0 ||
        items.length === 0 ||
        fiscalYears.length === 0
      ) {
        await fetchMasterData();
      }

      if (activeTab === "types") {
        const res = await api.getResourceTypes();
        if (res.success) setTypes(res.data || []);
      } else if (activeTab === "items") {
        const res = await api.getResourceItems();
        if (res.success) setItems(res.data || []);
      } else if (activeTab === "rates") {
        const res = await api.getResourceRates();
        if (res.success) setRates(res.data || []);
      } else if (activeTab === "stock") {
        const res = await api.getStock();
        if (res.success) setStock(res.data || []);
      }
    } catch (err) {
      addToast(`Failed to load ${activeTab}`, "error");
    } finally {
      setIsLoading(false);
      setRefreshing(false);
    }
  }, [
    activeTab,
    addToast,
    fetchMasterData,
    types.length,
    items.length,
    fiscalYears.length,
  ]);

  // Load master data on component mount
  useEffect(() => {
    fetchMasterData();
  }, [fetchMasterData]);

  // Load tab-specific data when tab changes
  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const handleSave = async () => {
    setSaving(true);
    try {
      let res;
      const payload = { ...form };

      // Convert numeric fields
      if (activeTab === "rates") {
        payload.resource_item_id = Number(form.resource_item_id);
        payload.fiscal_year_id = Number(form.fiscal_year_id);
        payload.rate_per_unit = Number(form.rate_per_unit);
      } else if (activeTab === "stock") {
        payload.resource_item_id = Number(form.resource_item_id);
        payload.fiscal_year_id = Number(form.fiscal_year_id);
        payload.total_quantity = Number(form.total_quantity);
      } else if (activeTab === "items") {
        payload.resource_type_id = Number(form.resource_type_id);
      }

      if (activeTab === "types") {
        if (editingItem) {
          res = await api.updateResourceType(editingItem.id, payload);
        } else {
          res = await api.createResourceType(payload);
        }
      } else if (activeTab === "items") {
        if (editingItem) {
          res = await api.updateResourceItem(editingItem.id, payload);
        } else {
          res = await api.createResourceItem(payload);
        }
      } else if (activeTab === "rates") {
        if (editingItem) {
          res = await api.updateResourceRate(editingItem.id, payload);
        } else {
          res = await api.createResourceRate(payload);
        }
      } else if (activeTab === "stock") {
        if (editingItem) {
          res = await api.updateStock(editingItem.id, payload);
        } else {
          res = await api.createStock(payload);
        }
      }

      if (res?.success) {
        addToast(
          `${activeTab.slice(0, -1)} ${editingItem ? "updated" : "created"} successfully`,
          "success",
        );
        setShowModal(false);
        setEditingItem(null);
        setForm({});

        // Refresh master data and current tab data
        await fetchMasterData();
        await fetchData();
      } else {
        addToast(res?.message || "Failed to save", "error");
      }
    } catch (err) {
      addToast(err.message || "Failed to save", "error");
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async () => {
    if (!deleteTarget) return;
    setSaving(true);
    try {
      let res;
      if (activeTab === "types") {
        res = await api.deleteResourceType(deleteTarget);
      } else if (activeTab === "items") {
        res = await api.deleteResourceItem(deleteTarget);
      } else if (activeTab === "rates") {
        res = await api.deleteResourceRate(deleteTarget);
      } else if (activeTab === "stock") {
        res = await api.deleteStock(deleteTarget);
      }

      if (res?.success) {
        addToast("Deleted successfully", "success");
        setDeleteTarget(null);

        // Refresh master data and current tab data
        await fetchMasterData();
        await fetchData();
      } else {
        addToast(res?.message || "Failed to delete", "error");
      }
    } catch (err) {
      addToast(err.message || "Failed to delete", "error");
    } finally {
      setSaving(false);
    }
  };

  const openCreate = () => {
    setEditingItem(null);
    if (activeTab === "types") {
      setForm({ name: "", unit: "" });
    } else if (activeTab === "items") {
      setForm({ name: "", resource_type_id: "" });
    } else if (activeTab === "rates") {
      setForm({ resource_item_id: "", fiscal_year_id: "", rate_per_unit: "" });
    } else if (activeTab === "stock") {
      setForm({ resource_item_id: "", fiscal_year_id: "", total_quantity: "" });
    }
    setShowModal(true);
  };

  const openEdit = (item) => {
    setEditingItem(item);
    if (activeTab === "types") {
      setForm({ name: item.name, unit: item.unit });
    } else if (activeTab === "items") {
      setForm({
        name: item.name,
        resource_type_id: String(item.resource_type_id || ""),
      });
    } else if (activeTab === "rates") {
      setForm({
        resource_item_id: String(item.resource_item_id || ""),
        fiscal_year_id: String(item.fiscal_year_id || ""),
        rate_per_unit: String(item.rate_per_unit || ""),
      });
    } else if (activeTab === "stock") {
      setForm({
        resource_item_id: String(item.resource_item_id || ""),
        fiscal_year_id: String(item.fiscal_year_id || ""),
        total_quantity: String(item.total_quantity || ""),
      });
    }
    setShowModal(true);
  };

  // Helper to get items grouped by type for better display
  const getItemsGroupedByType = () => {
    const grouped = {};
    items.forEach((item) => {
      const typeName = item.type?.name || "Uncategorized";
      if (!grouped[typeName]) {
        grouped[typeName] = [];
      }
      grouped[typeName].push(item);
    });
    return grouped;
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">
            Resources Management
          </h1>
          <p className="text-sm text-gray-500 dark:text-gray-400">
            Manage resource types, items, rates, and stock inventory
          </p>
        </div>
        <div className="flex gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => {
              setRefreshing(true);
              fetchMasterData();
              fetchData();
            }}
            isLoading={refreshing}
          >
            <RefreshCw size={14} className="mr-1" /> Refresh
          </Button>
          {canEdit && (
            <Button onClick={openCreate}>
              <Plus size={16} /> Add {activeTab.slice(0, -1)}
            </Button>
          )}
        </div>
      </div>

      {/* Tabs */}
      <div className="flex flex-wrap gap-1 bg-gray-100 dark:bg-gray-800/50 rounded-lg p-1 w-fit">
        {tabs.map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            className={`flex items-center gap-2 px-4 py-2 rounded-md text-sm font-medium transition-colors ${
              activeTab === tab.id
                ? "bg-white dark:bg-gray-900 text-emerald-600 dark:text-emerald-400 shadow-sm"
                : "text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200"
            }`}
          >
            <tab.icon size={16} />
            {tab.label}
          </button>
        ))}
      </div>

      {/* Content */}
      <Card>
        <CardContent className="p-0">
          {isLoading ? (
            <LoadingSpinner text={`Loading ${activeTab}...`} />
          ) : (
            <>
              {/* Resource Types Table */}
              {activeTab === "types" && (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>ID</TableHead>
                      <TableHead>Name</TableHead>
                      <TableHead>Unit</TableHead>
                      <TableHead>Items Count</TableHead>
                      {canEdit && <TableHead>Actions</TableHead>}
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {types.length === 0 ? (
                      <TableRow>
                        <TableCell
                          colSpan={canEdit ? 5 : 4}
                          className="text-center py-8 text-gray-500"
                        >
                          No resource types found. Click "Add Resource Type" to
                          create one.
                        </TableCell>
                      </TableRow>
                    ) : (
                      types.map((t) => (
                        <TableRow key={t.id}>
                          <TableCell className="font-mono text-xs">
                            #{t.id}
                          </TableCell>
                          <TableCell className="font-medium">
                            {t.name}
                          </TableCell>
                          <TableCell>
                            <Badge status="active">{t.unit}</Badge>
                          </TableCell>
                          <TableCell>{t.items?.length || 0}</TableCell>
                          {canEdit && (
                            <TableCell>
                              <div className="flex gap-1">
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  onClick={() => openEdit(t)}
                                >
                                  <Edit2 size={15} />
                                </Button>
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  onClick={() => setDeleteTarget(t.id)}
                                >
                                  <Trash2 size={15} className="text-red-500" />
                                </Button>
                              </div>
                            </TableCell>
                          )}
                        </TableRow>
                      ))
                    )}
                  </TableBody>
                </Table>
              )}

              {/* Resource Items Table */}
              {activeTab === "items" && (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>ID</TableHead>
                      <TableHead>Name</TableHead>
                      <TableHead>Resource Type</TableHead>
                      <TableHead>Unit</TableHead>
                      {canEdit && <TableHead>Actions</TableHead>}
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {items.length === 0 ? (
                      <TableRow>
                        <TableCell
                          colSpan={canEdit ? 5 : 4}
                          className="text-center py-8 text-gray-500"
                        >
                          No resource items found. Click "Add Resource Item" to
                          create one.
                        </TableCell>
                      </TableRow>
                    ) : (
                      items.map((i) => (
                        <TableRow key={i.id}>
                          <TableCell className="font-mono text-xs">
                            #{i.id}
                          </TableCell>
                          <TableCell className="font-medium">
                            {i.name}
                          </TableCell>
                          <TableCell>{i.type?.name || "-"}</TableCell>
                          <TableCell>{i.type?.unit || "-"}</TableCell>
                          {canEdit && (
                            <TableCell>
                              <div className="flex gap-1">
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  onClick={() => openEdit(i)}
                                >
                                  <Edit2 size={15} />
                                </Button>
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  onClick={() => setDeleteTarget(i.id)}
                                >
                                  <Trash2 size={15} className="text-red-500" />
                                </Button>
                              </div>
                            </TableCell>
                          )}
                        </TableRow>
                      ))
                    )}
                  </TableBody>
                </Table>
              )}

              {/* Rates Table */}
              {activeTab === "rates" && (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>ID</TableHead>
                      <TableHead>Resource Item</TableHead>
                      <TableHead>Fiscal Year</TableHead>
                      <TableHead>Rate Per Unit</TableHead>
                      {canEdit && <TableHead>Actions</TableHead>}
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {rates.length === 0 ? (
                      <TableRow>
                        <TableCell
                          colSpan={canEdit ? 5 : 4}
                          className="text-center py-8 text-gray-500"
                        >
                          No rates configured. Click "Add Rate" to set resource
                          prices.
                        </TableCell>
                      </TableRow>
                    ) : (
                      rates.map((r) => (
                        <TableRow key={r.id}>
                          <TableCell className="font-mono text-xs">
                            #{r.id}
                          </TableCell>
                          <TableCell className="font-medium">
                            {r.item?.name || "-"}
                          </TableCell>
                          <TableCell>{r.fiscal_year?.name || "-"}</TableCell>
                          <TableCell className="font-semibold text-emerald-600">
                            {formatCurrency(r.rate_per_unit)} per{" "}
                            {r.item?.type?.unit || "unit"}
                          </TableCell>
                          {canEdit && (
                            <TableCell>
                              <div className="flex gap-1">
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  onClick={() => openEdit(r)}
                                >
                                  <Edit2 size={15} />
                                </Button>
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  onClick={() => setDeleteTarget(r.id)}
                                >
                                  <Trash2 size={15} className="text-red-500" />
                                </Button>
                              </div>
                            </TableCell>
                          )}
                        </TableRow>
                      ))
                    )}
                  </TableBody>
                </Table>
              )}

              {/* Stock Table */}
              {activeTab === "stock" && (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>ID</TableHead>
                      <TableHead>Resource Item</TableHead>
                      <TableHead>Fiscal Year</TableHead>
                      <TableHead>Total Quantity</TableHead>
                      <TableHead>Remaining</TableHead>
                      <TableHead>Usage %</TableHead>
                      {canEdit && <TableHead>Actions</TableHead>}
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {stock.length === 0 ? (
                      <TableRow>
                        <TableCell
                          colSpan={canEdit ? 7 : 6}
                          className="text-center py-8 text-gray-500"
                        >
                          No stock entries found. Click "Add Stock" to record
                          inventory.
                        </TableCell>
                      </TableRow>
                    ) : (
                      stock.map((s) => {
                        const usagePercent =
                          s.total_quantity > 0
                            ? (
                                ((s.total_quantity - s.remaining_quantity) /
                                  s.total_quantity) *
                                100
                              ).toFixed(1)
                            : 0;
                        return (
                          <TableRow key={s.id}>
                            <TableCell className="font-mono text-xs">
                              #{s.id}
                            </TableCell>
                            <TableCell className="font-medium">
                              {s.item?.name || "-"}
                            </TableCell>
                            <TableCell>{s.fiscal_year?.name || "-"}</TableCell>
                            <TableCell>
                              {s.total_quantity} {s.item?.type?.unit || ""}
                            </TableCell>
                            <TableCell className="font-semibold">
                              {s.remaining_quantity} {s.item?.type?.unit || ""}
                            </TableCell>
                            <TableCell>
                              <div className="flex items-center gap-2">
                                <div className="w-20 h-2 bg-gray-200 rounded-full overflow-hidden">
                                  <div
                                    className="h-full bg-emerald-500 rounded-full"
                                    style={{ width: `${usagePercent}%` }}
                                  />
                                </div>
                                <span className="text-xs text-gray-500">
                                  {usagePercent}%
                                </span>
                              </div>
                            </TableCell>
                            {canEdit && (
                              <TableCell>
                                <div className="flex gap-1">
                                  <Button
                                    variant="ghost"
                                    size="icon"
                                    onClick={() => openEdit(s)}
                                  >
                                    <Edit2 size={15} />
                                  </Button>
                                  <Button
                                    variant="ghost"
                                    size="icon"
                                    onClick={() => setDeleteTarget(s.id)}
                                  >
                                    <Trash2
                                      size={15}
                                      className="text-red-500"
                                    />
                                  </Button>
                                </div>
                              </TableCell>
                            )}
                          </TableRow>
                        );
                      })
                    )}
                  </TableBody>
                </Table>
              )}
            </>
          )}
        </CardContent>
      </Card>

      {/* Add/Edit Modal */}
      <Modal
        isOpen={showModal}
        onClose={() => {
          setShowModal(false);
          setEditingItem(null);
        }}
        title={
          editingItem
            ? `Edit ${activeTab.slice(0, -1)}`
            : `Add ${activeTab.slice(0, -1)}`
        }
        size={activeTab === "rates" || activeTab === "stock" ? "lg" : "md"}
        footer={
          <>
            <Button variant="outline" onClick={() => setShowModal(false)}>
              Cancel
            </Button>
            <Button onClick={handleSave} isLoading={saving}>
              {editingItem ? "Update" : "Create"}
            </Button>
          </>
        }
      >
        <div className="space-y-4">
          {activeTab === "types" && (
            <>
              <Input
                label="Name"
                value={form.name || ""}
                onChange={(e) => setForm({ ...form, name: e.target.value })}
                placeholder="e.g., Timber, Firewood, Grass"
                required
              />
              <Input
                label="Unit"
                value={form.unit || ""}
                onChange={(e) => setForm({ ...form, unit: e.target.value })}
                placeholder="e.g., cft, kg, bundle, pieces"
                required
              />
            </>
          )}

          {activeTab === "items" && (
            <>
              <Select
                label="Resource Type"
                value={form.resource_type_id || ""}
                onChange={(e) =>
                  setForm({ ...form, resource_type_id: e.target.value })
                }
                options={types.map((t) => ({
                  value: String(t.id),
                  label: `${t.name} (${t.unit})`,
                }))}
                placeholder="Select resource type"
                required
              />
              <Input
                label="Item Name"
                value={form.name || ""}
                onChange={(e) => setForm({ ...form, name: e.target.value })}
                placeholder="e.g., Sal Wood, Pine Timber"
                required
              />
            </>
          )}

          {activeTab === "rates" && (
            <>
              <Select
                label="Resource Item"
                value={form.resource_item_id || ""}
                onChange={(e) =>
                  setForm({ ...form, resource_item_id: e.target.value })
                }
                options={items.map((i) => ({
                  value: String(i.id),
                  label: `${i.name} (${i.type?.name || "No Type"} - ${i.type?.unit || "unit"})`,
                }))}
                placeholder="Select resource item"
                required
              />
              <Select
                label="Fiscal Year"
                value={form.fiscal_year_id || ""}
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
              <Input
                label="Rate Per Unit (NPR)"
                type="number"
                step="0.01"
                value={form.rate_per_unit || ""}
                onChange={(e) =>
                  setForm({ ...form, rate_per_unit: e.target.value })
                }
                placeholder="e.g., 800"
                required
              />
            </>
          )}

          {activeTab === "stock" && (
            <>
              <Select
                label="Resource Item"
                value={form.resource_item_id || ""}
                onChange={(e) =>
                  setForm({ ...form, resource_item_id: e.target.value })
                }
                options={items.map((i) => ({
                  value: String(i.id),
                  label: `${i.name} (${i.type?.name || "No Type"} - ${i.type?.unit || "unit"})`,
                }))}
                placeholder="Select resource item"
                required
              />
              <Select
                label="Fiscal Year"
                value={form.fiscal_year_id || ""}
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
              <Input
                label="Total Quantity"
                type="number"
                step="0.01"
                value={form.total_quantity || ""}
                onChange={(e) =>
                  setForm({ ...form, total_quantity: e.target.value })
                }
                placeholder="Enter total available quantity"
                required
              />
              <p className="text-xs text-gray-500">
                Note: Remaining quantity will be automatically calculated based
                on sales/usage.
              </p>
            </>
          )}
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
            <Button variant="danger" onClick={handleDelete} isLoading={saving}>
              Delete
            </Button>
          </>
        }
      >
        <div className="flex items-start gap-3">
          <AlertTriangle size={20} className="text-red-500 shrink-0 mt-0.5" />
          <div>
            <p className="text-sm text-gray-700 dark:text-gray-300">
              Are you sure you want to delete this {activeTab.slice(0, -1)}?
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
