"use client";

import { motion } from "framer-motion";

const rows = [
  {
    requirement: "50,000+ concurrent agent workflows",
    answer: "Goroutines — lightweight, true parallelism, no GIL",
  },
  {
    requirement: "Network-heavy I/O (LLM calls, APIs, webhooks)",
    answer: "Built for large-scale networked services",
  },
  {
    requirement: "Predictable production behavior",
    answer: "Predictable memory, CPU, excellent profiling",
  },
  {
    requirement: "Simple deployment",
    answer: "Single static binary, tiny containers, fast startup",
  },
  {
    requirement: "Minimal dependencies",
    answer: "Rich standard library (HTTP, JSON, context, testing)",
  },
  {
    requirement: "Operational simplicity",
    answer: "Easy cross-compilation, DevOps teams love Go",
  },
];

export default function WhyGo() {
  return (
    <section className="py-32 px-6">
      <div className="max-w-5xl mx-auto">
        <motion.div
          className="text-center mb-16"
          initial={{ opacity: 0, y: 30 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, margin: "-100px" }}
          transition={{ duration: 0.7 }}
        >
          <h2 className="font-mono text-3xl md:text-5xl font-bold text-white mb-4">
            Why <span className="text-accent">Go</span>?
          </h2>
          <p className="text-text-muted text-lg max-w-2xl mx-auto">
            AI agents are infrastructure, not scripts. Infrastructure demands a
            language built for infrastructure.
          </p>
        </motion.div>

        {/* Terminal-style table */}
        <motion.div
          className="border border-border rounded-lg overflow-hidden bg-bg-card"
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, margin: "-100px" }}
          transition={{ duration: 0.6 }}
        >
          {/* Terminal title bar */}
          <div className="flex items-center gap-2 px-4 py-3 border-b border-border bg-bg-card-hover">
            <span className="w-3 h-3 rounded-full bg-red-500/80" />
            <span className="w-3 h-3 rounded-full bg-yellow-500/80" />
            <span className="w-3 h-3 rounded-full bg-green-500/80" />
            <span className="font-mono text-text-muted text-xs ml-2">
              go_vs_alternatives.md
            </span>
          </div>

          {/* Table header */}
          <div className="grid grid-cols-2 border-b border-border">
            <div className="px-6 py-3 font-mono text-accent text-xs tracking-wider uppercase">
              Requirement
            </div>
            <div className="px-6 py-3 font-mono text-accent text-xs tracking-wider uppercase">
              Go&apos;s Answer
            </div>
          </div>

          {/* Table rows */}
          {rows.map((row, i) => (
            <motion.div
              key={row.requirement}
              className={`grid grid-cols-2 ${
                i < rows.length - 1 ? "border-b border-border" : ""
              } hover:bg-bg-card-hover transition-colors`}
              initial={{ opacity: 0 }}
              whileInView={{ opacity: 1 }}
              viewport={{ once: true }}
              transition={{ duration: 0.4, delay: i * 0.08 }}
            >
              <div className="px-6 py-4 font-mono text-white text-sm">
                {row.requirement}
              </div>
              <div className="px-6 py-4 text-text-muted text-sm">
                {row.answer}
              </div>
            </motion.div>
          ))}
        </motion.div>

        <motion.p
          className="text-text-muted text-center mt-8 text-sm"
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6, delay: 0.4 }}
        >
          Kubernetes. Docker. Prometheus. Terraform. The most critical
          infrastructure of the modern internet runs on Go.
        </motion.p>
      </div>
    </section>
  );
}
