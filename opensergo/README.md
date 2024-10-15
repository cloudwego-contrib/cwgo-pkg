# opensergo (This is a community driven project)

## sentinel adapter
sentinel adapter for hertz

### How to use?

#### Server
```go
h := server.Default()
h.Use(SentinelServerMiddleware(
     WithServerResourceExtractor(func(c context.Context, ctx *app.RequestContext) string {
         return fmt.Sprintf("%v:%v", string(req.Method()), string(req.Path()))
     }),
     WithServerBlockFallback(func(c context.Context, ctx *app.RequestContext) {
         ctx.AbortWithStatusJSON(400, utils.H{
             "err":  "too many request; the quota used up",
             "code": 10222,
         })
     }),
))    
```
#### Client
```go
c, err := client.NewClient()
if err != nil {
    log.Fatalf("Unexpected error: %+v", err)
    return
}
c.Use(SentinelClientMiddleware(
    WithClientResourceExtractor(func(c context.Context, ctx *app.RequestContext) string {
        return "client_test"
    }),
    WithClientBlockFallback(func(c context.Context, ctx *app.RequestContext) {
        ctx.AbortWithStatusJSON(400, utils.H{
            "err":  "too many request; the quota used up",
            "code": 10222,
        }),
    }),
))
```

