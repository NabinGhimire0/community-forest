import { ArrowLeft, Home, TriangleAlert } from "lucide-react";
import { Link } from "react-router-dom";

export default function NotFound() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-slate-50 px-5 text-center dark:bg-slate-950">
      <div className="max-w-md">
        <div className="mx-auto flex h-16 w-16 items-center justify-center rounded-2xl bg-amber-100 text-amber-700 dark:bg-amber-950 dark:text-amber-300">
          <TriangleAlert size={30} />
        </div>
        <p className="mt-6 text-7xl font-black tracking-tight text-slate-950 dark:text-white">
          404
        </p>
        <h1 className="mt-3 text-2xl font-extrabold text-slate-900 dark:text-white">
          Page not found
        </h1>
        <p className="mt-3 text-sm leading-6 text-slate-500 dark:text-slate-400">
          The page does not exist or your account does not have access to it.
        </p>
        <div className="mt-7 flex flex-wrap justify-center gap-3">
          <Link
            to="/"
            className="inline-flex items-center gap-2 rounded-xl border border-slate-200 bg-white px-4 py-2.5 text-sm font-bold text-slate-700 hover:bg-slate-100 dark:border-white/10 dark:bg-slate-900 dark:text-slate-200"
          >
            <Home size={17} /> Public page
          </Link>
          <Link
            to="/dashboard"
            className="inline-flex items-center gap-2 rounded-xl bg-emerald-700 px-4 py-2.5 text-sm font-bold text-white hover:bg-emerald-800"
          >
            <ArrowLeft size={17} /> Dashboard
          </Link>
        </div>
      </div>
    </div>
  );
}
