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

package etcd

import (
	"bytes"
	"context"
	cwutils "github.com/cloudwego-contrib/cwgo-pkg/config/utils"
	"strconv"
	"sync"
	"text/template"
	"time"

	"go.etcd.io/etcd/api/v3/mvccpb"

	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/cloudwego/kitex/pkg/klog"
	"go.uber.org/zap"
)

type Key struct {
	Prefix string
	Path   string
}

type Client interface {
	SetParser(cwutils.ConfigParser)
	ClientConfigParam(cpc *cwutils.ConfigParamConfig, cfs ...CustomFunction) (Key, error)
	ServerConfigParam(cpc *cwutils.ConfigParamConfig, cfs ...CustomFunction) (Key, error)
	RegisterConfigCallback(ctx context.Context, key string, clientId int64, callback func(restoreDefault bool, data string, parser cwutils.ConfigParser))
	DeregisterConfig(key string, uniqueId int64)
}

type client struct {
	ecli *clientv3.Client
	// support customise parser
	parser             cwutils.ConfigParser
	etcdTimeout        time.Duration
	prefixTemplate     *template.Template
	serverPathTemplate *template.Template
	clientPathTemplate *template.Template
	cancelMap          map[string]context.CancelFunc
	m                  sync.Mutex
}

// Options etcd config options. All the fields have default value.
type Options struct {
	Node             []string
	Prefix           string
	ServerPathFormat string
	ClientPathFormat string
	Timeout          time.Duration
	LoggerConfig     *zap.Config
	ConfigParser     cwutils.ConfigParser
}

// NewClient Create a default etcd client
// It can create a client with default config by env variable.
// See: env.go
func NewClient(opts Options) (Client, error) {
	if opts.Node == nil {
		opts.Node = []string{EtcdDefaultNode}
	}
	if opts.ConfigParser == nil {
		opts.ConfigParser = cwutils.DefaultConfigParse()
	}
	if opts.Prefix == "" {
		opts.Prefix = EtcdDefaultConfigPrefix
	}
	if opts.Timeout == 0 {
		opts.Timeout = EtcdDefaultTimeout
	}
	if opts.ServerPathFormat == "" {
		opts.ServerPathFormat = EtcdDefaultServerPath
	}
	if opts.ClientPathFormat == "" {
		opts.ClientPathFormat = EtcdDefaultClientPath
	}
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints: opts.Node,
		LogConfig: opts.LoggerConfig,
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
	c := &client{
		ecli:               etcdClient,
		parser:             opts.ConfigParser,
		etcdTimeout:        opts.Timeout,
		prefixTemplate:     prefixTemplate,
		serverPathTemplate: serverNameTemplate,
		clientPathTemplate: clientNameTemplate,
		cancelMap:          make(map[string]context.CancelFunc),
	}
	return c, nil
}

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
	param := Key{}

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

// RegisterConfigCallback register the callback function to etcd client.
func (c *client) RegisterConfigCallback(ctx context.Context, key string, uniqueID int64, callback func(bool, string, cwutils.ConfigParser)) {
	go func() {
		clientCtx, cancel := context.WithCancel(context.Background())
		c.registerCancelFunc(key, uniqueID, cancel)
		watchChan := c.ecli.Watch(ctx, key)
		for {
			select {
			case <-clientCtx.Done():
				return
			case watchResp := <-watchChan:
				for _, event := range watchResp.Events {
					eventType := event.Type
					// check the event type
					if eventType == mvccpb.PUT {
						// config is updated
						value := string(event.Kv.Value)
						klog.Debugf("[etcd] config key: %s updated,value is %s", key, value)
						callback(false, value, c.parser)
					} else if eventType == mvccpb.DELETE {
						// config is deleted
						klog.Debugf("[etcd] config key: %s deleted", key)
						callback(true, "", c.parser)
					}
				}
			}
		}
	}()
	ctx2, cancel := context.WithTimeout(context.Background(), c.etcdTimeout)
	defer cancel()
	data, err := c.ecli.Get(ctx2, key)
	// the etcd client has handled the not exist cwerror.
	if err != nil {
		klog.Debugf("[etcd] key: %s config get value failed", key)
		return
	}
	if data.Count == 0 {
		return
	}
	callback(false, string(data.Kvs[0].Value), c.parser)
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
