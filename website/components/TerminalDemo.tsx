"use client";

import { motion, useInView } from "framer-motion";
import { useRef, useState, useEffect } from "react";

interface Line {
  type: "prompt" | "output" | "info" | "success" | "error" | "accent" | "dim" | "blank";
  text: string;
  delay: number; // ms after previous line
}

const demoLines: Line[] = [
  // 1. Scaffold
  { type: "prompt", text: "$ gogrid init --template team research-bot", delay: 0 },
  { type: "success", text: "  Created research-bot/gogrid.yaml", delay: 600 },
  { type: "success", text: "  Created research-bot/main.go", delay: 150 },
  { type: "success", text: "  Created research-bot/Makefile", delay: 150 },
  { type: "dim", text: "  Project ready. Run: cd research-bot && gogrid run", delay: 300 },
  { type: "blank", text: "", delay: 600 },

  // 2. Show YAML config
  { type: "prompt", text: "$ cat gogrid.yaml", delay: 0 },
  { type: "blank", text: "", delay: 400 },
  { type: "dim", text: '  version: "1"', delay: 100 },
  { type: "blank", text: "", delay: 60 },
  { type: "dim", text: "  agents:", delay: 80 },
  { type: "info", text: "    researcher:", delay: 80 },
  { type: "output", text: "      model: claude-sonnet-4-5", delay: 60 },
  { type: "output", text: "      provider: anthropic", delay: 60 },
  { type: "output", text: "      instructions: |", delay: 60 },
  { type: "accent", text: "        You are a research assistant. Provide thorough", delay: 60 },
  { type: "accent", text: "        analysis with key findings and evidence.", delay: 60 },
  { type: "output", text: "      config:", delay: 60 },
  { type: "output", text: "        max_turns: 10", delay: 60 },
  { type: "output", text: "        cost_budget: 0.50", delay: 60 },
  { type: "output", text: "        timeout: 2m", delay: 60 },
  { type: "blank", text: "", delay: 80 },
  { type: "info", text: "    reviewer:", delay: 80 },
  { type: "output", text: "      model: gpt-4o", delay: 60 },
  { type: "output", text: "      provider: openai", delay: 60 },
  { type: "output", text: "      instructions: |", delay: 60 },
  { type: "accent", text: "        You are a critical reviewer. Evaluate research", delay: 60 },
  { type: "accent", text: "        for accuracy, gaps, and logical consistency.", delay: 60 },
  { type: "output", text: "      config:", delay: 60 },
  { type: "output", text: "        max_turns: 5", delay: 60 },
  { type: "output", text: "        cost_budget: 0.25", delay: 60 },
  { type: "blank", text: "", delay: 700 },

  // 3. List agents
  { type: "prompt", text: "$ gogrid list", delay: 0 },
  { type: "info", text: "  NAME          PROVIDER    MODEL", delay: 500 },
  { type: "output", text: "  researcher    anthropic   claude-sonnet-4-5", delay: 100 },
  { type: "output", text: "  reviewer      openai      gpt-4o", delay: 100 },
  { type: "blank", text: "", delay: 600 },

  // 4. Run agent
  { type: "prompt", text: "$ gogrid run researcher -input \"Analyze Go's concurrency model\"", delay: 0 },
  { type: "blank", text: "", delay: 800 },
  { type: "dim", text: "  run: 019479a3c4e8  agent: researcher  model: claude-sonnet-4-5", delay: 200 },
  { type: "blank", text: "", delay: 200 },
  { type: "accent", text: "  [1/4] Loading memory...", delay: 300 },
  { type: "accent", text: "  [2/4] Calling LLM (prompt: 156 tokens)...", delay: 600 },
  { type: "accent", text: "  [3/4] Executing tool: web_search(\"Go concurrency goroutines\")...", delay: 1200 },
  { type: "accent", text: "  [4/4] Calling LLM (prompt: 892 tokens)...", delay: 800 },
  { type: "blank", text: "", delay: 400 },
  { type: "output", text: "  Go's concurrency model is built on goroutines and channels,", delay: 100 },
  { type: "output", text: "  inspired by Hoare's CSP. Goroutines are lightweight (~2KB)", delay: 80 },
  { type: "output", text: "  multiplexed onto OS threads. Channels provide typed,", delay: 80 },
  { type: "output", text: "  synchronized communication, eliminating shared-memory races.", delay: 80 },
  { type: "blank", text: "", delay: 400 },
  { type: "success", text: "  Done in 4.2s  |  2 LLM calls  |  1,048 tokens  |  $0.003", delay: 200 },
  { type: "blank", text: "", delay: 800 },

  // 5. Trace
  { type: "prompt", text: "$ gogrid trace 019479a3c4e8", delay: 0 },
  { type: "blank", text: "", delay: 600 },
  { type: "info", text: "  Agent: researcher | Model: claude-sonnet-4-5 | 4.2s", delay: 200 },
  { type: "blank", text: "", delay: 200 },
  { type: "output", text: "  agent.run (4.2s)", delay: 100 },
  { type: "dim", text: "  \u251C\u2500\u2500 memory.load (1ms)", delay: 80 },
  { type: "accent", text: "  \u251C\u2500\u2500 llm.complete (2.1s) [prompt: 156, completion: 89]", delay: 80 },
  { type: "accent", text: "  \u251C\u2500\u2500 tool.execute (1.8s) [\"web_search\"]", delay: 80 },
  { type: "accent", text: "  \u251C\u2500\u2500 llm.complete (0.3s) [prompt: 892, completion: 67]", delay: 80 },
  { type: "dim", text: "  \u2514\u2500\u2500 memory.save (2ms)", delay: 80 },
  { type: "blank", text: "", delay: 800 },

  // 6. Cost
  { type: "prompt", text: "$ gogrid cost", delay: 0 },
  { type: "blank", text: "", delay: 500 },
  { type: "info", text: "  RUN ID           AGENT       MODEL              COST", delay: 200 },
  { type: "output", text: "  019479a3c4e8     researcher  claude-sonnet-4-5  $0.0033", delay: 100 },
  { type: "dim", text: "  \u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500", delay: 80 },
  { type: "success", text: "  TOTAL            1 run                          $0.0033", delay: 100 },
];

