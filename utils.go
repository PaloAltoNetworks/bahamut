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

// HandleRecoveredPanic returns a well formatted elemental error and logs its if a panic occurred.
func HandleRecoveredPanic(r interface{}, req *elemental.Request) error {

	if r == nil {
		return nil
	}

	err := elemental.NewError("Internal Server Error", fmt.Sprintf("%v", r), "bahamut", http.StatusInternalServerError)
	st := string(debug.Stack())
	err.Data = map[string]interface{}{"stacktrace": st}
	zap.L().Error("panic", zap.String("stacktrace", st))

	sp := req.NewChildSpan("bahamut.result.panic")
	sp.SetTag("error", true)
	sp.SetTag("panic", true)
	sp.LogFields(
		log.String("panic", fmt.Sprintf("%v", r)),
		log.String("stacktrace", st),
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
		for _, ee := range e {
			if eee, ok := ee.(elemental.Error); ok {
				eee.Trace = request.RequestID
				outError = append(outError, eee)
			} else {
				outError = append(outError, ee)
			}
		}

	default:
		er := elemental.NewError("Internal Server Error", e.Error(), "bahamut", http.StatusInternalServerError)
		er.Trace = request.RequestID
		outError = elemental.NewErrors(er)
	}

	if request.Span() != nil {
		sp := request.NewChildSpan("bahamut.result.error")
		sp.SetTag("error", true)
		sp.LogFields(log.Object("elemental.error", outError))
		sp.Finish()
	}

	return outError
}
