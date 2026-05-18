import { useState, useEffect, useCallback, useRef } from "react";
import { useSelector } from "react-redux";
import {
  Plus,
  Search,
  Edit2,
  Trash2,
  Mail,
  MailOpen,
  FileText,
  Upload,
  Download,
  Eye,
  X,
  RefreshCw,
  Calendar,
  User,
  Building2,
  Tag,
  AlertTriangle,
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
import { formatDate } from "../../utils/helpers";

const emptyLetter = {
  type: "incoming",
  reference_no: "",
  title: "",
  subject: "",
  from_party: "",
  to_party: "",
  letter_date: new Date().toISOString().split("T")[0],
  received_date: "",
  sent_date: "",
  document_file: "",
  remarks: "",
};

export default function Letters() {
  const { user } = useSelector((state) => state.auth);
  const { addToast } = useToast();
  const canEdit = user?.role === "admin" || user?.role === "staff";
  const isAdmin = user?.role === "admin";

  const [letters, setLetters] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [typeFilter, setTypeFilter] = useState("");
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [refreshing, setRefreshing] = useState(false);

  const [showModal, setShowModal] = useState(false);
  const [showViewModal, setShowViewModal] = useState(false);
  const [editingLetter, setEditingLetter] = useState(null);
  const [viewingLetter, setViewingLetter] = useState(null);
  const [form, setForm] = useState(emptyLetter);
  const [saving, setSaving] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState(null);
  const fileInputRef = useRef(null);

  const fetchLetters = useCallback(async () => {
    setIsLoading(true);
    try {
      const res = await api.getLetters({
        page,
        limit: 10,
        search: search || undefined,
        type: typeFilter || undefined,
      });
      if (res.success) {
        setLetters(res.data || []);
        setTotalPages(res.meta?.total_pages || 1);
      }
    } catch (err) {
      addToast("Failed to load letters", "error");
    } finally {
      setIsLoading(false);
      setRefreshing(false);
    }
  }, [page, search, typeFilter, addToast]);

  useEffect(() => {
    fetchLetters();
  }, [fetchLetters]);

  const handleSave = async () => {
    setSaving(true);
    try {
      const payload = {
        type: form.type,
        reference_no: form.reference_no || null,
        title: form.title,
        subject: form.subject,
        from_party: form.from_party || null,
        to_party: form.to_party || null,
        letter_date: form.letter_date,
        received_date: form.received_date || null,
        sent_date: form.sent_date || null,
        document_file: form.document_file || null,
        remarks: form.remarks || null,
      };

      const res = editingLetter
        ? await api.updateLetter(editingLetter.id, payload)
        : await api.createLetter(payload);

      if (res.success) {
        addToast(
          `Letter ${editingLetter ? "updated" : "created"} successfully`,
          "success",
        );
        setShowModal(false);
        setEditingLetter(null);
        setForm(emptyLetter);
        fetchLetters();
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
      const res = await api.deleteLetter(deleteTarget);
      if (res.success) {
        addToast("Letter deleted successfully", "success");
        setDeleteTarget(null);
        fetchLetters();
      } else {
        addToast(res.message || "Failed to delete", "error");
      }
    } catch (err) {
      addToast(err.message || "Failed to delete", "error");
    } finally {
      setSaving(false);
    }
  };

  const handleDocumentUpload = async (letterId, file) => {
    const formData = new FormData();
    formData.append("document", file);

    setUploading(true);
    try {
      const res = await api.uploadLetterDocument(letterId, formData);
      if (res.success) {
        addToast("Document uploaded successfully", "success");
        fetchLetters();
        if (viewingLetter && viewingLetter.id === letterId) {
          const updated = await api.getLetter(letterId);
          if (updated.success) setViewingLetter(updated.data);
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

  const openView = async (letter) => {
    setViewingLetter(letter);
    setShowViewModal(true);
  };

  const openEdit = (letter) => {
    setEditingLetter(letter);
    setForm({
      type: letter.type,
      reference_no: letter.reference_no || "",
      title: letter.title,
      subject: letter.subject,
      from_party: letter.from_party || "",
      to_party: letter.to_party || "",
      letter_date: letter.letter_date?.split("T")[0] || "",
      received_date: letter.received_date?.split("T")[0] || "",
      sent_date: letter.sent_date?.split("T")[0] || "",
      document_file: letter.document_file || "",
      remarks: letter.remarks || "",
    });
    setShowModal(true);
  };

  const getDocumentIcon = (filename) => {
    if (!filename) return <FileText size={16} />;
    const ext = filename.split(".").pop()?.toLowerCase();
    if (ext === "pdf") return <FileText size={16} className="text-red-500" />;
    if (ext === "doc" || ext === "docx")
      return <FileText size={16} className="text-blue-500" />;
    return <FileText size={16} />;
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">
            Letters
          </h1>
          <p className="text-sm text-gray-500 dark:text-gray-400">
            Manage incoming and outgoing correspondence
          </p>
        </div>
        <div className="flex gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => {
              setRefreshing(true);
              fetchLetters();
            }}
            isLoading={refreshing}
          >
            <RefreshCw size={14} className="mr-1" /> Refresh
          </Button>
          {canEdit && (
            <Button
              onClick={() => {
                setEditingLetter(null);
                setForm({
                  ...emptyLetter,
                  letter_date: new Date().toISOString().split("T")[0],
                });
                setShowModal(true);
              }}
            >
              <Plus size={16} /> Add Letter
            </Button>
          )}
        </div>
      </div>

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
                placeholder="Search by title, subject, reference, from/to..."
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
                { value: "incoming", label: "Incoming" },
                { value: "outgoing", label: "Outgoing" },
              ]}
              className="w-40"
            />
          </div>
        </CardContent>
      </Card>

      {/* Table */}
      <Card>
        <CardContent className="p-0">
          {isLoading ? (
            <LoadingSpinner text="Loading letters..." />
          ) : letters.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 text-gray-400">
              <Mail size={48} className="mb-3 opacity-30" />
              <p className="text-lg font-medium">No letters found</p>
              {canEdit && (
                <Button
                  variant="outline"
                  className="mt-4"
                  onClick={() => setShowModal(true)}
                >
                  <Plus size={14} /> Add your first letter
                </Button>
              )}
            </div>
          ) : (
            <div className="overflow-x-auto">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Ref #</TableHead>
                    <TableHead>Type</TableHead>
                    <TableHead>Title</TableHead>
                    <TableHead>From/To</TableHead>
                    <TableHead>Date</TableHead>
                    <TableHead>Document</TableHead>
                    <TableHead>Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {letters.map((letter) => (
                    <TableRow key={letter.id}>
                      <TableCell className="font-mono text-xs">
                        {letter.reference_no || "-"}
                      </TableCell>
                      <TableCell>
                        <Badge status={letter.type}>
                          {letter.type === "incoming"
                            ? "📥 Incoming"
                            : "📤 Outgoing"}
                        </Badge>
                      </TableCell>
                      <TableCell className="font-medium max-w-50 truncate">
                        {letter.title}
                      </TableCell>
                      <TableCell className="max-w-37.5 truncate">
                        {letter.type === "incoming"
                          ? letter.from_party
                          : letter.to_party || "-"}
                      </TableCell>
                      <TableCell>{formatDate(letter.letter_date)}</TableCell>
                      <TableCell>
                        {letter.document_file ? (
                          <a
                            href={letter.document_file}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="inline-flex items-center gap-1 text-emerald-600 hover:text-emerald-700"
                            onClick={(e) => e.stopPropagation()}
                          >
                            {getDocumentIcon(letter.document_file)}
                            <span className="text-xs">View</span>
                          </a>
                        ) : (
                          <span className="text-gray-400 text-xs">No file</span>
                        )}
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center gap-1">
                          <Button
                            variant="ghost"
                            size="icon"
                            onClick={() => openView(letter)}
                            title="View Details"
                          >
                            <Eye size={15} />
                          </Button>
                          {canEdit && (
                            <>
                              <Button
                                variant="ghost"
                                size="icon"
                                onClick={() => openEdit(letter)}
                                title="Edit"
                              >
                                <Edit2 size={15} />
                              </Button>
                              <Button
                                variant="ghost"
                                size="icon"
                                onClick={() => setDeleteTarget(letter.id)}
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

      {/* Add/Edit Letter Modal */}
      <Modal
        isOpen={showModal}
        onClose={() => {
          setShowModal(false);
          setEditingLetter(null);
          setForm(emptyLetter);
        }}
        title={editingLetter ? "Edit Letter" : "Add Letter"}
        size="lg"
        footer={
          <>
            <Button variant="outline" onClick={() => setShowModal(false)}>
              Cancel
            </Button>
            <Button onClick={handleSave} isLoading={saving}>
              {editingLetter ? "Update" : "Create"}
            </Button>
          </>
        }
      >
        <div className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Select
              label="Letter Type"
              value={form.type}
              onChange={(e) => setForm({ ...form, type: e.target.value })}
              options={[
                { value: "incoming", label: "📥 Incoming" },
                { value: "outgoing", label: "📤 Outgoing" },
              ]}
              required
            />
            <Input
              label="Reference Number"
              value={form.reference_no}
              onChange={(e) =>
                setForm({ ...form, reference_no: e.target.value })
              }
              placeholder="e.g., BS-LET-2024-001"
            />
          </div>

          <Input
            label="Title"
            value={form.title}
            onChange={(e) => setForm({ ...form, title: e.target.value })}
            placeholder="Brief title of the letter"
            required
          />

          <Input
            label="Subject"
            value={form.subject}
            onChange={(e) => setForm({ ...form, subject: e.target.value })}
            placeholder="Subject of the letter"
            required
          />

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {form.type === "incoming" ? (
              <Input
                label="From Party"
                value={form.from_party}
                onChange={(e) =>
                  setForm({ ...form, from_party: e.target.value })
                }
                placeholder="Sender name/organization"
              />
            ) : (
              <Input
                label="To Party"
                value={form.to_party}
                onChange={(e) => setForm({ ...form, to_party: e.target.value })}
                placeholder="Recipient name/organization"
              />
            )}
            <Input
              label="Letter Date"
              type="date"
              value={form.letter_date}
              onChange={(e) =>
                setForm({ ...form, letter_date: e.target.value })
              }
              required
            />
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {form.type === "incoming" && (
              <Input
                label="Received Date"
                type="date"
                value={form.received_date}
                onChange={(e) =>
                  setForm({ ...form, received_date: e.target.value })
                }
              />
            )}
            {form.type === "outgoing" && (
              <Input
                label="Sent Date"
                type="date"
                value={form.sent_date}
                onChange={(e) =>
                  setForm({ ...form, sent_date: e.target.value })
                }
              />
            )}
          </div>

          <Textarea
            label="Remarks"
            value={form.remarks}
            onChange={(e) => setForm({ ...form, remarks: e.target.value })}
            placeholder="Additional notes..."
            rows={3}
          />

          <Input
            label="Document URL (optional)"
            value={form.document_file}
            onChange={(e) =>
              setForm({ ...form, document_file: e.target.value })
            }
            placeholder="https://... or /uploads/letters/..."
          />
          <p className="text-xs text-gray-500">
            You can also upload documents after creating the letter.
          </p>
        </div>
      </Modal>

      {/* View Letter Modal */}
      <Modal
        isOpen={showViewModal}
        onClose={() => {
          setShowViewModal(false);
          setViewingLetter(null);
        }}
        title="Letter Details"
        size="lg"
      >
        {viewingLetter && (
          <div className="space-y-6">
            {/* Header with Type Badge */}
            <div className="flex items-center justify-between">
              <Badge status={viewingLetter.type} className="text-sm px-3 py-1">
                {viewingLetter.type === "incoming"
                  ? "📥 Incoming"
                  : "📤 Outgoing"}
              </Badge>
              {viewingLetter.reference_no && (
                <span className="text-xs font-mono text-gray-500">
                  Ref: {viewingLetter.reference_no}
                </span>
              )}
            </div>

            {/* Title and Subject */}
            <div className="border-b border-gray-200 dark:border-white/10 pb-4">
              <h2 className="text-xl font-bold text-gray-900 dark:text-gray-100">
                {viewingLetter.title}
              </h2>
              <p className="text-gray-600 dark:text-gray-400 mt-2">
                {viewingLetter.subject}
              </p>
            </div>

            {/* Parties */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {viewingLetter.type === "incoming" ? (
                <div className="flex items-start gap-3">
                  <User size={18} className="text-gray-400 mt-0.5" />
                  <div>
                    <p className="text-xs text-gray-400">From</p>
                    <p className="font-medium text-gray-900 dark:text-gray-100">
                      {viewingLetter.from_party || "-"}
                    </p>
                  </div>
                </div>
              ) : (
                <div className="flex items-start gap-3">
                  <Building2 size={18} className="text-gray-400 mt-0.5" />
                  <div>
                    <p className="text-xs text-gray-400">To</p>
                    <p className="font-medium text-gray-900 dark:text-gray-100">
                      {viewingLetter.to_party || "-"}
                    </p>
                  </div>
                </div>
              )}
            </div>

            {/* Dates */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div className="flex items-start gap-3">
                <Calendar size={18} className="text-gray-400 mt-0.5" />
                <div>
                  <p className="text-xs text-gray-400">Letter Date</p>
                  <p className="text-gray-900 dark:text-gray-100">
                    {formatDate(viewingLetter.letter_date)}
                  </p>
                </div>
              </div>
              {viewingLetter.received_date && (
                <div className="flex items-start gap-3">
                  <Calendar size={18} className="text-gray-400 mt-0.5" />
                  <div>
                    <p className="text-xs text-gray-400">Received Date</p>
                    <p className="text-gray-900 dark:text-gray-100">
                      {formatDate(viewingLetter.received_date)}
                    </p>
                  </div>
                </div>
              )}
              {viewingLetter.sent_date && (
                <div className="flex items-start gap-3">
                  <Calendar size={18} className="text-gray-400 mt-0.5" />
                  <div>
                    <p className="text-xs text-gray-400">Sent Date</p>
                    <p className="text-gray-900 dark:text-gray-100">
                      {formatDate(viewingLetter.sent_date)}
                    </p>
                  </div>
                </div>
              )}
            </div>

            {/* Remarks */}
            {viewingLetter.remarks && (
              <div className="bg-gray-50 dark:bg-gray-800/50 rounded-lg p-4">
                <p className="text-xs text-gray-400 mb-1">Remarks</p>
                <p className="text-gray-700 dark:text-gray-300">
                  {viewingLetter.remarks}
                </p>
              </div>
            )}

            {/* Document Section */}
            <div className="border-t border-gray-200 dark:border-white/10 pt-4">
              <div className="flex items-center justify-between mb-3">
                <h4 className="font-medium text-gray-900 dark:text-gray-100 flex items-center gap-2">
                  <FileText size={16} />
                  Attached Document
                </h4>
                {canEdit && (
                  <div>
                    <input
                      type="file"
                      ref={fileInputRef}
                      accept=".pdf,.doc,.docx,.jpg,.jpeg,.png"
                      className="hidden"
                      onChange={(e) => {
                        const file = e.target.files?.[0];
                        if (file && viewingLetter) {
                          handleDocumentUpload(viewingLetter.id, file);
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
                      {viewingLetter.document_file ? "Replace" : "Upload"}
                    </Button>
                  </div>
                )}
              </div>
              {viewingLetter.document_file ? (
                <div className="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800/50 rounded-lg">
                  <div className="flex items-center gap-3">
                    {getDocumentIcon(viewingLetter.document_file)}
                    <span className="text-sm text-gray-700 dark:text-gray-300">
                      {viewingLetter.document_file.split("/").pop()}
                    </span>
                  </div>
                  <a
                    href={viewingLetter.document_file}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-emerald-600 hover:text-emerald-700 text-sm flex items-center gap-1"
                  >
                    <Download size={14} /> Download
                  </a>
                </div>
              ) : (
                <p className="text-sm text-gray-400 text-center py-4">
                  No document attached
                </p>
              )}
            </div>

            {/* Meta Info */}
            <div className="text-xs text-gray-400 border-t border-gray-200 dark:border-white/10 pt-4">
              <p>Created by: {viewingLetter.creator?.name || "Unknown"}</p>
              <p>Created at: {formatDate(viewingLetter.created_at)}</p>
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
              Are you sure you want to delete this letter?
            </p>
            <p className="text-xs text-gray-500 mt-1">
              This action cannot be undone. The attached document will also be
              deleted.
            </p>
          </div>
        </div>
      </Modal>
    </div>
  );
}
