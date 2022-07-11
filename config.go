package main

import (
	"os"

	"gopkg.in/yaml.v2"
)

// Config is the configuration that needs to run a Raft node.
type config struct {
	// ID is the identity of the node.
	ID string `yaml:"id"`

	// Quorum is a list of DNS addresses that includes all the nodes that
	// forms the Raft cluster.
	//
	// In the first version we use the cluster.local CNAME of each node.
	Quorum []string `yaml:"quorum"`
}

// loadConfig loads configurations from a config file.
func loadConfig(file string) (config, error) {
	f, err := os.Open(file)
	if err != nil {
		return config{}, err
	}
	defer f.Close()

	dec := yaml.NewDecoder(f)
	dec.SetStrict(true)

	var cfg config
	if err := dec.Decode(&cfg); err != nil {
		return config{}, err
	}

	return cfg, nil
}
