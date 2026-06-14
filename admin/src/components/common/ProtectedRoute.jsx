import { Navigate, useLocation } from "react-router-dom";
import { useSelector } from "react-redux";
import LoadingSpinner from "./LoadingSpinner";

export default function ProtectedRoute({ children, roles }) {
  const {
    isAuthenticated,
    user,
    isProfileLoading,
    hasProfileLoaded,
  } = useSelector((state) => state.auth);
  const location = useLocation();

  if (isProfileLoading || !hasProfileLoaded) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-slate-50 dark:bg-slate-950">
        <LoadingSpinner text="Loading your account..." />
      </div>
    );
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" state={{ from: location }} replace />;
  }

  if (roles && !roles.includes(user?.role)) {
    return <Navigate to="/dashboard" replace />;
  }

  const securitySetupRequired = user?.must_change_password || user?.mfa_setup_required;
  if (securitySetupRequired && location.pathname !== "/profile") {
    return <Navigate to="/profile" state={{ securitySetupRequired: true }} replace />;
  }

  return children;
}
