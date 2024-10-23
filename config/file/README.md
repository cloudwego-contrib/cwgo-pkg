# config-file (*This is a community driven project*)

[中文](README_CN.md)

Read, load, and listen to local configuration files

## Usage

### Supported file types

| json | yaml |
| ---  | --- |
| &#10004; | &#10004; |

### Basic

#### Server

```go
package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/cloudwego/kitex-examples/kitex_gen/api"
	"github.com/cloudwego/kitex-examples/kitex_gen/api/echo"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	kitexserver "github.com/cloudwego/kitex/server"
	"github.com/cloudwego-contrib/cwgo-pkg/config/file/filewatcher"
	"github.com/cloudwego-contrib/cwgo-pkg/config/file/parser"
	fileserver "github.com/cloudwego-contrib/cwgo-pkg/config/file/server"
)

var _ api.Echo = &EchoImpl{}

const (
	filepath    = "kitex_server.json"
	key         = "ServiceName"
	serviceName = "ServiceName"
)

// EchoImpl implements the last service interface defined in the IDL.
type EchoImpl struct{}

// Echo implements the Echo interface.
func (s *EchoImpl) Echo(ctx context.Context, req *api.Request) (resp *api.Response, err error) {
	klog.Info("echo called")
	return &api.Response{Message: req.Message}, nil
}

// customed by user
type MyParser struct{}

// one example for custom parser
// if the type of server config is json or yaml,just using default parser
func (p *MyParser) Decode(kind parser.ConfigType, data []byte, config interface{}) error {
	return json.Unmarshal(data, config)
}

func main() {
	klog.SetLevel(klog.LevelDebug)

	// create a file watcher object
	fw, err := filewatcher.NewFileWatcher(filepath)
	if err != nil {
		panic(err)
	}
	// start watching file changes
	if err = fw.StartWatching(); err != nil {
		panic(err)
	}
	defer fw.StopWatching()

	svr := echo.NewServer(
		new(EchoImpl),
		kitexserver.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: serviceName}),
		kitexserver.WithSuite(fileserver.NewSuite(key, fw)), // add watcher
	)
	if err := svr.Run(); err != nil {
		log.Println("server stopped with cwerror:", err)
	} else {
		log.Println("server stopped")
	}
}
```

#### Client

```go
package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/cloudwego/kitex-examples/kitex_gen/api"
	"github.com/cloudwego/kitex-examples/kitex_gen/api/echo"
	kitexclient "github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/klog"
	fileclient "github.com/cloudwego-contrib/cwgo-pkg/config/file/client"
	"github.com/cloudwego-contrib/cwgo-pkg/config/file/filewatcher"
	"github.com/cloudwego-contrib/cwgo-pkg/config/file/parser"
)

const (
	filepath    = "kitex_client.json"
	key         = "ClientName/ServiceName"
	serviceName = "ServiceName"
	clientName  = "echo"
)

// customed by user
type MyParser struct{}

// one example for custom parser
// if the type of client config is json or yaml,just using default parser
func (p *MyParser) Decode(kind parser.ConfigType, data []byte, config interface{}) error {
	return json.Unmarshal(data, config)
}

func main() {
	klog.SetLevel(klog.LevelDebug)

	// create a file watcher object
	fw, err := filewatcher.NewFileWatcher(filepath)
	if err != nil {
		panic(err)
	}
	// start watching file changes
	if err = fw.StartWatching(); err != nil {
		panic(err)
	}

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, os.Kill)
		<-sig
		fw.StopWatching()
		os.Exit(1)
	}()

	client, err := echo.NewClient(
		serviceName,
		kitexclient.WithHostPorts("0.0.0.0:8888"),
		kitexclient.WithSuite(fileclient.NewSuite(serviceName, key, fw)),
	)
	if err != nil {
		log.Fatal(err)
	}

	for {
		req := &api.Request{Message: "my request"}
		resp, err := client.Echo(context.Background(), req)
		if err != nil {
			klog.Errorf("take request cwerror: %v", err)
		} else {
			klog.Infof("receive response %v", resp)
		}
		time.Sleep(time.Second * 10)
	}
}
```

#### File Configuration

##### Custom Parser

The configuration format supports `json` and `yaml` by default. You can implement a custom parser by implementing the `ConfigParser` interface and pass in a custom function in `NewSuite`

Interface definition:
```go
// ConfigParser the parser for config file.
type ConfigParser interface {
	Decode(kind ConfigType, data []byte, config interface{}) error
}
```

Custom function example:
```go
func withParser(o *utils.Options) {
	o.Parser = &MyParser{}
	o.Params = &parser.ConfigParam{
		Type: parser.YAML,
	}
}
```

