package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strings"
	"time"

	"github.com/miekg/dns"
	"gopkg.in/yaml.v3"
)

func main() {

	var configFile = flag.String(
		"config", "domains.yaml", "input domain config")
	var ifaceName = flag.String(
		"iface", "ppp0", "interface which ipaddr is from")

	flag.Parse()

	log.Printf(">> reading ipaddr from iface %s", *ifaceName)

	iface, err := net.InterfaceByName(*ifaceName)
	if err != nil {
		log.Fatal(err)
	}
	addrs, err := iface.Addrs()
	if err != nil {
		log.Fatal(err)
	}

	// FIXME: multiple ip ?
	ipv4 := make([]string, 1)
	ipv6 := make([]string, 1)

	for _, addr := range addrs {
		var ip net.IP
		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		}
		if strings.Index(ip.String(), ":") == -1 {
			ipv4 = append(ipv4, ip.String())
		} else {
			ipv6 = append(ipv6, ip.String())
		}
	}

	log.Printf(">> ipv4 ipaddr: %s", ipv4)
	log.Printf(">> ipv6 ipaddr: %s", ipv6)

	log.Printf(">> reading config from %s", *configFile)

	content, err := ioutil.ReadFile(*configFile)
	if err != nil {
		log.Fatal(err)
	}

	config := new(Config)

	log.Printf(">> loading config from %s", *configFile)

	err = yaml.Unmarshal(content, &config)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf(">> config: %s", config)

	for _, domain := range config.Domains {
		if len(domain.Nameserver) <= 0 {
			log.Printf(
				"!!! no nameserver configured for zone %s",
				domain.Zone)
			continue
		}

		for _, hostname := range domain.Hostnames {
			c := new(dns.Client)
			c.Dialer = &net.Dialer{
				Timeout: time.Second * 1,
				// Timeout: time.Millisecond * 1,
			}

			// TODO
			nameserver := fmt.Sprintf("%s:53", domain.Nameserver)
			fqdn := fmt.Sprintf("%s.%s", hostname, domain.Zone)

			ml := new(dns.Msg)
			ml.Id = dns.Id()
			ml.Question = make([]dns.Question, 1)
			ml.Question[0] = dns.Question{
				dns.Fqdn(fqdn), dns.TypeA, dns.ClassINET}

			in, _, err := c.Exchange(ml, nameserver)

			if err != nil {
				log.Fatal(err)
			}

			if len(in.Answer) > 0 {
				if t, ok := in.Answer[0].(*dns.A); ok {
					// do something with t.Txt
					log.Printf(
						">>> resolved %s: %s",
						fqdn, t.A)
				}
			} else {
				log.Printf("!!! not found: %s", fqdn)
			}

		}

	}

}
