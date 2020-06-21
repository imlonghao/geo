package main

import (
	_ "github.com/coredns/coredns/plugin/file"
	_ "github.com/coredns/coredns/plugin/log"
	_ "github.com/coredns/coredns/plugin/reload"
	_ "github.com/imlonghao/geo"

	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/coremain"
)

var directives = []string{
	"reload",
	"log",
	"geo",
	"file",
}

func init() {
	dnsserver.Directives = directives
}

func main() {
	coremain.Run()
}
