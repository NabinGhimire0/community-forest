import { cn, getStatusClasses } from "../../utils/helpers";

export default function Badge({ status, children, className }) {
  return (
    <span
      className={cn(
        "inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium capitalize",
        getStatusClasses(status),
        className,
      )}
    >
      {children || status?.replace("_", " ")}
    </span>
  );
}
