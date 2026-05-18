export function cn(...classes) {
  return classes.filter(Boolean).join(" ");
}

export function formatCurrency(amount) {
  return new Intl.NumberFormat("en-NP", {
    style: "currency",
    currency: "NPR",
    minimumFractionDigits: 0,
    maximumFractionDigits: 2,
  }).format(amount);
}

export function formatNumber(num) {
  if (num >= 1000000) return `${(num / 1000000).toFixed(1)}M`;
  if (num >= 1000) return `${(num / 1000).toFixed(1)}K`;
  return num.toString();
}

export function formatDate(dateStr) {
  if (!dateStr) return "-";
  return new Date(dateStr).toLocaleDateString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

export function formatDateTime(dateStr) {
  if (!dateStr) return "-";
  return new Date(dateStr).toLocaleString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export function getInitials(name) {
  if (!name) return "?";
  return name
    .split(" ")
    .map((n) => n[0])
    .join("")
    .toUpperCase()
    .slice(0, 2);
}

// Tailwind class maps for status badges
export function getStatusClasses(status) {
  const map = {
    pending:
      "bg-amber-50 text-amber-700 dark:bg-amber-900/20 dark:text-amber-400",
    approved:
      "bg-emerald-50 text-emerald-700 dark:bg-emerald-900/20 dark:text-emerald-400",
    rejected: "bg-red-50 text-red-700 dark:bg-red-900/20 dark:text-red-400",
    completed:
      "bg-blue-50 text-blue-700 dark:bg-blue-900/20 dark:text-blue-400",
    paid: "bg-emerald-50 text-emerald-700 dark:bg-emerald-900/20 dark:text-emerald-400",
    waived: "bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400",
    active:
      "bg-emerald-50 text-emerald-700 dark:bg-emerald-900/20 dark:text-emerald-400",
    inactive: "bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400",
    credit:
      "bg-emerald-50 text-emerald-700 dark:bg-emerald-900/20 dark:text-emerald-400",
    debit: "bg-red-50 text-red-700 dark:bg-red-900/20 dark:text-red-400",
    incoming: "bg-blue-50 text-blue-700 dark:bg-blue-900/20 dark:text-blue-400",
    outgoing:
      "bg-orange-50 text-orange-700 dark:bg-orange-900/20 dark:text-orange-400",
    cash: "bg-emerald-50 text-emerald-700 dark:bg-emerald-900/20 dark:text-emerald-400",
    online: "bg-blue-50 text-blue-700 dark:bg-blue-900/20 dark:text-blue-400",
    bank_transfer:
      "bg-purple-50 text-purple-700 dark:bg-purple-900/20 dark:text-purple-400",
  };
  return (
    map[status] ||
    "bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400"
  );
}
