package platform

import (
	"context"
	"fmt"

	"github.com/go-stack/stack"
	"github.com/karlmutch/errors"

	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"

	"github.com/honeycombio/opencensus-exporter/honeycomb"
)

func StartOpenCensus(ctx context.Context, apiKey string, dataset string) (err errors.Error) {

	// Register views to collect data for the OpenCensus interceptor.
	if errGo := view.Register(ocgrpc.DefaultServerViews...); errGo != nil {
		logger.Fatal(fmt.Sprint(errors.Wrap(errGo).With("stack", stack.Trace().TrimRuntime())))
	}

	if len(apiKey) != 0 {
		exporter := honeycomb.NewExporter(apiKey, dataset)
		defer func() {
			go func() {
				select {
				case <-ctx.Done():
					exporter.Close()
				}
			}()
		}()

		trace.RegisterExporter(exporter)

		trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})

		return nil
	}

	return errors.New("apiKey is missing, please supply a value for this parameter").With("stack", stack.Trace().TrimRuntime())
}
