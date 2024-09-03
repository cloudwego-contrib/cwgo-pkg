# config-etcd (*This is a community driven project*)

[中文](./README_CN.md)

etcd as config center for service governance.

## Usage

### Basic

#### Server

```go
package main

import (
	"context"
	"log"

	"github.com/cloudwego/kitex-examples/kitex_gen/api"
	"github.com/cloudwego/kitex-examples/kitex_gen/api/echo"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	"github.com/kitex-contrib/config-etcd/etcd"
	etcdServer "github.com/kitex-contrib/config-etcd/server"
)

var _ api.Echo = &EchoImpl{}

// EchoImpl implements the last service interface defined in the IDL.
type EchoImpl struct{}

// Echo implements the Echo interface.
func (s *EchoImpl) Echo(ctx context.Context, req *api.Request) (resp *api.Response, err error) {
	klog.Info("echo called")
	return &api.Response{Message: req.Message}, nil
}

func main() {
	serviceName := "ServiceName"// your server-side service name
	etcdClient, err := etcd.NewClient(etcd.Options{})
	if err!=nil {
		panic(err)
	}
	svr := echo.NewServer(
		new(EchoImpl),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: serviceName}),
		server.WithSuite(etcdServer.NewSuite(serviceName, etcdClient)),
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

	"github.com/cloudwego/kitex-examples/kitex_gen/api"
	"github.com/cloudwego/kitex-examples/kitex_gen/api/echo"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/klog"
	etcdclient "github.com/kitex-contrib/config-etcd/client"
	"github.com/kitex-contrib/config-etcd/etcd"
	"github.com/kitex-contrib/config-etcd/utils"
)

type configLog struct{}

func (cl *configLog) Apply(opt *utils.Options) {
	fn := func(k *etcd.Key) {
		klog.Infof("etcd config %v", k)
	}
	opt.EtcdCustomFunctions = append(opt.EtcdCustomFunctions, fn)
}

func main() {
	etcdClient, err := etcd.NewClient(etcd.Options{})
	if err != nil {
		panic(err)
	}

	cl := &configLog{}

	serviceName := "ServiceName"// your server-side service name
	clientName := "ClientName"// your client-side service name
	client, err := echo.NewClient(
		serviceName,
		client.WithHostPorts("0.0.0.0:8888"),
		client.WithSuite(etcdclient.NewSuite(serviceName, clientName, etcdClient, cl)),
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

### Etcd Configuration

#### CustomFunction

Provide the mechanism to custom the etcd parameter `Key`.
```go
type Key struct {
    Prefix string
    Path   string
}
```

#### Options Variable

| Variable Name    | Default Value                                               | Introduction                                                                                                                                                                                    |
|------------------|-------------------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Node             | 127.0.0.1:2379                                              | Etcd server nodes                                                                                                                                                                               |
| Prefix           | /KitexConfig                                                | The prefix of Etcd                                                                                                                                                                              |
| ClientPathFormat | {{.ClientServiceName}}/{{.ServerServiceName}}/{{.Category}} | Use go [template](https://pkg.go.dev/text/template) syntax rendering to generate the appropriate ID, and use `ClientServiceName` `ServiceName` `Category` three metadata that can be customised |
| ServerPathFormat | {{.ServerServiceName}}/{{.Category}}                        | Use go [template](https://pkg.go.dev/text/template) syntax rendering to generate the appropriate ID, and use `ServiceName` `Category` two metadatas that can be customised                      |
| Timeout          | 5 * time.Second                                             | five seconds timeout                                                                                                                                                                            |
| LoggerConfig     | NULL                                                        | Default Logger                                                                                                                                                                                  |
| ConfigParser     | defaultConfigParser                                         | The default is the parser that parses json                                                                                                                                                      |

#### Governance Policy
> The configPath and configPrefix in the following example use default values, the service name is `ServiceName` and the client name is `ClientName`.

##### Rate Limit Category=limit
> Currently, current limiting only supports the server side, so ClientServiceName is empty.

[JSON Schema](https://github.com/cloudwego/kitex/blob/develop/pkg/limiter/item_limiter.go#L33)

| Variable         | Introduction                       |
|------------------|------------------------------------|
| connection_limit | Maximum concurrent connections     | 
| qps_limit        | Maximum request number every 100ms | 

Example:

> configPath: /KitexConfig/ServiceName/limit

```json
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

| Variable                      | Introduction                                   |
|-------------------------------|------------------------------------------------|
| type                          | 0: failure_policy 1: backup_policy             | 
| failure_policy.backoff_policy | Can only be set one of `fixed` `none` `random` | 

Example：

> configPath: /KitexConfig/ClientName/ServiceName/retry

```json
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

> configPath: /KitexConfig/ClientName/ServiceName/rpc_timeout

```json
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

| Variable   | Introduction                      |
|------------|-----------------------------------|
| min_sample | Minimum statistical sample number | 

Example：

The echo method uses the following configuration (0.3, 100) and other methods use the global default configuration (0.5, 200)

> configPath: /KitexConfig/ClientName/ServiceName/circuit_break

```json
{
  "echo": {
    "enable": true,
    "err_rate": 0.3, 
    "min_sample": 100 
  }
}
```
##### Degradation: Category=degradation

[JSON Schema](https://github.com/cloudwego/kitex/blob/develop/pkg/circuitbreak/item_circuit_breaker.go#L30)

| Variable   | Introduction                       |
|------------|------------------------------------|
| enable     | Whether to enable degradation      |
| percentage | The percentage of dropped requests | 

Example：

> configPath: /KitexConfig/ClientName/ServiceName/degradation

```json
{
  "enable": true,
  "percentage": 30
}
```
Note: Degradation is not enabled by default.
### More Info

Refer to [example](https://github.com/kitex-contrib/config-etcd/tree/main/example) for more usage.

## Compatibility

For grpc compatibility, the version of Go must >=1.19 

maintained by: [ViolaPioggia](https://github.com/violapioggia)

