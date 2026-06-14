import { useState, useEffect, useCallback, useRef } from "react";
import { useSelector } from "react-redux";
import {
  Plus,
  Search,
  Edit2,
  Trash2,
  Tag,
  Receipt,
  AlertTriangle,
  RefreshCw,
  Upload,
  Image as ImageIcon,
  Eye,
  X,
  Calendar,
  Building2,
  CreditCard,
  FileText,
  Download,
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

const emptyExpense = {
  fiscal_year_id: "",
  category_id: "",
  title: "",
  amount: "",
  expense_date: new Date().toISOString().split("T")[0],
  payment_method: "cash",
  paid_to: "",
  receipt_no: "",
  bill_photo: "",
  remarks: "",
};

const emptyCategory = {
  name: "",
  description: "",
};

export default function Expenses() {
  const { user } = useSelector((state) => state.auth);
  const { addToast } = useToast();
  const canEdit = user?.role === "admin" || user?.role === "staff";
  const isAdmin = user?.role === "admin";

  const [activeTab, setActiveTab] = useState("expenses");
  const [expenses, setExpenses] = useState([]);
  const [categories, setCategories] = useState([]);
  const [fiscalYears, setFiscalYears] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [categoryFilter, setCategoryFilter] = useState("");
  const [fiscalYearFilter, setFiscalYearFilter] = useState("");
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [refreshing, setRefreshing] = useState(false);

  // Expense Modal
  const [showExpenseModal, setShowExpenseModal] = useState(false);
  const [editingExpense, setEditingExpense] = useState(null);
  const [expenseForm, setExpenseForm] = useState(emptyExpense);
  const [saving, setSaving] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [viewingExpense, setViewingExpense] = useState(null);
  const [showViewModal, setShowViewModal] = useState(false);

  // Category Modal
  const [showCategoryModal, setShowCategoryModal] = useState(false);
  const [editingCategory, setEditingCategory] = useState(null);
  const [categoryForm, setCategoryForm] = useState(emptyCategory);

  const [deleteTarget, setDeleteTarget] = useState(null);
  const fileInputRef = useRef(null);

  const fetchMasterData = useCallback(async () => {
    try {
      const [categoriesRes, fiscalYearsRes] = await Promise.all([
        api.getExpenseCategories(),
        api.getFiscalYears(),
      ]);
      if (categoriesRes.success) setCategories(categoriesRes.data || []);
      if (fiscalYearsRes.success) {
        const years = fiscalYearsRes.data || [];
        const activeId = getActiveFiscalYearId(years);
        setFiscalYears(years);
        setFiscalYearFilter((current) => current || activeId);
        setExpenseForm((current) => ({
          ...current,
          fiscal_year_id: current.fiscal_year_id || activeId,
        }));
      }
    } catch (err) {
      console.error("Failed to fetch master data:", err);
    }
  }, []);

  const fetchExpenses = useCallback(async () => {
    setIsLoading(true);
    try {
      const res = await api.getExpenses({
        page,
        per_page: 10,
        search: search || undefined,
        category_id: categoryFilter || undefined,
        fiscal_year_id: fiscalYearFilter || undefined,
      });
      if (res.success) {
        setExpenses(res.data || []);
        setTotalPages(res.meta?.total_pages || 1);
      }
    } catch (_err) {
      addToast("Failed to load expenses", "error");
    } finally {
      setIsLoading(false);
      setRefreshing(false);
    }
  }, [page, search, categoryFilter, fiscalYearFilter, addToast]);

  useEffect(() => {
    fetchMasterData();
  }, [fetchMasterData]);

  useEffect(() => {
    if (activeTab === "expenses") {
      fetchExpenses();
    }
  }, [fetchExpenses, activeTab]);

  const newExpenseForm = () => ({
    ...emptyExpense,
    fiscal_year_id: getActiveFiscalYearId(fiscalYears),
    expense_date: new Date().toISOString().split("T")[0],
  });

  const openCreateExpense = () => {
    setEditingExpense(null);
    setExpenseForm(newExpenseForm());
    setShowExpenseModal(true);
  };

  const handleSaveExpense = async () => {
    setSaving(true);
    try {
      const payload = {
        fiscal_year_id: Number(expenseForm.fiscal_year_id),
        category_id: Number(expenseForm.category_id),
        title: expenseForm.title,
        amount: Number(expenseForm.amount),
        expense_date: expenseForm.expense_date,
        payment_method: expenseForm.payment_method,
        paid_to: expenseForm.paid_to,
        receipt_no: expenseForm.receipt_no || null,
        bill_photo: expenseForm.bill_photo || null,
        remarks: expenseForm.remarks || null,
      };

      const res = editingExpense
        ? await api.updateExpense(editingExpense.id, payload)
        : await api.createExpense(payload);

      if (res.success) {
        addToast(
          `Expense ${editingExpense ? "updated" : "created"} successfully`,
          "success",
        );
        setShowExpenseModal(false);
        setEditingExpense(null);
        setExpenseForm(newExpenseForm());
        fetchExpenses();
      } else {
        addToast(res.message || "Failed to save", "error");
      }
    } catch (err) {
      addToast(err.message || "Failed to save", "error");
    } finally {
      setSaving(false);
    }
  };

  const handleSaveCategory = async () => {
    setSaving(true);
    try {
      const res = editingCategory
        ? await api.updateExpenseCategory(editingCategory.id, categoryForm)
        : await api.createExpenseCategory(categoryForm);

      if (res.success) {
        addToast(
          `Category ${editingCategory ? "updated" : "created"} successfully`,
          "success",
        );
        setShowCategoryModal(false);
        setEditingCategory(null);
        setCategoryForm(emptyCategory);
        fetchMasterData();
      } else {
        addToast(res.message || "Failed to save", "error");
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
      if (deleteTarget.type === "expense") {
        res = await api.deleteExpense(deleteTarget.id);
      } else {
        res = await api.deleteExpenseCategory(deleteTarget.id);
      }

      if (res.success) {
        addToast("Deleted successfully", "success");
        setDeleteTarget(null);
        if (deleteTarget.type === "expense") {
          fetchExpenses();
        } else {
          fetchMasterData();
        }
      } else {
        addToast(res.message || "Failed to delete", "error");
      }
    } catch (err) {
      addToast(err.message || "Failed to delete", "error");
    } finally {
      setSaving(false);
    }
  };

  const handleBillPhotoUpload = async (expenseId, file) => {
    const formData = new FormData();
    formData.append("photo", file);

    setUploading(true);
    try {
      const res = await api.uploadExpenseBillPhoto(expenseId, formData);
      if (res.success) {
        addToast("Bill photo uploaded successfully", "success");
        fetchExpenses();
        if (viewingExpense && viewingExpense.id === expenseId) {
          const updated = await api.getExpense(expenseId);
          if (updated.success) setViewingExpense(updated.data);
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

  const openViewExpense = async (expense) => {
    setViewingExpense(expense);
    setShowViewModal(true);
  };

  const openEditExpense = (expense) => {
    setEditingExpense(expense);
    setExpenseForm({
      fiscal_year_id: String(expense.fiscal_year_id || ""),
      category_id: String(expense.category_id || ""),
      title: expense.title,
      amount: String(expense.amount),
      expense_date: expense.expense_date?.split("T")[0] || "",
      payment_method: expense.payment_method,
      paid_to: expense.paid_to,
      receipt_no: expense.receipt_no || "",
      bill_photo: expense.bill_photo || "",
      remarks: expense.remarks || "",
    });
    setShowExpenseModal(true);
  };

  const openEditCategory = (category) => {
    setEditingCategory(category);
    setCategoryForm({
      name: category.name,
      description: category.description || "",
    });
    setShowCategoryModal(true);
  };

  const paymentMethodOptions = [
    { value: "cash", label: "💰 Cash" },
    { value: "bank", label: "🏦 Bank Transfer" },
    { value: "online", label: "💳 Online Payment" },
  ];

  const getPaymentMethodIcon = (method) => {
    switch (method) {
      case "cash":
        return "💰";
      case "bank":
        return "🏦";
      case "online":
        return "💳";
      default:
        return "💵";
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">
            Expenses
          </h1>
          <p className="text-sm text-gray-500 dark:text-gray-400">
            Track and manage organization expenses
          </p>
        </div>
        <div className="flex gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => {
              setRefreshing(true);
              fetchExpenses();
            }}
            isLoading={refreshing}
          >
            <RefreshCw size={14} className="mr-1" /> Refresh
          </Button>
          {canEdit && (
            <Button
              onClick={openCreateExpense}
            >
              <Plus size={16} /> Add Expense
            </Button>
          )}
        </div>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 bg-gray-100 dark:bg-gray-800/50 rounded-lg p-1 w-fit">
        <button
          onClick={() => setActiveTab("expenses")}
          className={`flex items-center gap-2 px-4 py-2 rounded-md text-sm font-medium transition-colors ${
            activeTab === "expenses"
              ? "bg-white dark:bg-gray-900 text-emerald-600 dark:text-emerald-400 shadow-sm"
              : "text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200"
          }`}
        >
          <Receipt size={16} />
          Expenses
        </button>
        <button
          onClick={() => setActiveTab("categories")}
          className={`flex items-center gap-2 px-4 py-2 rounded-md text-sm font-medium transition-colors ${
            activeTab === "categories"
              ? "bg-white dark:bg-gray-900 text-emerald-600 dark:text-emerald-400 shadow-sm"
              : "text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200"
          }`}
        >
          <Tag size={16} />
          Categories
        </button>
      </div>

      {/* Expenses Tab */}
      {activeTab === "expenses" && (
        <>
          {/* Filters */}
          <Card>
            <CardContent className="p-4">
              <div className="flex flex-col sm:flex-row gap-3">
                <div className="flex-1 relative">
                  <Search
                    size={16}
                    className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400"
                  />
                  <input
                    type="text"
                    placeholder="Search expenses..."
                    value={search}
                    onChange={(e) => {
                      setSearch(e.target.value);
                      setPage(1);
                    }}
                    className="w-full pl-10 pr-3 py-2 border border-gray-200 dark:border-white/10 rounded-lg bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 text-sm focus:outline-none focus:border-emerald-500 focus:ring-2 focus:ring-emerald-500/20"
                  />
                </div>
                <Select
                  value={categoryFilter}
                  onChange={(e) => {
                    setCategoryFilter(e.target.value);
                    setPage(1);
                  }}
                  options={[
                    { value: "", label: "All Categories" },
                    ...categories.map((c) => ({
                      value: String(c.id),
                      label: c.name,
                    })),
                  ]}
                  className="w-48"
                />
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
              </div>
            </CardContent>
          </Card>

          {/* Expenses Table */}
          <Card>
            <CardContent className="p-0">
              {isLoading ? (
                <LoadingSpinner text="Loading expenses..." />
              ) : expenses.length === 0 ? (
                <div className="flex flex-col items-center justify-center py-12 text-gray-400">
                  <Receipt size={48} className="mb-3 opacity-30" />
                  <p className="text-lg font-medium">No expenses found</p>
                  {canEdit && (
                    <Button
                      variant="outline"
                      className="mt-4"
                      onClick={openCreateExpense}
                    >
                      <Plus size={14} /> Add your first expense
                    </Button>
                  )}
                </div>
              ) : (
                <div className="overflow-x-auto">
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>Date</TableHead>
                        <TableHead>Title</TableHead>
                        <TableHead>Category</TableHead>
                        <TableHead>Paid To</TableHead>
                        <TableHead>Amount</TableHead>
                        <TableHead>Method</TableHead>
                        <TableHead>Receipt</TableHead>
                        <TableHead>Actions</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {expenses.map((exp) => (
                        <TableRow key={exp.id}>
                          <TableCell>{formatDate(exp.expense_date)}</TableCell>
                          <TableCell className="font-medium max-w-50cate">
                            {exp.title}
                          </TableCell>
                          <TableCell>
                            <span className="inline-flex items-center gap-1">
                              <Tag size={12} className="text-gray-400" />
                              {exp.category?.name || "-"}
                            </span>
                          </TableCell>
                          <TableCell className="max-w-37.5 truncate">
                            {exp.paid_to}
                          </TableCell>
                          <TableCell className="font-semibold text-red-600">
                            -{formatCurrency(exp.amount)}
                          </TableCell>
                          <TableCell>
                            <span className="text-sm">
                              {getPaymentMethodIcon(exp.payment_method)}{" "}
                              {exp.payment_method}
                            </span>
                          </TableCell>
                          <TableCell>
                            {exp.receipt_no ? (
                              <span className="text-xs font-mono text-gray-500">
                                {exp.receipt_no}
                              </span>
                            ) : (
                              "-"
                            )}
                          </TableCell>
                          <TableCell>
                            <div className="flex items-center gap-1">
                              <Button
                                variant="ghost"
                                size="icon"
                                onClick={() => openViewExpense(exp)}
                                title="View Details"
                              >
                                <Eye size={15} />
                              </Button>
                              {canEdit && (
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  onClick={() => openEditExpense(exp)}
                                  title="Edit"
                                >
                                  <Edit2 size={15} />
                                </Button>
                              )}
                              {isAdmin && (
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  onClick={() =>
                                    setDeleteTarget({
                                      type: "expense",
                                      id: exp.id,
                                    })
                                  }
                                  title="Delete"
                                >
                                  <Trash2
                                    size={15}
                                    className="text-red-500"
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
        </>
      )}

      {/* Categories Tab */}
      {activeTab === "categories" && (
        <Card>
          <CardContent className="p-0">
            <div className="p-4 border-b border-gray-200 dark:border-white/10 flex justify-end">
              {isAdmin && (
                <Button
                  onClick={() => {
                    setEditingCategory(null);
                    setCategoryForm(emptyCategory);
                    setShowCategoryModal(true);
                  }}
                >
                  <Plus size={16} /> Add Category
                </Button>
              )}
            </div>
            {categories.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-12 text-gray-400">
                <Tag size={48} className="mb-3 opacity-30" />
                <p className="text-lg font-medium">No categories found</p>
                {isAdmin && (
                  <Button
                    variant="outline"
                    className="mt-4"
                    onClick={() => setShowCategoryModal(true)}
                  >
                    <Plus size={14} /> Add your first category
                  </Button>
                )}
              </div>
            ) : (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>ID</TableHead>
                    <TableHead>Name</TableHead>
                    <TableHead>Description</TableHead>
                    <TableHead>Expenses Count</TableHead>
                    {isAdmin && <TableHead>Actions</TableHead>}
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {categories.map((cat) => (
                    <TableRow key={cat.id}>
                      <TableCell className="font-mono text-xs">
                        #{cat.id}
                      </TableCell>
                      <TableCell className="font-medium">{cat.name}</TableCell>
                      <TableCell className="max-w-75 truncate">
                        {cat.description || "-"}
                      </TableCell>
                      <TableCell>{cat.expenses?.length || 0}</TableCell>
                      {isAdmin && (
                        <TableCell>
                          <div className="flex items-center gap-1">
                            <Button
                              variant="ghost"
                              size="icon"
                              onClick={() => openEditCategory(cat)}
                            >
                              <Edit2 size={15} />
                            </Button>
                            <Button
                              variant="ghost"
                              size="icon"
                              onClick={() =>
                                setDeleteTarget({
                                  type: "category",
                                  id: cat.id,
                                })
                              }
                            >
                              <Trash2 size={15} className="text-red-500" />
                            </Button>
                          </div>
                        </TableCell>
                      )}
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            )}
          </CardContent>
        </Card>
      )}

      {/* Add/Edit Expense Modal */}
      <Modal
        isOpen={showExpenseModal}
        onClose={() => {
          setShowExpenseModal(false);
          setEditingExpense(null);
          setExpenseForm(newExpenseForm());
        }}
        title={editingExpense ? "Edit Expense" : "Add Expense"}
        size="lg"
        footer={
          <>
            <Button
              variant="outline"
              onClick={() => setShowExpenseModal(false)}
            >
              Cancel
            </Button>
            <Button onClick={handleSaveExpense} isLoading={saving}>
              {editingExpense ? "Update" : "Create"}
            </Button>
          </>
        }
      >
        <div className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Select
              label="Fiscal Year"
              value={expenseForm.fiscal_year_id}
              onChange={(e) =>
                setExpenseForm({
                  ...expenseForm,
                  fiscal_year_id: e.target.value,
                })
              }
              options={buildFiscalYearOptions(fiscalYears)}
              placeholder="Select fiscal year"
              required
            />
            <Select
              label="Category"
              value={expenseForm.category_id}
              onChange={(e) =>
                setExpenseForm({ ...expenseForm, category_id: e.target.value })
              }
              options={categories.map((c) => ({
                value: String(c.id),
                label: c.name,
              }))}
              placeholder="Select category"
              required
            />
          </div>

          <Input
            label="Title"
            value={expenseForm.title}
            onChange={(e) =>
              setExpenseForm({ ...expenseForm, title: e.target.value })
            }
            placeholder="Brief description of the expense"
            required
          />

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Input
              label="Amount (NPR)"
              type="number"
              step="0.01"
              value={expenseForm.amount}
              onChange={(e) =>
                setExpenseForm({ ...expenseForm, amount: e.target.value })
              }
              placeholder="0.00"
              required
            />
            <Input
              label="Expense Date"
              type="date"
              value={expenseForm.expense_date}
              onChange={(e) =>
                setExpenseForm({ ...expenseForm, expense_date: e.target.value })
              }
              required
            />
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Select
              label="Payment Method"
              value={expenseForm.payment_method}
              onChange={(e) =>
                setExpenseForm({
                  ...expenseForm,
                  payment_method: e.target.value,
                })
              }
              options={paymentMethodOptions}
              required
            />
            <Input
              label="Paid To"
              value={expenseForm.paid_to}
              onChange={(e) =>
                setExpenseForm({ ...expenseForm, paid_to: e.target.value })
              }
              placeholder="Recipient name"
              required
            />
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Input
              label="Receipt Number"
              value={expenseForm.receipt_no}
              onChange={(e) =>
                setExpenseForm({ ...expenseForm, receipt_no: e.target.value })
              }
              placeholder="Optional receipt reference"
            />
            <Input
              label="Bill Photo URL"
              value={expenseForm.bill_photo}
              onChange={(e) =>
                setExpenseForm({ ...expenseForm, bill_photo: e.target.value })
              }
              placeholder="Image URL (or upload after creation)"
            />
          </div>

          <Textarea
            label="Remarks"
            value={expenseForm.remarks}
            onChange={(e) =>
              setExpenseForm({ ...expenseForm, remarks: e.target.value })
            }
            placeholder="Additional notes..."
            rows={3}
          />
        </div>
      </Modal>

      {/* Add/Edit Category Modal */}
      <Modal
        isOpen={showCategoryModal}
        onClose={() => {
          setShowCategoryModal(false);
          setEditingCategory(null);
          setCategoryForm(emptyCategory);
        }}
        title={editingCategory ? "Edit Category" : "Add Category"}
        footer={
          <>
            <Button
              variant="outline"
              onClick={() => setShowCategoryModal(false)}
            >
              Cancel
            </Button>
            <Button onClick={handleSaveCategory} isLoading={saving}>
              {editingCategory ? "Update" : "Create"}
            </Button>
          </>
        }
      >
        <div className="space-y-4">
          <Input
            label="Category Name"
            value={categoryForm.name}
            onChange={(e) =>
              setCategoryForm({ ...categoryForm, name: e.target.value })
            }
            placeholder="e.g., Salary, Office Rent, Maintenance"
            required
          />
          <Textarea
            label="Description"
            value={categoryForm.description}
            onChange={(e) =>
              setCategoryForm({ ...categoryForm, description: e.target.value })
            }
            placeholder="Optional description"
            rows={3}
          />
        </div>
      </Modal>

      {/* View Expense Modal */}
      <Modal
        isOpen={showViewModal}
        onClose={() => {
          setShowViewModal(false);
          setViewingExpense(null);
        }}
        title="Expense Details"
        size="lg"
      >
        {viewingExpense && (
          <div className="space-y-6">
            {/* Header with Amount */}
            <div className="flex items-center justify-between">
              <Badge status="debit" className="text-sm px-3 py-1">
                Expense
              </Badge>
              <span className="text-2xl font-bold text-red-600">
                -{formatCurrency(viewingExpense.amount)}
              </span>
            </div>

            {/* Title */}
            <div className="border-b border-gray-200 dark:border-white/10 pb-4">
              <h2 className="text-xl font-bold text-gray-900 dark:text-gray-100">
                {viewingExpense.title}
              </h2>
            </div>

            {/* Details Grid */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="flex items-start gap-3">
                <Calendar size={18} className="text-gray-400 mt-0.5" />
                <div>
                  <p className="text-xs text-gray-400">Expense Date</p>
                  <p className="text-gray-900 dark:text-gray-100">
                    {formatDate(viewingExpense.expense_date)}
                  </p>
                </div>
              </div>
              <div className="flex items-start gap-3">
                <Building2 size={18} className="text-gray-400 mt-0.5" />
                <div>
                  <p className="text-xs text-gray-400">Fiscal Year</p>
                  <p className="text-gray-900 dark:text-gray-100">
                    {viewingExpense.fiscal_year?.name || "-"}
                  </p>
                </div>
              </div>
              <div className="flex items-start gap-3">
                <Tag size={18} className="text-gray-400 mt-0.5" />
                <div>
                  <p className="text-xs text-gray-400">Category</p>
                  <p className="text-gray-900 dark:text-gray-100">
                    {viewingExpense.category?.name || "-"}
                  </p>
                </div>
              </div>
              <div className="flex items-start gap-3">
                <CreditCard size={18} className="text-gray-400 mt-0.5" />
                <div>
                  <p className="text-xs text-gray-400">Payment Method</p>
                  <p className="text-gray-900 dark:text-gray-100 capitalize">
                    {getPaymentMethodIcon(viewingExpense.payment_method)}{" "}
                    {viewingExpense.payment_method}
                  </p>
                </div>
              </div>
              <div className="flex items-start gap-3">
                <FileText size={18} className="text-gray-400 mt-0.5" />
                <div>
                  <p className="text-xs text-gray-400">Paid To</p>
                  <p className="text-gray-900 dark:text-gray-100">
                    {viewingExpense.paid_to}
                  </p>
                </div>
              </div>
              <div className="flex items-start gap-3">
                <Receipt size={18} className="text-gray-400 mt-0.5" />
                <div>
                  <p className="text-xs text-gray-400">Receipt Number</p>
                  <p className="text-gray-900 dark:text-gray-100 font-mono">
                    {viewingExpense.receipt_no || "-"}
                  </p>
                </div>
              </div>
            </div>

            {/* Remarks */}
            {viewingExpense.remarks && (
              <div className="bg-gray-50 dark:bg-gray-800/50 rounded-lg p-4">
                <p className="text-xs text-gray-400 mb-1">Remarks</p>
                <p className="text-gray-700 dark:text-gray-300">
                  {viewingExpense.remarks}
                </p>
              </div>
            )}

            {/* Bill Photo Section */}
            <div className="border-t border-gray-200 dark:border-white/10 pt-4">
              <div className="flex items-center justify-between mb-3">
                <h4 className="font-medium text-gray-900 dark:text-gray-100 flex items-center gap-2">
                  <ImageIcon size={16} />
                  Bill Photo / Document
                </h4>
                {canEdit && (
                  <div>
                    <input
                      type="file"
                      ref={fileInputRef}
                      accept="image/*,.pdf"
                      className="hidden"
                      onChange={(e) => {
                        const file = e.target.files?.[0];
                        if (file && viewingExpense) {
                          handleBillPhotoUpload(viewingExpense.id, file);
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
                      {viewingExpense.bill_photo ? "Replace" : "Upload"}
                    </Button>
                  </div>
                )}
              </div>
              {viewingExpense.bill_photo ? (
                <div className="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800/50 rounded-lg">
                  <div className="flex items-center gap-3">
                    <ImageIcon size={20} className="text-emerald-500" />
                    <span className="text-sm text-gray-700 dark:text-gray-300">
                      {viewingExpense.bill_photo.split("/").pop()}
                    </span>
                  </div>
                  <a
                    href={getImageUrl(viewingExpense.bill_photo)}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-emerald-600 hover:text-emerald-700 text-sm flex items-center gap-1"
                  >
                    <Download size={14} /> View
                  </a>
                </div>
              ) : (
                <p className="text-sm text-gray-400 text-center py-4">
                  No bill photo attached
                </p>
              )}
            </div>

            {/* Meta Info */}
            <div className="text-xs text-gray-400 border-t border-gray-200 dark:border-white/10 pt-4">
              <p>Created by: {viewingExpense.creator?.name || "Unknown"}</p>
              <p>Created at: {formatDate(viewingExpense.created_at)}</p>
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
              Are you sure you want to delete this {deleteTarget?.type}?
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
