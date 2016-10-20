package main

import (
	"github.com/docopt/docopt-go"

	"github.com/aporeto-inc/bahamut"
	"github.com/aporeto-inc/manipulate/manipmemory"

	"github.com/aporeto-inc/bahamut/example/handlers"
	"github.com/aporeto-inc/bahamut/example/models"
	"github.com/aporeto-inc/bahamut/example/processors"
	"github.com/aporeto-inc/bahamut/example/routes"

	"github.com/aporeto-inc/bahamut/example/db"
)

func main() {

	usage := `Demo Bahamut Server.

Usage:
    server -h | --help
    server -v | --version
    server [--listen=<addr>]

Options:
    -h --help               Show this screen.
    -v --version            Show the version.
    -L --listen=<addr>      Bind address:port.`

	arguments, _ := docopt.Parse(usage, nil, true, "Todo list application", false)
	listenAddr := ":9999"
	if arguments["--listen"] != nil {
		listenAddr = arguments["--listen"].(string)
	}

	// Print the Bahamut banner FTW!
	bahamut.PrintBanner()

	// Create the API configuration.
	// There is a lot more flag you can configure here.
	// Check the bahamut documentation for more information.
	apiConfig := bahamut.APIServerConfig{
		ListenAddress: listenAddr,
		Routes:        routes.Routes(),
	}

	// If you want to use kafka, use this line, otherwise you can use local channel backed push system.
	// pubsub = bahamut.NewKafkaPubSubServer([]string{locahost:9092})
	pubsub := bahamut.NewLocalPubSubServer(nil)

	// Connect to the pubsub.
	pubsub.Connect()

	// Create the API push configuration.
	// There is a lot more flag you can configure here.
	// Check the bahamut documentation for more information.
	pushConfig := bahamut.PushServerConfig{
		Service: pubsub,
	}

	// Create a bahamut serrver, and pass the api and push config.
	server := bahamut.NewServer(apiConfig, pushConfig)

	handlers.SetBahamutServer(server)

	manipulator := manipmemory.NewMemoryManipulator(db.Schema)

	// Then register your processors.
	server.RegisterProcessor(processors.NewListProcessor(manipulator), models.ListIdentity)
	server.RegisterProcessor(processors.NewTaskProcessor(manipulator), models.TaskIdentity)

	// And start the server.
	server.Start()
}
