package main

type Config struct {
	Domains []ZoneConfig `yaml:"domains"`
}

type ZoneConfig struct {
	Zone       string   `yaml:"zone"`
	Nameserver string   `yaml:nameserver`
	Hostnames  []string `yaml:"hostnames"`
}

type TisgConfig struct {
	Keys []TisgKey `yaml:"keys"`
}

type TisgKey struct {
	FQDN string `yaml:"fqdn"`
	Key  string `yaml:"key"`
}
