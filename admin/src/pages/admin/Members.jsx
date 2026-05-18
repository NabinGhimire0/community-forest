import { useState, useEffect, useCallback, useRef } from "react";
import { useSelector } from "react-redux";
import {
  Plus,
  Search,
  Edit2,
  Trash2,
  Eye,
  Users,
  UserPlus,
  Phone,
  Upload,
  Image as ImageIcon,
  X,
  RefreshCw,
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

const emptyMember = {
  membership_no: "",
  name: "",
  assistant_name: "",
  father_name: "",
  ward_no: "",
  tole: "",
  phone: "",
  photo: "",
  assistant_photo: "",
  joined_date: new Date().toISOString().split("T")[0],
  status: "active",
  remarks: "",
  family_members: [],
};

export default function Members() {
  const { user } = useSelector((state) => state.auth);
  const { addToast } = useToast();
  const canEdit = user?.role === "admin" || user?.role === "staff";

  const [members, setMembers] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState("");
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [showModal, setShowModal] = useState(false);
  const [showViewModal, setShowViewModal] = useState(false);
  const [showFamilyModal, setShowFamilyModal] = useState(false);
  const [editingMember, setEditingMember] = useState(null);
  const [viewingMember, setViewingMember] = useState(null);
  const [familyMembers, setFamilyMembers] = useState([]);
  const [form, setForm] = useState(emptyMember);
  const [familyForm, setFamilyForm] = useState({
    name: "",
    relation: "",
    age: "",
    gender: "male",
    citizenship_no: "",
    remarks: "",
  });
  const [saving, setSaving] = useState(false);
  const [uploadingPhoto, setUploadingPhoto] = useState(false);
  const [uploadingAssistantPhoto, setUploadingAssistantPhoto] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const photoInputRef = useRef(null);
  const assistantPhotoInputRef = useRef(null);

  const fetchMembers = useCallback(async () => {
    setIsLoading(true);
    try {
      const res = await api.getMembers({
        page,
        limit: 10,
        search,
        status: statusFilter || undefined,
      });
      if (res.success) {
        setMembers(res.data || []);
        setTotalPages(res.meta?.total_pages || 1);
      }
    } catch (err) {
      addToast("Failed to load members", "error");
    } finally {
      setIsLoading(false);
      setRefreshing(false);
    }
  }, [page, search, statusFilter, addToast]);

  useEffect(() => {
    fetchMembers();
  }, [fetchMembers]);

  const refreshMemberData = async (memberId) => {
    try {
      const res = await api.getMember(memberId);
      if (res.success && res.data) {
        // Update the member in the list
        setMembers((prevMembers) =>
          prevMembers.map((m) =>
            m.id === memberId ? { ...m, ...res.data } : m,
          ),
        );
        // If this is the currently viewed member, update that too
        if (viewingMember && viewingMember.id === memberId) {
          setViewingMember(res.data);
        }
      }
    } catch (err) {
      console.error("Failed to refresh member data:", err);
    }
  };

  const handleSave = async () => {
    setSaving(true);
    try {
      const payload = {
        ...form,
        ward_no: form.ward_no ? Number(form.ward_no) : undefined,
        phone: form.phone || null,
        photo: form.photo || null,
        assistant_photo: form.assistant_photo || null,
        family_members: (form.family_members || []).map((fm) => ({
          ...fm,
          age: fm.age ? Number(fm.age) : undefined,
        })),
      };

      const res = editingMember
        ? await api.updateMember(editingMember.id, payload)
        : await api.createMember(payload);
      if (res.success) {
        addToast(`Member ${editingMember ? "updated" : "created"}`, "success");
        setShowModal(false);
        setForm(emptyMember);
        setEditingMember(null);
        fetchMembers();
      }
    } catch (err) {
      addToast(err.message || "Failed", "error");
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async (id) => {
    if (!confirm("Delete this member?")) return;
    try {
      await api.deleteMember(id);
      addToast("Member deleted", "success");
      fetchMembers();
    } catch (err) {
      addToast(err.message || "Failed", "error");
    }
  };

  const handleView = async (member) => {
    setViewingMember(member);
    setShowViewModal(true);
    try {
      const res = await api.getFamilyMembers(member.id);
      if (res.success) setFamilyMembers(res.data || []);
    } catch {
      setFamilyMembers([]);
    }
  };

  const handleAddFamily = async () => {
    if (!viewingMember) return;
    setSaving(true);
    try {
      const res = await api.addFamilyMember(viewingMember.id, familyForm);
      if (res.success) {
        addToast("Family member added", "success");
        setShowFamilyModal(false);
        setFamilyForm({
          name: "",
          relation: "",
          age: "",
          gender: "male",
          citizenship_no: "",
          remarks: "",
        });
        const fres = await api.getFamilyMembers(viewingMember.id);
        if (fres.success) setFamilyMembers(fres.data || []);
        // Refresh member data to update family count
        refreshMemberData(viewingMember.id);
      }
    } catch (err) {
      addToast(err.message || "Failed", "error");
    } finally {
      setSaving(false);
    }
  };

  const handlePhotoUpload = async (memberId, file, type) => {
    const formData = new FormData();
    formData.append("photo", file);

    try {
      let res;
      if (type === "member") {
        setUploadingPhoto(true);
        res = await api.uploadMemberPhoto(memberId, formData);
      } else {
        setUploadingAssistantPhoto(true);
        res = await api.uploadAssistantPhoto(memberId, formData);
      }

      if (res.success) {
        addToast(
          `${type === "member" ? "Photo" : "Assistant photo"} uploaded successfully`,
          "success",
        );
        // Refresh the specific member data to show the new photo
        await refreshMemberData(memberId);
        // Also refresh the full list to update the table view
        fetchMembers();
      } else {
        addToast(res.message || "Upload failed", "error");
      }
    } catch (err) {
      addToast(err.message || "Upload failed", "error");
    } finally {
      setUploadingPhoto(false);
      setUploadingAssistantPhoto(false);
    }
  };

  const generateMembershipNo = () => {
    return `M-${new Date().getFullYear()}-${Math.floor(1000 + Math.random() * 9000)}`;
  };

  const openEdit = (m) => {
    setEditingMember(m);
    setForm({
      membership_no: m.membership_no,
      name: m.name,
      assistant_name: m.assistant_name || "",
      father_name: m.father_name || "",
      ward_no: m.ward_no || "",
      tole: m.tole || "",
      phone: m.phone || "",
      photo: m.photo || "",
      assistant_photo: m.assistant_photo || "",
      joined_date: m.joined_date ? m.joined_date.split("T")[0] : "",
      status: m.status || "active",
      remarks: m.remarks || "",
      family_members: m.family_members || [],
    });
    setShowModal(true);
  };

  // Helper to get full image URL
  const getImageUrl = (path) => {
    if (!path) return null;
    // If it's already a full URL, return it
    if (path.startsWith("http")) return path;
    // Otherwise, prepend the API base URL
    return path;
  };

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">
            Members
          </h1>
          <p className="text-sm text-gray-500 dark:text-gray-400">
            Manage community forestry members
          </p>
        </div>
        <div className="flex gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => {
              setRefreshing(true);
              fetchMembers();
            }}
            isLoading={refreshing}
          >
            <RefreshCw size={14} className="mr-1" /> Refresh
          </Button>
          {canEdit && (
            <Button
              onClick={() => {
                setEditingMember(null);
                setForm({
                  ...emptyMember,
                  membership_no: generateMembershipNo(),
                });
                setShowModal(true);
              }}
            >
              <Plus size={16} /> Add Member
            </Button>
          )}
        </div>
      </div>

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
                placeholder="Search members..."
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
                { value: "active", label: "Active" },
                { value: "inactive", label: "Inactive" },
              ]}
              className="w-40"
            />
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardContent className="p-0">
          {isLoading ? (
            <LoadingSpinner text="Loading members..." />
          ) : members.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 text-gray-400">
              <Users size={48} className="mb-3 opacity-30" />
              <p className="text-lg font-medium">No members found</p>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Photo</TableHead>
                    <TableHead>Member #</TableHead>
                    <TableHead>Name</TableHead>
                    <TableHead>Assistant</TableHead>
                    <TableHead>Ward / Tole</TableHead>
                    <TableHead>Phone</TableHead>
                    <TableHead>Join Date</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {members.map((m) => (
                    <TableRow key={m.id}>
                      <TableCell>
                        {m.photo ? (
                          <img
                            src={getImageUrl(m.photo)}
                            alt={m.name}
                            className="w-10 h-10 rounded-full object-cover"
                            onError={(e) => {
                              e.target.onerror = null;
                              e.target.src = "";
                              e.target.parentElement.innerHTML =
                                '<div class="w-10 h-10 rounded-full bg-gray-200 dark:bg-gray-700 flex items-center justify-center"><svg class="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" /></svg></div>';
                            }}
                          />
                        ) : (
                          <div className="w-10 h-10 rounded-full bg-gray-200 dark:bg-gray-700 flex items-center justify-center">
                            <ImageIcon size={16} className="text-gray-400" />
                          </div>
                        )}
                      </TableCell>
                      <TableCell className="font-mono text-xs">
                        {m.membership_no}
                      </TableCell>
                      <TableCell className="font-medium">{m.name}</TableCell>
                      <TableCell>{m.assistant_name || "-"}</TableCell>
                      <TableCell>
                        Ward {m.ward_no}, {m.tole}
                      </TableCell>
                      <TableCell>{m.phone || "-"}</TableCell>
                      <TableCell>{formatDate(m.joined_date)}</TableCell>
                      <TableCell>
                        <Badge status={m.status} />
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center gap-1">
                          <Button
                            variant="ghost"
                            size="icon"
                            onClick={() => handleView(m)}
                            title="View"
                          >
                            <Eye size={15} />
                          </Button>
                          {canEdit && (
                            <>
                              <Button
                                variant="ghost"
                                size="icon"
                                onClick={() => openEdit(m)}
                                title="Edit"
                              >
                                <Edit2 size={15} />
                              </Button>
                              <Button
                                variant="ghost"
                                size="icon"
                                onClick={() => handleDelete(m.id)}
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

      {/* Add/Edit Member Modal - Same as before */}
      <Modal
        isOpen={showModal}
        onClose={() => {
          setShowModal(false);
          setEditingMember(null);
          setForm(emptyMember);
        }}
        title={editingMember ? "Edit Member" : "Add Member"}
        size="lg"
        footer={
          <>
            <Button variant="outline" onClick={() => setShowModal(false)}>
              Cancel
            </Button>
            <Button onClick={handleSave} isLoading={saving}>
              {editingMember ? "Update" : "Create"}
            </Button>
          </>
        }
      >
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <Input
            label="Membership No."
            value={form.membership_no}
            onChange={(e) =>
              setForm({ ...form, membership_no: e.target.value })
            }
            required
          />
          <Input
            label="Full Name"
            value={form.name}
            onChange={(e) => setForm({ ...form, name: e.target.value })}
            required
          />
          <Input
            label="Assistant Name"
            value={form.assistant_name}
            onChange={(e) =>
              setForm({ ...form, assistant_name: e.target.value })
            }
            required
          />
          <Input
            label="Father Name"
            value={form.father_name}
            onChange={(e) => setForm({ ...form, father_name: e.target.value })}
            required
          />
          <Input
            label="Ward No"
            type="number"
            value={form.ward_no}
            onChange={(e) => setForm({ ...form, ward_no: e.target.value })}
            required
          />
          <Input
            label="Tole"
            value={form.tole}
            onChange={(e) => setForm({ ...form, tole: e.target.value })}
            required
          />
          <Input
            label="Phone"
            type="tel"
            value={form.phone}
            onChange={(e) => setForm({ ...form, phone: e.target.value })}
          />
          <Input
            label="Join Date"
            type="date"
            value={form.joined_date}
            onChange={(e) => setForm({ ...form, joined_date: e.target.value })}
          />
          <Select
            label="Status"
            value={form.status}
            onChange={(e) => setForm({ ...form, status: e.target.value })}
            options={[
              { value: "active", label: "Active" },
              { value: "inactive", label: "Inactive" },
            ]}
          />
          <Input
            label="Remarks"
            value={form.remarks}
            onChange={(e) => setForm({ ...form, remarks: e.target.value })}
          />
        </div>

        {/* Photo URLs */}
        <div className="mt-6 border-t border-gray-200 dark:border-white/10 pt-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Input
              label="Photo URL"
              value={form.photo || ""}
              onChange={(e) => setForm({ ...form, photo: e.target.value })}
              placeholder="Image URL (or upload after creation)"
            />
            <Input
              label="Assistant Photo URL"
              value={form.assistant_photo || ""}
              onChange={(e) =>
                setForm({ ...form, assistant_photo: e.target.value })
              }
              placeholder="Image URL (or upload after creation)"
            />
          </div>
        </div>

        {/* Family Members Section */}
        <div className="mt-6 border-t border-gray-200 dark:border-white/10 pt-4">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-sm font-semibold text-gray-900 dark:text-gray-100 uppercase tracking-wider">
              Family Members
            </h3>
            <Button
              size="sm"
              variant="outline"
              onClick={() =>
                setForm({
                  ...form,
                  family_members: [
                    ...(form.family_members || []),
                    {
                      name: "",
                      relation: "",
                      age: "",
                      gender: "male",
                      citizenship_no: "",
                      remarks: "",
                    },
                  ],
                })
              }
            >
              <UserPlus size={14} /> Add Family Member
            </Button>
          </div>

          {(form.family_members || []).length === 0 ? (
            <p className="text-sm text-gray-500 italic text-center py-4">
              No family members added yet. Click above to add.
            </p>
          ) : (
            <div className="space-y-4 max-h-75 overflow-y-auto pr-2">
              {(form.family_members || []).map((fm, idx) => (
                <div
                  key={idx}
                  className="grid grid-cols-1 md:grid-cols-12 gap-3 p-3 bg-gray-50 dark:bg-gray-800/50 rounded-lg relative border border-gray-100 dark:border-gray-800"
                >
                  <button
                    className="absolute -top-2 -right-2 bg-red-100 text-red-600 dark:bg-red-900/30 dark:text-red-400 rounded-full p-1.5 hover:bg-red-200 dark:hover:bg-red-900/50 transition-colors z-10 shadow-sm"
                    onClick={() => {
                      const updated = [...form.family_members];
                      updated.splice(idx, 1);
                      setForm({ ...form, family_members: updated });
                    }}
                    title="Remove"
                  >
                    <Trash2 size={12} />
                  </button>
                  <div className="md:col-span-4">
                    <Input
                      label="Name"
                      value={fm.name}
                      onChange={(e) => {
                        const updated = [...form.family_members];
                        updated[idx].name = e.target.value;
                        setForm({ ...form, family_members: updated });
                      }}
                      required
                    />
                  </div>
                  <div className="md:col-span-3">
                    <Input
                      label="Relation"
                      value={fm.relation}
                      placeholder="e.g. Spouse, Son"
                      onChange={(e) => {
                        const updated = [...form.family_members];
                        updated[idx].relation = e.target.value;
                        setForm({ ...form, family_members: updated });
                      }}
                      required
                    />
                  </div>
                  <div className="md:col-span-2">
                    <Input
                      label="Age"
                      type="number"
                      value={fm.age}
                      onChange={(e) => {
                        const updated = [...form.family_members];
                        updated[idx].age = e.target.value
                          ? parseInt(e.target.value)
                          : "";
                        setForm({ ...form, family_members: updated });
                      }}
                    />
                  </div>
                  <div className="md:col-span-3">
                    <Select
                      label="Gender"
                      value={fm.gender}
                      onChange={(e) => {
                        const updated = [...form.family_members];
                        updated[idx].gender = e.target.value;
                        setForm({ ...form, family_members: updated });
                      }}
                      options={[
                        { value: "male", label: "Male" },
                        { value: "female", label: "Female" },
                        { value: "other", label: "Other" },
                      ]}
                    />
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </Modal>

      {/* View Member Modal - With working photo upload and display */}
      <Modal
        isOpen={showViewModal}
        onClose={() => {
          setShowViewModal(false);
          setViewingMember(null);
        }}
        title="Member Details"
        size="lg"
      >
        {viewingMember && (
          <div className="space-y-6">
            {/* Photos Section */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div className="text-center">
                <div className="mb-2">
                  {viewingMember.photo ? (
                    <img
                      src={getImageUrl(viewingMember.photo)}
                      alt={viewingMember.name}
                      className="w-32 h-32 rounded-full object-cover mx-auto border-2 border-emerald-500"
                      onError={(e) => {
                        e.target.onerror = null;
                        e.target.src = "";
                        e.target.alt = "Failed to load";
                      }}
                    />
                  ) : (
                    <div className="w-32 h-32 rounded-full bg-gray-200 dark:bg-gray-700 flex items-center justify-center mx-auto">
                      <ImageIcon size={40} className="text-gray-400" />
                    </div>
                  )}
                </div>
                <h4 className="font-medium text-gray-900 dark:text-gray-100">
                  Member Photo
                </h4>
                {canEdit && (
                  <div className="mt-2">
                    <input
                      type="file"
                      ref={photoInputRef}
                      accept="image/*"
                      className="hidden"
                      onChange={(e) => {
                        const file = e.target.files?.[0];
                        if (file && viewingMember) {
                          handlePhotoUpload(viewingMember.id, file, "member");
                        }
                        if (photoInputRef.current)
                          photoInputRef.current.value = "";
                      }}
                    />
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={() => photoInputRef.current?.click()}
                      isLoading={uploadingPhoto}
                    >
                      <Upload size={14} className="mr-1" /> Upload Photo
                    </Button>
                  </div>
                )}
              </div>

              <div className="text-center">
                <div className="mb-2">
                  {viewingMember.assistant_photo ? (
                    <img
                      src={getImageUrl(viewingMember.assistant_photo)}
                      alt={viewingMember.assistant_name}
                      className="w-32 h-32 rounded-full object-cover mx-auto border-2 border-emerald-500"
                      onError={(e) => {
                        e.target.onerror = null;
                        e.target.src = "";
                        e.target.alt = "Failed to load";
                      }}
                    />
                  ) : (
                    <div className="w-32 h-32 rounded-full bg-gray-200 dark:bg-gray-700 flex items-center justify-center mx-auto">
                      <ImageIcon size={40} className="text-gray-400" />
                    </div>
                  )}
                </div>
                <h4 className="font-medium text-gray-900 dark:text-gray-100">
                  Assistant Photo
                </h4>
                {canEdit && (
                  <div className="mt-2">
                    <input
                      type="file"
                      ref={assistantPhotoInputRef}
                      accept="image/*"
                      className="hidden"
                      onChange={(e) => {
                        const file = e.target.files?.[0];
                        if (file && viewingMember) {
                          handlePhotoUpload(
                            viewingMember.id,
                            file,
                            "assistant",
                          );
                        }
                        if (assistantPhotoInputRef.current)
                          assistantPhotoInputRef.current.value = "";
                      }}
                    />
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={() => assistantPhotoInputRef.current?.click()}
                      isLoading={uploadingAssistantPhoto}
                    >
                      <Upload size={14} className="mr-1" /> Upload Assistant
                      Photo
                    </Button>
                  </div>
                )}
              </div>
            </div>

            {/* Member Info */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <p className="text-xs text-gray-400">Membership No.</p>
                <p className="font-medium text-gray-900 dark:text-gray-100">
                  {viewingMember.membership_no}
                </p>
              </div>
              <div>
                <p className="text-xs text-gray-400">Full Name</p>
                <p className="font-medium text-gray-900 dark:text-gray-100">
                  {viewingMember.name}
                </p>
              </div>
              <div>
                <p className="text-xs text-gray-400">Assistant Name</p>
                <p className="text-gray-900 dark:text-gray-100">
                  {viewingMember.assistant_name || "-"}
                </p>
              </div>
              <div>
                <p className="text-xs text-gray-400">Father Name</p>
                <p className="text-gray-900 dark:text-gray-100">
                  {viewingMember.father_name || "-"}
                </p>
              </div>
              <div className="flex items-center gap-2 text-gray-600 dark:text-gray-400">
                <Phone size={14} />
                {viewingMember.phone || "-"}
              </div>
              <div>
                <p className="text-xs text-gray-400">Status</p>
                <Badge status={viewingMember.status} />
              </div>
              <div>
                <p className="text-xs text-gray-400">Address (Ward, Tole)</p>
                <p className="text-gray-900 dark:text-gray-100">
                  Ward {viewingMember.ward_no}, {viewingMember.tole}
                </p>
              </div>
              <div>
                <p className="text-xs text-gray-400">Join Date</p>
                <p className="text-gray-900 dark:text-gray-100">
                  {formatDate(viewingMember.joined_date)}
                </p>
              </div>
              {viewingMember.remarks && (
                <div className="md:col-span-2">
                  <p className="text-xs text-gray-400">Remarks</p>
                  <p className="text-gray-900 dark:text-gray-100">
                    {viewingMember.remarks}
                  </p>
                </div>
              )}
            </div>

            {/* Family Members */}
            <div>
              <div className="flex items-center justify-between mb-3">
                <h4 className="font-semibold text-gray-900 dark:text-gray-100">
                  Family Members
                </h4>
                {canEdit && (
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() => setShowFamilyModal(true)}
                  >
                    <UserPlus size={14} /> Add
                  </Button>
                )}
              </div>
              {familyMembers.length === 0 ? (
                <p className="text-sm text-gray-400">No family members added</p>
              ) : (
                <div className="overflow-x-auto">
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>Name</TableHead>
                        <TableHead>Relation</TableHead>
                        <TableHead>Age</TableHead>
                        <TableHead>Gender</TableHead>
                        <TableHead>Citizenship No.</TableHead>
                        <TableHead>Remarks</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {familyMembers.map((fm) => (
                        <TableRow key={fm.id}>
                          <TableCell className="font-medium">
                            {fm.name}
                          </TableCell>
                          <TableCell className="capitalize">
                            {fm.relation}
                          </TableCell>
                          <TableCell>{fm.age || "-"}</TableCell>
                          <TableCell className="capitalize">
                            {fm.gender || "-"}
                          </TableCell>
                          <TableCell>{fm.citizenship_no || "-"}</TableCell>
                          <TableCell>{fm.remarks || "-"}</TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </div>
              )}
            </div>
          </div>
        )}
      </Modal>

      {/* Add Family Member Modal */}
      <Modal
        isOpen={showFamilyModal}
        onClose={() => setShowFamilyModal(false)}
        title="Add Family Member"
        footer={
          <>
            <Button variant="outline" onClick={() => setShowFamilyModal(false)}>
              Cancel
            </Button>
            <Button onClick={handleAddFamily} isLoading={saving}>
              Add
            </Button>
          </>
        }
      >
        <div className="space-y-4">
          <Input
            label="Name"
            value={familyForm.name}
            onChange={(e) =>
              setFamilyForm({ ...familyForm, name: e.target.value })
            }
            required
          />
          <Input
            label="Relation"
            value={familyForm.relation}
            onChange={(e) =>
              setFamilyForm({ ...familyForm, relation: e.target.value })
            }
            required
            placeholder="e.g. Spouse, Son, Daughter"
          />
          <Input
            label="Age"
            type="number"
            value={familyForm.age}
            onChange={(e) =>
              setFamilyForm({
                ...familyForm,
                age: e.target.value ? Number(e.target.value) : "",
              })
            }
          />
          <Select
            label="Gender"
            value={familyForm.gender}
            onChange={(e) =>
              setFamilyForm({ ...familyForm, gender: e.target.value })
            }
            options={[
              { value: "male", label: "Male" },
              { value: "female", label: "Female" },
              { value: "other", label: "Other" },
            ]}
          />
          <Input
            label="Citizenship No."
            value={familyForm.citizenship_no}
            onChange={(e) =>
              setFamilyForm({
                ...familyForm,
                citizenship_no: e.target.value,
              })
            }
          />
          <Textarea
            label="Remarks"
            value={familyForm.remarks}
            onChange={(e) =>
              setFamilyForm({ ...familyForm, remarks: e.target.value })
            }
          />
        </div>
      </Modal>
    </div>
  );
}
