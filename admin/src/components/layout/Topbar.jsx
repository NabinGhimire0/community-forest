import { Home, Menu, Moon, Sun } from "lucide-react";
import { Link } from "react-router-dom";
import { useDispatch, useSelector } from "react-redux";
import { toggleTheme } from "../../redux/slices/themeSlice";
import Button from "../ui/Button";
import { getOrganizationLocation, getRoleLabel } from "../../utils/helpers";

export default function Topbar({ onMenuClick }) {
  const dispatch = useDispatch();
  const theme = useSelector((state) => state.theme);
  const user = useSelector((state) => state.auth.user);
  const settings = useSelector((state) => state.appSettings.settings);
  const location = getOrganizationLocation(settings);

  return (
    <header className="sticky top-0 z-30 flex min-h-16 items-center justify-between gap-4 border-b border-slate-200/70 bg-white/80 px-4 backdrop-blur-xl dark:border-white/10 dark:bg-slate-950/75 lg:px-7">
      <div className="flex min-w-0 items-center gap-3">
        <Button
          variant="ghost"
          size="icon"
          className="lg:hidden"
          onClick={onMenuClick}
          aria-label="Open navigation"
        >
          <Menu size={20} />
        </Button>

        <div className="min-w-0">
          <h2 className="truncate text-sm font-bold text-slate-900 dark:text-white sm:text-base">
            {settings?.name}
          </h2>
          <p className="hidden truncate text-xs text-slate-500 dark:text-slate-400 sm:block">
            {location || "Community Forestry Management Portal"}
          </p>
        </div>
      </div>

      <div className="flex shrink-0 items-center gap-1.5">
        <Link
          to="/"
          className="hidden items-center gap-2 rounded-xl px-3 py-2 text-sm font-semibold text-slate-500 hover:bg-slate-100 hover:text-slate-900 dark:hover:bg-white/5 dark:hover:text-white sm:flex"
        >
          <Home size={17} />
          Public page
        </Link>

        <Button
          variant="ghost"
          size="icon"
          onClick={() => dispatch(toggleTheme())}
          aria-label={theme === "dark" ? "Use light theme" : "Use dark theme"}
          title={theme === "dark" ? "Light theme" : "Dark theme"}
        >
          {theme === "dark" ? <Sun size={18} /> : <Moon size={18} />}
        </Button>

        <Link
          to="/profile"
          title="View profile"
          className="ml-1 hidden items-center gap-2 rounded-xl border-l border-slate-200 py-1 pl-3 pr-2 transition hover:bg-slate-100 dark:border-white/10 dark:hover:bg-white/5 md:flex"
        >
          <div className="flex h-8 w-8 items-center justify-center rounded-full bg-emerald-100 text-xs font-bold text-emerald-700 dark:bg-emerald-950 dark:text-emerald-300">
            {user?.name?.charAt(0)?.toUpperCase() || "?"}
          </div>
          <div className="max-w-36 leading-tight">
            <p className="truncate text-sm font-semibold text-slate-900 dark:text-white">
              {user?.name}
            </p>
            <p className="text-[11px] text-slate-500 dark:text-slate-400">
              {getRoleLabel(user?.role)}
            </p>
          </div>
        </Link>
      </div>
    </header>
  );
}
