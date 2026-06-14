import { useEffect, useMemo, useState } from "react";
import {
  Archive,
  CalendarRange,
  DatabaseBackup,
  Download,
  FileArchive,
  FileSpreadsheet,
  KeyRound,
  ShieldCheck,
} from "lucide-react";
import { api } from "../../services/api";
import { useToast } from "../../components/common/Toast";
import Button from "../../components/ui/Button";
import { Card } from "../../components/ui/Card";
import Input from "../../components/ui/Input";
import Select from "../../components/ui/Select";
import {
  buildFiscalYearOptions,
  getActiveFiscalYearId,
} from "../../utils/fiscalYears";

const datasetLabels = {
  members: "Members",
  family_members: "Family members",
  membership_fees: "Membership / Gasti fees",
  sales: "Sales ledger",
  transactions: "All transactions",
  payments: "Payments",
  requests: "Resource requests",
  stock: "Stock",
  expenses: "Expenses",
  fines: "Fines",
  letters: "Letters",
  committee: "Committee",
  audit_logs: "Audit logs",
};

export default function SystemData() {
  const { addToast } = useToast();
  const [datasets, setDatasets] = useState([]);
  const [fiscalYears, setFiscalYears] = useState([]);
  const [dataset, setDataset] = useState("members");
  const [fiscalYearId, setFiscalYearId] = useState("");
  const [fromDate, setFromDate] = useState("");
  const [toDate, setToDate] = useState("");
  const [currentPassword, setCurrentPassword] = useState("");
  const [mfaCode, setMfaCode] = useState("");
  const [passphrase, setPassphrase] = useState("");
  const [confirmPassphrase, setConfirmPassphrase] = useState("");
  const [busyAction, setBusyAction] = useState("");

  useEffect(() => {
    Promise.all([api.getExportDatasets(), api.getFiscalYears({ per_page: 100 })])
      .then(([datasetResponse, fiscalResponse]) => {
        setDatasets(datasetResponse.data?.datasets || []);
        const years = fiscalResponse.data || [];
        setFiscalYears(years);
        setFiscalYearId((current) => current || getActiveFiscalYearId(years));
      })
      .catch((error) => addToast(error.message, "error"));
  }, [addToast]);

  const datasetOptions = useMemo(
    () => datasets.map((name) => ({ value: name, label: datasetLabels[name] || name })),
    [datasets],
  );
  const fiscalOptions = buildFiscalYearOptions(fiscalYears);

  const commonPayload = () => ({
    dataset,
    fiscal_year_id: fiscalYearId ? Number(fiscalYearId) : 0,
    from_date: fromDate,
    to_date: toDate,
    current_password: currentPassword,
    mfa_code: mfaCode,
  });

  const validateAdminPassword = () => {
    if (!currentPassword) {
      addToast("Enter your current administrator password to continue.", "warning");
      return false;
    }
    return true;
  };

  const validatePassphrase = () => {
    if (!validateAdminPassword()) return false;
    if (passphrase.length < 16) {
      addToast("The encryption passphrase must contain at least 16 characters.", "warning");
      return false;
    }
    if (passphrase !== confirmPassphrase) {
      addToast("The backup passphrases do not match.", "warning");
      return false;
    }
    return true;
  };

  const runDownload = async (action, operation, successMessage) => {
    setBusyAction(action);
    try {
      const download = await operation();
      api.saveDownload(download);
      addToast(successMessage, "success", 7000);
      setCurrentPassword("");
      setMfaCode("");
      if (action !== "csv") {
        setPassphrase("");
        setConfirmPassphrase("");
      }
    } catch (error) {
      addToast(error.message, "error", 7000);
    } finally {
      setBusyAction("");
    }
  };

  const exportCSV = () => {
    if (!validateAdminPassword()) return;
    runDownload("csv", () => api.exportDataset(commonPayload()), "CSV export downloaded.");
  };

  const exportAll = () => {
    if (!validatePassphrase()) return;
    runDownload(
      "all",
      () => api.exportAllData({ ...commonPayload(), passphrase }),
      "Encrypted all-data export downloaded. Store its passphrase separately.",
    );
  };

  const createBackup = (full) => {
    if (!validatePassphrase()) return;
    const payload = { current_password: currentPassword, mfa_code: mfaCode, passphrase };
    runDownload(
      full ? "full" : "database",
      () => (full ? api.createFullBackup(payload) : api.createDatabaseBackup(payload)),
      full
        ? "Encrypted full backup downloaded. It includes the database and uploaded files."
        : "Encrypted PostgreSQL backup downloaded.",
    );
  };

  return (
    <div className="space-y-6">
      <div>
        <p className="text-sm font-black uppercase tracking-[0.16em] text-emerald-700">Administration</p>
        <h1 className="mt-2 text-3xl font-black text-slate-950 dark:text-white">Exports & encrypted backups</h1>
        <p className="mt-2 max-w-3xl text-sm leading-6 text-slate-500 dark:text-slate-400">
          Export operational records for reporting or create disaster-recovery backups. Every operation requires fresh password and MFA verification and is written to the audit log.
        </p>
      </div>

      <div className="grid gap-5 xl:grid-cols-2">
        <Card className="p-5">
          <div className="flex items-start gap-3">
            <div className="rounded-xl bg-emerald-100 p-3 text-emerald-700 dark:bg-emerald-950"><FileSpreadsheet size={22} /></div>
            <div>
              <h2 className="text-lg font-bold text-slate-900 dark:text-white">Structured data export</h2>
              <p className="mt-1 text-sm text-slate-500">Download one CSV or an encrypted ZIP containing every supported dataset.</p>
            </div>
          </div>
          <div className="mt-5 grid gap-4 sm:grid-cols-2">
            <Select label="Dataset" value={dataset} onChange={(e) => setDataset(e.target.value)} options={datasetOptions} />
            <Select label="Fiscal year (optional)" value={fiscalYearId} onChange={(e) => setFiscalYearId(e.target.value)} placeholder="All fiscal years" options={fiscalOptions} />
            <Input label="From date (optional)" type="date" value={fromDate} onChange={(e) => setFromDate(e.target.value)} icon={<CalendarRange size={16} />} />
            <Input label="To date (optional)" type="date" value={toDate} onChange={(e) => setToDate(e.target.value)} icon={<CalendarRange size={16} />} />
          </div>
          <div className="mt-5 flex flex-wrap gap-3">
            <Button onClick={exportCSV} isLoading={busyAction === "csv"}><Download size={17} /> Export selected CSV</Button>
            <Button variant="outline" onClick={exportAll} isLoading={busyAction === "all"}><FileArchive size={17} /> Export all encrypted</Button>
          </div>
        </Card>

        <Card className="p-5">
          <div className="flex items-start gap-3">
            <div className="rounded-xl bg-blue-100 p-3 text-blue-700 dark:bg-blue-950"><DatabaseBackup size={22} /></div>
            <div>
              <h2 className="text-lg font-bold text-slate-900 dark:text-white">Disaster-recovery backup</h2>
              <p className="mt-1 text-sm text-slate-500">Create a PostgreSQL custom-format dump, or include all uploaded documents and photographs.</p>
            </div>
          </div>
          <div className="mt-5 rounded-xl border border-amber-200 bg-amber-50 p-4 text-sm leading-6 text-amber-900 dark:border-amber-900/60 dark:bg-amber-950/30 dark:text-amber-200">
            <strong>Important:</strong> the server must have a compatible <code>pg_dump</code> installed. Test restoration regularly on a separate database; a downloaded file alone is not a verified backup.
          </div>
          <div className="mt-5 flex flex-wrap gap-3">
            <Button onClick={() => createBackup(false)} isLoading={busyAction === "database"}><DatabaseBackup size={17} /> Database backup</Button>
            <Button variant="outline" onClick={() => createBackup(true)} isLoading={busyAction === "full"}><Archive size={17} /> Full database + files</Button>
          </div>
        </Card>
      </div>

      <Card className="p-5">
        <div className="flex items-start gap-3">
          <div className="rounded-xl bg-slate-100 p-3 text-slate-700 dark:bg-slate-800 dark:text-slate-200"><ShieldCheck size={22} /></div>
          <div>
            <h2 className="text-lg font-bold text-slate-900 dark:text-white">Security confirmation</h2>
            <p className="mt-1 text-sm text-slate-500">Passwords are sent only over the authenticated HTTPS session and are never included in the downloaded file.</p>
          </div>
        </div>
        <div className="mt-5 grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          <Input label="Current administrator password" type="password" autoComplete="current-password" value={currentPassword} onChange={(e) => setCurrentPassword(e.target.value)} icon={<KeyRound size={16} />} />
          <Input label="Fresh MFA / backup code" inputMode="numeric" autoComplete="one-time-code" value={mfaCode} onChange={(e) => setMfaCode(e.target.value)} icon={<ShieldCheck size={16} />} />
          <Input label="Backup encryption passphrase" type="password" autoComplete="new-password" value={passphrase} onChange={(e) => setPassphrase(e.target.value)} icon={<KeyRound size={16} />} />
          <Input label="Confirm encryption passphrase" type="password" autoComplete="new-password" value={confirmPassphrase} onChange={(e) => setConfirmPassphrase(e.target.value)} icon={<KeyRound size={16} />} />
        </div>
        <p className="mt-3 text-xs leading-5 text-slate-500">
          Use a fresh authenticator code for each sensitive operation and a unique backup passphrase of at least 16 characters. The organization cannot recover an encrypted backup when this passphrase is lost.
        </p>
      </Card>
    </div>
  );
}
