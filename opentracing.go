package bahamut

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"go.aporeto.io/elemental"

	opentracing "github.com/opentracing/opentracing-go"
)

var snipSlice = []string{"[snip]"}

func extractClaims(r *elemental.Request) string {

	tokenParts := strings.SplitN(r.Password, ".", 3)
	if len(tokenParts) != 3 {
		return "{}"
	}

	identity, err := base64.RawStdEncoding.DecodeString(tokenParts[1])
	if err != nil {
		return "{}"
	}

	return string(identity)
}

func tracingName(r *elemental.Request) string {

	switch r.Operation {

	case elemental.OperationCreate:
		return fmt.Sprintf("bahamut.handle.create.%s", r.Identity.Category)

	case elemental.OperationRetrieveMany:
		return fmt.Sprintf("bahamut.handle.retrieve_many.%s", r.Identity.Category)

	case elemental.OperationInfo:
		return fmt.Sprintf("bahamut.handle.info.%s", r.Identity.Category)

	case elemental.OperationUpdate:
		return fmt.Sprintf("bahamut.handle.update.%s", r.Identity.Category)

	case elemental.OperationDelete:
		return fmt.Sprintf("bahamut.handle.delete.%s", r.Identity.Category)

	case elemental.OperationRetrieve:
		return fmt.Sprintf("bahamut.handle.retrieve.%s", r.Identity.Category)

	case elemental.OperationPatch:
		return fmt.Sprintf("bahamut.handle.patch.%s", r.Identity.Category)
	}

	return fmt.Sprintf("Unknown operation: %s", r.Operation)
}

// StartTracing starts tracing the request.
func traceRequest(ctx context.Context, r *elemental.Request, tracer opentracing.Tracer) context.Context {

	if tracer == nil {
		return ctx
	}

	spanContext, _ := tracer.Extract(opentracing.TextMap, opentracing.HTTPHeadersCarrier(r.Headers))
	span := tracer.StartSpan(tracingName(r), ext.RPCServerOption(spanContext))
	trackingCtx := opentracing.ContextWithSpan(ctx, span)

	// Remove sensitive information from parameters.
	safeParameters := url.Values{}
	for k, v := range r.Parameters {
		lk := strings.ToLower(k)
		if lk == "token" || lk == "password" {
			safeParameters[k] = snipSlice
			continue
		}
		safeParameters[k] = v
	}

	// Remove sensitive information from headers.
	safeHeaders := http.Header{}
	for k, v := range r.Headers {
		lk := strings.ToLower(k)
		if lk == "authorization" {
			safeHeaders[k] = snipSlice
			continue
		}
		safeHeaders[k] = v
	}

	span.SetTag("req.api_version", r.Version)
	span.SetTag("req.id", r.RequestID)
	span.SetTag("req.identity", r.Identity.Name)
	span.SetTag("req.recursive", r.Recursive)
	span.SetTag("req.operation", r.Operation)
	span.SetTag("req.override_protection", r.OverrideProtection)

	if r.ExternalTrackingID != "" {
		span.SetTag("req.external_tracking_id", r.ExternalTrackingID)
	}

	if r.ExternalTrackingType != "" {
		span.SetTag("req.external_tracking_type", r.ExternalTrackingType)
	}

	if r.Namespace != "" {
		span.SetTag("req.namespace", r.Namespace)
	}

	if r.ObjectID != "" {
		span.SetTag("req.object.id", r.ObjectID)
	}

	if r.ParentID != "" {
		span.SetTag("req.parent.id", r.ParentID)
	}

	if !r.ParentIdentity.IsEmpty() {
		span.SetTag("req.parent.identity", r.ParentIdentity.Name)
	}

	span.LogFields(
		log.Int("req.page.number", r.Page),
		log.Int("req.page.size", r.PageSize),
		log.Object("req.headers", safeHeaders),
		log.Object("req.claims", extractClaims(r)),
		log.Object("req.client_ip", r.ClientIP),
		log.Object("req.parameters", safeParameters),
		log.Object("req.order_by", r.Order),
		log.String("req.payload", string(r.Data)),
	)

	return trackingCtx
}

func finishTracing(ctx context.Context) {

	span := opentracing.SpanFromContext(ctx)
	if span == nil {
		return
	}

	span.Finish()
}
