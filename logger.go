package bahamut

import (
	"github.com/Sirupsen/logrus"
)

// Logger contains the logger for bahamut.
var Logger = logrus.New()

var log = Logger.WithField("package", "github.com/aporeto-inc/bahamut")
