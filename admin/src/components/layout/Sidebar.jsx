import { useState } from "react";
import { NavLink, useNavigate } from "react-router-dom";
import { useSelector, useDispatch } from "react-redux";
import { motion, AnimatePresence } from "framer-motion";
import {
  TreePine,
  LayoutDashboard,
  Users,
  FileText,
  CreditCard,
  ArrowLeftRight,
  Receipt,
  AlertTriangle,
  Mail,
  Settings,
  Package,
  Calendar,
  BarChart3,
  ChevronDown,
  LogOut,
  X,
} from "lucide-react";
import { cn } from "../../utils/helpers";
import { logout } from "../../redux/slices/authSlice";

const navGroups = [
  {
    label: "Overview",
    items: [{ path: "/dashboard", label: "Dashboard", icon: LayoutDashboard }],
  },
  {
    label: "Management",
    items: [
      {
        path: "/members",
        label: "Members",
        icon: Users,
        roles: ["admin", "staff"],
      },
      { path: "/requests", label: "Requests", icon: FileText },
      { path: "/payments", label: "Payments", icon: CreditCard },
      { path: "/transactions", label: "Transactions", icon: ArrowLeftRight },
    ],
  },
  {
    label: "Finance",
    items: [
      {
        path: "/expenses",
        label: "Expenses",
        icon: Receipt,
        roles: ["admin", "staff"],
      },
      {
        path: "/fines",
        label: "Fines",
        icon: AlertTriangle,
        roles: ["admin", "staff"],
      },
    ],
  },
  {
    label: "Communication",
    items: [
      {
        path: "/letters",
        label: "Letters",
        icon: Mail,
        roles: ["admin", "staff"],
      },
    ],
  },
  {
    label: "Administration",
    items: [
      {
        path: "/samiti",
        label: "Samiti Settings",
        icon: Settings,
        roles: ["admin"],
      },
      {
        path: "/resources",
        label: "Resources",
        icon: Package,
        roles: ["admin", "staff"],
      },
      {
        path: "/fiscal-years",
        label: "Fiscal Years",
        icon: Calendar,
        roles: ["admin", "staff"],
      },
      {
        path: "/reports",
        label: "Reports",
        icon: BarChart3,
        roles: ["admin", "staff"],
      },
    ],
  },
];

