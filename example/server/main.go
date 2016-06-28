package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/docopt/docopt-go"

	"github.com/aporeto-inc/bahamut"
	"github.com/aporeto-inc/elemental"

	"github.com/aporeto-inc/bahamut/example/server/models"
	"github.com/aporeto-inc/bahamut/example/server/processors"
	"github.com/aporeto-inc/bahamut/example/server/routes"
)

type mySessionHandler struct{}

func (h *mySessionHandler) OnPushSessionStart(session *bahamut.PushSession) {}
func (h *mySessionHandler) OnPushSessionStop(session *bahamut.PushSession)  {}
func (h *mySessionHandler) ShouldPush(session *bahamut.PushSession, event *elemental.Event) bool {

	info := event.Entity.(map[string]interface{})
	return info["name"] != "ignore"
}

func main() {

	usage := `Demo Bahamut Server.

Usage:
    server -h | --help
    server -v | --version
    server [--log=<level>] [--format=<format>] [--profiling] [--no-push] [--no-api] [--listen=<addr>]

Arguments:
   level:  info | debug | warn
   format: text | json

Options:
    -h --help               Show this screen.
    -v --version            Show the version.
    -l --log=<level>        Set the log level [default: info].
    -f --format=<format>    Set the log level [default: text].
    -p --profiling          Enable the profiling server.
    -n --no-push            Disable the push server.
    -a --no-api             Disable the api server.
    -L --listen=<addr>      Bind address:port.`

	arguments, _ := docopt.Parse(usage, nil, true, "Test Cid Server", false)

	if arguments["--log"] == "debug" {
		log.SetLevel(log.DebugLevel)
	} else if arguments["--log"] == "info" {
		log.SetLevel(log.InfoLevel)
	} else if arguments["--log"] == "warn" {
		log.SetLevel(log.WarnLevel)
	}

	if arguments["--format"] == "json" {
		log.SetFormatter(&log.JSONFormatter{})
	}

	bahamut.PrintBanner()

	apiEnabled := true
	if arguments["-no-api"] == true {
		apiEnabled = false
	}

	profilingEnabled := false
	if arguments["--profiling"] == true {
		profilingEnabled = true
	}

	pushEnabled := true
	if arguments["--no-push"] == true {
		pushEnabled = false
	}

	address := ":9999"
	if arguments["--listen"] != nil {
		address = arguments["--listen"].(string)
	}

	pushConfig := bahamut.MakePushServerConfig([]string{"127.0.0.1:9092"}, "bahamut", &mySessionHandler{})
	server := bahamut.NewBahamut(address, routes.Routes(), pushConfig, apiEnabled, pushEnabled, profilingEnabled)
	server.RegisterProcessor(processors.NewListProcessor(), models.ListIdentity)
	server.RegisterProcessor(processors.NewTaskProcessor(), models.TaskIdentity)
	server.RegisterProcessor(processors.NewUserProcessor(), models.UserIdentity)

	server.Start()
}
