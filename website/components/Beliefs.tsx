"use client";

import { motion } from "framer-motion";

const beliefs = [
  {
    number: "01",
    title: "Clarity Over Cleverness",
    description:
      "Understanding and maintaining code is the bottleneck, not writing it. GoGrid follows the Go philosophy: explicit over implicit, readable over concise.",
  },
  {
    number: "02",
    title: "Production Is the Point",
    description:
      "Monitoring, cost tracking, error recovery, security, backward compatibility — built in from day one. Not bolted on after the demo works.",
  },
  {
    number: "03",
    title: "Memory Is First-Class",
    description:
      "Memory is as fundamental to an agent as a file system is to an OS. Shared memory pools, state ownership transfer, and monitorable storage — all primitives.",
  },
  {
    number: "04",
    title: "One Architecture Does Not Fit All",
    description:
      "Five composable patterns — single, team, pipeline, graph, dynamic — because we refuse to force your problem into our ideology.",
  },
  {
    number: "05",
    title: "Agents Are Infrastructure",
    description:
      "At scale, agents are long-running, stateful, concurrent, networked, mission-critical. Infrastructure demands a language built for infrastructure.",
  },
  {
    number: "06",
    title: "No Lock-In. Ever.",
    description:
      "Swapping models is a config change, not a rewrite. Open source. We will never paywall core features to push a managed platform.",
  },
  {
    number: "07",
    title: "Stability Is a Feature",
    description:
      "Backward-compatible, gradual updates. Your agents won't break when you upgrade. We version our APIs. We deprecate before we remove.",
  },
];

export default function Beliefs() {
  return (
    <section className="py-32 px-6">
      <div className="max-w-6xl mx-auto">
        <motion.div
          className="text-center mb-16"
          initial={{ opacity: 0, y: 30 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, margin: "-100px" }}
          transition={{ duration: 0.7 }}
        >
          <h2 className="font-mono text-3xl md:text-5xl font-bold text-white mb-4">
            What We <span className="text-accent">Believe</span>
          </h2>
          <p className="text-text-muted text-lg max-w-2xl mx-auto">
            Seven convictions that shape every decision in GoGrid.
          </p>
        </motion.div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {beliefs.map((belief, i) => (
            <motion.div
              key={belief.number}
              className={`border border-border rounded-lg p-6 bg-bg-card hover:bg-bg-card-hover hover:border-accent/30 transition-all ${
                i === 6 ? "md:col-span-2 lg:col-span-1" : ""
              }`}
              initial={{ opacity: 0, y: 30 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true, margin: "-50px" }}
              transition={{ duration: 0.5, delay: i * 0.08 }}
            >
              <span className="font-mono text-accent/50 text-xs mb-3 block">
                {belief.number}
              </span>
              <h3 className="font-mono text-white text-lg font-semibold mb-3">
                {belief.title}
              </h3>
              <p className="text-text-muted text-sm leading-relaxed">
                {belief.description}
              </p>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}