function useTypedLines(lines: Line[], active: boolean, loopPause = 3000) {
  const [visible, setVisible] = useState<number>(0);

  useEffect(() => {
    if (!active) return;
    setVisible(0);

    let timeout: ReturnType<typeof setTimeout>;
    let cancelled = false;
    let current = 0;

    function showNext() {
      if (cancelled) return;
      if (current >= lines.length) {
        // Pause at the end, then restart
        timeout = setTimeout(() => {
          if (cancelled) return;
          current = 0;
          setVisible(0);
          showNext();
        }, loopPause);
        return;
      }
      const delay = current === 0 ? 400 : lines[current].delay;
      timeout = setTimeout(() => {
        if (cancelled) return;
        current++;
        setVisible(current);
        showNext();
      }, delay);
    }

    showNext();
    return () => {
      cancelled = true;
      clearTimeout(timeout);
    };
  }, [active, lines, loopPause]);

  return visible;
}

function colorForType(type: Line["type"]) {
  switch (type) {
    case "prompt":
      return "text-white";
    case "output":
      return "text-[#e0e0e0]";
    case "info":
      return "text-[#60a5fa]";
    case "success":
      return "text-[#00ff88]";
    case "error":
      return "text-[#f87171]";
    case "accent":
      return "text-[#a78bfa]";
    case "dim":
      return "text-[#555555]";
    case "blank":
      return "";
  }
}

