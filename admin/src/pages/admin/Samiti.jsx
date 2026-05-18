import { useState, useEffect, useCallback, useRef } from "react";
import { useSelector } from "react-redux";
import {
  Save,
  Plus,
  Edit2,
  Trash2,
  User,
  Building2,
  AlertTriangle,
  RefreshCw,
  Upload,
  Image as ImageIcon,
  Download,
  Phone,
  Mail,
  MapPin,
  Calendar,
  Globe,
  Eye,
  X,
} from "lucide-react";
import { api } from "../../services/api";
import { Card, CardContent, CardHeader } from "../../components/ui/Card";
import Button from "../../components/ui/Button";
import Input from "../../components/ui/Input";
import Textarea from "../../components/ui/Textarea";
import Select from "../../components/ui/Select";
import Modal from "../../components/ui/Modal";
import LoadingSpinner from "../../components/common/LoadingSpinner";
import { useToast } from "../../components/common/Toast";
import { formatDate } from "../../utils/helpers";

const emptyHead = {
  name: "",
  post: "",
  phone: "",
  email: "",
  address: "",
  photo: "",
  tenure_start: "",
  tenure_end: "",
  is_active: true,
  remarks: "",
};

export default function Samiti() {
  const { user } = useSelector((state) => state.auth);
  const { addToast } = useToast();
  const isAdmin = user?.role === "admin";

  const [activeTab, setActiveTab] = useState("info");
  const [settings, setSettings] = useState(null);
  const [heads, setHeads] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [saving, setSaving] = useState(false);
  const [uploading, setUploading] = useState(false);

  const [settingsForm, setSettingsForm] = useState({});
  const [showHeadModal, setShowHeadModal] = useState(false);
  const [showViewModal, setShowViewModal] = useState(false);
  const [editingHead, setEditingHead] = useState(null);
  const [viewingHead, setViewingHead] = useState(null);
  const [headForm, setHeadForm] = useState(emptyHead);
  const [deleteTarget, setDeleteTarget] = useState(null);
  
  const logoInputRef = useRef(null);
  const headPhotoInputRef = useRef(null);

  const fetchData = useCallback(async () => {
    setIsLoading(true);
    try {
      const [settingsRes, headsRes] = await Promise.all([
        api.getSamitiSettings(),
        api.getSamitiHeads(),
      ]);
      if (settingsRes.success && settingsRes.data) {
        setSettings(settingsRes.data);
        setSettingsForm(settingsRes.data);
      }
      if (headsRes.success) setHeads(headsRes.data || []);
    } catch (err) {
      addToast("Failed to load samiti data", "error");
    } finally {
      setIsLoading(false);
      setRefreshing(false);
    }
  }, [addToast]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const handleSaveSettings = async () => {
    setSaving(true);
    try {
      const res = await api.updateSamitiSettings(settingsForm);
      if (res.success) {
        addToast("Settings saved successfully", "success");
        setSettings(res.data);
      } else {
        addToast(res.message || "Failed to save", "error");
      }
    } catch (err) {
      addToast(err.message || "Failed to save", "error");
    } finally {
      setSaving(false);
    }
  };

  const handleSaveHead = async () => {
    setSaving(true);
    try {
      const payload = {
        ...headForm,
        is_active: headForm.is_active,
        tenure_start: headForm.tenure_start || null,
        tenure_end: headForm.tenure_end || null,
      };
      
      const res = editingHead
        ? await api.updateSamitiHead(editingHead.id, payload)
        : await api.createSamitiHead(payload);
      
      if (res.success) {
        addToast(`Committee member ${editingHead ? "updated" : "added"} successfully`, "success");
        setShowHeadModal(false);
        setEditingHead(null);
        setHeadForm(emptyHead);
        fetchData();
      } else {
        addToast(res.message || "Failed to save", "error");
      }
    } catch (err) {
      addToast(err.message || "Failed to save", "error");
    } finally {
      setSaving(false);
    }
  };

  const handleDeleteHead = async () => {
    if (!deleteTarget) return;
    setSaving(true);
    try {
      const res = await api.deleteSamitiHead(deleteTarget);
      if (res.success) {
        addToast("Committee member removed", "success");
        setDeleteTarget(null);
        fetchData();
      } else {
        addToast(res.message || "Failed to delete", "error");
      }
    } catch (err) {
      addToast(err.message || "Failed to delete", "error");
    } finally {
      setSaving(false);
    }
  };

  const handleLogoUpload = async (file) => {
    const formData = new FormData();
    formData.append("logo", file);

    setUploading(true);
    try {
      const res = await api.uploadSamitiLogo(formData);
      if (res.success) {
        addToast("Logo uploaded successfully", "success");
        setSettingsForm({ ...settingsForm, logo: res.data.logo_url });
        setSettings({ ...settings, logo: res.data.logo_url });
      } else {
        addToast(res.message || "Upload failed", "error");
      }
    } catch (err) {
      addToast(err.message || "Upload failed", "error");
    } finally {
      setUploading(false);
      if (logoInputRef.current) logoInputRef.current.value = "";
    }
  };

  const handleHeadPhotoUpload = async (headId, file) => {
    const formData = new FormData();
    formData.append("photo", file);

    setUploading(true);
    try {
      const res = await api.uploadHeadPhoto(headId, formData);
      if (res.success) {
        addToast("Photo uploaded successfully", "success");
        fetchData();
        if (viewingHead && viewingHead.id === headId) {
          const updated = await api.getSamitiHead(headId);
          if (updated.success) setViewingHead(updated.data);
        }
      } else {
        addToast(res.message || "Upload failed", "error");
      }
    } catch (err) {
      addToast(err.message || "Upload failed", "error");
    } finally {
      setUploading(false);
      if (headPhotoInputRef.current) headPhotoInputRef.current.value = "";
    }
  };

  const openViewHead = (head) => {
    setViewingHead(head);
    setShowViewModal(true);
  };

  const openEditHead = (head) => {
    setEditingHead(head);
    setHeadForm({
      name: head.name,
      post: head.post,
      phone: head.phone || "",
      email: head.email || "",
      address: head.address || "",
      photo: head.photo || "",
      tenure_start: head.tenure_start?.split("T")[0] || "",
      tenure_end: head.tenure_end?.split("T")[0] || "",
      is_active: head.is_active,
      remarks: head.remarks || "",
    });
    setShowHeadModal(true);
  };

  const getPostBadge = (post) => {
    const styles = {
      chairperson: "bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400",
      secretary: "bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400",
      treasurer: "bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400",
      member: "bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-400",
    };
    return styles[post] || styles.member;
  };

  const postOptions = [
    { value: "chairperson", label: "Chairperson" },
    { value: "secretary", label: "Secretary" },
    { value: "treasurer", label: "Treasurer" },
    { value: "member", label: "Member" },
  ];

  if (isLoading) return <LoadingSpinner text="Loading samiti settings..." />;

  if (!isAdmin) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="text-center">
          <Building2 size={48} className="mx-auto mb-3 text-gray-400 opacity-30" />
          <p className="text-lg font-medium text-gray-500">
            Only administrators can access this page
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">
            Samiti Settings
          </h1>
          <p className="text-sm text-gray-500 dark:text-gray-400">
            Manage organization information and committee heads
          </p>
        </div>
        <div className="flex gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => {
              setRefreshing(true);
              fetchData();
            }}
            isLoading={refreshing}
          >
            <RefreshCw size={14} className="mr-1" /> Refresh
          </Button>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 bg-gray-100 dark:bg-gray-800/50 rounded-lg p-1 w-fit">
        <button
          onClick={() => setActiveTab("info")}
          className={`flex items-center gap-2 px-4 py-2 rounded-md text-sm font-medium transition-colors ${
            activeTab === "info"
              ? "bg-white dark:bg-gray-900 text-emerald-600 dark:text-emerald-400 shadow-sm"
              : "text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200"
          }`}
        >
          <Building2 size={16} />
          Organization Info
        </button>
        <button
          onClick={() => setActiveTab("heads")}
          className={`flex items-center gap-2 px-4 py-2 rounded-md text-sm font-medium transition-colors ${
            activeTab === "heads"
              ? "bg-white dark:bg-gray-900 text-emerald-600 dark:text-emerald-400 shadow-sm"
              : "text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200"
          }`}
        >
          <User size={16} />
          Committee Heads
        </button>
      </div>

      {/* Organization Info Tab */}
      {activeTab === "info" && (
        <Card>
          <CardHeader>
            <h3 className="font-semibold text-gray-900 dark:text-gray-100">
              Organization Information
            </h3>
          </CardHeader>
          <CardContent>
            {/* Logo Upload Section */}
            <div className="mb-6 flex items-center gap-4">
              <div className="relative">
                {settingsForm.logo ? (
                  <img
                    src={settingsForm.logo}
                    alt="Samiti Logo"
                    className="w-24 h-24 object-cover rounded-lg border border-gray-200 dark:border-white/10"
                    onError={(e) => {
                      e.target.onerror = null;
                      e.target.src = "";
                    }}
                  />
                ) : (
                  <div className="w-24 h-24 bg-gray-100 dark:bg-gray-800 rounded-lg flex items-center justify-center border border-gray-200 dark:border-white/10">
                    <Building2 size={32} className="text-gray-400" />
                  </div>
                )}
                <input
                  type="file"
                  ref={logoInputRef}
                  accept="image/*"
                  className="hidden"
                  onChange={(e) => {
                    const file = e.target.files?.[0];
                    if (file) handleLogoUpload(file);
                  }}
                />
              </div>
              <div>
                <Button
                  variant="outline"
                  onClick={() => logoInputRef.current?.click()}
                  isLoading={uploading}
                >
                  <Upload size={14} className="mr-1" />
                  {settingsForm.logo ? "Change Logo" : "Upload Logo"}
                </Button>
                <p className="text-xs text-gray-500 mt-1">
                  Recommended: Square image, max 2MB
                </p>
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <Input
                label="Organization Name"
                value={settingsForm.name || ""}
                onChange={(e) => setSettingsForm({ ...settingsForm, name: e.target.value })}
                placeholder="e.g., श्री पाञ्चकन्या सामुदायिक वन उपभोक्ता समूह"
              />
              <Input
                label="Registration Number"
                value={settingsForm.registration_no || ""}
                onChange={(e) => setSettingsForm({ ...settingsForm, registration_no: e.target.value })}
                placeholder="Registration number"
              />
              <Input
                label="Address / Location"
                value={settingsForm.address || ""}
                onChange={(e) => setSettingsForm({ ...settingsForm, address: e.target.value })}
                placeholder="Street address"
              />
              <div className="grid grid-cols-2 gap-2">
                <Input
                  label="Ward No"
                  type="number"
                  value={settingsForm.ward_no || ""}
                  onChange={(e) => setSettingsForm({ ...settingsForm, ward_no: Number(e.target.value) })}
                />
                <Input
                  label="Municipality"
                  value={settingsForm.municipality || ""}
                  onChange={(e) => setSettingsForm({ ...settingsForm, municipality: e.target.value })}
                />
              </div>
              <div className="grid grid-cols-2 gap-2">
                <Input
                  label="District"
                  value={settingsForm.district || ""}
                  onChange={(e) => setSettingsForm({ ...settingsForm, district: e.target.value })}
                />
                <Input
                  label="Province"
                  value={settingsForm.province || ""}
                  onChange={(e) => setSettingsForm({ ...settingsForm, province: e.target.value })}
                />
              </div>
              <Input
                label="Contact Phone"
                value={settingsForm.contact_phone || ""}
                onChange={(e) => setSettingsForm({ ...settingsForm, contact_phone: e.target.value })}
                placeholder="Phone number"
              />
              <Input
                label="Contact Email"
                type="email"
                value={settingsForm.contact_email || ""}
                onChange={(e) => setSettingsForm({ ...settingsForm, contact_email: e.target.value })}
                placeholder="Email address"
              />
              <Input
                label="Established Date"
                type="date"
                value={settingsForm.established_date?.split("T")[0] || ""}
                onChange={(e) => setSettingsForm({ ...settingsForm, established_date: e.target.value })}
              />
              <div className="grid grid-cols-2 gap-2">
                <Input
                  label="Latitude"
                  type="number"
                  step="any"
                  value={settingsForm.latitude || ""}
                  onChange={(e) => setSettingsForm({ ...settingsForm, latitude: parseFloat(e.target.value) })}
                  placeholder="Latitude"
                />
                <Input
                  label="Longitude"
                  type="number"
                  step="any"
                  value={settingsForm.longitude || ""}
                  onChange={(e) => setSettingsForm({ ...settingsForm, longitude: parseFloat(e.target.value) })}
                  placeholder="Longitude"
                />
              </div>
            </div>
            <div className="mt-4">
              <Textarea
                label="Description"
                value={settingsForm.description || ""}
                onChange={(e) => setSettingsForm({ ...settingsForm, description: e.target.value })}
                placeholder="Organization description..."
                rows={4}
              />
            </div>
            <div className="mt-6 flex justify-end">
              <Button onClick={handleSaveSettings} isLoading={saving}>
                <Save size={16} className="mr-1" /> Save Settings
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Committee Heads Tab */}
      {activeTab === "heads" && (
        <>
          <div className="flex justify-end">
            <Button
              onClick={() => {
                setEditingHead(null);
                setHeadForm(emptyHead);
                setShowHeadModal(true);
              }}
            >
              <Plus size={16} /> Add Committee Member
            </Button>
          </div>

          {heads.length === 0 ? (
            <Card>
              <CardContent className="py-12 text-center">
                <User size={48} className="mx-auto mb-3 text-gray-400 opacity-30" />
                <p className="text-lg font-medium text-gray-500">No committee members added</p>
                <Button variant="outline" className="mt-4" onClick={() => setShowHeadModal(true)}>
                  <Plus size={14} /> Add your first member
                </Button>
              </CardContent>
            </Card>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {heads.map((head) => (
                <Card key={head.id} hover>
                  <CardContent className="p-5">
                    <div className="flex items-start justify-between">
                      <div className="flex items-center gap-3">
                        <div className="relative">
                          {head.photo ? (
                            <img
                              src={head.photo}
                              alt={head.name}
                              className="w-14 h-14 rounded-full object-cover border-2 border-emerald-500"
                              onError={(e) => {
                                e.target.onerror = null;
                                e.target.src = "";
                              }}
                            />
                          ) : (
                            <div className="w-14 h-14 rounded-full bg-emerald-100 dark:bg-emerald-900/30 flex items-center justify-center">
                              <span className="text-xl font-semibold text-emerald-600 dark:text-emerald-400">
                                {head.name?.charAt(0).toUpperCase() || "?"}
                              </span>
                            </div>
                          )}
                        </div>
                        <div>
                          <h4 className="font-semibold text-gray-900 dark:text-gray-100">
                            {head.name}
                          </h4>
                          <span className={`inline-block px-2 py-0.5 rounded-full text-xs font-medium ${getPostBadge(head.post)}`}>
                            {head.post}
                          </span>
                        </div>
                      </div>
                      <div className="flex items-center gap-1">
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => openViewHead(head)}
                          title="View Details"
                        >
                          <Eye size={14} />
                        </Button>
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => openEditHead(head)}
                          title="Edit"
                        >
                          <Edit2 size={14} />
                        </Button>
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => setDeleteTarget(head.id)}
                          title="Delete"
                        >
                          <Trash2 size={14} className="text-red-500" />
                        </Button>
                      </div>
                    </div>
                    <div className="mt-4 space-y-1 text-sm text-gray-600 dark:text-gray-400">
                      {head.phone && (
                        <div className="flex items-center gap-2">
                          <Phone size={12} />
                          <span>{head.phone}</span>
                        </div>
                      )}
                      {head.email && (
                        <div className="flex items-center gap-2">
                          <Mail size={12} />
                          <span className="truncate">{head.email}</span>
                        </div>
                      )}
                      {head.address && (
                        <div className="flex items-center gap-2">
                          <MapPin size={12} />
                          <span className="truncate">{head.address}</span>
                        </div>
                      )}
                      <div className="flex items-center gap-2 pt-1">
                        <span className={`w-2 h-2 rounded-full ${head.is_active ? "bg-emerald-500" : "bg-gray-400"}`} />
                        <span className="text-xs">
                          {head.is_active ? "Active" : "Inactive"}
                        </span>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          )}
        </>
      )}

      {/* Add/Edit Head Modal */}
      <Modal
        isOpen={showHeadModal}
        onClose={() => {
          setShowHeadModal(false);
          setEditingHead(null);
          setHeadForm(emptyHead);
        }}
        title={editingHead ? "Edit Committee Member" : "Add Committee Member"}
        size="lg"
        footer={
          <>
            <Button variant="outline" onClick={() => setShowHeadModal(false)}>Cancel</Button>
            <Button onClick={handleSaveHead} isLoading={saving}>
              {editingHead ? "Update" : "Add"}
            </Button>
          </>
        }
      >
        <div className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Input
              label="Full Name"
              value={headForm.name}
              onChange={(e) => setHeadForm({ ...headForm, name: e.target.value })}
              required
            />
            <Select
              label="Position"
              value={headForm.post}
              onChange={(e) => setHeadForm({ ...headForm, post: e.target.value })}
              options={postOptions}
              required
            />
          </div>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Input
              label="Phone"
              type="tel"
              value={headForm.phone}
              onChange={(e) => setHeadForm({ ...headForm, phone: e.target.value })}
            />
            <Input
              label="Email"
              type="email"
              value={headForm.email}
              onChange={(e) => setHeadForm({ ...headForm, email: e.target.value })}
            />
          </div>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Input
              label="Tenure Start"
              type="date"
              value={headForm.tenure_start}
              onChange={(e) => setHeadForm({ ...headForm, tenure_start: e.target.value })}
            />
            <Input
              label="Tenure End"
              type="date"
              value={headForm.tenure_end}
              onChange={(e) => setHeadForm({ ...headForm, tenure_end: e.target.value })}
            />
          </div>
          <Input
            label="Address"
            value={headForm.address}
            onChange={(e) => setHeadForm({ ...headForm, address: e.target.value })}
          />
          <Input
            label="Photo URL"
            value={headForm.photo}
            onChange={(e) => setHeadForm({ ...headForm, photo: e.target.value })}
            placeholder="Image URL (or upload after creation)"
          />
          <Textarea
            label="Remarks"
            value={headForm.remarks}
            onChange={(e) => setHeadForm({ ...headForm, remarks: e.target.value })}
            rows={2}
          />
          <div className="flex items-center gap-2">
            <input
              type="checkbox"
              id="head-active"
              checked={headForm.is_active}
              onChange={(e) => setHeadForm({ ...headForm, is_active: e.target.checked })}
              className="w-4 h-4 rounded border-gray-300 dark:border-gray-600"
            />
            <label htmlFor="head-active" className="text-sm font-medium text-gray-700 dark:text-gray-300">
              Active
            </label>
          </div>
        </div>
      </Modal>

      {/* View Head Modal */}
      <Modal
        isOpen={showViewModal}
        onClose={() => {
          setShowViewModal(false);
          setViewingHead(null);
        }}
        title="Committee Member Details"
        size="lg"
      >
        {viewingHead && (
          <div className="space-y-6">
            {/* Photo Section */}
            <div className="flex flex-col items-center text-center">
              <div className="relative mb-3">
                {viewingHead.photo ? (
                  <img
                    src={viewingHead.photo}
                    alt={viewingHead.name}
                    className="w-32 h-32 rounded-full object-cover border-4 border-emerald-500"
                    onError={(e) => {
                      e.target.onerror = null;
                      e.target.src = "";
                    }}
                  />
                ) : (
                  <div className="w-32 h-32 rounded-full bg-emerald-100 dark:bg-emerald-900/30 flex items-center justify-center border-4 border-emerald-500">
                    <span className="text-4xl font-bold text-emerald-600 dark:text-emerald-400">
                      {viewingHead.name?.charAt(0).toUpperCase() || "?"}
                    </span>
                  </div>
                )}
                <input
                  type="file"
                  ref={headPhotoInputRef}
                  accept="image/*"
                  className="hidden"
                  onChange={(e) => {
                    const file = e.target.files?.[0];
                    if (file && viewingHead) {
                      handleHeadPhotoUpload(viewingHead.id, file);
                    }
                  }}
                />
              </div>
              <Button
                size="sm"
                variant="outline"
                onClick={() => headPhotoInputRef.current?.click()}
                isLoading={uploading}
              >
                <Upload size={14} className="mr-1" />
                {viewingHead.photo ? "Change Photo" : "Upload Photo"}
              </Button>
              <h2 className="text-xl font-bold text-gray-900 dark:text-gray-100 mt-4">
                {viewingHead.name}
              </h2>
              <span className={`inline-block px-3 py-1 rounded-full text-sm font-medium mt-1 ${getPostBadge(viewingHead.post)}`}>
                {viewingHead.post}
              </span>
            </div>

            {/* Contact Info */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 border-t border-gray-200 dark:border-white/10 pt-4">
              {viewingHead.phone && (
                <div className="flex items-start gap-3">
                  <Phone size={18} className="text-gray-400 mt-0.5" />
                  <div>
                    <p className="text-xs text-gray-400">Phone</p>
                    <p className="text-gray-900 dark:text-gray-100">{viewingHead.phone}</p>
                  </div>
                </div>
              )}
              {viewingHead.email && (
                <div className="flex items-start gap-3">
                  <Mail size={18} className="text-gray-400 mt-0.5" />
                  <div>
                    <p className="text-xs text-gray-400">Email</p>
                    <p className="text-gray-900 dark:text-gray-100">{viewingHead.email}</p>
                  </div>
                </div>
              )}
              {viewingHead.address && (
                <div className="flex items-start gap-3">
                  <MapPin size={18} className="text-gray-400 mt-0.5" />
                  <div>
                    <p className="text-xs text-gray-400">Address</p>
                    <p className="text-gray-900 dark:text-gray-100">{viewingHead.address}</p>
                  </div>
                </div>
              )}
              <div className="flex items-start gap-3">
                <Calendar size={18} className="text-gray-400 mt-0.5" />
                <div>
                  <p className="text-xs text-gray-400">Tenure</p>
                  <p className="text-gray-900 dark:text-gray-100">
                    {viewingHead.tenure_start ? formatDate(viewingHead.tenure_start) : "Start"} - 
                    {viewingHead.tenure_end ? formatDate(viewingHead.tenure_end) : "Present"}
                  </p>
                </div>
              </div>
              <div className="flex items-start gap-3">
                <Globe size={18} className="text-gray-400 mt-0.5" />
                <div>
                  <p className="text-xs text-gray-400">Status</p>
                  <p className={`font-medium ${viewingHead.is_active ? "text-emerald-600" : "text-gray-500"}`}>
                    {viewingHead.is_active ? "Active" : "Inactive"}
                  </p>
                </div>
              </div>
            </div>

            {/* Remarks */}
            {viewingHead.remarks && (
              <div className="bg-gray-50 dark:bg-gray-800/50 rounded-lg p-4">
                <p className="text-xs text-gray-400 mb-1">Remarks</p>
                <p className="text-gray-700 dark:text-gray-300">{viewingHead.remarks}</p>
              </div>
            )}
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
            <Button variant="danger" onClick={handleDeleteHead} isLoading={saving}>Delete</Button>
          </>
        }
      >
        <div className="flex items-start gap-3">
          <AlertTriangle size={20} className="text-red-500 shrink-0 mt-0.5" />
          <div>
            <p className="text-sm text-gray-700 dark:text-gray-300">
              Are you sure you want to remove this committee member?
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