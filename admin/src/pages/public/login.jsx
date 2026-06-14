import { useEffect, useState } from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import { useDispatch, useSelector } from "react-redux";
import { motion as Motion } from "framer-motion";
import {
  ArrowLeft,
  Eye,
  EyeOff,
  FileCheck2,
  Leaf,
  Lock,
  Phone,
  ShieldCheck,
  TreePine,
  Users,
  Wifi,
  WifiOff,
} from "lucide-react";
import { clearError, login } from "../../redux/slices/authSlice";
import { api } from "../../services/api";
import Button from "../../components/ui/Button";
import Input from "../../components/ui/Input";
import {
  getImageUrl,
  getOrganizationLocation,
} from "../../utils/helpers";

export default function Login() {
  const dispatch = useDispatch();
  const navigate = useNavigate();
  const location = useLocation();
  const { isLoading, error, errorCode, isAuthenticated } = useSelector(
    (state) => state.auth,
  );
  const settings = useSelector((state) => state.appSettings.settings);
  const [phone, setPhone] = useState("");
  const [password, setPassword] = useState("");
  const [otp, setOtp] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [serverStatus, setServerStatus] = useState("checking");
  const from = location.state?.from?.pathname || "/dashboard";
  const logoUrl = getImageUrl(settings?.logo);
  const organizationLocation = getOrganizationLocation(settings);

  useEffect(() => {
    if (isAuthenticated) navigate(from, { replace: true });
  }, [isAuthenticated, navigate, from]);

  useEffect(() => {
    let active = true;
    api.health()
      .then(() => active && setServerStatus("online"))
      .catch(() => active && setServerStatus("offline"));
    return () => {
      active = false;
    };
  }, []);

  useEffect(() => {
    return () => dispatch(clearError());
  }, [dispatch]);

  const handleSubmit = async (event) => {
    event.preventDefault();
    const result = await dispatch(login({
      phone: phone.trim(),
      password,
      otp: otp.trim(),
    }));
    if (result.meta.requestStatus === "fulfilled") {
      navigate(from, { replace: true });
    }
  };

  return (
    <div className="grid min-h-screen bg-slate-50 lg:grid-cols-[1.05fr_0.95fr]">
      <section className="relative hidden overflow-hidden bg-linear-to-br from-emerald-950 via-emerald-900 to-slate-950 p-12 text-white lg:flex lg:flex-col lg:justify-between">
        <div className="absolute inset-0 opacity-35 bg-[radial-gradient(circle_at_15%_20%,rgba(52,211,153,0.45),transparent_28%),radial-gradient(circle_at_85%_75%,rgba(16,185,129,0.3),transparent_32%)]" />
        <div className="relative z-10">
          <Link to="/" className="inline-flex items-center gap-2 text-sm font-semibold text-emerald-100/80 hover:text-white">
            <ArrowLeft size={17} /> Back to public page
          </Link>
        </div>

        <Motion.div
          initial={{ opacity: 0, y: 24 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6 }}
          className="relative z-10 max-w-xl"
        >
          {logoUrl ? (
            <img
              src={logoUrl}
              alt="Organization logo"
              className="h-20 w-20 rounded-2xl bg-white object-contain p-2 shadow-xl"
            />
          ) : (
            <div className="flex h-20 w-20 items-center justify-center rounded-2xl bg-white/12 text-emerald-200 backdrop-blur-sm">
              <TreePine size={40} />
            </div>
          )}
          <p className="mt-8 text-sm font-black uppercase tracking-[0.2em] text-emerald-300">
            Secure digital portal
          </p>
          <h1 className="mt-3 text-4xl font-black leading-tight">
            {settings?.name}
          </h1>
          {organizationLocation && (
            <p className="mt-3 text-sm font-semibold text-emerald-100/70">
              {organizationLocation}
            </p>
          )}
          <p className="mt-6 text-lg leading-8 text-emerald-50/75">
            One account provides the correct workspace for administrators,
            staff members and registered user-group members.
          </p>

          <div className="mt-9 grid gap-3 sm:grid-cols-3">
            {[
              { icon: ShieldCheck, text: "Role-based access" },
              { icon: Users, text: "Member records" },
              { icon: FileCheck2, text: "Service history" },
            ].map((item) => (
              <div key={item.text} className="rounded-2xl border border-white/10 bg-white/8 p-4 backdrop-blur-sm">
                <item.icon size={20} className="text-emerald-300" />
                <p className="mt-3 text-sm font-bold">{item.text}</p>
              </div>
            ))}
          </div>
        </Motion.div>

        <div className="relative z-10 flex items-center gap-2 text-xs text-emerald-100/55">
          <Leaf size={15} /> Community Forestry Management System
        </div>
      </section>

      <section className="flex items-center justify-center px-5 py-10 sm:px-10 lg:px-14">
        <Motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5 }}
          className="w-full max-w-md"
        >
          <Link to="/" className="mb-8 inline-flex items-center gap-2 text-sm font-semibold text-slate-500 hover:text-emerald-700 lg:hidden">
            <ArrowLeft size={17} /> Back to public page
          </Link>

          <div className="mb-8 flex items-center gap-3 lg:hidden">
            {logoUrl ? (
              <img src={logoUrl} alt="Organization logo" className="h-12 w-12 rounded-xl border bg-white object-contain p-1" />
            ) : (
              <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-emerald-700 text-white"><TreePine size={24} /></div>
            )}
            <div className="min-w-0">
              <p className="line-clamp-2 text-sm font-extrabold leading-5 text-slate-900">{settings?.name}</p>
              <p className="text-xs text-emerald-700">Secure portal</p>
            </div>
          </div>

          <div>
            <p className="text-sm font-black uppercase tracking-[0.16em] text-emerald-700">
              Account access
            </p>
            <h2 className="mt-2 text-3xl font-black tracking-tight text-slate-950">
              Sign in to your workspace
            </h2>
            <p className="mt-2 text-sm leading-6 text-slate-500">
              Use the phone number and password provided for your account.
            </p>
          </div>

          {error && (
            <Motion.div
              initial={{ opacity: 0, y: -8 }}
              animate={{ opacity: 1, y: 0 }}
              className="mt-6 rounded-xl border border-red-200 bg-red-50 p-3 text-sm font-medium text-red-700"
              role="alert"
            >
              {error}
            </Motion.div>
          )}

          <div
            className={`mt-6 flex items-center gap-2 rounded-xl border px-3 py-2 text-xs font-semibold ${
              serverStatus === "offline"
                ? "border-red-200 bg-red-50 text-red-700"
                : serverStatus === "online"
                  ? "border-emerald-200 bg-emerald-50 text-emerald-700"
                  : "border-slate-200 bg-white text-slate-500"
            }`}
          >
            {serverStatus === "offline" ? <WifiOff size={15} /> : <Wifi size={15} />}
            {serverStatus === "offline"
              ? "Backend server is not reachable"
              : serverStatus === "online"
                ? "Connected to the management server"
                : "Checking server connection..."}
          </div>

          <form onSubmit={handleSubmit} className="mt-5 space-y-5">
            <Input
              label="Phone number"
              type="tel"
              inputMode="tel"
              autoComplete="username"
              placeholder="Enter your phone number"
              value={phone}
              onChange={(event) => {
                setPhone(event.target.value);
                if (error) dispatch(clearError());
              }}
              icon={<Phone size={16} />}
              required
            />

            <div className="relative">
              <Input
                label="Password"
                type={showPassword ? "text" : "password"}
                autoComplete="current-password"
                placeholder="Enter your password"
                value={password}
                onChange={(event) => {
                  setPassword(event.target.value);
                  if (error) dispatch(clearError());
                }}
                icon={<Lock size={16} />}
                required
              />
              <button
                type="button"
                onClick={() => setShowPassword((value) => !value)}
                className="absolute bottom-2.5 right-3 rounded-md p-1 text-slate-400 hover:text-slate-700"
                aria-label={showPassword ? "Hide password" : "Show password"}
              >
                {showPassword ? <EyeOff size={18} /> : <Eye size={18} />}
              </button>
            </div>

            {(errorCode === "mfa_required" || errorCode === "invalid_mfa" || otp) && (
              <Input
                label="Authenticator or backup code"
                type="text"
                inputMode="numeric"
                autoComplete="one-time-code"
                placeholder="Enter the 6-digit code"
                value={otp}
                onChange={(event) => {
                  setOtp(event.target.value.replace(/\s/g, ""));
                  if (error) dispatch(clearError());
                }}
                icon={<ShieldCheck size={16} />}
                required
              />
            )}

            <Button
              type="submit"
              className="w-full"
              size="lg"
              isLoading={isLoading}
              disabled={!phone.trim() || !password || ((errorCode === "mfa_required" || errorCode === "invalid_mfa") && !otp.trim())}
            >
              Sign in
            </Button>
          </form>

          <div className="mt-6 rounded-xl bg-emerald-50 p-4 text-sm leading-6 text-emerald-900">
            Account unavailable? Contact the user-group office instead of sharing
            another member's login details.
          </div>
        </Motion.div>
      </section>
    </div>
  );
}
