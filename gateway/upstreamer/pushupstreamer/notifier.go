package pushupstreamer

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/shirou/gopsutil/process"
	"go.aporeto.io/bahamut"
	"go.uber.org/zap"
)

// A Notifier sends ServicePing to the Wutai gateways.
type Notifier struct {
	pubsub             bahamut.PubSubClient
	serviceName        string
	endpoint           string
	serviceStatusTopic string
}

// NewNotifier returns a new Wutai notifier.
func NewNotifier(pubsub bahamut.PubSubClient, serviceStatusTopic string, serviceName string, listenAddress string) *Notifier {

	_, port, err := net.SplitHostPort(listenAddress)
	if err != nil {
		zap.L().Fatal("Unable to parse listen address", zap.Error(err))
	}

	host, err := os.Hostname()
	if err != nil {
		zap.L().Fatal("Unable to retrieve hostname", zap.Error(err))
	}

	addrs, err := net.LookupHost(host)
	if err != nil {
		zap.L().Fatal("Unable to resolve hostname", zap.Error(err))
	}

	if len(addrs) == 0 {
		zap.L().Fatal("Unable to find any IP in resolved hostname", zap.Error(err))
	}

	var endpoint string
	for _, addr := range addrs {
		ip := net.ParseIP(addr)
		if len(ip.To4()) == net.IPv4len {
			endpoint = addr
			break
		}
	}

	if endpoint == "" {
		endpoint = addrs[0]
	}

	return &Notifier{
		pubsub:             pubsub,
		serviceName:        serviceName,
		endpoint:           fmt.Sprintf("%s:%s", endpoint, port),
		serviceStatusTopic: serviceStatusTopic,
	}
}

// MakeStartHook returns a bahamut start hook that sends the hello message to the Upstreamer periodically.
func (w *Notifier) MakeStartHook(ctx context.Context, zone int) func(server bahamut.Server) error {

	return func(server bahamut.Server) error {

		p, err := process.NewProcess(int32(os.Getpid()))
		if err != nil {
			return err
		}

		sp := ping{
			Name:         w.serviceName,
			Status:       serviceStatusHello,
			Endpoint:     w.endpoint,
			Routes:       server.RoutesInfo(),
			Versions:     server.VersionsInfo(),
			PushEndpoint: server.PushEndpoint(),
		}

		if sp.Load, err = p.CPUPercent(); err != nil {
			return err
		}

		pub := bahamut.NewPublication(fmt.Sprintf("%s-%d", w.serviceStatusTopic, zone))
		if err := pub.Encode(sp); err != nil {
			return err
		}

		if err := w.pubsub.Publish(pub); err != nil {
			return err
		}

		go func() {
			for {
				select {
				case <-time.After(5 * time.Second):

					if sp.Load, err = p.Percent(0); err != nil {
						zap.L().Error("Unable to retrieve cpu usage", zap.Error(err))
						continue
					}

					if err := pub.Encode(sp); err != nil {
						zap.L().Error("Unable to encode service ping", zap.Error(err))
						continue
					}

					if err := w.pubsub.Publish(pub); err != nil {
						zap.L().Error("Unable to send wutai up ping", zap.Error(err))
					}
				case <-ctx.Done():
					return
				}
			}
		}()

		return nil
	}
}

// MakeStopHook returns a bahamut stop hook that sends the goodbye message to the Upstreamer.
func (w *Notifier) MakeStopHook(zone int) func(server bahamut.Server) error {

	return func(server bahamut.Server) error {

		pub := bahamut.NewPublication(fmt.Sprintf("%s-%d", w.serviceStatusTopic, zone))
		if err := pub.Encode(ping{
			Name:     w.serviceName,
			Status:   serviceStatusGoodbye,
			Endpoint: w.endpoint,
		}); err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := w.pubsub.Publish(pub, bahamut.NATSOptPublishRequireAck(ctx)); err != nil {
			return err
		}

		<-time.After(time.Second)

		return w.pubsub.Disconnect()
	}
}
