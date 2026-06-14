import { useSelector } from "react-redux";
import AdminStaffDashboard from "./dashboard/AdminStaffDashboard";
import MemberDashboard from "./dashboard/MemberDashboard";

export default function Dashboard() {
  const role = useSelector((state) => state.auth.user?.role);

  if (role === "member") return <MemberDashboard />;
  return <AdminStaffDashboard />;
}
