package main

import (
	_ "github.com/imlonghao/geo"
	_ "github.com/coredns/coredns/plugin/file"

	"github.com/coredns/coredns/coremain"
	"github.com/coredns/coredns/core/dnsserver"
)

var directives = []string{
	"geo",
	"file",
}

func init() {
	dnsserver.Directives = directives
}

func main() {
	coremain.Run()
}