import { useCallback, useEffect, useMemo, useState } from "react";
import { useSelector } from "react-redux";
import {
  Banknote,
  Ban,
  Download,
  FileText,
  PlusCircle,
  Save,
  ShieldCheck,
  TrendingUp,
  X,
} from "lucide-react";
import { api } from "../../services/api";
import { Card, CardContent } from "../ui/Card";
import Button from "../ui/Button";
import Modal from "../ui/Modal";
import Badge from "../ui/Badge";
import Input from "../ui/Input";
import Select from "../ui/Select";
import Textarea from "../ui/Textarea";
import {
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
} from "../ui/Table";
import LoadingSpinner from "../common/LoadingSpinner";
import { useToast } from "../common/Toast";
import { formatCurrency, formatDate } from "../../utils/helpers";
import {
  buildFiscalYearOptions,
  getActiveFiscalYearId,
} from "../../utils/fiscalYears";

const EMPTY_HISTORICAL_FORM = {
  fiscal_year_id: "",
  sale_type: "timber",
  amount_remaining: "",
  record_date: "",
  physical_reference: "",
  remarks: "",
  save_as_draft: false,
};

const TYPE_LABELS = {
  membership_fee: "Membership Fee",
  legacy_gasti_fee: "Past Gasti Fee",
  resource_sale: "Resource Sale",
  legacy_timber_sale: "Past Timber Sale",
  legacy_firewood_sale: "Past Firewood Sale",
  legacy_other_sale: "Past Other Sale",
};

function transactionPaid(transaction) {
  return Number(transaction?.amount_paid ?? transaction?.paid_amount ?? 0);
}

function transactionRemaining(transaction) {
  return Number(
    transaction?.amount_remaining ?? transaction?.remaining ?? 0,
  );
}

function transactionFiscalYear(transaction) {
  return (
    transaction?.fiscal_year?.name ||
    transaction?.fiscal_year_object?.name ||
    transaction?.fiscal_year ||
    "-"
  );
}

function transactionStatus(transaction) {
  return transaction?.record_status || "verified";
}

function isLegacyTransaction(transaction) {
  return Boolean(transaction?.is_legacy) || transaction?.type?.startsWith("legacy_");
}

