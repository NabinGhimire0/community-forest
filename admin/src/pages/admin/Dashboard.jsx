import { useEffect } from "react";
import { useSelector, useDispatch } from "react-redux";
import { motion } from "framer-motion";
import {
  Users,
  FileText,
  CreditCard,
  Wallet,
  TrendingUp,
  Package,
  AlertCircle,
  CheckCircle,
  Clock,
} from "lucide-react";
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
  Legend,
  LineChart,
  Line,
} from "recharts";
import {
  fetchDashboard,
  fetchDashboardCharts,
} from "../../redux/slices/dashboardSlice";
import { Card, CardContent, CardHeader } from "../../components/ui/Card";
import {
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
} from "../../components/ui/Table";
import Badge from "../../components/ui/Badge";
import LoadingSpinner from "../../components/common/LoadingSpinner";
import { formatCurrency, formatNumber, formatDate } from "../../utils/helpers";

const COLORS = [
  "#10b981",
  "#3b82f6",
  "#f59e0b",
  "#8b5cf6",
  "#ec4899",
  "#06b6d4",
  "#ef4444",
];

export default function Dashboard() {
  const dispatch = useDispatch();
  const { data, charts, isLoading } = useSelector((state) => state.dashboard);

  useEffect(() => {
    dispatch(fetchDashboard());
    dispatch(fetchDashboardCharts());
  }, [dispatch]);

  if (isLoading && !data) return <LoadingSpinner text="Loading dashboard..." />;

  const statCards = [
    {
      key: "total_members",
      label: "Total Members",
      icon: Users,
      color: "text-emerald-600 dark:text-emerald-400",
      bg: "bg-emerald-50 dark:bg-emerald-900/30",
      format: "number",
    },
    {
      key: "pending_requests",
      label: "Pending Requests",
      icon: FileText,
      color: "text-amber-600 dark:text-amber-400",
      bg: "bg-amber-50 dark:bg-amber-900/30",
      format: "number",
    },
    {
      key: "total_revenue",
      label: "Total Revenue",
      icon: CreditCard,
      color: "text-emerald-600 dark:text-emerald-400",
      bg: "bg-emerald-50 dark:bg-emerald-900/30",
      format: "currency",
    },
    {
      key: "balance",
      label: "Net Balance",
      icon: Wallet,
      color: "text-emerald-600 dark:text-emerald-400",
      bg: "bg-emerald-50 dark:bg-emerald-900/30",
      format: "currency",
    },
  ];

  return (
    <div className="space-y-6">
      {/* Welcome Section */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">
            Dashboard
          </h1>
          <p className="text-sm text-gray-500 dark:text-gray-400">
            Welcome back! Here's what's happening with your community forest
            today.
          </p>
        </div>
        {data?.active_fiscal_year && (
          <div className="px-3 py-2 bg-emerald-50 dark:bg-emerald-900/20 rounded-lg">
            <p className="text-xs text-emerald-600 dark:text-emerald-400">
              Active Fiscal Year
            </p>
            <p className="text-sm font-semibold text-emerald-700 dark:text-emerald-300">
              {data.active_fiscal_year}
            </p>
          </div>
        )}
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {statCards.map((card, i) => (
          <motion.div
            key={card.key}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: i * 0.1, duration: 0.5, ease: "easeOut" }}
          >
            <Card hover>
              <CardContent className="p-5">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-gray-500 dark:text-gray-400">
                      {card.label}
                    </p>
                    <p className="text-3xl font-extrabold text-gray-900 dark:text-gray-100 mt-1 tracking-tight">
                      {card.format === "currency"
                        ? formatCurrency(data?.[card.key] || 0)
                        : formatNumber(data?.[card.key] || 0)}
                    </p>
                  </div>
                  <div className={`p-3.5 rounded-2xl ${card.bg}`}>
                    <card.icon size={24} className={card.color} />
                  </div>
                </div>
                <div className="flex items-center gap-1.5 mt-4">
                  <TrendingUp size={16} className="text-emerald-500" />
                  <span className="text-sm text-emerald-500 font-semibold">
                    +12%
                  </span>
                  <span className="text-sm text-gray-400">from last month</span>
                </div>
              </CardContent>
            </Card>
          </motion.div>
        ))}
      </div>

      {/* Charts Section */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Monthly Financials Chart */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.4 }}
        >
          <Card>
            <CardHeader>
              <h3 className="font-semibold text-gray-900 dark:text-gray-100">
                Monthly Financial Overview
              </h3>
            </CardHeader>
            <CardContent>
              <div className="h-80">
                <ResponsiveContainer width="100%" height="100%">
                  <BarChart data={charts?.monthly_financials || []}>
                    <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                    <XAxis
                      dataKey="month"
                      tick={{ fontSize: 12 }}
                      stroke="#9ca3af"
                    />
                    <YAxis tick={{ fontSize: 12 }} stroke="#9ca3af" />
                    <Tooltip
                      contentStyle={{
                        background: "#fff",
                        border: "1px solid #e5e7eb",
                        borderRadius: "8px",
                        fontSize: "12px",
                      }}
                      formatter={(value) => formatCurrency(value)}
                    />
                    <Legend />
                    <Bar
                      dataKey="income"
                      fill="#059669"
                      radius={[4, 4, 0, 0]}
                      name="Income"
                    />
                    <Bar
                      dataKey="expense"
                      fill="#ef4444"
                      radius={[4, 4, 0, 0]}
                      name="Expense"
                    />
                  </BarChart>
                </ResponsiveContainer>
              </div>
            </CardContent>
          </Card>
        </motion.div>

        {/* Resource Distribution Pie Chart */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.5 }}
        >
          <Card>
            <CardHeader>
              <h3 className="font-semibold text-gray-900 dark:text-gray-100">
                Resource Sales by Type
              </h3>
            </CardHeader>
            <CardContent>
              <div className="h-80">
                <ResponsiveContainer width="100%" height="100%">
                  <PieChart>
                    <Pie
                      data={charts?.resource_sales_by_type || []}
                      cx="50%"
                      cy="50%"
                      innerRadius={60}
                      outerRadius={90}
                      paddingAngle={4}
                      dataKey="total_amount"
                      nameKey="type_name"
                      label={({ type_name, percent }) =>
                        `${type_name} (${(percent * 100).toFixed(0)}%)`
                      }
                    >
                      {(charts?.resource_sales_by_type || []).map((_, i) => (
                        <Cell key={i} fill={COLORS[i % COLORS.length]} />
                      ))}
                    </Pie>
                    <Tooltip formatter={(value) => formatCurrency(value)} />
                    <Legend />
                  </PieChart>
                </ResponsiveContainer>
              </div>
            </CardContent>
          </Card>
        </motion.div>
      </div>

      {/* Second Row Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Member Growth Trend */}
        <Card>
          <CardHeader>
            <h3 className="font-semibold text-gray-900 dark:text-gray-100">
              Member Growth Trend
            </h3>
          </CardHeader>
          <CardContent>
            <div className="h-72">
              <ResponsiveContainer width="100%" height="100%">
                <LineChart data={charts?.member_growth || []}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                  <XAxis
                    dataKey="month"
                    tick={{ fontSize: 12 }}
                    stroke="#9ca3af"
                  />
                  <YAxis tick={{ fontSize: 12 }} stroke="#9ca3af" />
                  <Tooltip />
                  <Line
                    type="monotone"
                    dataKey="count"
                    stroke="#059669"
                    strokeWidth={2}
                    dot={{ fill: "#059669", r: 4 }}
                    name="New Members"
                  />
                </LineChart>
              </ResponsiveContainer>
            </div>
          </CardContent>
        </Card>

        {/* Request Status Distribution */}
        <Card>
          <CardHeader>
            <h3 className="font-semibold text-gray-900 dark:text-gray-100">
              Request Status Distribution
            </h3>
          </CardHeader>
          <CardContent>
            <div className="h-72">
              <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie
                    data={charts?.request_status_distribution || []}
                    cx="50%"
                    cy="50%"
                    innerRadius={60}
                    outerRadius={90}
                    paddingAngle={4}
                    dataKey="count"
                    nameKey="status"
                    label={({ status, percent }) =>
                      `${status} (${(percent * 100).toFixed(0)}%)`
                    }
                  >
                    {(charts?.request_status_distribution || []).map((_, i) => (
                      <Cell key={i} fill={COLORS[i % COLORS.length]} />
                    ))}
                  </Pie>
                  <Tooltip />
                  <Legend />
                </PieChart>
              </ResponsiveContainer>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Recent Activity Section */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Recent Requests */}
        <Card>
          <CardHeader>
            <h3 className="font-semibold text-gray-900 dark:text-gray-100">
              Recent Requests
            </h3>
          </CardHeader>
          <CardContent className="p-0">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Member</TableHead>
                  <TableHead>Resource</TableHead>
                  <TableHead>Qty</TableHead>
                  <TableHead>Status</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {(data?.recent_requests || []).map((req, i) => (
                  <TableRow key={i}>
                    <TableCell className="font-medium">
                      {req.member_name}
                    </TableCell>
                    <TableCell>{req.resource_name}</TableCell>
                    <TableCell>{req.quantity_requested}</TableCell>
                    <TableCell>
                      <Badge status={req.status} />
                    </TableCell>
                  </TableRow>
                ))}
                {(!data?.recent_requests ||
                  data.recent_requests.length === 0) && (
                  <TableRow>
                    <TableCell
                      colSpan={4}
                      className="text-center text-gray-400 py-8"
                    >
                      No recent requests
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </CardContent>
        </Card>

        {/* Recent Payments */}
        <Card>
          <CardHeader>
            <h3 className="font-semibold text-gray-900 dark:text-gray-100">
              Recent Payments
            </h3>
          </CardHeader>
          <CardContent className="p-0">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Member</TableHead>
                  <TableHead>Amount</TableHead>
                  <TableHead>Method</TableHead>
                  <TableHead>Status</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {(data?.recent_payments || []).map((payment, i) => (
                  <TableRow key={i}>
                    <TableCell className="font-medium">
                      {payment.member_name}
                    </TableCell>
                    <TableCell className="font-semibold text-emerald-600">
                      {formatCurrency(payment.amount)}
                    </TableCell>
                    <TableCell>
                      <Badge status={payment.payment_method}>
                        {payment.payment_method}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <Badge
                        status={
                          payment.status === "paid" ? "active" : "pending"
                        }
                      >
                        {payment.status === "paid" ? "Verified" : "Pending"}
                      </Badge>
                    </TableCell>
                  </TableRow>
                ))}
                {(!data?.recent_payments ||
                  data.recent_payments.length === 0) && (
                  <TableRow>
                    <TableCell
                      colSpan={4}
                      className="text-center text-gray-400 py-8"
                    >
                      No recent payments
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
