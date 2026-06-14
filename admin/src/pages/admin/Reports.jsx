import { useState, useEffect } from "react";
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
  Users,
  CreditCard,
  ArrowLeftRight,
  TrendingUp,
  BarChart3,
  FileText,
  DollarSign,
  Package,
  AlertCircle,
} from "lucide-react";
import { api } from "../../services/api";
import { Card, CardContent, CardHeader } from "../../components/ui/Card";
import Button from "../../components/ui/Button";
import Select from "../../components/ui/Select";
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
import { useToast } from "../../components/common/Toast";
import { formatCurrency, formatNumber, formatDate } from "../../utils/helpers";
import {
  buildFiscalYearOptions,
  getActiveFiscalYearId,
} from "../../utils/fiscalYears";

const COLORS = [
  "#059669",
  "#10b981",
  "#34d399",
  "#6ee7b7",
  "#a7f3d0",
  "#d1fae5",
  "#064e3b",
];

export default function Reports() {
  const { addToast } = useToast();
  const [activeTab, setActiveTab] = useState("dashboard");
  const [isLoading, setIsLoading] = useState(true);
  const [dashboardData, setDashboardData] = useState(null);
  const [memberReport, setMemberReport] = useState(null);
  const [resourceReport, setResourceReport] = useState(null);
  const [financialReport, setFinancialReport] = useState(null);
  const [fiscalYears, setFiscalYears] = useState([]);
  const [selectedFiscalYear, setSelectedFiscalYear] = useState("");

  useEffect(() => {
    const fetchFiscalYears = async () => {
      try {
        const res = await api.getFiscalYears();
        if (res.success && res.data) {
          setFiscalYears(res.data);
          setSelectedFiscalYear((current) =>
            current || getActiveFiscalYearId(res.data),
          );
        }
      } catch (err) {
        console.error("Failed to fetch fiscal years:", err);
      }
    };
    fetchFiscalYears();
  }, []);

  useEffect(() => {
    const fetchData = async () => {
      setIsLoading(true);
      try {
        if (activeTab === "dashboard") {
          const res = await api.getDashboardCharts();
          if (res.success) setDashboardData(res.data);
        } else if (activeTab === "members") {
          const res = await api.getMemberReport();
          if (res.success) setMemberReport(res.data);
        } else if (activeTab === "resources" && selectedFiscalYear) {
          const res = await api.getResourceReport({
            fiscal_year_id: selectedFiscalYear,
          });
          if (res.success) setResourceReport(res.data);
        } else if (activeTab === "financial" && selectedFiscalYear) {
          const res = await api.getFinancialReport({
            fiscal_year_id: selectedFiscalYear,
          });
          if (res.success) setFinancialReport(res.data);
        }
      } catch {
        addToast(`Failed to load ${activeTab} report`, "error");
      } finally {
        setIsLoading(false);
      }
    };
    fetchData();
  }, [activeTab, selectedFiscalYear, addToast]);

  const tabs = [
    { id: "dashboard", label: "Dashboard Overview", icon: BarChart3 },
    { id: "members", label: "Member Report", icon: Users },
    { id: "resources", label: "Resource Report", icon: Package },
    { id: "financial", label: "Financial Report", icon: DollarSign },
  ];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">
          Reports & Analytics
        </h1>
        <p className="text-sm text-gray-500 dark:text-gray-400">
          Comprehensive insights into your community forest operations
        </p>
      </div>

      {/* Tabs */}
      <div className="flex flex-wrap gap-1 bg-gray-100 dark:bg-gray-800/50 rounded-lg p-1 w-fit">
        {tabs.map((tab) => (
          <button
            key={tab.id}
            onClick={() => {
              setActiveTab(tab.id);
              setIsLoading(true);
            }}
            className={`flex items-center gap-2 px-4 py-2 rounded-md text-sm font-medium transition-colors ${
              activeTab === tab.id
                ? "bg-white dark:bg-gray-900 text-emerald-600 dark:text-emerald-400 shadow-sm"
                : "text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200"
            }`}
          >
            <tab.icon size={16} /> {tab.label}
          </button>
        ))}
      </div>

      {/* Fiscal Year Selector for Reports */}
      {(activeTab === "resources" || activeTab === "financial") && (
        <div className="flex justify-end">
          <Select
            value={selectedFiscalYear}
            onChange={(e) => setSelectedFiscalYear(e.target.value)}
            options={buildFiscalYearOptions(fiscalYears)}
            className="w-48"
          />
        </div>
      )}

      {isLoading ? (
        <LoadingSpinner text="Loading report..." />
      ) : (
        <>
          {/* Dashboard Overview Tab */}
          {activeTab === "dashboard" && dashboardData && (
            <div className="space-y-6">
              {/* Activity Timeline */}
              <Card>
                <CardHeader>
                  <h3 className="font-semibold text-gray-900 dark:text-gray-100">
                    Recent Activities
                  </h3>
                </CardHeader>
                <CardContent>
                  <div className="space-y-4">
                    {(dashboardData.recent_activities || []).map(
                      (activity, i) => (
                        <div
                          key={i}
                          className="flex items-start gap-3 pb-3 border-b border-gray-100 dark:border-gray-800 last:border-0"
                        >
                          {activity.type === "request" && (
                            <FileText
                              size={18}
                              className="text-blue-500 mt-0.5"
                            />
                          )}
                          {activity.type === "payment" && (
                            <CreditCard
                              size={18}
                              className="text-emerald-500 mt-0.5"
                            />
                          )}
                          <div className="flex-1">
                            <p className="font-medium text-gray-900 dark:text-gray-100">
                              {activity.action}
                            </p>
                            <p className="text-sm text-gray-500">
                              {activity.description}
                            </p>
                          </div>
                          <span className="text-xs text-gray-400">
                            {formatDate(activity.created_at)}
                          </span>
                        </div>
                      ),
                    )}
                  </div>
                </CardContent>
              </Card>

              {/* Charts */}
              <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                <Card>
                  <CardHeader>
                    <h3 className="font-semibold">
                      Ward-wise Member Distribution
                    </h3>
                  </CardHeader>
                  <CardContent>
                    <div className="h-80">
                      <ResponsiveContainer width="100%" height="100%">
                        <BarChart data={dashboardData.ward_wise_members || []}>
                          <CartesianGrid
                            strokeDasharray="3 3"
                            stroke="var(--border)"
                          />
                          <XAxis dataKey="ward_no" tick={{ fontSize: 12 }} />
                          <YAxis tick={{ fontSize: 12 }} />
                          <Tooltip />
                          <Bar
                            dataKey="count"
                            fill="#059669"
                            radius={[4, 4, 0, 0]}
                            name="Members"
                          />
                        </BarChart>
                      </ResponsiveContainer>
                    </div>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader>
                    <h3 className="font-semibold">
                      Payment Method Distribution
                    </h3>
                  </CardHeader>
                  <CardContent>
                    <div className="h-80">
                      <ResponsiveContainer width="100%" height="100%">
                        <PieChart>
                          <Pie
                            data={
                              dashboardData.payment_method_distribution || []
                            }
                            cx="50%"
                            cy="50%"
                            innerRadius={60}
                            outerRadius={90}
                            paddingAngle={4}
                            dataKey="total"
                            nameKey="payment_method"
                            label={({ payment_method, percent }) =>
                              `${payment_method} (${(percent * 100).toFixed(0)}%)`
                            }
                          >
                            {(
                              dashboardData.payment_method_distribution || []
                            ).map((_, i) => (
                              <Cell key={i} fill={COLORS[i % COLORS.length]} />
                            ))}
                          </Pie>
                          <Tooltip
                            formatter={(value) => formatCurrency(value)}
                          />
                          <Legend />
                        </PieChart>
                      </ResponsiveContainer>
                    </div>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader>
                    <h3 className="font-semibold">Expense by Category</h3>
                  </CardHeader>
                  <CardContent>
                    <div className="h-80">
                      <ResponsiveContainer width="100%" height="100%">
                        <PieChart>
                          <Pie
                            data={dashboardData.expense_by_category || []}
                            cx="50%"
                            cy="50%"
                            innerRadius={60}
                            outerRadius={90}
                            paddingAngle={4}
                            dataKey="total_amount"
                            nameKey="category_name"
                            label={({ category_name, percent }) =>
                              `${category_name} (${(percent * 100).toFixed(0)}%)`
                            }
                          >
                            {(dashboardData.expense_by_category || []).map(
                              (_, i) => (
                                <Cell
                                  key={i}
                                  fill={COLORS[i % COLORS.length]}
                                />
                              ),
                            )}
                          </Pie>
                          <Tooltip
                            formatter={(value) => formatCurrency(value)}
                          />
                          <Legend />
                        </PieChart>
                      </ResponsiveContainer>
                    </div>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader>
                    <h3 className="font-semibold">Fine Status Overview</h3>
                  </CardHeader>
                  <CardContent>
                    <div className="h-80">
                      <ResponsiveContainer width="100%" height="100%">
                        <PieChart>
                          <Pie
                            data={dashboardData.fine_status || []}
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
                            {(dashboardData.fine_status || []).map((_, i) => (
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
            </div>
          )}

          {/* Member Report Tab */}
          {activeTab === "members" && memberReport && (
            <div className="space-y-6">
              <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
                <Card>
                  <CardContent className="p-5">
                    <p className="text-sm text-gray-500">Total Members</p>
                    <p className="text-3xl font-bold text-gray-900">
                      {formatNumber(memberReport.total)}
                    </p>
                  </CardContent>
                </Card>
                <Card className="border-emerald-200 bg-emerald-50">
                  <CardContent className="p-5">
                    <p className="text-sm text-emerald-600">Active Members</p>
                    <p className="text-3xl font-bold text-emerald-600">
                      {formatNumber(memberReport.active)}
                    </p>
                  </CardContent>
                </Card>
                <Card className="border-red-200 bg-red-50">
                  <CardContent className="p-5">
                    <p className="text-sm text-red-600">Inactive Members</p>
                    <p className="text-3xl font-bold text-red-600">
                      {formatNumber(memberReport.inactive)}
                    </p>
                  </CardContent>
                </Card>
              </div>

              <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                <Card>
                  <CardHeader>
                    <h3 className="font-semibold">Ward-wise Distribution</h3>
                  </CardHeader>
                  <CardContent>
                    <div className="h-80">
                      <ResponsiveContainer width="100%" height="100%">
                        <BarChart data={memberReport.ward_wise || []}>
                          <CartesianGrid strokeDasharray="3 3" />
                          <XAxis dataKey="ward_no" />
                          <YAxis />
                          <Tooltip />
                          <Bar
                            dataKey="count"
                            fill="#059669"
                            radius={[4, 4, 0, 0]}
                            name="Members"
                          />
                        </BarChart>
                      </ResponsiveContainer>
                    </div>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader>
                    <h3 className="font-semibold">Monthly Joining Trend</h3>
                  </CardHeader>
                  <CardContent>
                    <div className="h-80">
                      <ResponsiveContainer width="100%" height="100%">
                        <LineChart data={memberReport.monthly_joining || []}>
                          <CartesianGrid strokeDasharray="3 3" />
                          <XAxis dataKey="month" />
                          <YAxis />
                          <Tooltip />
                          <Line
                            type="monotone"
                            dataKey="count"
                            stroke="#059669"
                            strokeWidth={2}
                            name="New Members"
                          />
                        </LineChart>
                      </ResponsiveContainer>
                    </div>
                  </CardContent>
                </Card>
              </div>
            </div>
          )}

          {/* Resource Report Tab */}
          {activeTab === "resources" && resourceReport && (
            <div className="space-y-6">
              <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                <Card>
                  <CardHeader>
                    <h3 className="font-semibold">Sales by Resource Type</h3>
                  </CardHeader>
                  <CardContent>
                    <div className="h-80">
                      <ResponsiveContainer width="100%" height="100%">
                        <PieChart>
                          <Pie
                            data={resourceReport.sales_by_type || []}
                            cx="50%"
                            cy="50%"
                            innerRadius={60}
                            outerRadius={90}
                            dataKey="total_amount"
                            nameKey="type_name"
                            label={({ type_name, percent }) =>
                              `${type_name} (${(percent * 100).toFixed(0)}%)`
                            }
                          >
                            {(resourceReport.sales_by_type || []).map(
                              (_, i) => (
                                <Cell
                                  key={i}
                                  fill={COLORS[i % COLORS.length]}
                                />
                              ),
                            )}
                          </Pie>
                          <Tooltip
                            formatter={(value) => formatCurrency(value)}
                          />
                          <Legend />
                        </PieChart>
                      </ResponsiveContainer>
                    </div>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader>
                    <h3 className="font-semibold">Stock Overview</h3>
                  </CardHeader>
                  <CardContent className="p-0">
                    <div className="overflow-x-auto">
                      <Table>
                        <TableHeader>
                          <TableRow>
                            <TableHead>Resource</TableHead>
                            <TableHead>Total</TableHead>
                            <TableHead>Reserved</TableHead>
                            <TableHead>Available</TableHead>
                            <TableHead>Remaining</TableHead>
                            <TableHead>Usage</TableHead>
                          </TableRow>
                        </TableHeader>
                        <TableBody>
                          {(resourceReport.stock || [])
                            .slice(0, 10)
                            .map((item, i) => (
                              <TableRow key={i}>
                                <TableCell className="font-medium">
                                  {item.item_name}
                                </TableCell>
                                <TableCell>
                                  {item.total_quantity} {item.unit}
                                </TableCell>
                                <TableCell className="font-medium text-amber-700">
                                  {item.reserved_quantity || 0} {item.unit}
                                </TableCell>
                                <TableCell className="font-medium text-emerald-700">
                                  {item.available_quantity || 0} {item.unit}
                                </TableCell>
                                <TableCell>
                                  {item.remaining_quantity} {item.unit}
                                </TableCell>
                                <TableCell>
                                  <div className="flex items-center gap-2">
                                    <div className="w-16 h-2 bg-gray-200 rounded-full overflow-hidden">
                                      <div
                                        className="h-full bg-emerald-500 rounded-full"
                                        style={{
                                          width: `${item.used_percent}%`,
                                        }}
                                      />
                                    </div>
                                    <span className="text-xs">
                                      {item.used_percent}%
                                    </span>
                                  </div>
                                </TableCell>
                              </TableRow>
                            ))}
                        </TableBody>
                      </Table>
                    </div>
                  </CardContent>
                </Card>
              </div>
            </div>
          )}

          {/* Financial Report Tab */}
          {activeTab === "financial" && financialReport && (
            <div className="space-y-6">
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
                <Card className="border-emerald-200 bg-emerald-50">
                  <CardContent className="p-4">
                    <p className="text-sm text-emerald-600">Total Revenue</p>
                    <p className="text-2xl font-bold text-emerald-600">
                      {formatCurrency(financialReport.total_revenue)}
                    </p>
                  </CardContent>
                </Card>
                <Card className="border-red-200 bg-red-50">
                  <CardContent className="p-4">
                    <p className="text-sm text-red-600">Total Expenses</p>
                    <p className="text-2xl font-bold text-red-600">
                      {formatCurrency(financialReport.total_expenses)}
                    </p>
                  </CardContent>
                </Card>
                <Card>
                  <CardContent className="p-4">
                    <p className="text-sm text-gray-500">Total Collected</p>
                    <p className="text-2xl font-bold text-gray-900">
                      {formatCurrency(financialReport.total_collected)}
                    </p>
                  </CardContent>
                </Card>
                <Card className="border-blue-200 bg-blue-50">
                  <CardContent className="p-4">
                    <p className="text-sm text-blue-600">Net Balance</p>
                    <p
                      className={`text-2xl font-bold ${financialReport.net_balance >= 0 ? "text-blue-600" : "text-red-600"}`}
                    >
                      {formatCurrency(Math.abs(financialReport.net_balance))}
                      {financialReport.net_balance >= 0 ? " CR" : " DR"}
                    </p>
                  </CardContent>
                </Card>
              </div>

              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <Card className="border-violet-200 bg-violet-50 dark:border-violet-900/50 dark:bg-violet-950/20">
                  <CardContent className="p-4">
                    <p className="text-sm text-violet-700 dark:text-violet-300">Historical Balance Recovered</p>
                    <p className="text-2xl font-bold text-violet-700 dark:text-violet-300">
                      {formatCurrency(financialReport.historical_collected || 0)}
                    </p>
                  </CardContent>
                </Card>
                <Card className="border-amber-200 bg-amber-50 dark:border-amber-900/50 dark:bg-amber-950/20">
                  <CardContent className="p-4">
                    <p className="text-sm text-amber-700 dark:text-amber-300">Historical Balance Outstanding</p>
                    <p className="text-2xl font-bold text-amber-700 dark:text-amber-300">
                      {formatCurrency(financialReport.historical_outstanding || 0)}
                    </p>
                  </CardContent>
                </Card>
              </div>

              <Card>
                <CardHeader>
                  <h3 className="font-semibold">Monthly Income vs Expense</h3>
                </CardHeader>
                <CardContent>
                  <div className="h-80">
                    <ResponsiveContainer width="100%" height="100%">
                      <LineChart data={financialReport.monthly_data || []}>
                        <CartesianGrid strokeDasharray="3 3" />
                        <XAxis dataKey="month" />
                        <YAxis />
                        <Tooltip formatter={(value) => formatCurrency(value)} />
                        <Legend />
                        <Line
                          type="monotone"
                          dataKey="income"
                          stroke="#059669"
                          strokeWidth={2}
                          name="Income"
                        />
                        <Line
                          type="monotone"
                          dataKey="expense"
                          stroke="#ef4444"
                          strokeWidth={2}
                          name="Expense"
                        />
                      </LineChart>
                    </ResponsiveContainer>
                  </div>
                </CardContent>
              </Card>

              <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                <Card>
                  <CardHeader>
                    <h3 className="font-semibold">Expense by Category</h3>
                  </CardHeader>
                  <CardContent>
                    <div className="h-80">
                      <ResponsiveContainer width="100%" height="100%">
                        <PieChart>
                          <Pie
                            data={financialReport.category_expenses || []}
                            cx="50%"
                            cy="50%"
                            innerRadius={60}
                            outerRadius={90}
                            dataKey="total_amount"
                            nameKey="category_name"
                            label={({ category_name, percent }) =>
                              `${category_name} (${(percent * 100).toFixed(0)}%)`
                            }
                          >
                            {(financialReport.category_expenses || []).map(
                              (_, i) => (
                                <Cell
                                  key={i}
                                  fill={COLORS[i % COLORS.length]}
                                />
                              ),
                            )}
                          </Pie>
                          <Tooltip
                            formatter={(value) => formatCurrency(value)}
                          />
                          <Legend />
                        </PieChart>
                      </ResponsiveContainer>
                    </div>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader>
                    <h3 className="font-semibold">Revenue Breakdown</h3>
                  </CardHeader>
                  <CardContent>
                    <div className="space-y-4">
                      <div className="flex justify-between items-center p-3 bg-gray-50 rounded-lg">
                        <span className="text-sm">Resource Sales</span>
                        <span className="font-semibold text-emerald-600">
                          {formatCurrency(financialReport.resource_sales)}
                        </span>
                      </div>
                      <div className="flex justify-between items-center p-3 bg-gray-50 rounded-lg">
                        <span className="text-sm">Membership Fees</span>
                        <span className="font-semibold text-emerald-600">
                          {formatCurrency(financialReport.membership_fees)}
                        </span>
                      </div>
                      <div className="flex justify-between items-center p-3 bg-gray-50 rounded-lg">
                        <span className="text-sm">Fines Collected</span>
                        <span className="font-semibold text-emerald-600">
                          {formatCurrency(financialReport.total_fines)}
                        </span>
                      </div>
                      <div className="flex justify-between items-center p-3 bg-violet-50 rounded-lg dark:bg-violet-950/20">
                        <span className="text-sm">Historical Amount Recovered</span>
                        <span className="font-semibold text-violet-700 dark:text-violet-300">
                          {formatCurrency(financialReport.historical_collected || 0)}
                        </span>
                      </div>
                      <div className="flex justify-between items-center p-3 bg-amber-50 rounded-lg dark:bg-amber-950/20">
                        <span className="text-sm">Historical Amount Still Due</span>
                        <span className="font-semibold text-amber-700 dark:text-amber-300">
                          {formatCurrency(financialReport.historical_outstanding || 0)}
                        </span>
                      </div>
                      <div className="flex justify-between items-center p-3 bg-emerald-50 rounded-lg">
                        <span className="font-semibold">Total Revenue</span>
                        <span className="font-bold text-emerald-700">
                          {formatCurrency(financialReport.total_revenue)}
                        </span>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  );
}
