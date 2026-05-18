import React from "react";
import { Link } from "react-router-dom";
import { motion } from "framer-motion";
import { TreePine, ArrowRight, ShieldCheck } from "lucide-react";
import Button from "../../components/ui/Button";

const stats = [
  { value: "50+", label: "Active Forest Committees" },
  { value: "12,000", label: "Community Members" },
  { value: "2.5M", label: "Trees Managed" },
  { value: "98%", label: "Financial Transparency" },
];

export default function Landing() {
  return (
    <div className="min-h-screen bg-linear-to-b from-emerald-50 via-white to-emerald-50 flex flex-col">
      {/* Hero Section */}
      <div className="flex-1 flex flex-col items-center justify-center px-6 py-20 text-center">
        {/* Brand Icon */}
        <motion.div
          className="bg-emerald-600 text-white p-5 rounded-3xl shadow-lg shadow-emerald-200 mb-8"
          initial={{ opacity: 0, scale: 0.5, y: -20 }}
          animate={{ opacity: 1, scale: 1, y: 0 }}
          transition={{ duration: 0.6, ease: "easeOut" }}
        >
          <TreePine className="h-14 w-14" />
        </motion.div>

        {/* Title */}
        <motion.h1
          className="text-5xl md:text-6xl font-extrabold text-gray-900 mb-3 tracking-tight"
          initial={{ opacity: 0, y: 30 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6, delay: 0.2 }}
        >
          श्री पाञ्चकन्या सामुदायिक वन उपभोक्ता समूह
        </motion.h1>

        {/* Subtitle */}
        <motion.p
          className="text-lg md:text-xl text-gray-500 mb-4 max-w-md"
          initial={{ opacity: 0, y: 30 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6, delay: 0.3 }}
        >
          Community Forestry Management System
        </motion.p>

        {/* Tagline */}
        <motion.div
          className="flex items-center gap-2 bg-emerald-100 text-emerald-700 px-4 py-2 rounded-full text-sm font-medium mb-10"
          initial={{ opacity: 0, y: 30 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6, delay: 0.4 }}
        >
          <ShieldCheck className="h-4 w-4" />
          Sustainable Forest Management
        </motion.div>

        {/* CTA Button */}
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6, delay: 0.5 }}
        >
          <Link to="/login">
            <Button className="bg-emerald-600 hover:bg-emerald-700 text-white px-8 py-3 text-base rounded-xl shadow-lg shadow-emerald-200 hover:shadow-emerald-300 transition-all inline-flex items-center gap-2">
              Go to Admin Panel
              <ArrowRight className="h-4 w-4" />
            </Button>
          </Link>
        </motion.div>
      </div>

      {/* Stats Section */}
      <motion.div
        className="border-t border-gray-100 bg-white"
        initial={{ opacity: 0, y: 40 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.7, delay: 0.7 }}
      >
        <div className="max-w-5xl mx-auto px-6 py-12">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-8">
            {stats.map((stat, index) => (
              <motion.div
                key={stat.label}
                className="text-center"
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.8 + index * 0.1 }}
              >
                <p className="text-3xl md:text-4xl font-bold text-emerald-600 mb-1">
                  {stat.value}
                </p>
                <p className="text-sm text-gray-500">{stat.label}</p>
              </motion.div>
            ))}
          </div>
        </div>
      </motion.div>

      {/* Footer */}
      <motion.footer
        className="text-center py-6 text-xs text-gray-400"
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ delay: 1.2 }}
      >
        &copy; {new Date().getFullYear()} Ban Samiti &mdash; Community Forestry
        Management System
      </motion.footer>
    </div>
  );
}
