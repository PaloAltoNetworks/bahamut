package bahamut

import (
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
	"strings"

	"github.com/aporeto-inc/addedeffect/tagutils"
	"github.com/aporeto-inc/elemental"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"go.uber.org/zap"
)

func handleRecoveredPanic(r interface{}, req *elemental.Request, recover bool) error {

	if r == nil {
		return nil
	}

	err := elemental.NewError("Internal Server Error", fmt.Sprintf("%v", r), "bahamut", http.StatusInternalServerError)

	st := string(debug.Stack())
	zap.L().Error("panic", zap.String("stacktrace", st))

	// Print the panic as it would have happened
	fmt.Fprintf(os.Stderr, "panic: %s\n\n%s", err, st) // nolint: errcheck

	sp := req.NewChildSpan("bahamut.result.panic")
	sp.SetTag("error", true)
	sp.SetTag("panic", true)
	sp.LogFields(
		log.String("panic", fmt.Sprintf("%v", r)),
		log.String("stack", st),
	)
	sp.Finish()

	if !recover {
		panic(err)
	}

	return err
}

func processError(err error, request *elemental.Request, span opentracing.Span) elemental.Errors {

	var outError elemental.Errors

	spanID := request.RequestID
	if stringer, ok := request.Span().(fmt.Stringer); ok {
		spanID = strings.SplitN(stringer.String(), ":", 2)[0]
	}

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
				outError = append(outError, err)
			}
		}

	default:
		eerr := elemental.NewError("Internal Server Error", e.Error(), "bahamut", http.StatusInternalServerError)
		eerr.Trace = spanID
		outError = elemental.NewErrors(eerr)
		zap.L().Error("Internal Server Error", zap.Error(eerr), zap.String("trace", spanID))
	}

	if span != nil {
		sp := opentracing.StartSpan("bahamut.result.error", opentracing.ChildOf(span.Context()))
		sp.SetTag("error", true)
		sp.SetTag("status.code", outError.Code())
		sp.LogFields(log.Object("elemental.error", outError))
		sp.Finish()
	}

	return outError
}

func claimsToMap(claims []string) map[string]string {

	claimsMap := map[string]string{}

	var k, v string

	for _, claim := range claims {
		if err := tagutils.SplitPtr(claim, &k, &v); err != nil {
			panic(err)
		}
		claimsMap[k] = v
	}

	return claimsMap
}
