package {{ handlers_package_name }}

import "github.com/aporeto-inc/bahamut"

var currentBahamut bahamut.Server

// SetBahamutServer sets the bahamut server to use for the handlers.
func SetBahamutServer(server bahamut.Server) {

	currentBahamut = server
}

func currentBahamutServer() bahamut.Server {

  if currentBahamut == nil {
    panic("You must set the current bahamut server using {{ handlers_package_name }}.SetBahamutServer()")
  }

  return currentBahamut
}
