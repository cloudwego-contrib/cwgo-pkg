// Copyright 2023 CloudWeGo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package apollo

import (
	"bytes"
	"runtime/debug"
	"sync"
	"text/template"

	cwutils "github.com/cloudwego-contrib/cwgo-pkg/config/utils"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/shima-park/agollo"
)

// Client the wrapper of apollo client.
type Client interface {
	SetParser(cwutils.ConfigParser)
	ClientConfigParam(cpc *cwutils.ConfigParamConfig) (ConfigParam, error)
	ServerConfigParam(cpc *cwutils.ConfigParamConfig) (ConfigParam, error)
	RegisterConfigCallback(ConfigParam, func(string, cwutils.ConfigParser), int64)
	DeregisterConfig(ConfigParam, int64) error
}

type ConfigParam struct {
	Key       string
	nameSpace string
	Cluster   string
	Type      cwutils.ConfigType
}

type callbackHandler func(namespace, cluster, key, data string)

type configParamKey struct {
	Key       string
	NameSpace string
	Cluster   string
}

// namespace: category
// key: ClientService.ServerService
func getConfigParamKey(in *ConfigParam) configParamKey {
	return configParamKey{
		Key:       in.Key,
		NameSpace: in.nameSpace,
		Cluster:   in.Cluster,
	}
}

type client struct {
	acli agollo.Agollo
	// support customise parser
	parser            cwutils.ConfigParser
	stop              chan bool
	clusterTemplate   *template.Template
	serverKeyTemplate *template.Template
	clientKeyTemplate *template.Template
	handlerMutex      sync.RWMutex
	handlers          map[configParamKey]map[int64]callbackHandler
}

const (
	RetryConfigName          = "retry"
	RpcTimeoutConfigName     = "rpc_timeout"
	CircuitBreakerConfigName = "circuit_break"

	LimiterConfigName = "limit"
)

var Close sync.Once

type Options struct {
	ConfigServerURL string
	AppID           string
	Cluster         string
	ServerKeyFormat string
	ClientKeyFormat string
	ApolloOptions   []agollo.Option
	ConfigParser    cwutils.ConfigParser
}

type OptionFunc func(option *Options)

func NewClient(opts Options, optsfunc ...OptionFunc) (Client, error) {
	if opts.ConfigServerURL == "" {
		opts.ConfigServerURL = ApolloDefaultConfigServerURL
	}
	if opts.ConfigParser == nil {
		opts.ConfigParser = cwutils.DefaultConfigParse()
	}
	if opts.AppID == "" {
		opts.AppID = ApolloDefaultAppId
	}
	if opts.Cluster == "" {
		opts.Cluster = ApolloDefaultCluster
	}
	if opts.ServerKeyFormat == "" {
		opts.ServerKeyFormat = ApolloDefaultServerKey
	}
	if opts.ClientKeyFormat == "" {
		opts.ClientKeyFormat = ApolloDefaultClientKey
	}
	opts.ApolloOptions = append(opts.ApolloOptions,
		agollo.Cluster(opts.Cluster),
		agollo.AutoFetchOnCacheMiss(),
		agollo.FailTolerantOnBackupExists(),
	)
	for _, option := range optsfunc {
		option(&opts)
	}
	apolloCli, err := agollo.New(opts.ConfigServerURL, opts.AppID, opts.ApolloOptions...)
	if err != nil {
		return nil, err
	}
	clusterTemplate, err := template.New("cluster").Parse(opts.Cluster)
	if err != nil {
		return nil, err
	}
	serverKeyTemplate, err := template.New("serverKey").Parse(opts.ServerKeyFormat)
	if err != nil {
		return nil, err
	}
	clientKeyTemplate, err := template.New("clientKey").Parse(opts.ClientKeyFormat)
	if err != nil {
		return nil, err
	}
	cli := &client{
		acli:              apolloCli,
		parser:            opts.ConfigParser,
		stop:              make(chan bool),
		clusterTemplate:   clusterTemplate,
		serverKeyTemplate: serverKeyTemplate,
		clientKeyTemplate: clientKeyTemplate,
		handlers:          make(map[configParamKey]map[int64]callbackHandler),
	}

	return cli, nil
}

func WithApolloOption(apolloOption ...agollo.Option) OptionFunc {
	return func(option *Options) {
		option.ApolloOptions = append(option.ApolloOptions, apolloOption...)
	}
}

func (c *client) SetParser(parser cwutils.ConfigParser) {
	c.parser = parser
}

func (c *client) render(cpc *cwutils.ConfigParamConfig, t *template.Template) (string, error) {
	var tpl bytes.Buffer
	err := t.Execute(&tpl, cpc)
	if err != nil {
		return "", err
	}
	return tpl.String(), nil
}

func (c *client) ServerConfigParam(cpc *cwutils.ConfigParamConfig) (ConfigParam, error) {
	return c.configParam(cpc, c.serverKeyTemplate)
}

// ClientConfigParam render client config parameters
func (c *client) ClientConfigParam(cpc *cwutils.ConfigParamConfig) (ConfigParam, error) {
	return c.configParam(cpc, c.clientKeyTemplate)
}

