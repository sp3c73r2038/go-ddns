package main

type Config struct {
	Domains []ZoneConfig `yaml:"domains"`
}

type ZoneConfig struct {
	Zone       string   `yaml:"zone"`
	Nameserver string   `yaml:nameserver`
	Hostnames  []string `yaml:"hostnames"`
}
