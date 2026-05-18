import { cn, getInitials } from "../../utils/helpers";

export default function Avatar({ src, name, size = "md", className }) {
  const sizes = {
    sm: "w-8 h-8 text-xs",
    md: "w-10 h-10 text-sm",
    lg: "w-12 h-12 text-base",
    xl: "w-16 h-16 text-lg",
  };

  if (src)
    return (
      <img
        src={src}
        alt={name || "Avatar"}
        className={cn("rounded-full object-cover", sizes[size], className)}
      />
    );

  return (
    <div
      className={cn(
        "rounded-full flex items-center justify-center font-semibold bg-emerald-50 dark:bg-emerald-900/20 text-emerald-700 dark:text-emerald-400",
        sizes[size],
        className,
      )}
    >
      {getInitials(name)}
    </div>
  );
}
