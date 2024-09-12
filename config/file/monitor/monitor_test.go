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
	"testing"

	"github.com/cloudwego-contrib/cwgo-pkg/config/file/filewatcher"
	"github.com/cloudwego-contrib/cwgo-pkg/config/file/mock"
	"github.com/cloudwego-contrib/cwgo-pkg/config/file/parser"
)

const (
	filepath = "./../testdata/test.json"
)

func TestNewConfigMonitor(t *testing.T) {
	m := mock.NewMockFileWatcher()
	if _, err := NewConfigMonitor("test", m); err != nil {
		t.Errorf("NewConfigMonitor() error = %v", err)
	}
}

func TestNewConfigMonitorFailed(t *testing.T) {
	m := mock.NewMockFileWatcher()
	if _, err := NewConfigMonitor("", m); err == nil {
		t.Errorf("NewConfigMonitor() error = %v", err)
	}
	if _, err := NewConfigMonitor("test", nil); err == nil {
		t.Errorf("NewConfigMonitor() error = %v", err)
	}
}

func TestKey(t *testing.T) {
	m := mock.NewMockFileWatcher()
	cm, err := NewConfigMonitor("test", m)
	if err != nil {
		t.Errorf("NewConfigMonitor() error = %v", err)
	}
	if cm.Key() != "test" {
		t.Errorf("Key() error")
	}
}

func TestSetManager(t *testing.T) {
	m := mock.NewMockFileWatcher()
	cm, err := NewConfigMonitor("test", m)
	if err != nil {
		t.Errorf("NewConfigMonitor() error = %v", err)
	}
	cm.SetManager(&parser.ServerFileManager{})
}

func TestSetParser(t *testing.T) {
	m := mock.NewMockFileWatcher()
	cm, err := NewConfigMonitor("test", m)
	if err != nil {
		t.Errorf("NewConfigMonitor() error = %v", err)
	}
	cm.SetParser(&parser.Parser{})

	// use json format test ConfigParse
	kind := parser.JSON
	jsonData := []byte(`{"key": "value"}`)
	var config struct{}
	err = cm.ConfigParse(kind, jsonData, &config)
	if err != nil {
		t.Errorf("ConfigParse() error = %v", err)
	}
}

func TestSetParams(t *testing.T) {
	m := mock.NewMockFileWatcher()
	cm, err := NewConfigMonitor("test", m)
	if err != nil {
		t.Errorf("NewConfigMonitor() error = %v", err)
	}

	cm.SetParams(&parser.ConfigParam{})
}

func TestRegisterCallback(t *testing.T) {
	m := mock.NewMockFileWatcher()
	cm, err := NewConfigMonitor("test", m)
	if err != nil {
		t.Errorf("NewConfigMonitor() error = %v", err)
	}
	cm.RegisterCallback(nil)
}

func TestDeregisterCallback(t *testing.T) {
	m := mock.NewMockFileWatcher()
	cm, err := NewConfigMonitor("test", m)
	if err != nil {
		t.Errorf("NewConfigMonitor() error = %v", err)
	}
	cm.DeregisterCallback(1)
}

func TestStartFailed(t *testing.T) {
	m := mock.NewMockFileWatcher()
	cm, err := NewConfigMonitor("test", m)
	if err != nil {
		t.Errorf("NewConfigMonitor() error = %v", err)
	}
	if err := cm.Start(); err == nil {
		t.Errorf("filewatcher not sert manager, Start() should error, but not")
	}
}

func TestStartSuccess(t *testing.T) {
	m := mock.NewMockFileWatcher()
	cm, err := NewConfigMonitor("test", m)
	if err != nil {
		t.Errorf("NewConfigMonitor() error = %v", err)
	}
	cm.SetManager(&parser.ServerFileManager{})
	if err := cm.Start(); err != nil {
		t.Errorf("Start() error = %v", err)
	}
}

func TestStop(t *testing.T) {
	m := mock.NewMockFileWatcher()
	cm, err := NewConfigMonitor("test", m)
	if err != nil {
		t.Errorf("NewConfigMonitor() error = %v", err)
	}
	cm.Stop()
}

func TestEntireProgressWithSingleKey(t *testing.T) {
	testProcess(filepath, []string{"Test1"}, t)
}

func TestEntireProgressWithDifferentKey(t *testing.T) {
	testProcess(filepath, []string{"Test1", "Test2"}, t)
}

func TestEntireProcessWithSameKey(t *testing.T) {
	testProcess(filepath, []string{"Test1", "Test1"}, t)
}

func createConfigManager(fw filewatcher.FileWatcher, key string, t *testing.T) ConfigMonitor {
	// create a config monitor object
	cm, err := NewConfigMonitor(key, fw)
	if err != nil {
		t.Errorf("NewConfigMonitor() error = %v", err)
	}
	// set manager
	cm.SetManager(&parser.ServerFileManager{})

	// register callback
	id := cm.RegisterCallback(func() {
		t.Errorf("THIS CALLBACK SHOULD NOT BE INVOKED")
	})

	cm.RegisterCallback(func() {
		t.Logf("INVOKE CALLBACK ON CONFIG MANAGER, index: %v\n", cm.WatcherID())
	})
	cm.DeregisterCallback(id)

	// start monitoring
	if err = cm.Start(); err != nil {
		t.Errorf("Start() error = %v", err)
	}

	// call specific callback
	t.Log("call specific ConfigManager0")
	fw.CallOnceSpecific(id)

	return cm
}

func testProcess(filepath string, mapKey []string, t *testing.T) {
	if len(mapKey) < 1 {
		t.Errorf("mapKey is empty")
	}

	// create a file watcher object
	fw, err := filewatcher.NewFileWatcher(filepath)
	if err != nil {
		t.Errorf("NewFileWatcher() error = %v", err)
	}
	// start watching file changes
	if err = fw.StartWatching(); err != nil {
		t.Errorf("StartWatching() error = %v", err)
	}

	cm := createConfigManager(fw, mapKey[0], t)

	// not have enough key
	if len(mapKey) < 2 {
		t.Log("Without enough map keys, do not execute multi-key listening test.")
		cm.Stop()
		fw.StopWatching()
		return
	}

	// CREATE ANOTHER NEW CONFIG MONITOR
	cm1 := createConfigManager(fw, mapKey[1], t)

	t.Log("call all ConfigManager")
	fw.CallOnceAll()

	t.Log("DeregisterCallback ConfigManager0 and CallOnceAll")
	cm.Stop()
	fw.CallOnceAll()

	t.Log("DeregisterCallback ConfigManager1 and CallOnceAll")
	cm1.Stop()
	fw.CallOnceAll()

	fw.StopWatching()
}
