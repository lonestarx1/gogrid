"use client";

import { motion } from "framer-motion";

const patterns = [
  {
    name: "Single Agent",
    icon: "●",
    description:
      "A well-scoped agent with a small number of tools. The recommended starting point for any GoGrid project.",
  },
  {
    name: "Team (Chat Room)",
    icon: "●●●",
    description:
      "Multiple domain experts collaborating in real-time — concurrent execution, debate, and consensus via pub/sub messaging with shared memory. Optional coordinator agent synthesizes the final decision.",
  },
  {
    name: "Pipeline (Linear)",
    icon: "● → ● → ●",
    description:
      "Sequential handoff between specialists. Each agent completes its work, yields state to the next, and terminates cleanly.",
  },
  {
    name: "Graph",
    icon: "● ⇄ ●",
    description:
      "Like a pipeline with loops (re-do, clarify) and parallel branches that merge. Bounded agents with visible data flow.",
  },
  {
    name: "Dynamic Orchestration",
    icon: "● → ✱",
    description:
      "Agents spawn child agents, teams, pipelines, or graphs at runtime. Unlimited scaling, minimal assumptions.",
  },
];

export default function Architecture() {
  return (
    <section className="relative py-32 px-6 dot-grid">
      <div className="absolute inset-0 bg-gradient-to-b from-bg via-transparent to-bg" />

      <div className="relative z-10 max-w-5xl mx-auto">
        <motion.div
          className="text-center mb-16"
          initial={{ opacity: 0, y: 30 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, margin: "-100px" }}
          transition={{ duration: 0.7 }}
        >
          <h2 className="font-mono text-3xl md:text-5xl font-bold text-white mb-4">
            5 Orchestration{" "}
            <span className="text-accent">Patterns</span>
          </h2>
          <p className="text-text-muted text-lg max-w-2xl mx-auto">
            Different problems demand different architectures. GoGrid supports
            all five — and they compose.
          </p>
        </motion.div>

        <div className="space-y-4">
          {patterns.map((pattern, i) => (
            <motion.div
              key={pattern.name}
              className="border border-border rounded-lg p-6 bg-bg-card hover:border-accent/30 transition-all flex flex-col md:flex-row md:items-center gap-4"
              initial={{ opacity: 0, x: -30 }}
              whileInView={{ opacity: 1, x: 0 }}
              viewport={{ once: true, margin: "-50px" }}
              transition={{ duration: 0.5, delay: i * 0.1 }}
            >
              <div className="font-mono text-accent text-2xl md:text-3xl w-32 shrink-0 text-center">
                {pattern.icon}
              </div>
              <div>
                <h3 className="font-mono text-white text-lg font-semibold mb-1">
                  {pattern.name}
                </h3>
                <p className="text-text-muted text-sm leading-relaxed">
                  {pattern.description}
                </p>
              </div>
            </motion.div>
          ))}
        </div>

        <motion.div
          className="mt-12 text-center"
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6, delay: 0.5 }}
        >
          <p className="font-mono text-accent text-sm border border-accent/30 inline-block px-6 py-3 rounded-full">
            All patterns are composable. A graph node can contain a team. A team
            member can spawn a pipeline.
          </p>
        </motion.div>
      </div>
    </section>
  );
}
