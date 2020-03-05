// Source: metrics/metrics.go

// Package mock_metrics is a generated GoMock package.
package mock_metrics

import (
	metrics "github.com/supplyon/gremcos/metrics"
)

// StubCounter is a mock of Counter interface
type StubCounter struct {
}

// NewStubCounter creates a new mock instance
func NewStubCounter() *StubCounter {
	stub := &StubCounter{}
	return stub
}

// Inc mocks base method
func (m *StubCounter) Inc() {
}

// Add mocks base method
func (m *StubCounter) Add(arg0 float64) {
}

// StubGauge is a mock of Gauge interface
type StubGauge struct {
}

// NewStubGauge creates a new mock instance
func NewStubGauge() *StubGauge {
	stub := &StubGauge{}
	return stub
}

// Set mocks base method
func (m *StubGauge) Set(arg0 float64) {
}

// Add mocks base method
func (m *StubGauge) Add(arg0 float64) {
}

// StubGaugeVec is a mock of GaugeVec interface
type StubGaugeVec struct {
}

// NewStubGaugeVec creates a new mock instance
func NewStubGaugeVec() *StubGaugeVec {
	stub := &StubGaugeVec{}
	return stub
}

var nopGauge = NewStubGauge()

// WithLabelValues mocks base method
func (m *StubGaugeVec) WithLabelValues(lvs ...string) metrics.Gauge {
	return nopGauge
}

// StubCounterVec is a mock of CounterVec interface
type StubCounterVec struct {
}

// NewStubCounterVec creates a new mock instance
func NewStubCounterVec() *StubCounterVec {
	stub := &StubCounterVec{}
	return stub
}

var nopCounter = NewStubCounter()

// WithLabelValues mocks base method
func (m *StubCounterVec) WithLabelValues(lvs ...string) metrics.Counter {
	return nopCounter
}

// StubHistogram is a mock of Histogram interface
type StubHistogram struct {
}

// NewStubHistogram creates a new mock instance
func NewStubHistogram() *StubHistogram {
	stub := &StubHistogram{}
	return stub
}

// Observe mocks base method
func (m *StubHistogram) Observe(arg0 float64) {
}
