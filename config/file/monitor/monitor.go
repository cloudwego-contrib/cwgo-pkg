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

package monitor

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/cloudwego-contrib/cwgo-pkg/config/file/filewatcher"
	"github.com/cloudwego-contrib/cwgo-pkg/config/file/parser"
	"github.com/cloudwego-contrib/cwgo-pkg/config/file/utils"
	"github.com/cloudwego/kitex/pkg/klog"
)

type ConfigMonitor interface {
	Key() string
	Config() interface{}
	CallbackSize() int
	Start() error
	WatcherID() int64
	Stop()
	SetManager(manager parser.ConfigManager)
	SetParser(parser parser.ConfigParser)
	SetParams(params *parser.ConfigParam)
	ConfigParse(kind parser.ConfigType, data []byte, config interface{}) error
	RegisterCallback(callback func()) int64
	DeregisterCallback(uniqueID int64)
}

type configMonitor struct {
	// support customise parser
	parser      parser.ConfigParser     // Parser for the config file
	params      *parser.ConfigParam     // params for the config file
	manager     parser.ConfigManager    // Manager for the config file
	config      interface{}             // config details
	fileWatcher filewatcher.FileWatcher // local config file watcher
	callbacks   map[int64]func()        // callbacks when config file changed
	key         string                  // key of the config in the config file
	id          int64                   // unique id for filewatcher to register/deregister
	lock        sync.RWMutex            // mutex
	counter     atomic.Int64            // unique id for callbacks, only increase
}

// NewConfigMonitor init a monitor for the config file
func NewConfigMonitor(key string, watcher filewatcher.FileWatcher, opts ...utils.Option) (ConfigMonitor, error) {
	if key == "" {
		return nil, errors.New("empty config key")
	}
	if watcher == nil {
		return nil, errors.New("filewatcher is nil")
	}

	option := &utils.Options{
		Parser: parser.DefaultConfigParser(),
		Params: parser.DefaultConfigParam(),
	}

	for _, opt := range opts {
		opt(option)
	}

	return &configMonitor{
		fileWatcher: watcher,
		key:         key,
		callbacks:   make(map[int64]func(), 0),
		parser:      option.Parser,
		params:      option.Params,
	}, nil
}

// Key return the key of the config file
func (c *configMonitor) Key() string { return c.key }

// Config return the config details
func (c *configMonitor) Config() interface{} { return c.config }

// CallbackSize return the size of the callbacks
func (c *configMonitor) CallbackSize() int {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return len(c.callbacks)
}

// WatcherID return the unique id of the filewatcher
func (c *configMonitor) WatcherID() int64 { return c.id }

// Start starts the file watch progress
func (c *configMonitor) Start() error {
	if c.manager == nil {
		return errors.New("not set manager for config file")
	}

	c.id = c.fileWatcher.RegisterCallback(c.parseHandler)

	return c.fileWatcher.CallOnceSpecific(c.id)
}

// Stop stops the file watch progress
func (c *configMonitor) Stop() {
	for k := range c.callbacks {
		c.DeregisterCallback(k)
	}

	// deregister current object's parseHandler from filewatcher
	c.fileWatcher.DeregisterCallback(c.id)
}

// SetManager set the manager for the config file
func (c *configMonitor) SetManager(manager parser.ConfigManager) { c.manager = manager }

// SetParser set the parser for the config file
func (c *configMonitor) SetParser(parser parser.ConfigParser) {
	c.parser = parser
}

// SetParams set the params for the config file, such as file type
func (c *configMonitor) SetParams(params *parser.ConfigParam) {
	c.params = params
}

// ConfigParse call configMonitor.parser.Decode()
func (c *configMonitor) ConfigParse(kind parser.ConfigType, data []byte, config interface{}) error {
	return c.parser.Decode(kind, data, config)
}

// RegisterCallback add callback function, it will be called when file changed, return key for deregister
func (c *configMonitor) RegisterCallback(callback func()) int64 {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.callbacks == nil {
		c.callbacks = make(map[int64]func())
	}

	key := c.counter.Add(1)
	c.callbacks[key] = callback

	klog.Debugf("[local] config monitor registered callback, id: %v\n", key)
	return key
}

// DeregisterCallback remove callback function.
func (c *configMonitor) DeregisterCallback(key int64) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if _, exists := c.callbacks[key]; !exists {
		klog.Warnf("[local] ConfigMonitor callback %s not registered", key)
		return
	}
	delete(c.callbacks, key)
}

// parseHandler parse and invoke each function in the callbacks array
func (c *configMonitor) parseHandler(data []byte) {
	resp := c.manager

	kind := c.params
	err := c.parser.Decode(kind.Type, data, resp)
	if err != nil {
		klog.Errorf("[local] failed to parse the config file: %v\n", err)
		return
	}

	c.config = resp.GetConfig(c.key)
	if c.config == nil {
		klog.Warnf("[local] not matching key found, skip. current key: %v\n", c.key)
		return
	}

	if len(c.callbacks) > 0 {
		for key, callback := range c.callbacks {
			if callback == nil {
				c.DeregisterCallback(key) // When encountering Nil's callback function, directly cancel it here.
				klog.Warnf("[local] filewatcher callback %v is nil, deregister it", key)
				continue
			}
			callback()
		}
	}
	klog.Infof("[local] config parse and update complete \n")
}
