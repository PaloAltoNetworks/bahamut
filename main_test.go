package bahamut

import "github.com/Sirupsen/logrus"

func init() {
	Logger.Level = logrus.FatalLevel
}
