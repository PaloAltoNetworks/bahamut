package bahamut

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
	"strings"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"go.aporeto.io/elemental"
	"go.uber.org/zap"
)

func handleRecoveredPanic(ctx context.Context, r interface{}, disablePanicRecovery bool) error {

	if r == nil {
		return nil
	}

	err := elemental.NewError("Internal Server Error", fmt.Sprintf("%v", r), "bahamut", http.StatusInternalServerError)

	st := string(debug.Stack())
	zap.L().Error("panic", zap.String("stacktrace", st))

	// Print the panic as it would have happened
	fmt.Fprintf(os.Stderr, "panic: %s\n\n%s", err, st) // nolint: errcheck

	sp := opentracing.SpanFromContext(ctx)
	if sp != nil {
		sp.SetTag("error", true)
		sp.SetTag("panic", true)
		sp.LogFields(
			log.String("panic", fmt.Sprintf("%v", r)),
			log.String("stack", st),
		)
	}

	if disablePanicRecovery {
		if sp != nil {
			sp.Finish()
		}
		panic(err)
	}

	return err
}

func extractSpanID(span opentracing.Span) string {

	spanID := "unknown"
	if stringer, ok := span.(fmt.Stringer); ok {
		spanID = strings.SplitN(stringer.String(), ":", 2)[0]
	}

	return spanID
}

func processError(ctx context.Context, err error) (outError elemental.Errors) {

	span := opentracing.SpanFromContext(ctx)

	spanID := extractSpanID(span)

	switch e := err.(type) {

	case elemental.Error:
		e.Trace = spanID
		outError = elemental.NewErrors(e)

	case elemental.Errors:
		for _, err := range e {
			if eerr, ok := err.(elemental.Error); ok {
				eerr.Trace = spanID
				outError = append(outError, eerr)
			} else {
				cerr := elemental.NewError("Internal Server Error", err.Error(), "bahamut", http.StatusInternalServerError)
				cerr.Trace = spanID
				outError = append(outError, cerr)
			}
		}

	default:
		eerr := elemental.NewError("Internal Server Error", e.Error(), "bahamut", http.StatusInternalServerError)
		eerr.Trace = spanID
		outError = elemental.NewErrors(eerr)
		zap.L().Error("Internal Server Error", zap.Error(eerr), zap.String("trace", spanID))
	}

	if span != nil {
		span.SetTag("error", true)
		span.SetTag("status.code", outError.Code())
		span.LogFields(log.Object("elemental.error", outError))
	}

	return outError
}

func claimsToMap(claims []string) map[string]string {

	claimsMap := map[string]string{}

	var k, v string

	for _, claim := range claims {
		if err := splitPtr(claim, &k, &v); err != nil {
			panic(err)
		}
		claimsMap[k] = v
	}

	return claimsMap
}

func splitPtr(tag string, key *string, value *string) (err error) {

	l := len(tag)
	if l < 3 {
		err = fmt.Errorf("Invalid tag: invalid length '%s'", tag)
		return
	}

	if tag[0] == '=' {
		err = fmt.Errorf("Invalid tag: missing key '%s'", tag)
		return
	}

	for i := 0; i < l; i++ {
		if tag[i] == '=' {
			if i+1 >= l {
				return fmt.Errorf("Invalid tag: missing value '%s'", tag)
			}
			*key = tag[:i]
			*value = tag[i+1:]
			return
		}
	}

	return fmt.Errorf("Invalid tag: missing equal symbol '%s'", tag)
}
