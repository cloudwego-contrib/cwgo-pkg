// Code generated by Kitex v0.5.1. DO NOT EDIT.
package hello

import (
	api "github.com/cloudwego-contrib/cwgo-pkg/registry/consul/consulkitex/example/hello/kitex_gen/api"
	server "github.com/cloudwego/kitex/server"
)

// NewServer creates a server.Server with the given handler and options.
func NewServer(handler api.Hello, opts ...server.Option) server.Server {
	var options []server.Option

	options = append(options, opts...)

	svr := server.NewServer(options...)
	if err := svr.RegisterService(serviceInfo(), handler); err != nil {
		panic(err)
	}
	return svr
}
