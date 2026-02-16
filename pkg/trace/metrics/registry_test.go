package metrics

import (
	"strings"
	"sync"
	"testing"
)

func TestCounterInc(t *testing.T) {
	r := NewRegistry()
	c := r.Counter("test_counter", "A test counter")

	labels := map[string]string{"method": "GET"}
	c.Inc(labels)
	c.Inc(labels)

	if got := c.Value(labels); got != 2 {
		t.Errorf("Value = %f, want 2", got)
	}
}

func TestCounterAdd(t *testing.T) {
	r := NewRegistry()
	c := r.Counter("test_counter", "")

	labels := map[string]string{"status": "200"}
	c.Add(5.5, labels)
	c.Add(3.25, labels)

	if got := c.Value(labels); got != 8.75 {
		t.Errorf("Value = %f, want 8.75", got)
	}
}

func TestCounterDistinctLabels(t *testing.T) {
	r := NewRegistry()
	c := r.Counter("test_counter", "")

	c.Inc(map[string]string{"status": "200"})
	c.Inc(map[string]string{"status": "500"})
	c.Inc(map[string]string{"status": "200"})

	if got := c.Value(map[string]string{"status": "200"}); got != 2 {
		t.Errorf("status=200 Value = %f, want 2", got)
	}
	if got := c.Value(map[string]string{"status": "500"}); got != 1 {
		t.Errorf("status=500 Value = %f, want 1", got)
	}
}

func TestCounterNilLabels(t *testing.T) {
	r := NewRegistry()
	c := r.Counter("test_counter", "")
	c.Inc(nil)
	c.Inc(nil)

	if got := c.Value(nil); got != 2 {
		t.Errorf("Value = %f, want 2", got)
	}
}

func TestGaugeSetAndValue(t *testing.T) {
	r := NewRegistry()
	g := r.Gauge("test_gauge", "A test gauge")

	labels := map[string]string{"host": "a"}
	g.Set(42, labels)
	if got := g.Value(labels); got != 42 {
		t.Errorf("Value = %f, want 42", got)
	}

	g.Set(0, labels)
	if got := g.Value(labels); got != 0 {
		t.Errorf("Value = %f, want 0", got)
	}
}

func TestGaugeIncDec(t *testing.T) {
	r := NewRegistry()
	g := r.Gauge("test_gauge", "")

	labels := map[string]string{"host": "a"}
	g.Inc(labels)
	g.Inc(labels)
	g.Dec(labels)

	if got := g.Value(labels); got != 1 {
		t.Errorf("Value = %f, want 1", got)
	}
}

func TestGaugeAdd(t *testing.T) {
	r := NewRegistry()
	g := r.Gauge("test_gauge", "")

	labels := map[string]string{"host": "a"}
	g.Add(10, labels)
	g.Add(-3, labels)

	if got := g.Value(labels); got != 7 {
		t.Errorf("Value = %f, want 7", got)
	}
}

func TestHistogramObserve(t *testing.T) {
	r := NewRegistry()
	h := r.Histogram("test_hist", "A test histogram", 1, 5, 10)

	labels := map[string]string{"op": "read"}
	h.Observe(0.5, labels)
	h.Observe(3, labels)
	h.Observe(7, labels)
	h.Observe(15, labels)

	if got := h.Count(labels); got != 4 {
		t.Errorf("Count = %d, want 4", got)
	}
	if got := h.Sum(labels); got != 25.5 {
		t.Errorf("Sum = %f, want 25.5", got)
	}
}

func TestHistogramDefaultBuckets(t *testing.T) {
	r := NewRegistry()
	h := r.Histogram("test_hist", "")

	labels := map[string]string{"op": "write"}
	h.Observe(0.001, labels)
	h.Observe(1, labels)

	if got := h.Count(labels); got != 2 {
		t.Errorf("Count = %d, want 2", got)
	}
}

func TestRegistryDeduplication(t *testing.T) {
	r := NewRegistry()
	c1 := r.Counter("requests", "help1")
	c2 := r.Counter("requests", "help2")

	if c1 != c2 {
		t.Error("Counter should return the same instance for the same name")
	}

	g1 := r.Gauge("mem", "help1")
	g2 := r.Gauge("mem", "help2")
	if g1 != g2 {
		t.Error("Gauge should return the same instance for the same name")
	}

	h1 := r.Histogram("lat", "help1")
	h2 := r.Histogram("lat", "help2")
	if h1 != h2 {
		t.Error("Histogram should return the same instance for the same name")
	}
}

