// Package metrics provides Prometheus-compatible metrics for GoGrid.
//
// A Registry holds counters, gauges, and histograms. The Export method
// returns all metrics in Prometheus exposition format, suitable for
// scraping by Prometheus or compatible systems.
//
// The Collector wraps any trace.Tracer and automatically populates
// metrics from GoGrid trace spans — agent runs, LLM calls, tool
// executions, and cost data are tracked without manual instrumentation.
//
// Usage:
//
//	reg := metrics.NewRegistry()
//	collector := metrics.NewCollector(innerTracer, reg)
//
//	// Use collector as the tracer for agents, teams, etc.
//	a := agent.New("assistant", agent.WithTracer(collector))
//
//	// Export metrics for Prometheus
//	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
//	    w.Header().Set("Content-Type", "text/plain; version=0.0.4")
//	    fmt.Fprint(w, reg.Export())
//	})
package metrics
