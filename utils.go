package bahamut

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/aporeto-inc/elemental"
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
func HandleRecoveredPanic(r interface{}) error {

	if r == nil {
		return nil
	}

	err := elemental.NewError("Internal Server Error", fmt.Sprintf("%v", r), "bahamut", http.StatusInternalServerError)
	st := string(debug.Stack())
	err.Data = st
	zap.L().Error("panic", zap.String("stacktrace", st))

	return err
}
