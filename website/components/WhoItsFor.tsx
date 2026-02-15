"use client";

import { motion } from "framer-motion";

const audiences = [
  {
    label: "Infrastructure Engineers",
    description: "Building production AI systems, not prototypes",
  },
  {
    label: "Platform Teams",
    description: "Deploying multi-tenant agent workloads at scale",
  },
  {
    label: "Companies",
    description: "That need agents they can monitor, audit, secure, and trust",
  },
  {
    label: "Frustrated Developers",
    description:
      "Tired of framework churn, abstraction bloat, and the demo-to-production gap",
  },
];

export default function WhoItsFor() {
  return (
    <section className="relative py-32 px-6 dot-grid">
      <div className="absolute inset-0 bg-gradient-to-b from-bg via-transparent to-bg" />

      <div className="relative z-10 max-w-4xl mx-auto">
        <motion.div
          className="text-center mb-16"
          initial={{ opacity: 0, y: 30 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, margin: "-100px" }}
          transition={{ duration: 0.7 }}
        >
          <h2 className="font-mono text-3xl md:text-5xl font-bold text-white mb-4">
            Who It&apos;s <span className="text-accent">For</span>
          </h2>
        </motion.div>

        <div className="space-y-4 mb-12">
          {audiences.map((audience, i) => (
            <motion.div
              key={audience.label}
              className="flex items-start gap-4 border border-border rounded-lg p-5 bg-bg-card hover:border-accent/30 transition-colors"
              initial={{ opacity: 0, x: -20 }}
              whileInView={{ opacity: 1, x: 0 }}
              viewport={{ once: true, margin: "-50px" }}
              transition={{ duration: 0.5, delay: i * 0.1 }}
            >
              <span className="text-accent font-mono text-lg mt-0.5">
                &gt;
              </span>
              <div>
                <p className="font-mono text-white font-semibold">
                  {audience.label}
                </p>
                <p className="text-text-muted text-sm mt-1">
                  {audience.description}
                </p>
              </div>
            </motion.div>
          ))}
        </div>

        <motion.div
          className="border border-accent/20 rounded-lg p-6 bg-bg-card text-center"
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6, delay: 0.4 }}
        >
          <p className="text-text-muted text-sm">
            If you&apos;re building a weekend hackathon project, there are
            simpler tools.
            <br />
            <span className="text-white font-medium">
              If you&apos;re building something that needs to run in production,
              at scale, for real users — GoGrid is for you.
            </span>
          </p>
        </motion.div>
      </div>
    </section>
  );
}