func TestExportCounter(t *testing.T) {
	r := NewRegistry()
	c := r.Counter("http_requests_total", "Total HTTP requests")
	c.Inc(map[string]string{"method": "GET", "status": "200"})
	c.Add(3, map[string]string{"method": "POST", "status": "201"})

	out := r.Export()

	if !strings.Contains(out, "# HELP http_requests_total Total HTTP requests") {
		t.Error("missing HELP line")
	}
	if !strings.Contains(out, "# TYPE http_requests_total counter") {
		t.Error("missing TYPE line")
	}
	if !strings.Contains(out, `http_requests_total{method="GET",status="200"} 1`) {
		t.Errorf("missing GET counter line, got:\n%s", out)
	}
	if !strings.Contains(out, `http_requests_total{method="POST",status="201"} 3`) {
		t.Errorf("missing POST counter line, got:\n%s", out)
	}
}

func TestExportGauge(t *testing.T) {
	r := NewRegistry()
	g := r.Gauge("memory_bytes", "Memory usage")
	g.Set(1024, map[string]string{"host": "a"})

	out := r.Export()

	if !strings.Contains(out, "# TYPE memory_bytes gauge") {
		t.Error("missing TYPE line")
	}
	if !strings.Contains(out, `memory_bytes{host="a"} 1024`) {
		t.Errorf("missing gauge line, got:\n%s", out)
	}
}

func TestExportHistogram(t *testing.T) {
	r := NewRegistry()
	h := r.Histogram("request_duration_seconds", "Request duration", 0.1, 0.5, 1)

	labels := map[string]string{"handler": "api"}
	h.Observe(0.05, labels)
	h.Observe(0.3, labels)
	h.Observe(0.8, labels)
	h.Observe(2.0, labels)

	out := r.Export()

	if !strings.Contains(out, "# TYPE request_duration_seconds histogram") {
		t.Error("missing TYPE line")
	}
	if !strings.Contains(out, `request_duration_seconds_bucket{handler="api",le="0.1"} 1`) {
		t.Errorf("missing bucket 0.1, got:\n%s", out)
	}
	if !strings.Contains(out, `request_duration_seconds_bucket{handler="api",le="0.5"} 2`) {
		t.Errorf("missing bucket 0.5, got:\n%s", out)
	}
	if !strings.Contains(out, `request_duration_seconds_bucket{handler="api",le="1"} 3`) {
		t.Errorf("missing bucket 1, got:\n%s", out)
	}
	if !strings.Contains(out, `request_duration_seconds_bucket{handler="api",le="+Inf"} 4`) {
		t.Errorf("missing +Inf bucket, got:\n%s", out)
	}
	if !strings.Contains(out, `request_duration_seconds_count{handler="api"} 4`) {
		t.Errorf("missing count, got:\n%s", out)
	}
}

func TestExportEmpty(t *testing.T) {
	r := NewRegistry()
	_ = r.Counter("empty_counter", "")
	_ = r.Gauge("empty_gauge", "")

	out := r.Export()
	if out != "" {
		t.Errorf("Export of empty registry should be empty, got: %q", out)
	}
}

func TestCounterConcurrency(t *testing.T) {
	r := NewRegistry()
	c := r.Counter("concurrent", "")
	labels := map[string]string{"worker": "pool"}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Inc(labels)
		}()
	}
	wg.Wait()

	if got := c.Value(labels); got != 100 {
		t.Errorf("Value = %f, want 100", got)
	}
}

func TestHistogramConcurrency(t *testing.T) {
	r := NewRegistry()
	h := r.Histogram("concurrent_hist", "", 1, 5, 10)
	labels := map[string]string{"op": "test"}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			h.Observe(2.5, labels)
		}()
	}
	wg.Wait()

	if got := h.Count(labels); got != 100 {
		t.Errorf("Count = %d, want 100", got)
	}
	if got := h.Sum(labels); got != 250 {
		t.Errorf("Sum = %f, want 250", got)
	}
}

func TestLabelKey(t *testing.T) {
	// Same labels in different order should produce the same key.
	k1 := labelKey(map[string]string{"a": "1", "b": "2"})
	k2 := labelKey(map[string]string{"b": "2", "a": "1"})
	if k1 != k2 {
		t.Errorf("labelKey not order-independent: %q vs %q", k1, k2)
	}

	// Empty labels produce empty key.
	if k := labelKey(nil); k != "" {
		t.Errorf("labelKey(nil) = %q, want empty", k)
	}
}

func TestValueMissingLabels(t *testing.T) {
	r := NewRegistry()
	c := r.Counter("test", "")

	if got := c.Value(map[string]string{"nonexistent": "true"}); got != 0 {
		t.Errorf("Value for missing labels = %f, want 0", got)
	}

	g := r.Gauge("test_g", "")
	if got := g.Value(nil); got != 0 {
		t.Errorf("Gauge Value for missing labels = %f, want 0", got)
	}

	h := r.Histogram("test_h", "")
	if got := h.Count(nil); got != 0 {
		t.Errorf("Histogram Count for missing labels = %d, want 0", got)
	}
	if got := h.Sum(nil); got != 0 {
		t.Errorf("Histogram Sum for missing labels = %f, want 0", got)
	}
}
