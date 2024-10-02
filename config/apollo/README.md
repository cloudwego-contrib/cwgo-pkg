# config-apollo (*This is a community driven project*)

[中文](README_CN.md)

Apollo as config centre.

## How to use?

### Basic

#### Server

```go
package main

import (
	"context"
	"log"
	"net"

	cwserver "github.com/cloudwego-contrib/cwgo-pkg/config/apollo/server"

	"github.com/cloudwego-contrib/cwgo-pkg/config/apollo/apollo"
	"github.com/cloudwego-contrib/cwgo-pkg/config/apollo/utils"
	"github.com/cloudwego/kitex-examples/kitex_gen/api"
	"github.com/cloudwego/kitex-examples/kitex_gen/api/echo"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
)

// Customed by user
type configLog struct{}

func (cl *configLog) Apply(opt *utils.Options) {
	fn := func(cp *apollo.ConfigParam) {
		klog.Infof("apollo config %v", cp)
	}
	opt.ApolloCustomFunctions = append(opt.ApolloCustomFunctions, fn)
}

var _ api.Echo = &EchoImpl{}

// EchoImpl implements the last service interface defined in the IDL.
type EchoImpl struct{}

// Echo implements the Echo interface.
func (s *EchoImpl) Echo(ctx context.Context, req *api.Request) (resp *api.Response, err error) {
	klog.Info("echo called")
	return &api.Response{Message: req.Message}, nil
}

func main() {
	klog.SetLevel(klog.LevelDebug)
	apolloClient, err := apollo.NewClient(apollo.Options{})
	if err != nil {
		panic(err)
	}
	serviceName := "ServiceName" // server-side service name
	addr, err := net.ResolveTCPAddr("tcp", "localhost:8899")
	if err != nil {
		panic(err)
	}

	cl := &configLog{}

	svr := echo.NewServer(
		new(EchoImpl),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: serviceName}),
		server.WithSuite(cwserver.NewSuite(serviceName, apolloClient, cl)),
		server.WithServiceAddr(addr),
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
	"log"
	"time"

	"github.com/cloudwego-contrib/cwgo-pkg/config/apollo/apollo"
	cwclient "github.com/cloudwego-contrib/cwgo-pkg/config/apollo/client"

	"github.com/cloudwego/kitex-examples/kitex_gen/api"
	"github.com/cloudwego/kitex-examples/kitex_gen/api/echo"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/klog"

	"github.com/cloudwego-contrib/cwgo-pkg/config/apollo/utils"
)

// Customed by user
type configLog struct{}

func (cl *configLog) Apply(opt *utils.Options) {
	fn := func(cp *apollo.ConfigParam) {
		klog.Infof("apollo config %v", cp)
	}
	opt.ApolloCustomFunctions = append(opt.ApolloCustomFunctions, fn)
}

func main() {
	klog.SetLevel(klog.LevelDebug)
	apolloClient, err := apollo.NewClient(apollo.Options{})
	if err != nil {
		panic(err)
	}
	cl := &configLog{}

	serviceName := "ServiceName" // your server-side service name
	clientName := "ClientName"   // your client-side service name
	client, err := echo.NewClient(
		serviceName,
		client.WithHostPorts("localhost:8899"),
		client.WithSuite(cwclient.NewSuite(serviceName, clientName, apolloClient, cl)),
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

### Apollo Configuration

Initialize the client according to the parameters in Options. After establishing the link, the suite will subscribe to the corresponding configuration based on 'AppId', 'NameSpace', 'ServiceName', or 'ClientName' and dynamically update its own policy. Please refer to the following ·Options· variables for specific parameters.

The configured format supports' json 'by default, and can be customized using the function [SetParser] for format parsing. In' NewSuite ', the field function in the instance that implements the' Option 'interface is used to modify the format of the subscription function.
####

#### CustomFunction

Allow users to use instances of custom implementation Option interfaces to customize Apollo parameters

#### Options Variable

| 参数            |                  变量默认值                   | 作用                                                         |
| :-------------- | :-------------------------------------------: | ------------------------------------------------------------ |
| ConfigServerURL |                127.0.0.1:8080                 | apollo config service address                                |
| AppID           |                   KitexApp                    | appid of apollo (Uniqueness constraint / Length limit of 32 characters) |
| ClientKeyFormat | {{.ClientServiceName}}.{{.ServerServiceName}} | Using the go [template](https://pkg.go.dev/text/template) syntax to render and generate the corresponding ID, using two metadata: `ClientServiceName` and `ServiceName` (Length limit of 128 characters) |
| ServerKeyFormat |            {{.ServerServiceName}}             | Using the go [template](https://pkg.go.dev/text/template) Syntax rendering generates corresponding IDs, using 'ServiceName' as a single metadata (Length limit of 128 characters) |
| Cluster         |                    default                    | Using default values, users can assign values as needed (Length limit of 32 characters) |

#### Governance Policy

> The namespace in the following example uses fixed policy values, with default values for APPID and Cluster. The service name is ServiceName and the client name is ClientName

##### Rate Limit Category=limit
> Currently, current limiting only supports the server side, so ClientServiceName is empty.

[JSON Schema](https://github.com/cloudwego/kitex/blob/develop/pkg/limiter/item_limiter.go#L33)

|Variable|Introduction|
|----|----|
|connection_limit| Maximum concurrent connections | 
|qps_limit| Maximum request number every 100ms | 
Example:
```json
namespace: `limit`
key: `ServiceName`

{
  "connection_limit": 100,
  "qps_limit": 2000
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
namespace: `retry`
key: `ClientName.ServiceName`
{
    "*": {  
        "enable": true,
        "type": 0,                 
        "failure_policy": {
            "stop_policy": {
                "max_retry_times": 3,
                "max_duration_ms": 2000,
                "cb_policy": {
                    "error_rate": 0.3
                }
            },
            "backoff_policy": {
                "backoff_type": "fixed", 
                "cfg_items": {
                    "fix_ms": 50
                }
            },
            "retry_same_node": false
        }
    },
    "echo": { 
        "enable": true,
        "type": 1,                 
        "backup_policy": {
            "retry_delay_ms": 100,
            "retry_same_node": false,
            "stop_policy": {
                "max_retry_times": 2,
                "max_duration_ms": 300,
                "cb_policy": {
                    "error_rate": 0.2
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
namespace: `rpc_timeout`
key: `ClientName.ServiceName`
{
  "*": {
    "conn_timeout_ms": 100,
    "rpc_timeout_ms": 3000
  },
  "echo": {
    "conn_timeout_ms": 50,
    "rpc_timeout_ms": 1000
  }
}
```
Note: The circuit breaker implementation of kitex does not currently support changing the global default configuration (see [initServiceCB](https://github.com/cloudwego/kitex/blob/v0.5.1/pkg/circuitbreak/cbsuite.go#L195) for details).

##### Circuit Break: Category=circuit_break

[JSON Schema](https://github.com/cloudwego/kitex/blob/develop/pkg/circuitbreak/item_circuit_breaker.go#L30)

|Variable|Introduction|
|----|----|
|min_sample| Minimum statistical sample number| 
Example：
```json
The echo method uses the following configuration (0.3, 100) and other methods use the global default configuration (0.5, 200)
namespace: `circuit_break`
key: `ClientName.ServiceName`
{
  "echo": {
    "enable": true,
    "err_rate": 0.3, 
    "min_sample": 100 
  }
}
```
### More Info

Refer to [example](https://github.com/kitex-contrib/examples/config/apollo) for more usage.

maintained by: [xz](https://github.com/XZ0730)

