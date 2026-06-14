import { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { useNavigate } from "react-router-dom";
import { KeyRound, LogOut, ShieldCheck, Smartphone, TriangleAlert } from "lucide-react";
import { api } from "../../services/api";
import { loadProfile, sessionExpired } from "../../redux/slices/authSlice";
import { useToast } from "../common/Toast";
import Button from "../ui/Button";
import Input from "../ui/Input";
import { Card, CardContent, CardHeader } from "../ui/Card";

export default function AccountSecurity() {
  const account = useSelector((state) => state.auth.user);
  const dispatch = useDispatch();
  const navigate = useNavigate();
  const { addToast } = useToast();
  const [passwords, setPasswords] = useState({ old_password: "", new_password: "", confirm: "" });
  const [mfaPassword, setMfaPassword] = useState("");
  const [mfaCode, setMfaCode] = useState("");
  const [disableCode, setDisableCode] = useState("");
  const [setup, setSetup] = useState(null);
  const [backupCodes, setBackupCodes] = useState([]);
  const [sessions, setSessions] = useState([]);
  const [busy, setBusy] = useState("");

  useEffect(() => {
    if (!account?.must_change_password && !account?.mfa_setup_required) {
      api.getSessions().then((response) => setSessions(response.data || [])).catch(() => {});
    }
  }, [account?.must_change_password, account?.mfa_setup_required, account?.mfa_enabled]);

  const changePassword = async (event) => {
    event.preventDefault();
    if (passwords.new_password !== passwords.confirm) {
      addToast("The new passwords do not match.", "warning");
      return;
    }
    setBusy("password");
    try {
      await api.changePassword({ old_password: passwords.old_password, new_password: passwords.new_password });
      dispatch(sessionExpired());
      addToast("Password changed. Sign in again with your new password.", "success", 7000);
      navigate("/login", { replace: true });
    } catch (error) {
      addToast(error.message, "error");
    } finally {
      setBusy("");
    }
  };

  const beginMFA = async () => {
    setBusy("mfa-setup");
    try {
      const response = await api.beginMFA(mfaPassword);
      setSetup(response.data);
      setBackupCodes([]);
      addToast("Authenticator enrollment started.", "success");
    } catch (error) {
      addToast(error.message, "error");
    } finally {
      setBusy("");
    }
  };

  const enableMFA = async () => {
    setBusy("mfa-enable");
    try {
      const response = await api.enableMFA(mfaCode);
      setBackupCodes(response.data?.backup_codes || []);
      setSetup(null);
      setMfaCode("");
      setMfaPassword("");
      await dispatch(loadProfile());
      addToast("Multi-factor authentication enabled.", "success", 7000);
    } catch (error) {
      addToast(error.message, "error");
    } finally {
      setBusy("");
    }
  };

  const disableMFA = async () => {
    setBusy("mfa-disable");
    try {
      await api.disableMFA(mfaPassword, disableCode);
      setMfaPassword("");
      setDisableCode("");
      await dispatch(loadProfile());
      addToast("Multi-factor authentication disabled.", "success");
    } catch (error) {
      addToast(error.message, "error");
    } finally {
      setBusy("");
    }
  };

  const revokeAll = async () => {
    setBusy("sessions");
    try {
      await api.revokeAllSessions();
      dispatch(sessionExpired());
      navigate("/login", { replace: true });
    } catch (error) {
      addToast(error.message, "error");
    } finally {
      setBusy("");
    }
  };

  return (
    <div className="space-y-6">
      {(account?.must_change_password || account?.mfa_setup_required) && (
        <div className="flex gap-3 rounded-2xl border border-amber-300 bg-amber-50 p-4 text-amber-900 dark:border-amber-900 dark:bg-amber-950/30 dark:text-amber-200">
          <TriangleAlert className="mt-0.5 shrink-0" size={20} />
          <div>
            <p className="font-bold">Security setup is required</p>
            <p className="mt-1 text-sm leading-6">
              {account?.must_change_password
                ? "Change the temporary password before using the rest of the system."
                : "Set up an authenticator app before using privileged administrator or staff features."}
            </p>
          </div>
        </div>
      )}

      <Card>
        <CardHeader className="flex items-center gap-2"><KeyRound size={20} className="text-emerald-600" /><h2 className="font-bold">Change password</h2></CardHeader>
        <CardContent>
          <form onSubmit={changePassword} className="grid gap-4 md:grid-cols-3">
            <Input label="Current password" type="password" autoComplete="current-password" value={passwords.old_password} onChange={(e) => setPasswords((v) => ({ ...v, old_password: e.target.value }))} required />
            <Input label="New strong password" type="password" autoComplete="new-password" value={passwords.new_password} onChange={(e) => setPasswords((v) => ({ ...v, new_password: e.target.value }))} required />
            <Input label="Confirm new password" type="password" autoComplete="new-password" value={passwords.confirm} onChange={(e) => setPasswords((v) => ({ ...v, confirm: e.target.value }))} required />
            <div className="md:col-span-3"><Button type="submit" isLoading={busy === "password"}>Change password</Button></div>
          </form>
          <p className="mt-3 text-xs text-slate-500">Use at least 12 characters with upper/lowercase letters, a number and a symbol. All active sessions are revoked after the change.</p>
        </CardContent>
      </Card>

      {!account?.must_change_password && (
        <Card>
          <CardHeader className="flex items-center gap-2"><Smartphone size={20} className="text-emerald-600" /><h2 className="font-bold">Multi-factor authentication</h2></CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center gap-2 text-sm"><ShieldCheck size={18} className={account?.mfa_enabled ? "text-emerald-600" : "text-slate-400"} /><strong>{account?.mfa_enabled ? "Enabled" : "Not enabled"}</strong></div>
            {!account?.mfa_enabled ? (
              <>
                {!setup && (
                  <div className="flex max-w-xl flex-col gap-3 sm:flex-row sm:items-end">
                    <Input label="Current password" type="password" autoComplete="current-password" value={mfaPassword} onChange={(e) => setMfaPassword(e.target.value)} />
                    <Button type="button" onClick={beginMFA} isLoading={busy === "mfa-setup"} disabled={!mfaPassword}>Start setup</Button>
                  </div>
                )}
                {setup && (
                  <div className="space-y-4 rounded-xl border border-emerald-200 bg-emerald-50 p-4 dark:border-emerald-900 dark:bg-emerald-950/20">
                    <p className="text-sm leading-6">Add the account to Google Authenticator, Microsoft Authenticator, 1Password or another TOTP app. Use the URI below or enter the secret manually.</p>
                    <div className="break-all rounded-lg bg-white p-3 font-mono text-xs dark:bg-slate-900">{setup.otpauth_uri}</div>
                    <div><span className="text-xs font-bold uppercase text-slate-500">Secret</span><p className="mt-1 break-all font-mono font-bold">{setup.secret}</p></div>
                    <div className="flex max-w-lg flex-col gap-3 sm:flex-row sm:items-end">
                      <Input label="Six-digit authenticator code" inputMode="numeric" autoComplete="one-time-code" value={mfaCode} onChange={(e) => setMfaCode(e.target.value.replace(/\s/g, ""))} />
                      <Button type="button" onClick={enableMFA} isLoading={busy === "mfa-enable"} disabled={!mfaCode}>Verify and enable</Button>
                    </div>
                  </div>
                )}
              </>
            ) : (
              <div className="grid max-w-2xl gap-3 sm:grid-cols-2">
                <Input label="Current password" type="password" autoComplete="current-password" value={mfaPassword} onChange={(e) => setMfaPassword(e.target.value)} />
                <Input label="Authenticator or backup code" value={disableCode} onChange={(e) => setDisableCode(e.target.value.replace(/\s/g, ""))} />
                <div className="sm:col-span-2"><Button type="button" variant="danger" onClick={disableMFA} isLoading={busy === "mfa-disable"} disabled={!mfaPassword || !disableCode}>Disable MFA</Button></div>
              </div>
            )}
            {backupCodes.length > 0 && (
              <div className="rounded-xl border border-red-200 bg-red-50 p-4 dark:border-red-900 dark:bg-red-950/20">
                <p className="font-bold text-red-800 dark:text-red-200">Save these one-time backup codes now</p>
                <p className="mt-1 text-sm text-red-700 dark:text-red-300">They will not be shown again. Store them offline and never share them.</p>
                <div className="mt-3 grid gap-2 sm:grid-cols-2">{backupCodes.map((code) => <code key={code} className="rounded bg-white px-3 py-2 text-center dark:bg-slate-900">{code}</code>)}</div>
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {!account?.must_change_password && !account?.mfa_setup_required && (
        <Card>
          <CardHeader className="flex items-center gap-2"><LogOut size={20} className="text-emerald-600" /><h2 className="font-bold">Active sessions</h2></CardHeader>
          <CardContent>
            <div className="space-y-2">{sessions.map((session) => (
              <div key={session.id} className="rounded-xl border border-slate-200 p-3 text-sm dark:border-white/10">
                <p className="font-semibold">{session.user_agent || "Unknown browser"}</p>
                <p className="mt-1 text-xs text-slate-500">IP: {session.ip_address || "unknown"} · Last seen: {new Date(session.last_seen_at).toLocaleString()}</p>
              </div>
            ))}</div>
            <Button className="mt-4" variant="danger" onClick={revokeAll} isLoading={busy === "sessions"}>Sign out all devices</Button>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
