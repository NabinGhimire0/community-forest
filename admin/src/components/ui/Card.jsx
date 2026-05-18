import { cn } from "../../utils/helpers";

export function Card({ className, children, hover, ...props }) {
  return (
    <div
      className={cn(
        "glass rounded-2xl shadow-[0_8px_30px_rgb(0,0,0,0.04)] dark:shadow-none",
        hover &&
          "transition-all duration-300 hover:-translate-y-1 hover:shadow-[0_8px_30px_rgb(0,0,0,0.08)] cursor-pointer",
        className,
      )}
      {...props}
    >
      {children}
    </div>
  );
}

export function CardHeader({ className, children, ...props }) {
  return (
    <div
      className={cn(
        "px-6 py-4 border-b border-gray-200 dark:border-white/10",
        className,
      )}
      {...props}
    >
      {children}
    </div>
  );
}

export function CardContent({ className, children, ...props }) {
  return (
    <div className={cn("px-6 py-4", className)} {...props}>
      {children}
    </div>
  );
}

export function CardFooter({ className, children, ...props }) {
  return (
    <div
      className={cn(
        "px-6 py-4 border-t border-gray-200 dark:border-white/10",
        className,
      )}
      {...props}
    >
      {children}
    </div>
  );
}
