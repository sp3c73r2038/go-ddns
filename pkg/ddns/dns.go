package ddns

import (
	"fmt"
	"net"
	"time"

	"github.com/miekg/dns"
)

func Query(domain string, server string, timeout int) (rv []string, err error) {
	rv = make([]string, 0)

	c := new(dns.Client)
	c.Dialer = &net.Dialer{
		Timeout: time.Millisecond * time.Duration(timeout),
	}

	m := new(dns.Msg)
	m.Id = dns.Id()
	m.Question = make([]dns.Question, 1)
	m.Question[0] = dns.Question{
		dns.Fqdn(domain), dns.TypeA, dns.ClassINET,
	}

	in, _, err := c.Exchange(m, server)
	if err != nil {
		return
	}

	if in.Rcode != dns.RcodeSuccess {
		err = fmt.Errorf("Rcode: %d", in.Rcode)
		return
	}

	if len(in.Answer) <= 0 {
		return
	}

	for _, ans := range in.Answer {
		if t, ok := ans.(*dns.A); ok {
			rv = append(rv, t.A.String())
		}
	}
	return
}

func Update(
	hostname string, zone string, ip string, ttl uint32,
	server string, timeout int, tsig map[string]string) (rv bool, err error) {

	rv = false

	fqdn := dns.Fqdn(fmt.Sprintf("%s.%s", hostname, zone))

	// DELETE A RRset first
	oldRRs := make([]dns.RR, 1)
	oldRR := new(dns.A)
	oldRR.Hdr = dns.RR_Header{
		Name:   fqdn,
		Rrtype: dns.TypeA,
		Class:  dns.ClassANY,
	}
	oldRRs[0] = oldRR

	newRRs := make([]dns.RR, 1)
	newRR := new(dns.A)
	newRR.Hdr = dns.RR_Header{
		Name:   fqdn,
		Rrtype: dns.TypeA,
		Class:  dns.ClassINET,
		Ttl:    ttl,
	}
	newRR.A = net.ParseIP(ip)
	newRRs[0] = newRR

	m := new(dns.Msg)
	m.Id = dns.Id()

	m.SetUpdate(dns.Fqdn(zone))
	m.RemoveName(oldRRs)
	m.Insert(newRRs)

	c := new(dns.Client)
	c.Dialer = &net.Dialer{
		Timeout: time.Millisecond * time.Duration(timeout),
	}

	if tsig != nil && len(tsig) > 0 {
		c.TsigSecret = tsig
		for k, _ := range tsig {
			m.SetTsig(k, dns.HmacMD5, 300, time.Now().Unix())
		}
	}

	// log.Println(m)

	in, _, err := c.Exchange(m, server)
	if err != nil {
		return
	}

	if in.Rcode != dns.RcodeSuccess {
		err = fmt.Errorf("Rcode: %d", in.Rcode)
		return
	}
	rv = true
	return
}