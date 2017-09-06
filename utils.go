package bahamut

import (
	"fmt"
	"net/http"
	"runtime/debug"

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

	switch e := err.(type) {

	case elemental.Error:
		e.Trace = request.RequestID
		outError = elemental.NewErrors(e)

	case elemental.Errors:
		for _, err := range e {
			if eerr, ok := err.(elemental.Error); ok {
				eerr.Trace = request.RequestID
				outError = append(outError, eerr)
			} else {
				outError = append(outError, err)
			}
		}

	default:
		eerr := elemental.NewError("Internal Server Error", e.Error(), "bahamut", http.StatusInternalServerError)
		eerr.Trace = request.RequestID
		outError = elemental.NewErrors(eerr)
		zap.L().Error("Internal Server Error", zap.Error(eerr), zap.String("trace", request.RequestID))
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
