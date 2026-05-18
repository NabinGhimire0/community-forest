import { Link } from "react-router-dom";
import { motion } from "framer-motion";
import { AlertTriangle, Home, ArrowLeft } from "lucide-react";

import Button from "../components/ui/Button";

// ---------------------------------------------------------------------------
// Animation variants
// ---------------------------------------------------------------------------
const containerVariants = {
  hidden: { opacity: 0 },
  visible: {
    opacity: 1,
    transition: { staggerChildren: 0.1 },
  },
};
const itemVariants = {
  hidden: { opacity: 0, y: 20 },
  visible: { opacity: 1, y: 0, transition: { duration: 0.5, ease: "easeOut" } },
};

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------
export default function NotFound() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-slate-50 px-4">
      <motion.div
        className="text-center max-w-md mx-auto"
        variants={containerVariants}
        initial="hidden"
        animate="visible"
      >
        {/* Icon */}
        <motion.div variants={itemVariants} className="flex justify-center mb-6">
          <div className="p-4 rounded-2xl bg-amber-50">
            <AlertTriangle className="h-12 w-12 text-amber-500" />
          </div>
        </motion.div>

        {/* 404 Number */}
        <motion.div variants={itemVariants}>
          <h1 className="text-8xl font-extrabold text-slate-900 tracking-tight">
            404
          </h1>
        </motion.div>

        {/* Title */}
        <motion.div variants={itemVariants} className="mt-4">
          <h2 className="text-2xl font-bold text-slate-800">
            Page not found
          </h2>
        </motion.div>

        {/* Description */}
        <motion.div variants={itemVariants} className="mt-3">
          <p className="text-sm text-slate-500 leading-relaxed">
            The page you are looking for doesn't exist or has been moved.
            Please check the URL or navigate back to the homepage.
          </p>
        </motion.div>

        {/* Action Buttons */}
        <motion.div variants={itemVariants} className="mt-8 flex items-center justify-center gap-3">
          <Link to="/">
            <Button variant="outline">
              <Home className="h-4 w-4 mr-2" />
              Go Home
            </Button>
          </Link>
          <Link to="/admin/dashboard">
            <Button variant="primary">
              <ArrowLeft className="h-4 w-4 mr-2" />
              Dashboard
            </Button>
          </Link>
        </motion.div>
      </motion.div>
    </div>
  );
}