#### Governance Policy
> The service name is `ServiceName` and the client name is `ClientName`.

##### Rate Limit Category=limit
> Currently, current limiting only supports the server side, so ClientServiceName is empty.

[JSON Schema](https://github.com/cloudwego/kitex/blob/develop/pkg/limiter/item_limiter.go#L33)

|Variable|Introduction|
|----|----|
|connection_limit| Maximum concurrent connections |
|qps_limit| Maximum request number every 100ms |

Example:
```json
{
    "ServiceName": {
        "limit": {
            "connection_limit": 300,
            "qps_limit": 200
        }
    }
}
```

Note:

- The granularity of the current limit configuration is server global, regardless of client or method.
- Not configured or value is 0 means not enabled.
- connection_limit and qps_limit can be configured independently, e.g. connection_limit = 100, qps_limit = 0

##### Retry Policy Category=retry
[JSON Schema](https://github.com/cloudwego/kitex/blob/develop/pkg/retry/policy.go#L63)

|Variable|Introduction|
|----|----|
|type| 0: failure_policy 1: backup_policy|
|failure_policy.backoff_policy| Can only be set one of `fixed` `none` `random` |

Example：
```json
{
    "ClientName/ServiceName": {
        "retry": {
            "*": {
                "enable": true,
                "type": 0,
                "failure_policy": {
                    "stop_policy": {
                        "max_retry_times": 3,
                        "max_duration_ms": 2000,
                        "cb_policy": {
                            "error_rate": 0.2
                        }
                    }
                }
            },
            "Echo": {
                "enable": true,
                "type": 1,
                "backup_policy": {
                    "retry_delay_ms": 200,
                    "stop_policy": {
                        "max_retry_times": 2,
                        "max_duration_ms": 1000,
                        "cb_policy": {
                            "error_rate": 0.3
                        }
                    }
                }
            }
        }
    }
}
```
Note: retry.Container has built-in support for specifying the default configuration using the `*` wildcard (see the [getRetryer](https://github.com/cloudwego/kitex/blob/v0.5.1/pkg/retry/retryer.go#L240) method for details).

##### RPC Timeout Category=rpc_timeout

[JSON Schema](https://github.com/cloudwego/kitex/blob/develop/pkg/rpctimeout/item_rpc_timeout.go#L42)

Example：
```json
{
    "ClientName/ServiceName": {
        "timeout": {
            "*": {
                "conn_timeout_ms": 100,
                "rpc_timeout_ms": 2000
            },
            "Pay": {
                "conn_timeout_ms": 50,
                "rpc_timeout_ms": 1000
            }
        },
    }
}
```

##### Circuit Break: Category=circuit_break

[JSON Schema](https://github.com/cloudwego/kitex/blob/develop/pkg/circuitbreak/item_circuit_breaker.go#L30)

|Variable|Introduction|
|----|----|
|min_sample| Minimum statistical sample number|

The echo method uses the following configuration (0.3, 100) and other methods use the global default configuration (0.5, 200)

Example：
```json

{
    "ClientName/ServiceName": {
        "circuitbreaker": {
            "Echo": {
                "enable": true,
                "err_rate": 0.3,
                "min_sample": 100
            }
        },
    }
}
```

Note: The circuit breaker implementation of kitex does not currently support changing the global default configuration (see [initServiceCB](https://github.com/cloudwego/kitex/blob/v0.5.1/pkg/circuitbreak/cbsuite.go#L195) for details).
### More Info

Refer to [example](https://github.com/kitex-contrib/config-file/tree/main/example) for more usage.

## Note

For client configuration, you should write all their configurations in the same pair of `$UserServiceName/$ServerServiceName`, for example

```json
{
    "ClientName/ServiceName": {
        "timeout": {
            "*": {
                "conn_timeout_ms": 100,
                "rpc_timeout_ms": 2000
            },
            "Pay": {
                "conn_timeout_ms": 50,
                "rpc_timeout_ms": 1000
            }
        },
        "circuitbreaker": {
            "Echo": {
                "enable": true,
                "err_rate": 0.3,
                "min_sample": 100
            }
        },
        "retry": {
            "*": {
                "enable": true,
                "type": 0,
                "failure_policy": {
                    "stop_policy": {
                        "max_retry_times": 3,
                        "max_duration_ms": 2000,
                        "cb_policy": {
                            "error_rate": 0.2
                        }
                    }
                }
            },
            "Echo": {
                "enable": true,
                "type": 1,
                "backup_policy": {
                    "retry_delay_ms": 200,
                    "stop_policy": {
                        "max_retry_times": 2,
                        "max_duration_ms": 1000,
                        "cb_policy": {
                            "error_rate": 0.3
                        }
                    }
                }
            }
        }
    }
}
```