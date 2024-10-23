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

package filewatcher

import (
	"errors"
	"os"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/cloudwego-contrib/cwgo-pkg/config/file/utils"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/fsnotify/fsnotify"
)

type FileWatcher interface {
	FilePath() string
	CallbackSize() int
	RegisterCallback(callback func(data []byte)) int64
	DeregisterCallback(uniqueID int64)
	StartWatching() error
	StopWatching()
	CallOnceAll() error
	CallOnceSpecific(uniqueID int64) error
}

// FileWatcher is used for file monitoring
type fileWatcher struct {
	filePath  string                      // The path to the file to be monitored.
	callbacks map[int64]func(data []byte) // Custom functions to be executed when the file changes.
	watcher   *fsnotify.Watcher           // fsnotify file change watcher.
	done      chan struct{}               // A channel for signaling the watcher to stop.
	lock      sync.RWMutex                // mutex
	counter   atomic.Int64                // unique id for callbacks, only increase
}

// NewFileWatcher creates a new FileWatcher instance.
// filePath should be a path to a file, not a directory.
func NewFileWatcher(filePath string) (FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	exist, err := utils.PathExists(filePath)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, errors.New("file [" + filePath + "] not exist")
	}

	fw := &fileWatcher{
		filePath:  filePath,
		watcher:   watcher,
		done:      make(chan struct{}),
		callbacks: make(map[int64]func(data []byte), 0),
	}

	return fw, nil
}

// FilePath returns the file address that the current object is listening to
func (fw *fileWatcher) FilePath() string { return fw.filePath }

// CallbackSize returns the number of callback functions.
func (fw *fileWatcher) CallbackSize() int {
	fw.lock.RLock()
	defer fw.lock.RUnlock()
	return len(fw.callbacks)
}

// RegisterCallback sets the callback function.
func (fw *fileWatcher) RegisterCallback(callback func(data []byte)) int64 {
	fw.lock.Lock()
	defer fw.lock.Unlock()

	if fw.callbacks == nil {
		fw.callbacks = make(map[int64]func(data []byte), 0)
	}

	klog.Debugf("[local] filewatcher to %v registered callback\n", fw.filePath)

	uniqueID := fw.counter.Add(1)
	fw.callbacks[uniqueID] = callback
	return uniqueID
}

// DeregisterCallback remove callback function.
func (fw *fileWatcher) DeregisterCallback(uniqueID int64) {
	fw.lock.Lock()
	defer fw.lock.Unlock()

	if _, exists := fw.callbacks[uniqueID]; !exists {
		klog.Warnf("[local] FileWatcher callback %s not registered", uniqueID)
		return
	}
	delete(fw.callbacks, uniqueID)
	klog.Infof("[local] filewatcher to %v deregistered callback id: %v\n", fw.filePath, uniqueID)
}

// Start starts monitoring file changes.
// This method will add the file to be monitored to the watcher and start the monitoring process instanctly.
func (fw *fileWatcher) StartWatching() error {
	fw.lock.Lock()
	if err := fw.watcher.Add(fw.filePath); err != nil {
		return err
	}
	fw.lock.Unlock()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				klog.Errorf("[local] file watcher panic: %v\n", r)
			}
		}()
		fw.start()
	}()

	return nil
}

// StopWatching Stop stops monitoring file changes.
// Stop watching will close the done channel, and do not restart again.
func (fw *fileWatcher) StopWatching() {
	fw.lock.Lock()
	defer fw.lock.Unlock()
	klog.Infof("[local] stop watching file: %s", fw.filePath)
	close(fw.done)
}

// start responsible for handling fsnotify event information.
func (fw *fileWatcher) start() {
	defer fw.watcher.Close()
	for {
		select {
		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}
			if event.Has(fsnotify.Write) {
				if err := fw.CallOnceAll(); err != nil {
					klog.Errorf("[local] read config file failed: %v\n", err)
				}
			}
			if event.Has(fsnotify.Remove) {
				klog.Warnf("[local] file %s is removed, stop watching", fw.filePath)
				fw.StopWatching()
			}
		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
			klog.Errorf("[local] file watcher meet error: %v\n", err)
		case <-fw.done:
			return
		}
	}
}

// CallOnceAll calls the callback function list once.
func (fw *fileWatcher) CallOnceAll() error {
	data, err := os.ReadFile(fw.filePath)
	if err != nil {
		return err
	}

	for key, callback := range fw.callbacks {
		if callback == nil {
			fw.DeregisterCallback(key) // When encountering Nil's callback function, directly cancel it here.
			klog.Warnf("[local] filewatcher callback %v is nil, deregister it", key)
			continue
		}
		callback(data)
	}
	return nil
}

// CallOnceSpecific calls the callback function once by uniqueID.
func (fw *fileWatcher) CallOnceSpecific(uniqueID int64) error {
	data, err := os.ReadFile(fw.filePath)
	if err != nil {
		return err
	}

	if callback, ok := fw.callbacks[uniqueID]; ok {
		callback(data)
	} else {
		return errors.New("not found callback for id: " + strconv.FormatInt(uniqueID, 10))
	}
	return nil
}
