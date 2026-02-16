package metrics

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
)

// DefaultBuckets are the default histogram bucket boundaries (in seconds).
var DefaultBuckets = []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}

// Registry holds a collection of metrics and exports them in
// Prometheus exposition format.
type Registry struct {
	mu         sync.RWMutex
	counters   map[string]*Counter
	gauges     map[string]*Gauge
	histograms map[string]*Histogram
}

// NewRegistry creates an empty metrics registry.
func NewRegistry() *Registry {
	return &Registry{
		counters:   make(map[string]*Counter),
		gauges:     make(map[string]*Gauge),
		histograms: make(map[string]*Histogram),
	}
}

// Counter returns the named counter, creating it if necessary.
func (r *Registry) Counter(name, help string) *Counter {
	r.mu.Lock()
	defer r.mu.Unlock()
	if c, ok := r.counters[name]; ok {
		return c
	}
	c := &Counter{name: name, help: help, values: make(map[string]*labeledValue)}
	r.counters[name] = c
	return c
}

// Gauge returns the named gauge, creating it if necessary.
func (r *Registry) Gauge(name, help string) *Gauge {
	r.mu.Lock()
	defer r.mu.Unlock()
	if g, ok := r.gauges[name]; ok {
		return g
	}
	g := &Gauge{name: name, help: help, values: make(map[string]*labeledValue)}
	r.gauges[name] = g
	return g
}

// Histogram returns the named histogram, creating it if necessary.
// If no buckets are provided, DefaultBuckets are used.
func (r *Registry) Histogram(name, help string, buckets ...float64) *Histogram {
	r.mu.Lock()
	defer r.mu.Unlock()
	if h, ok := r.histograms[name]; ok {
		return h
	}
	if len(buckets) == 0 {
		buckets = make([]float64, len(DefaultBuckets))
		copy(buckets, DefaultBuckets)
	}
	sort.Float64s(buckets)
	h := &Histogram{name: name, help: help, buckets: buckets, values: make(map[string]*histValue)}
	r.histograms[name] = h
	return h
}

// Export returns all metrics in Prometheus exposition format.
func (r *Registry) Export() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var b strings.Builder

	for _, name := range sortedMapKeys(r.counters) {
		c := r.counters[name]
		c.writeTo(&b)
	}
	for _, name := range sortedMapKeys(r.gauges) {
		g := r.gauges[name]
		g.writeTo(&b)
	}
	for _, name := range sortedMapKeys(r.histograms) {
		h := r.histograms[name]
		h.writeTo(&b)
	}

	return b.String()
}

// --- Counter ---

type labeledValue struct {
	labels map[string]string
	value  float64
}

// Counter is a monotonically increasing metric.
type Counter struct {
	mu     sync.Mutex
	name   string
	help   string
	values map[string]*labeledValue
}

// Inc increments the counter by 1.
func (c *Counter) Inc(labels map[string]string) { c.Add(1, labels) }

// Add adds the given value to the counter.
func (c *Counter) Add(v float64, labels map[string]string) {
	key := labelKey(labels)
	c.mu.Lock()
	defer c.mu.Unlock()
	lv, ok := c.values[key]
	if !ok {
		lv = &labeledValue{labels: copyLabels(labels)}
		c.values[key] = lv
	}
	lv.value += v
}

// Value returns the current counter value for the given labels.
func (c *Counter) Value(labels map[string]string) float64 {
	key := labelKey(labels)
	c.mu.Lock()
	defer c.mu.Unlock()
	if lv, ok := c.values[key]; ok {
		return lv.value
	}
	return 0
}

func (c *Counter) writeTo(b *strings.Builder) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.values) == 0 {
		return
	}
	if c.help != "" {
		fmt.Fprintf(b, "# HELP %s %s\n", c.name, c.help)
	}
	fmt.Fprintf(b, "# TYPE %s counter\n", c.name)
	for _, key := range sortedMapKeys(c.values) {
		lv := c.values[key]
		fmt.Fprintf(b, "%s%s %s\n", c.name, formatLabels(lv.labels), formatFloat(lv.value))
	}
}

// --- Gauge ---

// Gauge is a metric that can go up and down.
type Gauge struct {
	mu     sync.Mutex
	name   string
	help   string
	values map[string]*labeledValue
}

// Set sets the gauge to the given value.
func (g *Gauge) Set(v float64, labels map[string]string) {
	key := labelKey(labels)
	g.mu.Lock()
	defer g.mu.Unlock()
	lv, ok := g.values[key]
	if !ok {
		lv = &labeledValue{labels: copyLabels(labels)}
		g.values[key] = lv
	}
	lv.value = v
}

// Inc increments the gauge by 1.
func (g *Gauge) Inc(labels map[string]string) { g.Add(1, labels) }

