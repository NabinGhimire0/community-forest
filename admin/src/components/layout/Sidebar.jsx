import { useState } from "react";
import { NavLink, useNavigate } from "react-router-dom";
import { useDispatch, useSelector } from "react-redux";
import { AnimatePresence, motion as Motion } from "framer-motion";
import {
  AlertTriangle,
  ArrowLeftRight,
  BarChart3,
  Calendar,
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  CreditCard,
  FileText,
  LayoutDashboard,
  LogOut,
  Mail,
  Package,
  Receipt,
  Settings,
  DatabaseBackup,
  TreePine,
  Users,
  UserRound,
  X,
} from "lucide-react";
import { logout } from "../../redux/slices/authSlice";
import { cn, getImageUrl, getRoleLabel } from "../../utils/helpers";

const groupDefinitions = [
  {
    label: "Overview",
    items: [
      {
        path: "/dashboard",
        label: "Dashboard",
        memberLabel: "My Dashboard",
        icon: LayoutDashboard,
        roles: ["admin", "staff", "member"],
      },
      {
        path: "/profile",
        label: "My Profile",
        icon: UserRound,
        roles: ["admin", "staff", "member"],
      },
    ],
  },
  {
    label: "Member Services",
    items: [
      {
        path: "/members",
        label: "Member Register",
        icon: Users,
        roles: ["admin", "staff"],
      },
      {
        path: "/requests",
        label: "Requests",
        memberLabel: "My Requests",
        icon: FileText,
        roles: ["admin", "staff", "member"],
      },
      {
        path: "/payments",
        label: "Payments",
        memberLabel: "My Payments",
        icon: CreditCard,
        roles: ["admin", "staff", "member"],
      },
      {
        path: "/transactions",
        label: "Transactions",
        memberLabel: "My Ledger",
        icon: ArrowLeftRight,
        roles: ["admin", "staff", "member"],
      },
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
    label: "Operations",
    items: [
      {
        path: "/letters",
        label: "Letters",
        icon: Mail,
        roles: ["admin", "staff"],
      },
      {
        path: "/resources",
        label: "Resources & Stock",
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
      {
        path: "/system-data",
        label: "Exports & Backups",
        icon: DatabaseBackup,
        roles: ["admin"],
      },
      {
        path: "/samiti",
        label: "Organization Settings",
        icon: Settings,
        roles: ["admin"],
      },
    ],
  },
];

export default function Sidebar({
  isOpen,
  onClose,
  collapsed,
  onToggleCollapse,
}) {
  const dispatch = useDispatch();
  const navigate = useNavigate();
  const user = useSelector((state) => state.auth.user);
  const settings = useSelector((state) => state.appSettings.settings);
  const [expandedGroups, setExpandedGroups] = useState({
    Overview: true,
    "Member Services": true,
    Finance: true,
    Operations: true,
  });

  const role = user?.role;
  const groups = groupDefinitions
    .map((group) => ({
      ...group,
      items: group.items
        .filter((item) => item.roles.includes(role))
        .map((item) => ({
          ...item,
          displayLabel:
            role === "member" && item.memberLabel
              ? item.memberLabel
              : item.label,
        })),
    }))
    .filter((group) => group.items.length > 0);

  const handleLogout = async () => {
    await dispatch(logout());
    navigate("/login", { replace: true });
  };

  const logoUrl = getImageUrl(settings?.logo);

  return (
    <>
      <AnimatePresence>
        {isOpen && (
          <Motion.button
            type="button"
            aria-label="Close navigation"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="fixed inset-0 z-40 bg-slate-950/55 backdrop-blur-sm lg:hidden"
            onClick={onClose}
          />
        )}
      </AnimatePresence>

      <aside
        className={cn(
          "fixed inset-y-0 left-0 z-50 flex flex-col border-r border-slate-200/70 bg-white/90 shadow-xl shadow-slate-950/5 backdrop-blur-xl transition-all duration-300 dark:border-white/10 dark:bg-slate-950/90",
          collapsed ? "w-19" : "w-69",
          isOpen ? "translate-x-0" : "-translate-x-full lg:translate-x-0",
        )}
      >
        <div className="relative flex min-h-20 items-center gap-3 border-b border-slate-200/70 px-4 dark:border-white/10">
          <div className="flex min-w-0 flex-1 items-center gap-3">
            {logoUrl ? (
              <img
                src={logoUrl}
                alt="Organization logo"
                className="h-10 w-10 shrink-0 rounded-xl border border-emerald-100 bg-white object-contain p-1 dark:border-emerald-900/50"
              />
            ) : (
              <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-linear-to-br from-emerald-500 to-teal-700 text-white shadow-lg shadow-emerald-500/20">
                <TreePine size={21} />
              </div>
            )}

            {!collapsed && (
              <div className="min-w-0">
                <h1 className="line-clamp-2 text-sm font-bold leading-5 text-slate-900 dark:text-white">
                  {settings?.name}
                </h1>
                <p className="mt-0.5 truncate text-[11px] font-medium text-emerald-700 dark:text-emerald-400">
                  Community Forestry Portal
                </p>
              </div>
            )}
          </div>

          <button
            type="button"
            onClick={onClose}
            className="rounded-lg p-2 text-slate-500 hover:bg-slate-100 lg:hidden dark:hover:bg-white/5"
            aria-label="Close menu"
          >
            <X size={18} />
          </button>
        </div>

        <nav className="flex-1 space-y-2 overflow-y-auto px-3 py-4">
          {groups.map((group) => {
            const isExpanded = expandedGroups[group.label];
            return (
              <div key={group.label}>
                {!collapsed && (
                  <button
                    type="button"
                    onClick={() =>
                      setExpandedGroups((previous) => ({
                        ...previous,
                        [group.label]: !previous[group.label],
                      }))
                    }
                    className="mb-1 flex w-full items-center justify-between rounded-lg px-3 py-2 text-[10px] font-bold uppercase tracking-[0.16em] text-slate-400 hover:bg-slate-50 hover:text-slate-600 dark:hover:bg-white/5 dark:hover:text-slate-300"
                  >
                    {group.label}
                    <ChevronDown
                      size={13}
                      className={cn(
                        "transition-transform",
                        isExpanded && "rotate-180",
                      )}
                    />
                  </button>
                )}

                <AnimatePresence initial={false}>
                  {(collapsed || isExpanded) && (
                    <Motion.div
                      initial={{ height: 0, opacity: 0 }}
                      animate={{ height: "auto", opacity: 1 }}
                      exit={{ height: 0, opacity: 0 }}
                      className="space-y-1 overflow-hidden"
                    >
                      {group.items.map((item) => (
                        <NavLink
                          key={item.path}
                          to={item.path}
                          onClick={onClose}
                          title={collapsed ? item.displayLabel : undefined}
                          className={({ isActive }) =>
                            cn(
                              "group flex min-h-11 items-center gap-3 rounded-xl px-3 text-sm font-semibold transition-all",
                              collapsed && "justify-center px-2",
                              isActive
                                ? "bg-emerald-600 text-white shadow-md shadow-emerald-600/20"
                                : "text-slate-600 hover:bg-emerald-50 hover:text-emerald-800 dark:text-slate-400 dark:hover:bg-emerald-950/40 dark:hover:text-emerald-300",
                            )
                          }
                        >
                          <item.icon size={19} className="shrink-0" />
                          {!collapsed && (
                            <span className="truncate">{item.displayLabel}</span>
                          )}
                        </NavLink>
                      ))}
                    </Motion.div>
                  )}
                </AnimatePresence>
              </div>
            );
          })}
        </nav>

        <div className="border-t border-slate-200/70 p-3 dark:border-white/10">
          <div
            className={cn(
              "flex items-center gap-3 rounded-xl bg-slate-50 p-2.5 dark:bg-white/5",
              collapsed && "justify-center",
            )}
          >
            <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-emerald-100 font-bold text-emerald-700 dark:bg-emerald-950 dark:text-emerald-300">
              {user?.name?.charAt(0)?.toUpperCase() || "?"}
            </div>
            {!collapsed && (
              <div className="min-w-0 flex-1">
                <p className="truncate text-sm font-semibold text-slate-900 dark:text-white">
                  {user?.name}
                </p>
                <p className="text-[11px] text-slate-500 dark:text-slate-400">
                  {getRoleLabel(role)}
                </p>
              </div>
            )}
            {!collapsed && (
              <button
                type="button"
                onClick={handleLogout}
                className="rounded-lg p-2 text-slate-400 hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-950/40"
                title="Sign out"
                aria-label="Sign out"
              >
                <LogOut size={17} />
              </button>
            )}
          </div>

          <button
            type="button"
            onClick={onToggleCollapse}
            className="mt-2 hidden w-full items-center justify-center gap-2 rounded-lg py-2 text-xs font-semibold text-slate-500 hover:bg-slate-100 hover:text-slate-800 lg:flex dark:hover:bg-white/5 dark:hover:text-white"
          >
            {collapsed ? <ChevronRight size={16} /> : <ChevronLeft size={16} />}
            {!collapsed && "Collapse sidebar"}
          </button>

          {collapsed && (
            <button
              type="button"
              onClick={handleLogout}
              className="mt-1 hidden w-full items-center justify-center rounded-lg py-2 text-slate-400 hover:bg-red-50 hover:text-red-600 lg:flex dark:hover:bg-red-950/40"
              title="Sign out"
              aria-label="Sign out"
            >
              <LogOut size={18} />
            </button>
          )}
        </div>
      </aside>
    </>
  );
}
