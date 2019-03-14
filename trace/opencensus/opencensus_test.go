package opencensus

import (
	"testing"

	"github.com/iofog/iofog-kubelet/trace"
)

func TestTracerImplementsTracer(t *testing.T) {
	// ensure that Adapter implements trace.Tracer
	if tt := trace.Tracer(Adapter{}); tt == nil {
	}
}
