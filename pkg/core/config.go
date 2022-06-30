package core

import (
	"fmt"
	"os"
	"strings"
)

// A Config contains the items a service uses to configure itself and its
// transport provider.
type Config struct {
	Mode        MessageType // Service runs as Mode (ping, pong, ding, or dong)
	ApiVersion  string      // ApiVersion is the version of the server's API
	Namespace   string      // localhost or Kubernetes Namespace
	ServicePort string      // Port to run the service on
	PingSvc     string      // Fully qualified name of the service, minus the
	PongSvc     string      // the transport.
	DingSvc     string
	DongSvc     string
}

// Init reads environment variables and creates a proper Config object.
func (c *Config) Init() {
	mode := getenv("PPDD_MODE", "ping")
	switch {
	case strings.EqualFold(mode, "ping"):
		c.Mode = Ping
	case strings.EqualFold(mode, "pong"):
		c.Mode = Pong
	case strings.EqualFold(mode, "ding"):
		c.Mode = Ding
	case strings.EqualFold(mode, "dong"):
		c.Mode = Dong
	default:
		c.Mode = Undefined
	}
	c.ApiVersion = getenv("API_VERSION", "v1")
	c.Namespace = getenv("NAMESPACE", "localhost")
	c.ServicePort = getenv("SVC_PORT", "8080")
	pingName := getenv("PING_SVC", "ping")
	pingPort := getenv("PING_PORT", "8080")
	pongName := getenv("PONG_SVC", "pong")
	pongPort := getenv("PONG_PORT", "8080")
	dingName := getenv("DING_SVC", "ding")
	dingPort := getenv("DING_PORT", "8080")
	dongName := getenv("DONG_SVC", "dong")
	dongPort := getenv("DONG_PORT", "8080")

	if strings.EqualFold(c.Namespace, "localhost") {
		c.PingSvc = fmt.Sprintf("localhost:%s", pingPort)
		c.PongSvc = fmt.Sprintf("localhost:%s", pongPort)
		c.DingSvc = fmt.Sprintf("localhost:%s", dingPort)
		c.DongSvc = fmt.Sprintf("localhost:%s", dongPort)
		return
	}
	c.PingSvc = fmt.Sprintf("%s.%s.svc.cluster.local:%s",
		pingName, c.Namespace, pingPort)
	c.PongSvc = fmt.Sprintf("%s.%s.svc.cluster.local:%s",
		pongName, c.Namespace, pongPort)
	c.DingSvc = fmt.Sprintf("%s.%s.svc.cluster.local:%s",
		dingName, c.Namespace, dingPort)
	c.DongSvc = fmt.Sprintf("%s.%s.svc.cluster.local:%s",
		dongName, c.Namespace, dongPort)
}

// getenv is a helper function that returns a default value if an environment
// variable is not set, or if it is set - but it is empty.
func getenv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && len(value) != 0 {
		return value
	}
	return fallback
}
