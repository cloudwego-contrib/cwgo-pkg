# nacos (*This is a community driven project*)

[中文](README_CN.md)

Nacos as service discovery for Hertz.

## How to use?

### 
- The nacos/v2 version of hertz does not currently support creating multiple port examples in the same group multiple times.
- Service registration and discovery in nacos/v2 is compatible with previous versions.
- `CustomLogger` type in constant.ClientConfig has been removed in nacos-sdk-go v2. Instead, use the `(github.com/nacos-group/nacos-sdk-go/v2/common/logger).SetLogger` to customize the log.
- nacos/v2 only supports nacos 2.X version.

### Server

**[example/standard/server/main.go](examples/standard/server/main.go)**

```go
import (
	"context"
	"log"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/app/server/registry-hertz"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/cloudwego-contrib/cwgo-pkg/registry/nacos/nacoshertz/v2"
)

func main() {
	addr := "127.0.0.1:8888"
	r, err := nacos.NewDefaultNacosRegistry()
	if err != nil {
		log.Fatal(err)
		return
	}
	h := server.Default(
		server.WithHostPorts(addr),
		server.WithRegistry(r, &registry.Info{
			ServiceName: "hertz.test.demo",
			Addr:        utils.NewNetAddr("tcp", addr),
			Weight:      10,
			Tags:        nil,
		}),
	)
	h.GET("/ping", func(c context.Context, ctx *app.RequestContext) {
		ctx.JSON(consts.StatusOK, utils.H{"ping": "pong"})
	})
	h.Spin()
}
```

### Client

**[example/standard/client/main.go](examples/standard/client/main.go)**

```go
import (
	"context"
	"log"

	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/app/middlewares/client/sd"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego-contrib/cwgo-pkg/registry/nacos/nacoshertz/v2"
)

func main() {
client, err := client.NewClient()
	if err != nil {
		panic(err)
	}
	r, err := nacos.NewDefaultNacosResolver()
	if err != nil {
		log.Fatal(err)
		return
	}
	client.Use(sd.Discovery(r))
	// ...
}

```

### Custom Logger

**[examples/logger/main.go](examples/logger/main.go)**

```go
package main

import (
	"github.com/cloudwego-contrib/cwgo-pkg/registry/nacos/nacoshertz/v2/common"
	"github.com/cloudwego-contrib/cwgo-pkg/registry/nacos/nacoshertz/v2/common"
	"github.com/nacos-group/nacos-sdk-go/v2/common/logger"
)

func main() {
	logger.InitLogger(logger.Config{
		Level: "debug",
	})
	logger.SetLogger(common.NewCustomNacosLogger())
	logger.Info("info")
}

```

## How to run example?

### run docker

- make prepare

```bash
make prepare
```

### run server

```go
go run ./examples/standard/server/main.go
```

### run client

```go
go run ./examples/standard/client/main.go
```

```go
2022/07/26 13:52:47.310617 main.go:46: [Info] code = 200, body ={"ping":"pong"}
2022/07/26 13:52:47.311019 main.go:46: [Info] code = 200, body ={"ping":"pong"}
2022/07/26 13:52:47.311186 main.go:46: [Info] code = 200, body ={"ping":"pong"}
2022/07/26 13:52:47.311318 main.go:46: [Info] code = 200, body ={"ping":"pong"}
2022/07/26 13:52:47.311445 main.go:46: [Info] code = 200, body ={"ping":"pong"}
2022/07/26 13:52:47.311585 main.go:46: [Info] code = 200, body ={"ping":"pong"}
2022/07/26 13:52:47.311728 main.go:46: [Info] code = 200, body ={"ping":"pong"}
2022/07/26 13:52:47.311858 main.go:46: [Info] code = 200, body ={"ping":"pong"}
2022/07/26 13:52:47.311977 main.go:46: [Info] code = 200, body ={"ping":"pong"}
2022/07/26 13:52:47.312107 main.go:46: [Info] code = 200, body ={"ping":"pong"}
```

## Custom Nacos Client Configuration

### Server

**[example/custom_config/server/main.go](examples/custom_config/server/main.go)**

```go
import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/app/server/registry-hertz"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/cloudwego-contrib/cwgo-pkg/registry/nacos/nacoshertz/v2"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

func main() {
	sc := []constant.ServerConfig{
		*constant.NewServerConfig("127.0.0.1", 8848),
	}
	
	cc := constant.ClientConfig{
		NamespaceId:         "public",
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "/tmp/nacos/log",
		CacheDir:            "/tmp/nacos/cache",
		LogLevel:            "info",
	}
	
	cli, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		},
	)
	if err != nil {
		panic(err)
	}
	
	addr := "127.0.0.1:8888"
	r := nacos.NewNacosRegistry(cli)
	h := server.Default(
		server.WithHostPorts(addr),
		server.WithRegistry(r, &registry.Info{
			ServiceName: "hertz.test.demo",
			Addr:        utils.NewNetAddr("tcp", addr),
			Weight:      10,
			Tags:        nil,
		}),
	)
	// ...
	h.Spin()
}

```

### Client

**[example/custom_config/client/main.go](examples/custom_config/client/main.go)**

```go
import (
	"context"
	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/app/middlewares/client/sd"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego-contrib/cwgo-pkg/registry/nacos/nacoshertz/v2"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

func main() {
	cli, err := client.NewClient()
	if err != nil {
		panic(err)
	}
	sc := []constant.ServerConfig{
		*constant.NewServerConfig("127.0.0.1", 8848),
	}
	cc := constant.ClientConfig{
		NamespaceId:         "public",
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "/tmp/nacos/log",
		CacheDir:            "/tmp/nacos/cache",
		LogLevel:            "info",
	}

	nacosCli, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		}
  )
	if err != nil {
		panic(err)
	}
	r := nacos.NewNacosResolver(nacosCli)
	cli.Use(sd.Discovery(r))
	// ...
}

```

## Environment Variable

| Environment Variable Name | Environment Variable Default Value | Environment Variable Introduction |
| ------------------------- | ---------------------------------- | --------------------------------- |
| serverAddr               | 127.0.0.1                          | nacos server address              |
| serverPort               | 8848                               | nacos server port                 |
| namespace                 |                                    | the namespaceId of nacos          |


## Compatibility

- This package use Nacos2.x client.

- Nacos2.x detail [see](https://nacos.io/en-us/docs/v2/upgrading/2.0.0-compatibility.html)

- Supported Go version over 1.16

- Supported Nacos version over 2.x
