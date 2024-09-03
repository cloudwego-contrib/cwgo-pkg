# registry-servicecomb (*This is a community driven project*)

Use [service-comb](https://github.com/apache/servicecomb-service-center) as service registry for `Kitex`.

## How to use?

### Server
```go
import (
    // ...
    "github.com/cloudwego/kitex/pkg/rpcinfo"
    "github.com/cloudwego/kitex/server"
    "github.com/kitex-contrib/registry-servicecomb/registry"
)

// ...

func main() {
    r, err := registry.NewDefaultSCRegistry()
    if err != nil {
        panic(err)
    }
    svr := hello.NewServer(
        new(HelloImpl),
        server.WithRegistry(r),
        server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: "Hello"}),
        server.WithServiceAddr(&net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8080}),
    )
    if err := svr.Run(); err != nil {
        log.Println("server stopped with cwerror:", err)
    } else {
        log.Println("server stopped")
    }
    // ...
}
```

### Client
```go
import (
    // ...
    "github.com/cloudwego/kitex/client"
    "github.com/kitex-contrib/registry-servicecomb/resolver"
)

func main() {
    r, err := resolver.NewDefaultSCResolver()
    if err != nil {
        panic(err)
    }
    newClient := hello.MustNewClient("Hello", client.WithResolver(r))
    // ...
}
```

## Compatibility
Compatible with Service Comb Center v4.

maintained by: [bodhisatan](https://github.com/bodhisatan)