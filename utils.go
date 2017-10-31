package bahamut

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/aporeto-inc/elemental"
	"github.com/opentracing/opentracing-go/log"
	"go.uber.org/zap"
)

// PrintBanner prints the Bahamut Banner.
//
// Yey!
func PrintBanner() {
	fmt.Println(`
   ____        _                           _           .
  | __ )  __ _| |__   __ _ _ __ ___  _   _| |_.   .>   )\;'a__
  |  _ \ / _. | '_ \ / _. | '_ ' _ \| | | | __|  (  _ _)/ /-." ~~
  | |_) | (_| | | | | (_| | | | | | | |_| | |_    '( )_ )/
  |____/ \__,_|_| |_|\__,_|_| |_| |_|\__,_|\__|    <_  <_

___________________________________________________________________
                                                     🚀  by Aporeto
`)
}

func handleRecoveredPanic(r interface{}, req *elemental.Request) error {

	if r == nil {
		return nil
	}

	err := elemental.NewError("Internal Server Error", fmt.Sprintf("%v", r), "bahamut", http.StatusInternalServerError)
	st := string(debug.Stack())
	zap.L().Error("panic", zap.String("stacktrace", st))

	sp := req.NewChildSpan("bahamut.result.panic")
	sp.SetTag("error", true)
	sp.SetTag("panic", true)
	sp.LogFields(
		log.String("panic", fmt.Sprintf("%v", r)),
		log.String("stack", st),
	)
	sp.Finish()

	return err
}

func processError(err error, request *elemental.Request) elemental.Errors {

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

	if request.Span() != nil {
		sp := request.NewChildSpan("bahamut.result.error")
		sp.SetTag("error", true)
		sp.SetTag("error.code", outError.Code())
		sp.LogFields(log.Object("elemental.error", outError))
		sp.Finish()
	}

	return outError
}

func claimsToMap(claims []string) map[string]string {

	claimsMap := map[string]string{}

	for _, claim := range claims {
		parts := strings.SplitN(claim, "=", 2)
		if len(parts) == 2 {
			claimsMap[parts[0]] = parts[1]
		}
	}

	return claimsMap
}
