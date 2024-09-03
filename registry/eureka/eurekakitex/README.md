# registry-eureka (*This is a community driven project*)

## How to use?

### Server

```go
import (
    ...
    euregistry "github.com/kitex-contrib/registry-eureka/registry"
    "github.com/cloudwego/kitex/server"
    "github.com/cloudwego/kitex/pkg/rpcinfo"
    ...
)

func main() {
    ...
    r = euregistry.NewEurekaRegistry([]string{"http://127.0.0.1:8080/eureka"}, 15*time.Second)
	svr := echo.NewServer(new(EchoImpl), server.WithRegistry(r),
    server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: "test"}), 
    )
    if err := svr.Run(); err != nil {
        log.Println("server stopped with cwerror:", err)
    } else {
        log.Println("server stopped")
    }
    ...
}
```

### Client

```go
import (
    ...
    "github.com/kitex-contrib/registry-eureka/resolver"
    "github.com/cloudwego/kitex/client"
    ...
)

func main() {
    ...
    r = resolver.NewEurekaResolver([]string{"http://127.0.0.1:8080/eureka"})
    client, err := echo.NewClient("echo", 
        client.WithResolver(r),
    )
    if err != nil {
        log.Fatal(err)
    }
    ...
}
```

## Authentication

use `HTTP Basic Auth`

### Server

```go
import (
    ...
    euregistry "github.com/kitex-contrib/registry-eureka/registry"
    "github.com/cloudwego/kitex/server"
    "github.com/cloudwego/kitex/pkg/rpcinfo"
    ...
)

func main() {
    ...
    r = euregistry.NewEurekaRegistry([]string{"http://username:password@127.0.0.1:8080/eureka"}, 15*time.Second)
	svr := echo.NewServer(new(EchoImpl), server.WithRegistry(r),
    ...
}
```

### Client

```go
import (
    ...
    "github.com/kitex-contrib/registry-eureka/resolver"
    "github.com/cloudwego/kitex/client"
    ...
)

func main() {
    ...
    r = resolver.NewEurekaResolver([]string{"http://username:password@127.0.0.1:8080/eureka"})
    client, err := echo.NewClient("echo", 
        client.WithResolver(r),
    )
    ...
}
```

## Test

use `spring-cloud-starter-netflix-eureka-server` in Java.

```xml
<dependency>
    <groupId>org.springframework.cloud</groupId>
    <artifactId>spring-cloud-starter-netflix-eureka-server</artifactId>
    <version>2.1.0.RELEASE</version>
</dependency>
```
use `docker-compose up` and run unit tests.

## How to set log level

Eureka client for this project uses [fargo](https://github.com/hudl/fargo) ,fargo use `github.com/op/go-logging` as logger but does not provide an interface to adjust the log level.

Here are instructions on how to adjust the log level.

```go
import (
    ...
    "github.com/op/go-logging"
    "github.com/kitex-contrib/registry-eureka/resolver"
    "github.com/cloudwego/kitex/client"
    ...
)

func main() {
...
    // set log level
    logging.SetLevel(logging.WARNING, "fargo")
    logging.SetLevel(logging.WARNING, "fargo.metadata")
    logging.SetLevel(logging.WARNING, "fargo.marshal")

    r = resolver.NewEurekaResolver([]string{"http://127.0.0.1:8080/eureka"})
    client, err := echo.NewClient("echo",
    client.WithResolver(r))
    if err != nil {
    	log.Fatal(err)
    }
...
}
```


## More info

See example.

## Compatibility
Compatible with eureka server v1.

maintained by: [kinggo](https://github.com/li-jin-gou)