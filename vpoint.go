package core

import (
	"io/ioutil"

	"github.com/v2ray/v2ray-core/log"
	v2net "github.com/v2ray/v2ray-core/net"
)

var (
	inboundFactories  = make(map[string]InboundConnectionHandlerFactory)
	outboundFactories = make(map[string]OutboundConnectionHandlerFactory)
)

func RegisterInboundConnectionHandlerFactory(name string, factory InboundConnectionHandlerFactory) error {
	// TODO check name
	inboundFactories[name] = factory
	return nil
}

func RegisterOutboundConnectionHandlerFactory(name string, factory OutboundConnectionHandlerFactory) error {
	// TODO check name
	outboundFactories[name] = factory
	return nil
}

// VPoint is an single server in V2Ray system.
type VPoint struct {
	port       uint16
	ichFactory InboundConnectionHandlerFactory
	ichConfig  []byte
	ochFactory OutboundConnectionHandlerFactory
	ochConfig  []byte
}

// NewVPoint returns a new VPoint server based on given configuration.
// The server is not started at this point.
func NewVPoint(config VConfig) (*VPoint, error) {
	var vpoint = new(VPoint)
	vpoint.port = config.Port

	ichFactory, ok := inboundFactories[config.InboundConfig.Protocol]
	if !ok {
		panic(log.Error("Unknown inbound connection handler factory %s", config.InboundConfig.Protocol))
	}
	vpoint.ichFactory = ichFactory
	if len(config.InboundConfig.File) > 0 {
		ichConfig, err := ioutil.ReadFile(config.InboundConfig.File)
		if err != nil {
			panic(log.Error("Unable to read config file %v", err))
		}
		vpoint.ichConfig = ichConfig
	}

	ochFactory, ok := outboundFactories[config.OutboundConfig.Protocol]
	if !ok {
		panic(log.Error("Unknown outbound connection handler factory %s", config.OutboundConfig.Protocol))
	}

	vpoint.ochFactory = ochFactory
	if len(config.OutboundConfig.File) > 0 {
		ochConfig, err := ioutil.ReadFile(config.OutboundConfig.File)
		if err != nil {
			panic(log.Error("Unable to read config file %v", err))
		}
		vpoint.ochConfig = ochConfig
	}

	return vpoint, nil
}

type InboundConnectionHandlerFactory interface {
	Create(vp *VPoint, config []byte) (InboundConnectionHandler, error)
}

type InboundConnectionHandler interface {
	Listen(port uint16) error
}

type OutboundConnectionHandlerFactory interface {
	Create(VP *VPoint, config []byte, dest v2net.VAddress) (OutboundConnectionHandler, error)
}

type OutboundConnectionHandler interface {
	Start(vray OutboundVRay) error
}

// Start starts the VPoint server, and return any error during the process.
// In the case of any errors, the state of the server is unpredicatable.
func (vp *VPoint) Start() error {
	if vp.port <= 0 {
		return log.Error("Invalid port %d", vp.port)
	}
	inboundConnectionHandler, err := vp.ichFactory.Create(vp, vp.ichConfig)
	if err != nil {
		return err
	}
	err = inboundConnectionHandler.Listen(vp.port)
	return nil
}

func (vp *VPoint) NewInboundConnectionAccepted(destination v2net.VAddress) InboundVRay {
	ray := NewVRay()
	// TODO: handle error
	och, _ := vp.ochFactory.Create(vp, vp.ochConfig, destination)
	_ = och.Start(ray)
	return ray
}
