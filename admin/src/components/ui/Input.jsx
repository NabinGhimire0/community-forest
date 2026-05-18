import { forwardRef } from "react";
import { cn } from "../../utils/helpers";

const Input = forwardRef(({ className, label, error, icon, ...props }, ref) => (
  <div className="w-full">
    {label && (
      <label className="block text-sm font-medium text-gray-600 dark:text-gray-400 mb-1.5">
        {label}
      </label>
    )}
    <div className="relative">
      {icon && (
        <div className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400">
          {icon}
        </div>
      )}
      <input
        ref={ref}
        className={cn(
          "w-full px-4 py-2.5 border border-gray-200/80 dark:border-white/10 rounded-xl bg-white/50 backdrop-blur-sm dark:bg-gray-900/40 text-gray-900 dark:text-gray-100 text-sm transition-all duration-300 placeholder:text-gray-400 dark:placeholder:text-gray-500 focus:outline-none focus:border-emerald-500 focus:ring-4 focus:ring-emerald-500/15 focus:bg-white dark:focus:bg-gray-900 shadow-sm",
          icon && "pl-10",
          error && "border-red-500 focus:border-red-500 focus:ring-red-500/20",
          className,
        )}
        {...props}
      />
    </div>
    {error && <p className="mt-1 text-xs text-red-500">{error}</p>}
  </div>
));

Input.displayName = "Input";
export default Input;
