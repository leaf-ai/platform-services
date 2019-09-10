package platform

import (
	"context"

	"github.com/honeycombio/opencensus-exporter/honeycomb"
	"github.com/karlmutch/errors"

	"go.opencensus.io/trace"
)

func StartOpenCensus(ctx context.Context, apiKey string, dataset string) (err errors.Error) {

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

	return nil
}
