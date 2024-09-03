package common

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/polarismesh/polaris-go/api"
	"github.com/polarismesh/polaris-go/pkg/config"
)

var mutexPolarisContext sync.Mutex

// GetPolarisConfig get polaris config from endpoints.
func GetPolarisConfig(configFile ...string) (api.SDKContext, error) {
	mutexPolarisContext.Lock()
	defer mutexPolarisContext.Unlock()
	var (
		cfg config.Configuration
		err error
	)

	if len(configFile) != 0 {
		cfg, err = config.LoadConfigurationByFile(configFile[0])
	} else {
		cfg, err = config.LoadConfigurationByDefaultFile()
	}

	if err != nil {
		return nil, err
	}

	sdkCtx, err := api.InitContextByConfig(cfg)
	if err != nil {
		return nil, err
	}

	return sdkCtx, nil
}

// SplitDescription splits description to namespace and serviceName.
func SplitDescription(description string) (string, string) {
	str := strings.Split(description, ":")
	namespace, serviceName := str[0], str[1]
	return namespace, serviceName
}

// SplitCachedKey splits description to namespace and serviceName.
func SplitCachedKey(cachedKey string) (string, string) {
	str := strings.Split(cachedKey, ":")
	namespace, serviceName := str[1], str[2]
	return namespace, serviceName
}

// GetLocalIPv4Address gets local ipv4 address when info host is empty.
func GetLocalIPv4Address() (string, error) {
	addr, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addr {
		ipNet, isIpNet := addr.(*net.IPNet)
		if isIpNet && !ipNet.IP.IsLoopback() {
			ipv4 := ipNet.IP.To4()
			if ipv4 != nil {
				return ipv4.String(), nil
			}
		}
	}
	return "", fmt.Errorf("not found ipv4 address")
}

// GetInfoHostAndPort gets Host and port from info.Addr.
func GetInfoHostAndPort(Addr string) (string, int, error) {
	infoHost, port, err := net.SplitHostPort(Addr)
	if err != nil {
		return "", 0, err
	} else {
		if port == "" {
			return infoHost, 0, fmt.Errorf("registry info addr missing port")
		}
		if infoHost == "" {
			ipv4, err := GetLocalIPv4Address()
			if err != nil {
				return "", 0, fmt.Errorf("get local ipv4 cwerror, cause %v", err)
			}
			infoHost = ipv4
		}
	}
	infoPort, err := strconv.Atoi(port)
	if err != nil {
		return "", 0, err
	}
	return infoHost, infoPort, nil
}

// GetInstanceKey generates instanceKey  for one instance.
func GetInstanceKey(namespace, serviceName, host, port string) string {
	var instanceKey strings.Builder
	instanceKey.WriteString(namespace)
	instanceKey.WriteString(":")
	instanceKey.WriteString(serviceName)
	instanceKey.WriteString(":")
	instanceKey.WriteString(host)
	instanceKey.WriteString(":")
	instanceKey.WriteString(port)
	return instanceKey.String()
}
