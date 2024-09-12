// Copyright 2024 CloudWeGo Authors
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

package consul

import (
	"bytes"
	"context"
	cwutils "github.com/cloudwego-contrib/cwgo-pkg/config/utils"
	"html/template"
	"strconv"
	"sync"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
	"go.uber.org/zap"
)

const WatchByKey = "key"

type Key struct {
	Type   cwutils.ConfigType
	Prefix string
	Path   string
}
type ListenConfig struct {
	Key        string
	Type       string
	DataCenter string
	Token      string
	ConsulAddr string
	Namespace  string
	Partition  string
}
type Client interface {
	SetParser(configParser cwutils.ConfigParser)
	ClientConfigParam(cpc *cwutils.ConfigParamConfig, cfs ...CustomFunction) (Key, error)
	ServerConfigParam(cpc *cwutils.ConfigParamConfig, cfs ...CustomFunction) (Key, error)
	RegisterConfigCallback(key string, uniqueID int64, callback func(string, cwutils.ConfigParser))
	DeregisterConfig(key string, uniqueID int64)
}

type Options struct {
	Addr             string
	Prefix           string
	ServerPathFormat string
	ClientPathFormat string
	DataCenter       string
	TimeOut          time.Duration
	NamespaceId      string
	Token            string
	Partition        string
	LoggerConfig     *zap.Config
	ConfigParser     cwutils.ConfigParser
}

type client struct {
	consulCli          *api.Client
	lconfig            *ListenConfig
	parser             cwutils.ConfigParser
	consulTimeout      time.Duration
	prefixTemplate     *template.Template
	serverPathTemplate *template.Template
	clientPathTemplate *template.Template
	cancelMap          map[string]context.CancelFunc
	m                  sync.Mutex
}

func NewClient(opts Options) (Client, error) {
	if opts.Addr == "" {
		opts.Addr = ConsulDefaultConfigAddr
	}
	if opts.Prefix == "" {
		opts.Prefix = ConsulDefaultConfiGPrefix
	}
	if opts.ConfigParser == nil {
		opts.ConfigParser = cwutils.DefaultConfigParse()
	}
	if opts.TimeOut == 0 {
		opts.TimeOut = ConsulDefaultTimeout
	}
	if opts.ClientPathFormat == "" {
		opts.ClientPathFormat = ConsulDefaultClientPath
	}
	if opts.ServerPathFormat == "" {
		opts.ServerPathFormat = ConsulDefaultServerPath
	}
	if opts.DataCenter == "" {
		opts.DataCenter = ConsulDefaultDataCenter
	}
	consulClient, err := api.NewClient(&api.Config{
		Address:    opts.Addr,
		Datacenter: opts.DataCenter,
		Token:      opts.Token,
		Namespace:  opts.NamespaceId,
		Partition:  opts.Partition,
	})
	if err != nil {
		return nil, err
	}
	prefixTemplate, err := template.New("prefix").Parse(opts.Prefix)
	if err != nil {
		return nil, err
	}
	serverNameTemplate, err := template.New("serverName").Parse(opts.ServerPathFormat)
	if err != nil {
		return nil, err
	}
	clientNameTemplate, err := template.New("clientName").Parse(opts.ClientPathFormat)
	if err != nil {
		return nil, err
	}
	lconfig := &ListenConfig{
		Type:       WatchByKey,
		DataCenter: opts.DataCenter,
		Token:      opts.Token,
		ConsulAddr: opts.Addr,
		Namespace:  opts.NamespaceId,
		Partition:  opts.Partition,
	}
	c := &client{
		consulCli:          consulClient,
		parser:             opts.ConfigParser,
		consulTimeout:      opts.TimeOut,
		prefixTemplate:     prefixTemplate,
		serverPathTemplate: serverNameTemplate,
		clientPathTemplate: clientNameTemplate,
		lconfig:            lconfig,
		cancelMap:          make(map[string]context.CancelFunc),
	}
	return c, nil
}

// SetParser support customise parser
func (c *client) SetParser(parser cwutils.ConfigParser) {
	c.parser = parser
}

func (c *client) ClientConfigParam(cpc *cwutils.ConfigParamConfig, cfs ...CustomFunction) (Key, error) {
	return c.configParam(cpc, c.clientPathTemplate, cfs...)
}