export default function TerminalDemo() {
  const ref = useRef<HTMLDivElement>(null);
  const isInView = useInView(ref, { once: true, margin: "-100px" });
  const [started, setStarted] = useState(false);
  const visibleCount = useTypedLines(demoLines, started);

  const terminalRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (isInView && !started) setStarted(true);
  }, [isInView, started]);

  // Auto-scroll terminal
  useEffect(() => {
    if (terminalRef.current) {
      terminalRef.current.scrollTop = terminalRef.current.scrollHeight;
    }
  }, [visibleCount]);

  return (
    <section ref={ref} className="relative py-32 px-6">
      <div className="max-w-5xl mx-auto">
        <motion.div
          className="text-center mb-12"
          initial={{ opacity: 0, y: 30 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, margin: "-100px" }}
          transition={{ duration: 0.7 }}
        >
          <h2 className="font-mono text-3xl md:text-5xl font-bold text-white mb-4">
            See It in <span className="text-accent">Action</span>
          </h2>
          <p className="text-text-muted text-lg max-w-2xl mx-auto">
            Define agents in YAML. Run, trace, and track costs — one CLI, zero friction.
          </p>
        </motion.div>

        <motion.div
          className="relative rounded-xl overflow-hidden border border-border shadow-2xl shadow-black/50"
          initial={{ opacity: 0, y: 40 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, margin: "-80px" }}
          transition={{ duration: 0.8 }}
        >
          {/* Terminal title bar */}
          <div className="flex items-center gap-2 px-4 py-3 bg-[#161b22] border-b border-border">
            <div className="flex gap-2">
              <span className="w-3 h-3 rounded-full bg-[#ff5f56]" />
              <span className="w-3 h-3 rounded-full bg-[#ffbd2e]" />
              <span className="w-3 h-3 rounded-full bg-[#27c93f]" />
            </div>
            <span className="flex-1 text-center font-mono text-xs text-[#555]">
              gogrid &mdash; research-bot
            </span>
          </div>

          {/* Terminal body */}
          <div
            ref={terminalRef}
            className="bg-[#0d1117] p-5 font-mono text-sm leading-6 h-[480px] overflow-y-auto scroll-smooth"
          >
            {demoLines.slice(0, visibleCount).map((line, i) => {
              if (line.type === "blank") {
                return <div key={i} className="h-3" />;
              }

              const isPrompt = line.type === "prompt";

              return (
                <motion.div
                  key={i}
                  initial={{ opacity: 0, x: isPrompt ? -8 : 0 }}
                  animate={{ opacity: 1, x: 0 }}
                  transition={{ duration: isPrompt ? 0.3 : 0.15 }}
                  className={`whitespace-pre ${colorForType(line.type)} ${
                    isPrompt ? "font-semibold" : ""
                  }`}
                >
                  {line.text}
                </motion.div>
              );
            })}

            {/* Blinking cursor */}
            {visibleCount < demoLines.length && started && (
              <span className="inline-block w-2 h-4 bg-accent animate-pulse" />
            )}
            {visibleCount >= demoLines.length && (
              <div className="mt-1">
                <span className="text-white font-semibold">$ </span>
                <span className="inline-block w-2 h-4 bg-accent animate-pulse align-middle" />
              </div>
            )}
          </div>

          {/* Subtle gradient at top when scrolled */}
          <div className="absolute top-[44px] left-0 right-0 h-8 bg-gradient-to-b from-[#0d1117] to-transparent pointer-events-none z-10 opacity-0" />
        </motion.div>

        {/* Feature callouts below terminal */}
        <motion.div
          className="grid grid-cols-2 md:grid-cols-4 gap-4 mt-8"
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6, delay: 0.3 }}
        >
          {[
            { label: "Scaffold a project", desc: "gogrid init" },
            { label: "Configure in YAML", desc: "gogrid.yaml" },
            { label: "Run agents", desc: "gogrid run" },
            { label: "Trace & cost", desc: "gogrid trace" },
          ].map((item) => (
            <div
              key={item.label}
              className="border border-border rounded-lg p-4 bg-bg-card text-center"
            >
              <p className="font-mono text-accent text-xs mb-1">{item.desc}</p>
              <p className="text-text-muted text-xs">{item.label}</p>
            </div>
          ))}
        </motion.div>
      </div>
    </section>
  );
}