// configParam render config parameters. All the parameters can be customized with CustomFunction.
// ConfigParam explain:
//  1. Type: key format, support JSON and YAML, JSON by default. Could extend it by implementing the ConfigParser interface.
//  2. Content: empty by default. Customize with CustomFunction.
//  3. nameSpace: select by user (retry / circuit_breaker / rpc_timeout / limit).
//  4. ServerKey: {{.ServerServiceName}} by default.
//     ClientKey: {{.ClientServiceName}}.{{.ServerServiceName}} by default.
//  5. Cluster: default by default
func (c *client) configParam(cpc *cwutils.ConfigParamConfig, t *template.Template) (ConfigParam, error) {
	param := ConfigParam{
		Type:      cwutils.JSON,
		nameSpace: cpc.Category,
	}
	var err error
	param.Key, err = c.render(cpc, t)
	if err != nil {
		return param, err
	}
	param.Cluster, err = c.render(cpc, c.clusterTemplate)
	if err != nil {
		return param, err
	}
	return param, nil
}

// DeregisterConfig deregister the config.
func (c *client) DeregisterConfig(cfg ConfigParam, uniqueID int64) error {
	configKey := getConfigParamKey(&cfg)
	klog.Debugf("deregister key %v for uniqueID %d", configKey, uniqueID)
	c.handlerMutex.Lock()
	defer c.handlerMutex.Unlock()
	handlers, ok := c.handlers[configKey]
	if ok {
		delete(handlers, uniqueID)
	}
	// Stop when users is null
	if len(handlers) == 0 {
		Close.Do(func() {
			// close listen goroutine
			close(c.stop)
		})
		// close longpoll
		c.acli.Stop()
	}
	return nil
}

// Read and execute callback functions for unique value binding
func (c *client) onChange(namespace, cluster, key, data string) {
	handlers := make([]callbackHandler, 0, 5)

	c.handlerMutex.RLock()
	configKey := configParamKey{
		Key:       key,
		NameSpace: namespace,
		Cluster:   cluster,
	}
	for _, handler := range c.handlers[configKey] {
		handlers = append(handlers, handler)
	}
	c.handlerMutex.RUnlock()
	for _, handler := range handlers {
		handler(namespace, cluster, key, data)
	}
}

// RegisterConfigCallback register the callback function to apollo client.
func (c *client) RegisterConfigCallback(param ConfigParam,
	callback func(string, cwutils.ConfigParser), uniqueID int64,
) {
	onChange := func(namespace, cluster, key, data string) {
		klog.Debugf("[apollo] uniqueID %d config %s updated, namespace %s cluster %s key %s data %s",
			uniqueID, namespace, namespace, cluster, key, data)
		callback(data, c.parser)
	}

	configMap := c.acli.GetNameSpace(param.nameSpace)
	data, ok := configMap[param.Key]
	if !ok {
		klog.Warnf("[apollo] key not found | key :%s", param.Key)
		klog.Warnf("[apollo] configMap: %v", configMap)
	} else {
		callback(data.(string), c.parser)
	}

	go c.listenConfig(param, c.stop, onChange, uniqueID)
}

func (c *client) listenConfig(param ConfigParam, stop chan bool, callback func(namespace, cluster, key, data string), uniqueID int64) {
	defer func() {
		if err := recover(); err != nil {
			klog.Error("[apollo] listen goroutine cwerror: %v, stack: %s", err, string(debug.Stack()))
		}
	}()

	configKey := getConfigParamKey(&param)
	klog.Debugf("register key %v for uniqueID %d", configKey, uniqueID)
	c.handlerMutex.Lock()
	handlers, ok := c.handlers[configKey]
	if !ok {
		handlers = make(map[int64]callbackHandler)
		c.handlers[configKey] = handlers
	}
	handlers[uniqueID] = callback
	c.handlerMutex.Unlock()

	if !ok {
		errorsCh := c.acli.Start()
		apolloRespCh := c.acli.WatchNamespace(param.nameSpace, stop)

		for {
			select {
			case resp := <-apolloRespCh:
				data, ok := resp.NewValue[param.Key]
				if !ok {
					// Deal with delete config
					klog.Warnf("[apollo] config %s cwerror, namespace %s cluster %s key %s : cwerror : key not found | please recover key from remote config",
						param.nameSpace, param.nameSpace, param.Cluster, param.Key)
					c.onChange(param.nameSpace, param.Cluster, param.Key, emptyConfig)
					continue
				}
				c.onChange(param.nameSpace, param.Cluster, param.Key, data.(string))
			case err := <-errorsCh:
				klog.Errorf("[apollo] config %s cwerror, namespace %s cluster %s key %s : cwerror %s",
					param.nameSpace, param.nameSpace, param.Cluster, param.Key, err.Err.Error())
				return
			case <-stop:
				klog.Warnf("[apollo] config %s exit,namespace %s cluster %s key %s : exit",
					param.nameSpace, param.nameSpace, param.Cluster, param.Key)
				return
			}
		}
	}
}
