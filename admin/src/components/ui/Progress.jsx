import { cn } from "../../utils/helpers";

export default function Progress({ value = 0, max = 100, className }) {
  const percentage = Math.min(Math.max((value / max) * 100, 0), 100);
  return (
    <div
      className={cn(
        "w-full h-2 bg-gray-200 dark:bg-gray-700 rounded-full overflow-hidden",
        className,
      )}
    >
      <div
        className="h-full rounded-full bg-emerald-500 transition-all duration-500"
        style={{ width: `${percentage}%` }}
      />
    </div>
  );
}
