import { cn } from "../../utils/helpers";

export function Table({ className, children, ...props }) {
  return (
    <div className="overflow-x-auto">
      <table className={cn("w-full text-sm", className)} {...props}>
        {children}
      </table>
    </div>
  );
}

export function TableHeader({ className, children, ...props }) {
  return (
    <thead
      className={cn("bg-gray-50/80 dark:bg-[#0B1120]/50 backdrop-blur-sm", className)}
      {...props}
    >
      {children}
    </thead>
  );
}

export function TableBody({ className, children, ...props }) {
  return (
    <tbody className={className} {...props}>
      {children}
    </tbody>
  );
}

export function TableRow({ className, children, hover = true, ...props }) {
  return (
    <tr
      className={cn(
        "border-b border-gray-200/60 dark:border-white/5 transition-colors",
        hover && "hover:bg-emerald-50/40 dark:hover:bg-white/5",
        className,
      )}
      {...props}
    >
      {children}
    </tr>
  );
}

export function TableHead({ className, children, ...props }) {
  return (
    <th
      className={cn(
        "px-4 py-3 text-left text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider",
        className,
      )}
      {...props}
    >
      {children}
    </th>
  );
}

export function TableCell({ className, children, ...props }) {
  return (
    <td
      className={cn("px-4 py-3 text-gray-900 dark:text-gray-100", className)}
      {...props}
    >
      {children}
    </td>
  );
}
