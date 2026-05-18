import { Moon, Sun, Menu, Bell, Search } from "lucide-react";
import { useSelector, useDispatch } from "react-redux";
import { toggleTheme } from "../../redux/slices/themeSlice";
import Button from "../ui/Button";

export default function Topbar({ onMenuClick, sidebarCollapsed }) {
  const dispatch = useDispatch();
  const theme = useSelector((state) => state.theme);
  const { user } = useSelector((state) => state.auth);

  return (
    <header className="sticky top-0 z-30 h-16 border-b border-gray-200/50 dark:border-white/5 glass flex items-center justify-between px-4 lg:px-8">
      <div className="flex items-center gap-3">
        <Button
          variant="ghost"
          size="icon"
          className="lg:hidden"
          onClick={onMenuClick}
        >
          <Menu size={20} />
        </Button>
        <div>
          <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">
            श्री पाञ्चकन्या सामुदायिक वन उपभोक्ता समूह
          </h2>
          <p className="text-xs text-gray-400 hidden sm:block">
            वन समिति प्रबन्धन प्रणाली
          </p>
        </div>
      </div>
      <div className="flex items-center gap-2">
        <div className="hidden md:flex items-center bg-gray-100 dark:bg-gray-800 rounded-lg px-3 py-1.5">
          <Search size={16} className="text-gray-400 mr-2" />
          <input
            type="text"
            placeholder="Search..."
            className="bg-transparent border-none outline-none text-sm text-gray-900 dark:text-gray-100 placeholder:text-gray-400 w-40"
          />
        </div>
        <Button variant="ghost" size="icon" className="relative">
          <Bell size={18} />
          <span className="absolute top-1.5 right-1.5 w-2 h-2 bg-emerald-500 rounded-full" />
        </Button>
        <Button
          variant="ghost"
          size="icon"
          onClick={() => dispatch(toggleTheme())}
        >
          {theme === "dark" ? <Sun size={18} /> : <Moon size={18} />}
        </Button>
        <div className="hidden sm:flex items-center gap-2 ml-2 pl-2 border-l border-gray-200 dark:border-white/10">
          <div className="w-8 h-8 rounded-full bg-emerald-50 dark:bg-emerald-900/20 flex items-center justify-center">
            <span className="text-xs font-semibold text-emerald-700 dark:text-emerald-400">
              {user?.name?.charAt(0).toUpperCase() || "?"}
            </span>
          </div>
          <div className="hidden md:block">
            <p className="text-sm font-medium text-gray-900 dark:text-gray-100">
              {user?.name}
            </p>
            <p className="text-[11px] text-gray-400 capitalize">{user?.role}</p>
          </div>
        </div>
      </div>
    </header>
  );
}
