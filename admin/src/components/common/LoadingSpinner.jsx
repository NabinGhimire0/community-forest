import { Loader2 } from "lucide-react";
import { cn } from "../../utils/helpers";

export default function LoadingSpinner({ size = 24, className, text }) {
  return (
    <div
      className={cn(
        "flex flex-col items-center justify-center gap-3 py-12",
        className,
      )}
    >
      <Loader2 size={size} className="animate-spin text-emerald-600" />
      {text && <p className="text-sm text-gray-400">{text}</p>}
    </div>
  );
}
