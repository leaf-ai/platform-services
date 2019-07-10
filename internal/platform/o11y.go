package platform

import (
	"context"
	"os"

	grpc_metadata "google.golang.org/grpc/metadata"

	"github.com/honeycombio/libhoney-go"
	"github.com/honeycombio/opencensus-exporter/honeycomb"
	"go.opencensus.io/trace"
)

func StartOpenCensus(apiKey string, dataset string) {
	exporter := honeycomb.NewExporter(apiKey, dataset)
	defer exporter.Close()

	trace.RegisterExporter(exporter)

	trace.ApplyConfig(trace.Config{DefaultSampler: trace.ProbabilitySampler(1.0)})
}

func CreateEvent(ctx context.Context, service string, name string) (ev *libhoney.Event) {

	ev = libhoney.NewEvent()

	ev.Add(map[string]interface{}{
		"name":         name,
		"service_name": service,
	})

	if md, ok := grpc_metadata.FromIncomingContext(ctx); ok { // check to see if this request is already part of a trace
		headers := []string{
			"x-request-id",
			"x-b3-traceid",
			"x-b3-spanid",
			"x-b3-parentspanid",
			"x-b3-sampled",
			"x-b3-flags",
			"x-ot-span-context",
		}
		for _, header := range headers {
			if v, isPresent := md[header]; isPresent {
				ev.AddField(header, v)
			}
		}
		if tmp, ok := md["x-b3-traceid"]; ok {
			if len(tmp) > 0 {
				// it's from the gateway
				ev.AddField("x-b3-traceid", tmp)
			} else {
				ev.AddField("x-b3-traceid", GetPseudoUUID())
			}
		}

		if tmp, ok := md["x-b3-spanid"]; ok && len(tmp) > 0 {
			ev.AddField("x-b3-parentspanid", tmp)
		}
	} else {
		ev.AddField("x-b3-traceid", GetPseudoUUID())
	}

	ev.AddField("x-b3-spanid", GetPseudoUUID())
	if host, errGo := os.Hostname(); errGo == nil {
		ev.AddField("hostname", host)
	}
	return ev
}
