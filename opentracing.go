// Copyright 2019 Aporeto Inc.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bahamut

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"go.aporeto.io/elemental"
)

var snipSlice = []string{"[snip]"}

func extractClaims(r *elemental.Request) string {

	if r.Password == "" {
		return "{}"
	}

	tokenParts := strings.SplitN(r.Password, ".", 3)
	if len(tokenParts) != 3 {
		return fmt.Sprintf("invalid token format: %s", r.Password)
	}

	identity, err := base64.RawStdEncoding.DecodeString(tokenParts[1])
	if err != nil {
		return fmt.Sprintf("invalid token encoding: %s: %s", r.Password, err)
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
func traceRequest(ctx context.Context, r *elemental.Request, tracer opentracing.Tracer, exludedIdentities map[string]struct{}, cleaner TraceCleaner) context.Context {

	if tracer == nil {
		return ctx
	}

	if _, ok := exludedIdentities[r.Identity.Name]; ok {
		return ctx
	}

	spanContext, _ := tracer.Extract(opentracing.TextMap, opentracing.HTTPHeadersCarrier(r.Headers))
	span := tracer.StartSpan(tracingName(r), ext.RPCServerOption(spanContext))
	trackingCtx := opentracing.ContextWithSpan(ctx, span)

	// Remove sensitive information from parameters.
	safeParameters := url.Values{}
	for k, p := range r.Parameters {
		lk := strings.ToLower(k)
		if lk == "token" || lk == "password" {
			safeParameters[k] = snipSlice
			continue
		}
		safeParameters[k] = []string{fmt.Sprintf("%v", p.Values())}
	}

	// Remove sensitive information from headers.
	safeHeaders := http.Header{}
	for k, v := range r.Headers {
		lk := strings.ToLower(k)
		if lk == "authorization" || lk == "cookie" {
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

	data := append([]byte{}, r.Data...)
	if cleaner != nil {
		data = cleaner(r.Identity, data)
	}

	span.LogFields(
		log.Int("req.page.number", r.Page),
		log.Int("req.page.size", r.PageSize),
		log.Object("req.headers", safeHeaders),
		log.Object("req.claims", extractClaims(r)),
		log.Object("req.client_ip", r.ClientIP),
		log.Object("req.parameters", safeParameters),
		log.Object("req.order_by", r.Order),
		log.String("req.payload", string(data)),
	)

	return trackingCtx
}

func finishTracing(ctx context.Context, onlyErrors bool) {

	span := opentracing.SpanFromContext(ctx)
	if span == nil {
		return
	}

	if span.BaggageItem(errorsOnlyFlagBaggageItem) == "" && onlyErrors {
		return
	}

	span.Finish()
}
