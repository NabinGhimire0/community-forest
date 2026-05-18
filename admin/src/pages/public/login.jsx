import { useState, useEffect } from "react";
import { useNavigate, useLocation } from "react-router-dom";
import { useSelector, useDispatch } from "react-redux";
import { motion } from "framer-motion";
import { Phone, Lock, Eye, EyeOff, TreePine } from "lucide-react";
import { login, clearError } from "../../redux/slices/authSlice";
import Button from "../../components/ui/Button";
import Input from "../../components/ui/Input";

export default function Login() {
  const dispatch = useDispatch();
  const navigate = useNavigate();
  const location = useLocation();
  const { isLoading, error, isAuthenticated } = useSelector(
    (state) => state.auth,
  );
  const [phone, setPhone] = useState("");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const from = location.state?.from?.pathname || "/dashboard";

  useEffect(() => {
    if (isAuthenticated) navigate(from, { replace: true });
  }, [isAuthenticated, navigate, from]);
  useEffect(() => {
    return () => dispatch(clearError());
  }, [dispatch]);

  const handleSubmit = async (e) => {
    e.preventDefault();
    const result = await dispatch(login({ phone, password }));
    if (result.meta.requestStatus === "fulfilled")
      navigate(from, { replace: true });
  };

  return (
    <div className="min-h-screen flex">
      {/* Left panel */}
      <motion.div
        initial={{ opacity: 0, x: -50 }}
        animate={{ opacity: 1, x: 0 }}
        transition={{ duration: 0.6 }}
        className="hidden lg:flex lg:w-1/2 relative overflow-hidden bg-linear-to-br from-emerald-600 via-emerald-700 to-emerald-900"
      >
        <div className="absolute inset-0 opacity-10">
          <div className="absolute top-20 left-10 text-8xl">🌲</div>
          <div className="absolute top-40 right-20 text-6xl">🌳</div>
          <div className="absolute bottom-20 left-20 text-7xl">🌿</div>
          <div className="absolute bottom-40 right-10 text-5xl">🍃</div>
          <div className="absolute top-1/2 left-1/3 text-9xl">🏔️</div>
        </div>
        <div className="relative z-10 flex flex-col justify-center items-center p-12 text-white">
          <div className="w-20 h-20 rounded-2xl bg-white/20 backdrop-blur-sm flex items-center justify-center mb-8">
            <TreePine size={40} className="text-white" />
          </div>
          <h1 className="text-4xl font-bold mb-3">वन समिति</h1>
          <h2 className="text-2xl font-light mb-6 opacity-90">
            Community Forestry
          </h2>
          <p className="text-lg text-center opacity-80 max-w-md leading-relaxed">
            Management System for sustainable community forestry operations,
            resource tracking, and financial management.
          </p>
          <div className="mt-12 flex gap-8 text-center">
            <div>
              <p className="text-3xl font-bold">245+</p>
              <p className="text-sm opacity-70">Members</p>
            </div>
            <div>
              <p className="text-3xl font-bold">12</p>
              <p className="text-sm opacity-70">Resources</p>
            </div>
            <div>
              <p className="text-3xl font-bold">5K+</p>
              <p className="text-sm opacity-70">Transactions</p>
            </div>
          </div>
        </div>
      </motion.div>

      {/* Right form */}
      <motion.div
        initial={{ opacity: 0, y: 30 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5, delay: 0.2 }}
        className="flex-1 flex items-center justify-center p-8 bg-white dark:bg-gray-900"
      >
        <div className="w-full max-w-md">
          <div className="lg:hidden flex items-center gap-3 mb-8">
            <div className="w-10 h-10 rounded-xl bg-emerald-600 flex items-center justify-center">
              <TreePine size={22} className="text-white" />
            </div>
            <div>
              <h1 className="text-lg font-bold text-gray-900 dark:text-gray-100">
                श्री पाञ्चकन्या सामुदायिक वन उपभोक्ता समूह
              </h1>
              <p className="text-xs text-gray-400">Community Forestry</p>
            </div>
          </div>
          <h2 className="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-1">
            Welcome back
          </h2>
          <p className="text-gray-500 dark:text-gray-400 mb-8">
            Sign in to your account
          </p>

          {error && (
            <motion.div
              initial={{ opacity: 0, y: -10 }}
              animate={{ opacity: 1, y: 0 }}
              className="mb-4 p-3 rounded-lg bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-red-700 dark:text-red-400 text-sm"
            >
              {error}
            </motion.div>
          )}

          <form onSubmit={handleSubmit} className="space-y-5">
            <Input
              label="Phone Number"
              type="tel"
              placeholder="Enter phone number"
              value={phone}
              onChange={(e) => setPhone(e.target.value)}
              icon={<Phone size={16} />}
              required
            />
            <div className="relative">
              <Input
                label="Password"
                type={showPassword ? "text" : "password"}
                placeholder="Enter password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                icon={<Lock size={16} />}
                required
              />
              <button
                type="button"
                onClick={() => setShowPassword(!showPassword)}
                className="absolute right-3 top-9.5 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
              >
                {showPassword ? <EyeOff size={16} /> : <Eye size={16} />}
              </button>
            </div>
            <Button
              type="submit"
              isLoading={isLoading}
              className="w-full"
              size="lg"
            >
              Sign In
            </Button>
          </form>
          <p className="mt-6 text-center text-sm text-gray-400">
            Community Forestry Management System v1.0
          </p>
        </div>
      </motion.div>
    </div>
  );
}
