import { lazy, Suspense, useEffect } from "react";
import { Navigate, Route, Routes } from "react-router-dom";
import { useDispatch, useSelector } from "react-redux";
import { loadProfile } from "./redux/slices/authSlice";
import {
  fetchAppSettings,
  fetchPublicCommittee,
} from "./redux/slices/appSettingsSlice";
import AdminLayout from "./components/layout/AdminLayout";
import ProtectedRoute from "./components/common/ProtectedRoute";
import LoadingSpinner from "./components/common/LoadingSpinner";
import { ToastProvider } from "./components/common/Toast";
import Landing from "./pages/public/landing";
import Login from "./pages/public/login";
import EsewaResult from "./pages/public/EsewaResult";
import NotFound from "./pages/NotFound";

const Dashboard = lazy(() => import("./pages/admin/Dashboard"));
const Members = lazy(() => import("./pages/admin/Members"));
const Requests = lazy(() => import("./pages/admin/Requests"));
const Payments = lazy(() => import("./pages/admin/Payments"));
const Transactions = lazy(() => import("./pages/admin/Transactions"));
const Expenses = lazy(() => import("./pages/admin/Expenses"));
const Fines = lazy(() => import("./pages/admin/Fines"));
const Letters = lazy(() => import("./pages/admin/Letters"));
const Samiti = lazy(() => import("./pages/admin/Samiti"));
const Resources = lazy(() => import("./pages/admin/Resources"));
const FiscalYears = lazy(() => import("./pages/admin/FiscalYears"));
const Reports = lazy(() => import("./pages/admin/Reports"));
const Profile = lazy(() => import("./pages/admin/Profile"));
const SystemData = lazy(() => import("./pages/admin/SystemData"));

const adminAndStaff = ["admin", "staff"];

function RoleProtected({ roles, children }) {
  return <ProtectedRoute roles={roles}>{children}</ProtectedRoute>;
}

function PageFallback() {
  return <LoadingSpinner text="Loading page..." />;
}

export default function App() {
  const dispatch = useDispatch();
  const settings = useSelector((state) => state.appSettings.settings);

  useEffect(() => {
    dispatch(fetchAppSettings());
    dispatch(fetchPublicCommittee());
  }, [dispatch]);

  useEffect(() => {
    dispatch(loadProfile());
  }, [dispatch]);

  useEffect(() => {
    document.title = settings?.name || "Community Forestry Management System";
  }, [settings?.name]);

  return (
    <ToastProvider>
      <Suspense fallback={<PageFallback />}>
        <Routes>
          <Route path="/" element={<Landing />} />
          <Route path="/login" element={<Login />} />
          <Route path="/payments/esewa/result" element={<EsewaResult />} />

          <Route
            element={
              <ProtectedRoute>
                <AdminLayout />
              </ProtectedRoute>
            }
          >
            <Route path="/dashboard" element={<Dashboard />} />
            <Route path="/profile" element={<Profile />} />
            <Route
              path="/members"
              element={
                <RoleProtected roles={adminAndStaff}>
                  <Members />
                </RoleProtected>
              }
            />
            <Route path="/requests" element={<Requests />} />
            <Route path="/payments" element={<Payments />} />
            <Route path="/transactions" element={<Transactions />} />
            <Route
              path="/expenses"
              element={
                <RoleProtected roles={adminAndStaff}>
                  <Expenses />
                </RoleProtected>
              }
            />
            <Route
              path="/fines"
              element={
                <RoleProtected roles={adminAndStaff}>
                  <Fines />
                </RoleProtected>
              }
            />
            <Route
              path="/letters"
              element={
                <RoleProtected roles={adminAndStaff}>
                  <Letters />
                </RoleProtected>
              }
            />
            <Route
              path="/resources"
              element={
                <RoleProtected roles={adminAndStaff}>
                  <Resources />
                </RoleProtected>
              }
            />
            <Route
              path="/fiscal-years"
              element={
                <RoleProtected roles={adminAndStaff}>
                  <FiscalYears />
                </RoleProtected>
              }
            />
            <Route
              path="/reports"
              element={
                <RoleProtected roles={adminAndStaff}>
                  <Reports />
                </RoleProtected>
              }
            />
            <Route
              path="/system-data"
              element={
                <RoleProtected roles={["admin"]}>
                  <SystemData />
                </RoleProtected>
              }
            />
            <Route
              path="/samiti"
              element={
                <RoleProtected roles={["admin"]}>
                  <Samiti />
                </RoleProtected>
              }
            />
          </Route>

          <Route path="/app" element={<Navigate to="/dashboard" replace />} />
          <Route path="*" element={<NotFound />} />
        </Routes>
      </Suspense>
    </ToastProvider>
  );
}
