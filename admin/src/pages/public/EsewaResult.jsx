import { CheckCircle2, XCircle, AlertTriangle, ArrowRight } from "lucide-react";
import { Link, useSearchParams } from "react-router-dom";
import { useSelector } from "react-redux";
import Button from "../../components/ui/Button";
import { Card, CardContent } from "../../components/ui/Card";

export default function EsewaResult() {
  const [params] = useSearchParams();
  const status = params.get("status") || "error";
  const message = params.get("message");
  const paymentId = params.get("payment_id");
  const token = useSelector((state) => state.auth.token);

  const success = status === "success";
  const failed = status === "failed";
  const Icon = success ? CheckCircle2 : failed ? XCircle : AlertTriangle;
  const title = success
    ? "Payment completed successfully"
    : failed
      ? "Payment was not completed"
      : "Payment could not be verified";

  return (
    <main className="min-h-screen bg-gray-50 px-4 py-16 dark:bg-gray-950">
      <Card className="mx-auto max-w-lg shadow-xl">
        <CardContent className="space-y-6 p-8 text-center">
          <Icon
            size={72}
            className={
              success
                ? "mx-auto text-emerald-600"
                : failed
                  ? "mx-auto text-red-600"
                  : "mx-auto text-amber-600"
            }
          />
          <div>
            <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">
              {title}
            </h1>
            <p className="mt-2 text-sm text-gray-500 dark:text-gray-400">
              {message ||
                (success
                  ? "eSewa confirmed the transaction and the request ledger has been updated."
                  : "No amount was settled. You can return to the request and try again.")}
            </p>
            {paymentId && (
              <p className="mt-3 text-xs font-mono text-gray-400">
                Payment #{paymentId}
              </p>
            )}
          </div>
          <div className="flex flex-col justify-center gap-3 sm:flex-row">
            <Link to={token ? "/requests" : "/login"}>
              <Button className="w-full sm:w-auto">
                {token ? "Back to Requests" : "Log in"}
                <ArrowRight size={16} />
              </Button>
            </Link>
            {token && (
              <Link to="/payments">
                <Button variant="outline" className="w-full sm:w-auto">
                  View Payments
                </Button>
              </Link>
            )}
          </div>
        </CardContent>
      </Card>
    </main>
  );
}
