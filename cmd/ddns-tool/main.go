package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

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

	var domain = flag.String("domain", "", "whole domain")
	var hostname = flag.String("hostname", "", "hostname")
	var zone = flag.String("zone", "", "zone")
	var recordType = flag.String("type", "A", "DNS record type")
	var payload = flag.String("payload", "", "payload")
	var ttl = flag.Int("ttl", 300, "ttl in seconds")
	var nameserver = flag.String("nameserver", "", "nameserver")
	var keyFile = flag.String(
		"tisg", "tisg.yaml", "tisg key config")

	flag.Parse()

	if len(*domain) > 0 {
		re, _ := regexp.Compile(
			`([-a-z0-9\.]+)\.([-a-z0-9]+\.[-a-z0-9]+)`)
		groups := re.FindAllStringSubmatch(*domain, -1)
		if groups == nil {
			panic(fmt.Sprintf("invalid domain: %s", *domain))
		}
		zone = &groups[len(groups)-1][2]
		hostname = &groups[len(groups)-1][1]
	} else {
		if len(*hostname) <= 0 {
			askFor("hostname", hostname)
		}
		if len(*zone) <= 0 {
			askFor("zone", zone)
		}
	}
	if len(*payload) <= 0 {
		askFor("payload", payload)
	}
	if len(*nameserver) <= 0 {
		askFor("nameserver", nameserver)
	}

	log.Printf("hostname: %s", *hostname)
	log.Printf("zone: %s", *zone)
	log.Printf("payload: %s", *payload)
	log.Printf("type: %s", *recordType)
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

	var ok bool
	timeout := time.Second * 10
	switch strings.ToLower(*recordType) {
	case "a":
		ok, err = ddns.Update(
			*hostname, *zone, *payload,
			uint32(*ttl), *nameserver, timeout, tisg)
		if err != nil {
			log.Fatal(err)
		}
	case "txt":
		txt := []string{*payload}
		ok, err = ddns.UpdateTXT(
			*hostname, *zone, txt,
			uint32(*ttl), *nameserver, timeout, tisg)
		if err != nil {
			log.Fatal(err)
		}

	default:
		log.Fatalf("unknown type: %s", recordType)
	}

	if ok {
		fqdn := fmt.Sprintf("%s.%s", *hostname, *zone)
		log.Printf("%s updated to %s", fqdn, *payload)
	}

}