func (c *client) ServerConfigParam(cpc *cwutils.ConfigParamConfig, cfs ...CustomFunction) (Key, error) {
	return c.configParam(cpc, c.serverPathTemplate, cfs...)
}

// configParam render config parameters. All the parameters can be customized with CustomFunction.
// ConfigParam explain:
//  1. Prefix: KitexConfig by default.
//  2. ServerPath: {{.ServerServiceName}}/{{.Category}} by default.
//     ClientPath: {{.ClientServiceName}}/{{.ServerServiceName}}/{{.Category}} by default.
func (c *client) configParam(cpc *cwutils.ConfigParamConfig, t *template.Template, cfs ...CustomFunction) (Key, error) {
	param := Key{Type: cwutils.JSON}
	var err error
	param.Path, err = c.render(cpc, t)
	if err != nil {
		return param, err
	}
	param.Prefix, err = c.render(cpc, c.prefixTemplate)
	if err != nil {
		return param, err
	}

	for _, cf := range cfs {
		cf(&param)
	}
	return param, nil
}

func (c *client) render(cpc *cwutils.ConfigParamConfig, t *template.Template) (string, error) {
	var tpl bytes.Buffer
	err := t.Execute(&tpl, cpc)
	if err != nil {
		return "", err
	}
	return tpl.String(), nil
}

// RegisterConfigCallback register the callback function to consul client.
func (c *client) RegisterConfigCallback(key string, uniqueID int64, callback func(string, cwutils.ConfigParser)) {
	go func() {
		clientCtx, cancel := context.WithCancel(context.Background())
		params := make(map[string]interface{})
		params["datacenter"] = c.lconfig.DataCenter
		params["token"] = c.lconfig.Token
		params["type"] = c.lconfig.Type
		params["key"] = key
		c.lconfig.Key = key
		kv := c.consulCli.KV()
		get, _, _ := kv.Get(c.lconfig.Key, nil)
		if get == nil {
			klog.Debugf("[consul]  key:%s doesn't exist", key)
			_, err := kv.Put(&api.KVPair{
				Key:   c.lconfig.Key,
				Value: []byte("{}"),
			}, nil)
			if err != nil {
				klog.Errorf("[consul] Add key: %s failed,cwerror: %s", key, err.Error())
			}
		}
		c.registerCancelFunc(key, uniqueID, cancel)
		w, err := watch.Parse(params)
		if err != nil {
			klog.Debugf("[consul] key:add listen for %s failed", key)
		}
		if w == nil {
			klog.Debugf("[consul] key:add listen for %s failed", key)
			return
		}
		klog.Debugf("[consul] key:add listen for %s successfully", key)
		w.Handler = func(u uint64, i interface{}) {
			if i == nil {
				return
			}
			kv := i.(*api.KVPair)
			v := string(kv.Value)
			klog.Debugf("[consul] config key: %s updated,value is %s", key, v)
			callback(v, c.parser)
		}

		go func() {
			err := w.Run(c.lconfig.ConsulAddr)
			if err != nil {
				klog.Errorf("[consul] listen key: %s failed,cwerror: %s", key, err.Error())
			}
		}()
		for range clientCtx.Done() {
			w.Stop()
			return
		}
	}()
	_, cancel := context.WithTimeout(context.Background(), c.consulTimeout)
	defer cancel()
	kv := c.consulCli.KV()
	get, _, err := kv.Get(key, nil)
	if err != nil {
		klog.Debugf("[consul] key: %s config get value failed", c.lconfig.Key)
		return
	}
	if get == nil {
		return
	}
	if get.Value == nil {
		return
	}
	callback(string(get.Value), c.parser)
}

func (c *client) DeregisterConfig(key string, uniqueID int64) {
	c.deregisterCancelFunc(key, uniqueID)
}

func (c *client) deregisterCancelFunc(key string, uniqueID int64) {
	c.m.Lock()
	clientKey := key + "/" + strconv.FormatInt(uniqueID, 10)
	cancel := c.cancelMap[clientKey]
	cancel()
	delete(c.cancelMap, clientKey)
	c.m.Unlock()
}

func (c *client) registerCancelFunc(key string, uniqueID int64, cancel context.CancelFunc) {
	c.m.Lock()
	clientKey := key + "/" + strconv.FormatInt(uniqueID, 10)
	c.cancelMap[clientKey] = cancel
	c.m.Unlock()
}