export default function MemberFinancialModals({
  memberId,
  memberName,
  onClose,
  initialModal,
}) {
  const { addToast } = useToast();
  const role = useSelector((state) => state.auth.user?.role);
  const canDigitize = role === "admin" || role === "staff";
  const isAdmin = role === "admin";

  const [activeModal, setActiveModal] = useState(initialModal || null);
  const [details, setDetails] = useState(null);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [showHistoricalForm, setShowHistoricalForm] = useState(false);
  const [fiscalYears, setFiscalYears] = useState([]);
  const [historicalForm, setHistoricalForm] = useState(
    EMPTY_HISTORICAL_FORM,
  );
  const [evidenceFile, setEvidenceFile] = useState(null);

  const [cashTarget, setCashTarget] = useState(null);
  const [cashAmount, setCashAmount] = useState("");
  const [cashRemarks, setCashRemarks] = useState("");
  const [reverseTarget, setReverseTarget] = useState(null);
  const [reverseReason, setReverseReason] = useState("");

  const fiscalYearOptions = useMemo(
    () => buildFiscalYearOptions(fiscalYears),
    [fiscalYears],
  );

  const fetchDetails = useCallback(async () => {
    if (!memberId || !activeModal) return;
    setLoading(true);
    try {
      const response =
        activeModal === "fee"
          ? await api.getMemberFeeDetails(memberId)
          : await api.getMemberSalesDetails(memberId);
      if (!response.success) throw new Error(response.message);
      setDetails(response.data);
    } catch (error) {
      addToast(error.message || "Failed to load financial details", "error");
    } finally {
      setLoading(false);
    }
  }, [activeModal, addToast, memberId]);

  useEffect(() => {
    setActiveModal(initialModal || null);
  }, [initialModal]);

  useEffect(() => {
    setShowHistoricalForm(false);
    setHistoricalForm(EMPTY_HISTORICAL_FORM);
    setEvidenceFile(null);
    fetchDetails();
  }, [activeModal, fetchDetails]);

  const closeModal = () => {
    setActiveModal(null);
    setShowHistoricalForm(false);
    setCashTarget(null);
    setReverseTarget(null);
    onClose();
  };

  const loadFiscalYears = async () => {
    if (fiscalYears.length > 0) return fiscalYears;
    try {
      const response = await api.getFiscalYears();
      if (!response.success) throw new Error(response.message);
      const years = response.data || [];
      setFiscalYears(years);
      return years;
    } catch (error) {
      addToast(error.message || "Failed to load fiscal years", "error");
      return [];
    }
  };

  const openHistoricalForm = async () => {
    const years = await loadFiscalYears();
    setHistoricalForm({
      ...EMPTY_HISTORICAL_FORM,
      fiscal_year_id: getActiveFiscalYearId(years),
    });
    setEvidenceFile(null);
    setShowHistoricalForm(true);
  };

  const saveHistoricalBalance = async () => {
    const amount = Number(historicalForm.amount_remaining);
    if (!historicalForm.fiscal_year_id) {
      addToast("Please select a fiscal year", "error");
      return;
    }
    if (!Number.isFinite(amount) || amount <= 0) {
      addToast("Past remaining amount must be greater than zero", "error");
      return;
    }

    setSaving(true);
    try {
      const payload = {
        category: activeModal === "fee" ? "fee" : "sales",
        fiscal_year_id: Number(historicalForm.fiscal_year_id),
        sale_type:
          activeModal === "sales" ? historicalForm.sale_type : undefined,
        amount_remaining: amount,
        record_date: historicalForm.record_date || undefined,
        physical_reference:
          historicalForm.physical_reference.trim() || undefined,
        remarks: historicalForm.remarks.trim() || undefined,
        save_as_draft: role === "staff" || historicalForm.save_as_draft,
      };
      const response = await api.createHistoricalTransaction(memberId, payload);
      if (!response.success) throw new Error(response.message);

      let uploadWarning = "";
      if (evidenceFile) {
        try {
          await api.uploadTransactionDocument(response.data.id, evidenceFile);
        } catch (uploadError) {
          uploadWarning = ` Balance saved, but evidence upload failed: ${uploadError.message}`;
        }
      }

      addToast(
        `${response.message || "Historical balance saved."}${uploadWarning}`,
        uploadWarning ? "warning" : "success",
      );
      setShowHistoricalForm(false);
      setHistoricalForm({
        ...EMPTY_HISTORICAL_FORM,
        fiscal_year_id: getActiveFiscalYearId(fiscalYears),
      });
      setEvidenceFile(null);
      await fetchDetails();
    } catch (error) {
      addToast(error.message || "Failed to save historical balance", "error");
    } finally {
      setSaving(false);
    }
  };

  const verifyTransaction = async (transaction) => {
    setSaving(true);
    try {
      const response = await api.verifyHistoricalTransaction(transaction.id);
      addToast(response.message || "Historical balance verified", "success");
      await fetchDetails();
    } catch (error) {
      addToast(error.message || "Verification failed", "error");
    } finally {
      setSaving(false);
    }
  };

  const openCashPayment = (transaction) => {
    setCashTarget(transaction);
    setCashAmount(String(transactionRemaining(transaction)));
    setCashRemarks("");
  };

  const recordCashPayment = async () => {
    const amount = Number(cashAmount);
    const remaining = transactionRemaining(cashTarget);
    if (!Number.isFinite(amount) || amount <= 0 || amount > remaining) {
      addToast(`Enter an amount between Rs. 0.01 and Rs. ${remaining.toFixed(2)}`, "error");
      return;
    }
    setSaving(true);
    try {
      const response = await api.createCashPayment({
        member_id: Number(memberId),
        ledger_transaction_id: Number(cashTarget.id),
        amount,
        remarks: cashRemarks.trim() || undefined,
      });
      addToast(response.message || "Cash payment recorded", "success");
      setCashTarget(null);
      await fetchDetails();
      if (response.data?.id) {
        try {
          await api.downloadPaymentReceipt(response.data.id);
        } catch {
          // The payment is already safely recorded; receipt remains downloadable
          // from the Payments page if the browser blocks the automatic download.
        }
      }
    } catch (error) {
      addToast(error.message || "Failed to record cash payment", "error");
    } finally {
      setSaving(false);
    }
  };

  const reverseTransaction = async () => {
    if (!reverseReason.trim()) {
      addToast("Enter a reason for the reversal", "error");
      return;
    }
    setSaving(true);
    try {
      const response = await api.reverseHistoricalTransaction(
        reverseTarget.id,
        reverseReason.trim(),
      );
      addToast(response.message || "Historical balance reversed", "success");
      setReverseTarget(null);
      setReverseReason("");
      await fetchDetails();
    } catch (error) {
      addToast(error.message || "Failed to reverse balance", "error");
    } finally {
      setSaving(false);
    }
  };

  const downloadDocument = async (document) => {
    try {
      await api.downloadUploadedFile(document);
    } catch (error) {
      addToast(error.message || "Failed to download document", "error");
    }
  };

  const historicalFormPanel = (
    <Card className="border-amber-200 bg-amber-50/70 dark:border-amber-900/40 dark:bg-amber-900/10">
      <CardContent className="space-y-4 p-4">
        <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <h4 className="font-semibold text-gray-900 dark:text-gray-100">
              {activeModal === "fee"
                ? "Add Past Gasti Fee Balance"
                : "Add Past Sales Balance"}
            </h4>
            <p className="mt-1 text-xs text-gray-500">
              Digitize the unpaid amount exactly as shown in the physical
              register. Attach the scanned page or receipt when available.
            </p>
          </div>
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={() => setShowHistoricalForm(false)}
            disabled={saving}
          >
            <X size={15} /> Close
          </Button>
        </div>

        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
          <Select
            label="Fiscal Year / आर्थिक वर्ष"
            value={historicalForm.fiscal_year_id}
            onChange={(event) =>
              setHistoricalForm((current) => ({
                ...current,
                fiscal_year_id: event.target.value,
              }))
            }
            options={fiscalYearOptions}
            placeholder="Select fiscal year"
            required
          />
          {activeModal === "sales" && (
            <Select
              label="Sale Type / बिक्री प्रकार"
              value={historicalForm.sale_type}
              onChange={(event) =>
                setHistoricalForm((current) => ({
                  ...current,
                  sale_type: event.target.value,
                }))
              }
              options={[
                { value: "timber", label: "काठ / Timber" },
                { value: "firewood", label: "दाउरा / Firewood" },
                { value: "other", label: "अन्य / Other" },
              ]}
            />
          )}
          <Input
            label="Past Remaining Amount / पुरानो बाँकी रकम"
            type="number"
            min="0.01"
            step="0.01"
            value={historicalForm.amount_remaining}
            onChange={(event) =>
              setHistoricalForm((current) => ({
                ...current,
                amount_remaining: event.target.value,
              }))
            }
            required
          />
          <Input
            label="Register Record Date"
            type="date"
            value={historicalForm.record_date}
            onChange={(event) =>
              setHistoricalForm((current) => ({
                ...current,
                record_date: event.target.value,
              }))
            }
          />
          <Input
            label="Physical Reference"
            value={historicalForm.physical_reference}
            onChange={(event) =>
              setHistoricalForm((current) => ({
                ...current,
                physical_reference: event.target.value,
              }))
            }
            placeholder="Register page / old receipt number"
          />
          <label className="block text-sm text-gray-700 dark:text-gray-300">
            <span className="mb-1.5 block font-medium">
              Scanned Evidence (image or PDF)
            </span>
            <input
              type="file"
              accept="image/jpeg,image/png,image/webp,application/pdf"
              onChange={(event) => setEvidenceFile(event.target.files?.[0] || null)}
              className="block w-full rounded-lg border border-gray-200 bg-white p-2 text-sm dark:border-white/10 dark:bg-gray-900"
            />
          </label>
        </div>

        <Textarea
          label="Remarks"
          value={historicalForm.remarks}
          onChange={(event) =>
            setHistoricalForm((current) => ({
              ...current,
              remarks: event.target.value,
            }))
          }
          rows={3}
          placeholder="Any note copied from the physical register"
        />

        {isAdmin && (
          <label className="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-300">
            <input
              type="checkbox"
              checked={historicalForm.save_as_draft}
              onChange={(event) =>
                setHistoricalForm((current) => ({
                  ...current,
                  save_as_draft: event.target.checked,
                }))
              }
            />
            Save as draft for verification later
          </label>
        )}
        {role === "staff" && (
          <p className="rounded-lg bg-blue-50 p-3 text-xs text-blue-700 dark:bg-blue-900/20 dark:text-blue-300">
            Staff entries are saved as drafts. An administrator must verify
            them before they affect totals or accept payment.
          </p>
        )}

        <div className="flex justify-end gap-2">
          <Button
            type="button"
            variant="outline"
            onClick={() => setShowHistoricalForm(false)}
            disabled={saving}
          >
            Cancel
          </Button>
          <Button type="button" onClick={saveHistoricalBalance} isLoading={saving}>
            <Save size={16} /> Save Historical Balance
          </Button>
        </div>
      </CardContent>
    </Card>
  );

  const transactions = details?.transactions || [];
  const title =
    activeModal === "fee"
      ? `Fee Details - ${memberName}`
      : `Sales Details - ${memberName}`;

  return (
    <>
      <Modal isOpen={Boolean(activeModal)} onClose={closeModal} title={title} size="xl">
        {loading ? (
          <LoadingSpinner text="Loading financial details..." />
        ) : (
          <div className="space-y-5">
            <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
              <div className="flex items-center gap-2 text-sm text-gray-500">
                {activeModal === "fee" ? <FileText size={17} /> : <TrendingUp size={17} />}
                Current and historical ledger records
              </div>
              {canDigitize && !showHistoricalForm && (
                <Button type="button" size="sm" onClick={openHistoricalForm}>
                  <PlusCircle size={16} />
                  {activeModal === "fee"
                    ? "Add Past Fee Balance"
                    : "Add Past Sales Balance"}
                </Button>
              )}
            </div>

            {showHistoricalForm && historicalFormPanel}

            {activeModal === "fee" ? (
              <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
                <SummaryCard label="Total Charge" value={details?.total_amount} />
                <SummaryCard label="Total Paid" value={details?.total_paid} tone="paid" />
                <SummaryCard label="Remaining" value={details?.total_remaining} tone="due" />
              </div>
            ) : (
              <div className="grid grid-cols-1 gap-3 sm:grid-cols-5">
                <SummaryCard label="Timber" value={details?.timber_total} />
                <SummaryCard label="Firewood" value={details?.firewood_total} />
                <SummaryCard label="Other" value={details?.other_total} />
                <SummaryCard label="Received" value={details?.total_received} tone="paid" />
                <SummaryCard label="Remaining" value={details?.total_remaining} tone="due" />
              </div>
            )}

            <div className="overflow-x-auto rounded-lg border border-gray-200 dark:border-white/10">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Date</TableHead>
                    <TableHead>Fiscal Year</TableHead>
                    <TableHead>Type</TableHead>
                    <TableHead>Total</TableHead>
                    <TableHead>Paid</TableHead>
                    <TableHead>Remaining</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Evidence</TableHead>
                    <TableHead>Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {transactions.length === 0 ? (
                    <TableRow>
                      <TableCell colSpan={9} className="py-8 text-center text-gray-500">
                        No records found
                      </TableCell>
                    </TableRow>
                  ) : (
                    transactions.map((transaction) => {
                      const legacy = isLegacyTransaction(transaction);
                      const status = transactionStatus(transaction);
                      const remaining = transactionRemaining(transaction);
                      const paid = transactionPaid(transaction);
                      const documents = transaction.documents || [];
                      return (
                        <TableRow key={transaction.id}>
                          <TableCell>{formatDate(transaction.date)}</TableCell>
                          <TableCell>{transactionFiscalYear(transaction)}</TableCell>
                          <TableCell>
                            <div className="space-y-1">
                              <span>{TYPE_LABELS[transaction.type] || transaction.description || transaction.type}</span>
                              {legacy && <Badge status="pending">Legacy Register</Badge>}
                              {transaction.physical_reference && (
                                <p className="max-w-40 text-xs text-gray-500">
                                  Ref: {transaction.physical_reference}
                                </p>
                              )}
                            </div>
                          </TableCell>
                          <TableCell>{formatCurrency(transaction.total_amount || 0)}</TableCell>
                          <TableCell className="text-emerald-600">
                            {formatCurrency(paid)}
                          </TableCell>
                          <TableCell className="text-red-600">
                            {formatCurrency(remaining)}
                          </TableCell>
                          <TableCell>
                            <Badge
                              status={
                                status === "verified"
                                  ? "success"
                                  : status === "reversed"
                                    ? "failed"
                                    : "pending"
                              }
                            >
                              {status}
                            </Badge>
                          </TableCell>
                          <TableCell>
                            {documents.length === 0 ? (
                              <span className="text-xs text-gray-400">None</span>
                            ) : (
                              <div className="flex flex-wrap gap-1">
                                {documents.map((document) => (
                                  <Button
                                    key={document.id}
                                    type="button"
                                    variant="ghost"
                                    size="sm"
                                    title={document.original_name}
                                    onClick={() => downloadDocument(document)}
                                  >
                                    <Download size={14} />
                                    {documents.length === 1 ? "Open" : document.id}
                                  </Button>
                                ))}
                              </div>
                            )}
                          </TableCell>
                          <TableCell>
                            {legacy ? (
                              <div className="flex flex-wrap gap-1">
                                {isAdmin && status === "draft" && (
                                  <Button
                                    type="button"
                                    variant="success"
                                    size="sm"
                                    onClick={() => verifyTransaction(transaction)}
                                    disabled={saving}
                                  >
                                    <ShieldCheck size={14} /> Verify
                                  </Button>
                                )}
                                {isAdmin && status === "verified" && remaining > 0 && (
                                  <Button
                                    type="button"
                                    size="sm"
                                    onClick={() => openCashPayment(transaction)}
                                  >
                                    <Banknote size={14} /> Cash
                                  </Button>
                                )}
                                {isAdmin && status !== "reversed" && paid <= 0 && (
                                  <Button
                                    type="button"
                                    variant="danger"
                                    size="sm"
                                    onClick={() => {
                                      setReverseTarget(transaction);
                                      setReverseReason("");
                                    }}
                                  >
                                    <Ban size={14} /> Reverse
                                  </Button>
                                )}
                              </div>
                            ) : (
                              <span className="text-xs text-gray-400">System record</span>
                            )}
                          </TableCell>
                        </TableRow>
                      );
                    })
                  )}
                </TableBody>
              </Table>
            </div>
          </div>
        )}
      </Modal>

      <Modal
        isOpen={Boolean(cashTarget)}
        onClose={() => setCashTarget(null)}
        title="Record Cash Payment"
        footer={
          <>
            <Button variant="outline" onClick={() => setCashTarget(null)} disabled={saving}>
              Cancel
            </Button>
            <Button onClick={recordCashPayment} isLoading={saving}>
              Record Cash & Generate Receipt
            </Button>
          </>
        }
      >
        {cashTarget && (
          <div className="space-y-4">
            <div className="rounded-lg bg-gray-50 p-3 text-sm dark:bg-gray-800/50">
              Remaining balance: <strong>{formatCurrency(transactionRemaining(cashTarget))}</strong>
            </div>
            <Input
              label="Cash Amount"
              type="number"
              min="0.01"
              max={transactionRemaining(cashTarget)}
              step="0.01"
              value={cashAmount}
              onChange={(event) => setCashAmount(event.target.value)}
              required
            />
            <Textarea
              label="Remarks"
              value={cashRemarks}
              onChange={(event) => setCashRemarks(event.target.value)}
              placeholder="Physical cash receipt or note"
              rows={3}
            />
            <p className="text-xs text-gray-500">
              Partial payment is supported. The remaining historical balance
              will update automatically.
            </p>
          </div>
        )}
      </Modal>

      <Modal
        isOpen={Boolean(reverseTarget)}
        onClose={() => setReverseTarget(null)}
        title="Reverse Historical Balance"
        footer={
          <>
            <Button variant="outline" onClick={() => setReverseTarget(null)} disabled={saving}>
              Cancel
            </Button>
            <Button variant="danger" onClick={reverseTransaction} isLoading={saving}>
              Confirm Reversal
            </Button>
          </>
        }
      >
        <div className="space-y-4">
          <p className="text-sm text-gray-600 dark:text-gray-300">
            This does not delete the record. It preserves the audit trail and
            removes the amount from financial totals.
          </p>
          <Textarea
            label="Reason for reversal"
            value={reverseReason}
            onChange={(event) => setReverseReason(event.target.value)}
            rows={4}
            required
          />
        </div>
      </Modal>
    </>
  );
}

function SummaryCard({ label, value, tone = "default" }) {
  const toneClass =
    tone === "paid"
      ? "border-emerald-200 bg-emerald-50 text-emerald-700 dark:border-emerald-900/40 dark:bg-emerald-900/10 dark:text-emerald-300"
      : tone === "due"
        ? "border-red-200 bg-red-50 text-red-700 dark:border-red-900/40 dark:bg-red-900/10 dark:text-red-300"
        : "border-gray-200 bg-gray-50 text-gray-800 dark:border-white/10 dark:bg-gray-800/40 dark:text-gray-100";
  return (
    <Card className={toneClass}>
      <CardContent className="p-4">
        <p className="text-xs opacity-75">{label}</p>
        <p className="mt-1 text-xl font-bold">{formatCurrency(Number(value || 0))}</p>
      </CardContent>
    </Card>
  );
}
