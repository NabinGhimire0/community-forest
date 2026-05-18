import { Routes, Route, Navigate } from "react-router-dom";
import { useEffect } from "react";
import { useSelector, useDispatch } from "react-redux";
import { loadProfile } from "./redux/slices/authSlice";
import AdminLayout from "./components/layout/AdminLayout";
import ProtectedRoute from "./components/common/ProtectedRoute";
import { ToastProvider } from "./components/common/Toast";
import Login from "./pages/public/Login";
import Dashboard from "./pages/admin/Dashboard";
import Members from "./pages/admin/Members";
import Requests from "./pages/admin/Requests";
import Payments from "./pages/admin/Payments";
import Transactions from "./pages/admin/Transactions";
import Expenses from "./pages/admin/Expenses";
import Fines from "./pages/admin/Fines";
import Letters from "./pages/admin/Letters";
import Samiti from "./pages/admin/Samiti";
import Resources from "./pages/admin/Resources";
import FiscalYears from "./pages/admin/FiscalYears";
import Reports from "./pages/admin/Reports";

export default function App() {
  const dispatch = useDispatch();
  const { token } = useSelector((state) => state.auth);

  useEffect(() => {
    if (token) dispatch(loadProfile());
  }, [token, dispatch]);

  return (
    <ToastProvider>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route
          path="/"
          element={
            <ProtectedRoute>
              <AdminLayout />
            </ProtectedRoute>
          }
        >
          <Route index element={<Navigate to="/dashboard" replace />} />
          <Route path="dashboard" element={<Dashboard />} />
          <Route path="members" element={<Members />} />
          <Route path="requests" element={<Requests />} />
          <Route path="payments" element={<Payments />} />
          <Route path="transactions" element={<Transactions />} />
          <Route path="expenses" element={<Expenses />} />
          <Route path="fines" element={<Fines />} />
          <Route path="letters" element={<Letters />} />
          <Route path="samiti" element={<Samiti />} />
          <Route path="resources" element={<Resources />} />
          <Route path="fiscal-years" element={<FiscalYears />} />
          <Route path="reports" element={<Reports />} />
        </Route>
        <Route path="*" element={<Navigate to="/dashboard" replace />} />
      </Routes>
    </ToastProvider>
  );
}
