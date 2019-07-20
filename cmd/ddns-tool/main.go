package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"reflect"
	"strconv"

	"github.com/aleiphoenix/go-ddns/pkg/ddns"
	"github.com/aleiphoenix/go-ddns/tools/input"
	"gopkg.in/yaml.v3"
)

func askFor(name string, v interface{}) {
	read := input.MustReadInput(fmt.Sprintf("%s ?\n> ", name))
	switch v.(type) {
	case *int:
		i, err := strconv.Atoi(read)
		if err != nil {
			log.Fatal(err)
		}
		v = &i
	case *string:
		v = &read
	default:
		log.Fatal("unknown type interface: %s", reflect.TypeOf(v))
	}
}

func main() {

	var hostname = flag.String("hostname", "", "hostname")
	var zone = flag.String("zone", "", "zone")
	var ip = flag.String("ip", "", "ip")
	var ttl = flag.Int("ttl", 300, "ttl in seconds")
	var nameserver = flag.String("nameserver", "", "nameserver")
	var keyFile = flag.String(
		"tisg", "tisg.yaml", "tisg key config")

	flag.Parse()

	if len(*hostname) <= 0 {
		askFor("hostname", hostname)
	}
	if len(*zone) <= 0 {
		askFor("zone", zone)
	}
	if len(*ip) <= 0 {
		askFor("ip", ip)
	}
	if len(*nameserver) <= 0 {
		askFor("nameserver", nameserver)
	}

	log.Printf("hostname: %s", *hostname)
	log.Printf("zone: %s", *zone)
	log.Printf("ip: %s", *ip)
	log.Printf("ttl: %d", *ttl)
	log.Printf("nameserver: %s", *nameserver)

	key := new(ddns.TisgConfig)
	log.Printf(">> loading key from %s", *keyFile)
	content, err := ioutil.ReadFile(*keyFile)
	err = yaml.Unmarshal(content, &key)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf(">> key: %s", key)
	tisg := make(map[string]string)
	for _, t := range key.Keys {
		tisg[t.FQDN] = t.Key
	}

	ok, err := ddns.Update(
		*hostname, *zone, *ip,
		uint32(*ttl), *nameserver, 3000, tisg)
	if err != nil {
		log.Fatal(err)
	}
	if ok {
		fqdn := fmt.Sprintf("%s.%s", *hostname, *zone)
		log.Printf("%s updated to %s", fqdn, *ip)
	}
}
