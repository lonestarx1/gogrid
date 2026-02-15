"use client";

import { motion } from "framer-motion";

const frameworks = [
  {
    name: "LangChain",
    issue: "Ecosystem is vast, but drowns you in abstraction layers",
  },
  {
    name: "CrewAI",
    issue: "Great mental model, but falls apart in production",
  },
  {
    name: "AutoGen",
    issue: "Pioneered multi-agent — then got rewritten twice",
  },
  {
    name: "Most frameworks",
    issue: "Python-only, prototype-first, production as an afterthought",
  },
];

export default function Problem() {
  return (
    <section id="manifesto" className="relative py-32 dot-grid">
      <div className="absolute inset-0 bg-gradient-to-b from-bg via-transparent to-bg" />

      <div className="relative z-10 max-w-4xl mx-auto px-6">
        <motion.div
          initial={{ opacity: 0, y: 40 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, margin: "-100px" }}
          transition={{ duration: 0.7 }}
        >
          <h2 className="font-mono text-3xl md:text-5xl font-bold text-white mb-6">
            The Problem Is Not Intelligence.
            <br />
            <span className="text-accent">It&apos;s Infrastructure.</span>
          </h2>

          <p className="text-text-muted text-lg md:text-xl leading-relaxed mb-12 max-w-3xl">
            The world doesn&apos;t need another way to call an LLM. What it
            needs is a way to run thousands of AI agents in production —
            reliably, securely, affordably, and at scale.
          </p>
        </motion.div>

        {/* Stat callout */}
        <motion.div
          className="border border-border rounded-lg p-8 mb-12 bg-bg-card glow-box"
          initial={{ opacity: 0, scale: 0.95 }}
          whileInView={{ opacity: 1, scale: 1 }}
          viewport={{ once: true, margin: "-100px" }}
          transition={{ duration: 0.6, delay: 0.2 }}
        >
          <p className="font-mono text-4xl md:text-6xl font-bold text-accent mb-3">
            80–90%
          </p>
          <p className="text-text-muted text-lg">
            of AI agent projects never leave the pilot phase.{" "}
            <span className="text-text-muted/60 text-sm">(RAND, 2025)</span>
          </p>
          <p className="text-text mt-4">
            This isn&apos;t a failure of AI — it&apos;s a failure of
            infrastructure.
          </p>
        </motion.div>

        {/* Framework critiques */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {frameworks.map((fw, i) => (
            <motion.div
              key={fw.name}
              className="border border-border rounded-lg p-5 bg-bg-card hover:border-accent/30 transition-colors"
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true, margin: "-50px" }}
              transition={{ duration: 0.5, delay: i * 0.1 }}
            >
              <p className="font-mono text-accent text-sm mb-2">{fw.name}</p>
              <p className="text-text-muted text-sm">{fw.issue}</p>
            </motion.div>
          ))}
        </div>

        <motion.p
          className="font-mono text-accent text-lg mt-12 text-center"
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6, delay: 0.5 }}
        >
          GoGrid is here to fix that.
        </motion.p>
      </div>
    </section>
  );
}