export default function Sidebar({ isOpen, onClose, collapsed }) {
  const dispatch = useDispatch();
  const navigate = useNavigate();
  const { user } = useSelector((state) => state.auth);
  const [expandedGroups, setExpandedGroups] = useState({
    Overview: true,
    Management: true,
    Finance: true,
    Communication: true,
    Administration: true,
  });

  const toggleGroup = (label) =>
    setExpandedGroups((prev) => ({ ...prev, [label]: !prev[label] }));
  const handleLogout = () => {
    dispatch(logout());
    navigate("/login");
  };

  const filteredGroups = navGroups
    .map((group) => ({
      ...group,
      items: group.items.filter(
        (item) => !item.roles || (user && item.roles.includes(user.role)),
      ),
    }))
    .filter((group) => group.items.length > 0);

  return (
    <>
      <AnimatePresence>
        {isOpen && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="fixed inset-0 bg-black/50 z-40 lg:hidden"
            onClick={onClose}
          />
        )}
      </AnimatePresence>

      <aside
        className={cn(
          "fixed left-0 top-0 z-50 h-full bg-white/70 dark:bg-[#0B1120]/70 backdrop-blur-xl border-r border-gray-200/50 dark:border-white/5 transition-all duration-300 flex flex-col shadow-[4px_0_24px_rgba(0,0,0,0.02)]",
          collapsed ? "w-17" : "w-65",
          isOpen ? "translate-x-0" : "-translate-x-full lg:translate-x-0",
        )}
      >
        {/* Header */}
        <div className="flex items-center justify-between h-16 px-4 border-b border-gray-200 dark:border-white/10">
          <div
            className={cn(
              "flex items-center gap-3 overflow-hidden",
              collapsed && "justify-center",
            )}
          >
            <div className="shrink-0 w-8 h-8 rounded-lg bg-emerald-600 flex items-center justify-center">
              <TreePine size={18} className="text-white" />
            </div>
            {!collapsed && (
              <div className="overflow-hidden whitespace-nowrap">
                <h1 className="text-sm font-bold text-gray-900 dark:text-gray-100">
                  श्री पाञ्चकन्या सामुदायिक वन उपभोक्ता समूह
                </h1>
                <p className="text-[10px] text-gray-400">Community Forestry</p>
              </div>
            )}
          </div>
          <button
            onClick={onClose}
            className="lg:hidden p-1 rounded-md hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-400"
          >
            <X size={18} />
          </button>
        </div>

        {/* Nav */}
        <nav className="flex-1 overflow-y-auto py-4 px-3 space-y-1">
          {filteredGroups.map((group) => (
            <div key={group.label} className="mb-2">
              {!collapsed && (
                <button
                  onClick={() => toggleGroup(group.label)}
                  className="flex items-center justify-between w-full px-2 py-1.5 text-[11px] font-semibold text-gray-400 uppercase tracking-wider hover:text-gray-600 dark:hover:text-gray-300 transition-colors"
                >
                  {group.label}
                  <ChevronDown
                    size={14}
                    className={cn(
                      "transition-transform",
                      expandedGroups[group.label] && "rotate-180",
                    )}
                  />
                </button>
              )}
              <AnimatePresence>
                {(expandedGroups[group.label] || collapsed) && (
                  <motion.div
                    initial={{ height: 0, opacity: 0 }}
                    animate={{ height: "auto", opacity: 1 }}
                    exit={{ height: 0, opacity: 0 }}
                    transition={{ duration: 0.2 }}
                    className="overflow-hidden"
                  >
                    {group.items.map((item) => (
                      <NavLink
                        key={item.path}
                        to={item.path}
                        onClick={onClose}
                        className={({ isActive }) =>
                          cn(
                            "flex items-center gap-3 w-full px-3 py-2.5 rounded-xl text-sm font-medium transition-all duration-300 relative group overflow-hidden",
                            isActive
                              ? "bg-linear-to-r from-emerald-50 to-teal-50/50 text-emerald-700 dark:from-emerald-900/40 dark:to-teal-900/20 dark:text-emerald-300 shadow-sm"
                              : "text-gray-500 hover:bg-white/60 hover:text-gray-900 dark:text-gray-400 dark:hover:bg-white/5 dark:hover:text-gray-100",
                            collapsed && "justify-center px-2",
                          )
                        }
                        title={collapsed ? item.label : undefined}
                      >
                        <item.icon size={20} className="shrink-0" />
                        {!collapsed && <span>{item.label}</span>}
                      </NavLink>
                    ))}
                  </motion.div>
                )}
              </AnimatePresence>
            </div>
          ))}
        </nav>

        {/* User */}
        <div className="border-t border-gray-200 dark:border-white/10 p-3">
          <div
            className={cn(
              "flex items-center gap-3",
              collapsed && "justify-center",
            )}
          >
            <div className="w-8 h-8 rounded-full bg-emerald-50 dark:bg-emerald-900/20 flex items-center justify-center shrink-0">
              <span className="text-xs font-semibold text-emerald-700 dark:text-emerald-400">
                {user?.name?.charAt(0).toUpperCase() || "?"}
              </span>
            </div>
            {!collapsed && (
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium text-gray-900 dark:text-gray-100 truncate">
                  {user?.name}
                </p>
                <p className="text-[11px] text-gray-400 capitalize">
                  {user?.role}
                </p>
              </div>
            )}
            <button
              onClick={handleLogout}
              className="p-1.5 rounded-md hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-400 hover:text-red-500 transition-colors"
              title="Logout"
            >
              <LogOut size={16} />
            </button>
          </div>
        </div>
      </aside>
    </>
  );
}
