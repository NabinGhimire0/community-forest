import { useState, useEffect } from "react";
import { Outlet } from "react-router-dom";
import { useDispatch } from "react-redux";
import { setTheme } from "../../redux/slices/themeSlice";
import { logout } from "../../redux/slices/authSlice";
import Sidebar from "./Sidebar";
import Topbar from "./Topbar";

export default function AdminLayout() {
  const dispatch = useDispatch();
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);

  useEffect(() => {
    const saved = localStorage.getItem("theme");
    if (saved) dispatch(setTheme(saved));
    else if (window.matchMedia("(prefers-color-scheme: dark)").matches)
      dispatch(setTheme("dark"));
  }, [dispatch]);

  useEffect(() => {
    const handleUnauthorized = () => dispatch(logout());
    window.addEventListener("auth:unauthorized", handleUnauthorized);
    return () =>
      window.removeEventListener("auth:unauthorized", handleUnauthorized);
  }, [dispatch]);

  return (
    <div className="min-h-screen bg-transparent text-gray-900 dark:text-gray-100">
      <Sidebar
        isOpen={sidebarOpen}
        onClose={() => setSidebarOpen(false)}
        collapsed={sidebarCollapsed}
        onToggleCollapse={() => setSidebarCollapsed(!sidebarCollapsed)}
      />
      <div
        className={`transition-all duration-300 flex-1 ${
          sidebarCollapsed ? "lg:ml-[68px]" : "lg:ml-[260px]"
        }`}
      >
        <Topbar
          onMenuClick={() => setSidebarOpen(!sidebarOpen)}
          sidebarCollapsed={sidebarCollapsed}
        />
        <main className="p-4 lg:p-6">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
