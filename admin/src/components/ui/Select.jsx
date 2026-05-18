import { forwardRef } from "react";
import { ChevronDown } from "lucide-react";
import { cn } from "../../utils/helpers";

const Select = forwardRef(
  ({ className, label, error, options, placeholder, ...props }, ref) => (
    <div className="w-full">
      {label && (
        <label className="block text-sm font-medium text-gray-600 dark:text-gray-400 mb-1.5">
          {label}
        </label>
      )}
      <div className="relative">
        <select
          ref={ref}
          className={cn(
            "w-full px-3 py-2 border border-gray-200 dark:border-white/10 rounded-lg bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 text-sm appearance-none pr-10 transition-colors focus:outline-none focus:border-emerald-500 focus:ring-2 focus:ring-emerald-500/20",
            error && "border-red-500",
            className,
          )}
          {...props}
        >
          {placeholder && <option value="">{placeholder}</option>}
          {options?.map((opt) => (
            <option key={opt.value} value={opt.value}>
              {opt.label}
            </option>
          ))}
        </select>
        <ChevronDown
          size={16}
          className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 pointer-events-none"
        />
      </div>
      {error && <p className="mt-1 text-xs text-red-500">{error}</p>}
    </div>
  ),
);

Select.displayName = "Select";
export default Select;
