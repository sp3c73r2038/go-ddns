package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strings"
	"sync/atomic"
	"time"

	"github.com/aleiphoenix/go-ddns/pkg/ddns"
	"gopkg.in/yaml.v3"
)

var lock uint32

func main() {

	var configFile = flag.String(
		"config", "domains.yaml", "input domain config")
	var keyFile = flag.String(
		"tisg", "tisg.yaml", "tisg key config")
	var ifaceName = flag.String(
		"iface", "ppp0", "interface which ipaddr is from")
	var timeout = flag.Int("timeout", 10, "timeout in seconds")

	flag.Parse()

	for {
		update(configFile, keyFile, ifaceName, timeout)
		time.Sleep(time.Second * 60)
	}
}

func update(
	configFile *string, keyFile *string, ifaceName *string, timeout *int) {

	// CAS to 1
	if !atomic.CompareAndSwapUint32(&lock, 0, 1) {
		log.Println("!! another routine is updating, skip...")
	}

	defer func() {
		atomic.StoreUint32(&lock, 0)
	}()

	_timeout := time.Second * time.Duration(*timeout)
	log.Printf("")
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
	ipv4 := make([]string, 0)
	ipv6 := make([]string, 0)

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

	config := new(ddns.Config)

	log.Printf(">> loading config from %s", *configFile)

	err = yaml.Unmarshal(content, &config)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf(">> config: %s", config)

	log.Printf(">> read key from %s", *keyFile)
	if err != nil {
		log.Fatal(err)
	}

	content, err = ioutil.ReadFile(*keyFile)

	key := new(ddns.TisgConfig)
	log.Printf(">> loading key from %s", *keyFile)
	err = yaml.Unmarshal(content, &key)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf(">> key: %s", key)

	for _, domain := range config.Domains {
		if len(domain.Nameserver) <= 0 {
			log.Printf(
				"!!! no nameserver configured for zone %s",
				domain.Zone)
			continue
		}

		for _, hostname := range domain.Hostnames {
			nameserver := fmt.Sprintf("%s:53", domain.Nameserver)
			// resolves, err := Query(fqdn, nameserver, 1000)

			tisg := make(map[string]string)
			for _, t := range key.Keys {
				tisg[t.FQDN] = t.Key
			}

			for _, ip := range ipv4 {
				// log.Println(ip)
				ok, err := ddns.Update(
					hostname, domain.Zone, ip,
					uint32(60), nameserver, _timeout, tisg)
				if err != nil {
					log.Fatal(err)
				}
				if ok {
					fqdn := fmt.Sprintf(
						"%s.%s", hostname, domain.Zone)
					log.Printf("%s updated to %s", fqdn, ip)
				}
			}

		}

	}

}
