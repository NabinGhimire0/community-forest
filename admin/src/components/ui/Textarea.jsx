import { forwardRef } from "react";
import { cn } from "../../utils/helpers";

const Textarea = forwardRef(({ className, label, error, ...props }, ref) => (
  <div className="w-full">
    {label && (
      <label className="block text-sm font-medium text-gray-600 dark:text-gray-400 mb-1.5">
        {label}
      </label>
    )}
    <textarea
      ref={ref}
      className={cn(
        "w-full px-3 py-2 border border-gray-200 dark:border-white/10 rounded-lg bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 text-sm min-h-20 resize-y transition-colors placeholder:text-gray-400 dark:placeholder:text-gray-500 focus:outline-none focus:border-emerald-500 focus:ring-2 focus:ring-emerald-500/20",
        error && "border-red-500",
        className,
      )}
      {...props}
    />
    {error && <p className="mt-1 text-xs text-red-500">{error}</p>}
  </div>
));

Textarea.displayName = "Textarea";
export default Textarea;