// Dec decrements the gauge by 1.
func (g *Gauge) Dec(labels map[string]string) { g.Add(-1, labels) }

// Add adds the given value to the gauge.
func (g *Gauge) Add(v float64, labels map[string]string) {
	key := labelKey(labels)
	g.mu.Lock()
	defer g.mu.Unlock()
	lv, ok := g.values[key]
	if !ok {
		lv = &labeledValue{labels: copyLabels(labels)}
		g.values[key] = lv
	}
	lv.value += v
}

// Value returns the current gauge value for the given labels.
func (g *Gauge) Value(labels map[string]string) float64 {
	key := labelKey(labels)
	g.mu.Lock()
	defer g.mu.Unlock()
	if lv, ok := g.values[key]; ok {
		return lv.value
	}
	return 0
}

func (g *Gauge) writeTo(b *strings.Builder) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if len(g.values) == 0 {
		return
	}
	if g.help != "" {
		fmt.Fprintf(b, "# HELP %s %s\n", g.name, g.help)
	}
	fmt.Fprintf(b, "# TYPE %s gauge\n", g.name)
	for _, key := range sortedMapKeys(g.values) {
		lv := g.values[key]
		fmt.Fprintf(b, "%s%s %s\n", g.name, formatLabels(lv.labels), formatFloat(lv.value))
	}
}

// --- Histogram ---

type histValue struct {
	labels       map[string]string
	bucketCounts []uint64
	count        uint64
	sum          float64
}

// Histogram tracks the distribution of observed values.
type Histogram struct {
	mu      sync.Mutex
	name    string
	help    string
	buckets []float64
	values  map[string]*histValue
}

// Observe records a value in the histogram.
func (h *Histogram) Observe(v float64, labels map[string]string) {
	key := labelKey(labels)
	h.mu.Lock()
	defer h.mu.Unlock()
	hv, ok := h.values[key]
	if !ok {
		hv = &histValue{
			labels:       copyLabels(labels),
			bucketCounts: make([]uint64, len(h.buckets)),
		}
		h.values[key] = hv
	}
	for i, bound := range h.buckets {
		if v <= bound {
			hv.bucketCounts[i]++
		}
	}
	hv.count++
	hv.sum += v
}

// Count returns the total observation count for the given labels.
func (h *Histogram) Count(labels map[string]string) uint64 {
	key := labelKey(labels)
	h.mu.Lock()
	defer h.mu.Unlock()
	if hv, ok := h.values[key]; ok {
		return hv.count
	}
	return 0
}

// Sum returns the sum of observed values for the given labels.
func (h *Histogram) Sum(labels map[string]string) float64 {
	key := labelKey(labels)
	h.mu.Lock()
	defer h.mu.Unlock()
	if hv, ok := h.values[key]; ok {
		return hv.sum
	}
	return 0
}

func (h *Histogram) writeTo(b *strings.Builder) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if len(h.values) == 0 {
		return
	}
	if h.help != "" {
		fmt.Fprintf(b, "# HELP %s %s\n", h.name, h.help)
	}
	fmt.Fprintf(b, "# TYPE %s histogram\n", h.name)
	for _, key := range sortedMapKeys(h.values) {
		hv := h.values[key]
		for i, bound := range h.buckets {
			bl := copyLabels(hv.labels)
			bl["le"] = formatFloat(bound)
			fmt.Fprintf(b, "%s_bucket%s %d\n", h.name, formatLabels(bl), hv.bucketCounts[i])
		}
		infLabels := copyLabels(hv.labels)
		infLabels["le"] = "+Inf"
		fmt.Fprintf(b, "%s_bucket%s %d\n", h.name, formatLabels(infLabels), hv.count)
		fmt.Fprintf(b, "%s_sum%s %s\n", h.name, formatLabels(hv.labels), formatFloat(hv.sum))
		fmt.Fprintf(b, "%s_count%s %d\n", h.name, formatLabels(hv.labels), hv.count)
	}
}

// --- helpers ---

func labelKey(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for i, k := range keys {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(labels[k])
	}
	return b.String()
}

func formatLabels(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	b.WriteByte('{')
	for i, k := range keys {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(k)
		b.WriteString(`="`)
		b.WriteString(labels[k])
		b.WriteByte('"')
	}
	b.WriteByte('}')
	return b.String()
}

func formatFloat(v float64) string {
	if v == float64(int64(v)) && !math.IsInf(v, 0) {
		return fmt.Sprintf("%g", v)
	}
	return fmt.Sprintf("%g", v)
}

func copyLabels(m map[string]string) map[string]string {
	if len(m) == 0 {
		return nil
	}
	cp := make(map[string]string, len(m))
	for k, v := range m {
		cp[k] = v
	}
	return cp
}

func sortedMapKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
