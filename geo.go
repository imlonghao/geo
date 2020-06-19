package geo

import (
	"context"
	"github.com/coredns/coredns/plugin"
	ip2location "github.com/ip2location/ip2location-go"
	"github.com/coredns/coredns/plugin/pkg/fall"
	"github.com/coredns/coredns/request"
	"math"
	"net"

	"github.com/miekg/dns"
)

type Geo struct {
	Next plugin.Handler
	Fall fall.F

	IP2Location  *ip2location.DB
	Zone         map[string][]GeoType
	ZoneFilename string
}

type GeoType struct {
	Type  uint16  `yaml:"type"`
	TTL   uint32  `yaml:"ttl"`
	Value string  `yaml:"value"`
	Lat   float64 `yaml:"lat"`
	Lon   float64 `yaml:"lon"`
}

type GeoFile struct {
	Zone  string  `yaml:"zone"`
	TTL   uint32  `yaml:"ttl"`
	Type  string  `yaml:"type"`
	Value string  `yaml:"value"`
	Lat   float64 `yaml:"lat"`
	Lon   float64 `yaml:"lon"`
}

// ServeDNS implements the plugin.Handler interface.
func (g Geo) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}

	a := new(dns.Msg)
	a.SetReply(r)
	a.Authoritative = true

	ip := state.IP()
	location, err := g.IP2Location.Get_all(ip)
	if err != nil {
		return plugin.NextOrFailure(g.Name(), g.Next, ctx, w, r)
	}

	geoZone, ok := g.Zone[state.QName()]
	if !ok {
		return plugin.NextOrFailure(g.Name(), g.Next, ctx, w, r)
	}

	min := math.MaxFloat64
	var result GeoType
	for _, zone := range geoZone {
		d := Distance(float64(location.Latitude), float64(location.Longitude), zone.Lat, zone.Lon)
		if d < min {
			min = d
			result = zone
		}
	}
	if result.Value == "" {
		return plugin.NextOrFailure(g.Name(), g.Next, ctx, w, r)
	}

	var rr dns.RR

	switch result.Type {
	case dns.TypeA:
		rr = new(dns.A)
		rr.(*dns.A).Hdr = dns.RR_Header{Name: state.QName(), Rrtype: dns.TypeA, Class: state.QClass(), Ttl: result.TTL}
		rr.(*dns.A).A = net.ParseIP(result.Value)
	case dns.TypeCNAME:
		rr = new(dns.CNAME)
		rr.(*dns.CNAME).Hdr = dns.RR_Header{Name: state.QName(), Rrtype: dns.TypeCNAME, Class: state.QClass(), Ttl: result.TTL}
		rr.(*dns.CNAME).Target = result.Value
	default:
		return plugin.NextOrFailure(g.Name(), g.Next, ctx, w, r)
	}

	a.Answer = []dns.RR{rr}

	err = w.WriteMsg(a)
	if err != nil {
		log.Error(err)
	}

	return dns.RcodeSuccess, nil
}

// Name implements the Handler interface.
func (g Geo) Name() string { return "geo" }
