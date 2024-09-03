package common

import (
	"os"
)

const (
	NacosEnvServerAddr     = "serverAddr"
	NacosEnvServerPort     = "serverPort"
	NacosEnvNamespaceID    = "namespace"
	NacosDefaultServerAddr = "127.0.0.1"
	NacosDefaultPort       = 8848
	NacosDefaultRegionID   = "cn-hangzhou"
)

// NacosAddr Get Nacos addr from environment variables.
func NacosAddr() string {
	addr := os.Getenv(NacosEnvServerAddr)
	if len(addr) == 0 {
		return NacosDefaultServerAddr
	}
	return addr
}

// NacosNameSpaceID Get Nacos namespace id from environment variables.
func NacosNameSpaceID() string {
	return os.Getenv(NacosEnvNamespaceID)
}
