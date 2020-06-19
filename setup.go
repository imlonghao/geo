package geo

import (
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	i2l "github.com/ip2location/ip2location-go"
	"github.com/miekg/dns"
	"io/ioutil"
	"time"

	"github.com/caddyserver/caddy"
	"gopkg.in/yaml.v2"
)

var (
	log = clog.NewWithPlugin("hosts")
)

func init() { plugin.Register("geo", setup) }

func setup(c *caddy.Controller) error {
	geo := Geo{}
	geo.Zone = make(map[string][]GeoType)

	c.Next() // 'geo'

	if !c.Next() {
		log.Error("geo", "IP2Location path missing")
	}
	ip2location, err := i2l.OpenDB(c.Val())
	if err != nil {
		log.Error("geo", err)
	}
	geo.IP2Location = ip2location

	if !c.Next() {
		log.Error("geo", "Geo zone path missing")
	}
	geo.ZoneFilename = c.Val()

	if c.NextArg() {
		return plugin.Error("geo", c.ArgErr())
	}

	parseChan := periodicZoneUpdate(&geo)
	c.OnStartup(func() error {
		geo.reload()
		return nil
	})
	c.OnShutdown(func() error {
		close(parseChan)
		return nil
	})

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		geo.Next = next
		return &geo
	})

	return nil
}

func periodicZoneUpdate(g *Geo) chan bool {
	parseChan := make(chan bool)

	go func() {
		ticker := time.NewTicker(time.Second * 30)
		for {
			select {
			case <-parseChan:
				return
			case <-ticker.C:
				g.reload()
			}
		}
	}()
	return parseChan
}

func (g *Geo) reload() {
	zone := make(map[string][]GeoType)
	file, err := ioutil.ReadFile(g.ZoneFilename)
	if err != nil {
		log.Error(err)
	}
	var geoFile []GeoFile
	err = yaml.Unmarshal(file, &geoFile)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	for _, geoRule := range geoFile {
		z := zone[geoRule.Zone]
		var t uint16
		switch geoRule.Type {
		case "a":
			t = dns.TypeA
		case "aaaa":
			t = dns.TypeAAAA
		case "cname":
			t = dns.TypeCNAME
		}
		z = append(z, GeoType{
			TTL:   geoRule.TTL,
			Type:  t,
			Value: geoRule.Value,
			Lat:   geoRule.Lat,
			Lon:   geoRule.Lon,
		})
		zone[geoRule.Zone] = z
	}
	g.Zone = zone
}
