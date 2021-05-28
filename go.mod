module go.aporeto.io/bahamut

go 1.13

require (
	go.aporeto.io/elemental v1.100.1-0.20210428215439-6059ff91f9f7
	go.aporeto.io/tg v1.34.1-0.20210427202027-51db463efa40
	go.aporeto.io/wsc v1.36.1-0.20210422182307-cde7d2b8a7eb
)

require (
	github.com/NYTimes/gziphandler v1.1.1
	github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d // indirect
	github.com/armon/go-proxyproto v0.0.0-20200108142055-f0b8253b1507
	github.com/cespare/xxhash v1.1.0
	github.com/go-ole/go-ole v1.2.4 // indirect
	github.com/go-zoo/bone v1.3.0
	github.com/gofrs/uuid v3.3.0+incompatible
	github.com/golang/mock v1.4.4
	github.com/google/go-cmp v0.5.4 // indirect
	github.com/gorilla/websocket v1.4.2
	github.com/jonboulle/clockwork v0.2.0 // indirect
	github.com/karlseguin/ccache/v2 v2.0.6
	github.com/kr/text v0.2.0 // indirect
	github.com/mailgun/multibuf v0.0.0-20150714184110-565402cd71fb
	github.com/mitchellh/reflectwalk v1.0.1 // indirect
	github.com/nats-io/nats-server/v2 v2.1.7
	github.com/nats-io/nats.go v1.10.0
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/opentracing/opentracing-go v1.2.0
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.7.1
	github.com/shirou/gopsutil v2.20.6+incompatible
	github.com/sirupsen/logrus v1.8.1
	github.com/smartystreets/goconvey v1.6.4
	github.com/stretchr/testify v1.7.0 // indirect
	github.com/valyala/tcplisten v0.0.0-20161114210144-ceec8f93295a
	github.com/vulcand/oxy v1.1.0
	go.uber.org/zap v1.15.0
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83 // indirect
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b // indirect
	golang.org/x/sys v0.0.0-20210525143221-35b2ab0089ea // indirect
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e
	golang.org/x/tools v0.1.2 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	honnef.co/go/tools v0.1.4 // indirect
)

// Oxy
replace github.com/vulcand/oxy => github.com/aporeto-inc/oxy v1.10.1-0.20210528215002-c1399da9883f
