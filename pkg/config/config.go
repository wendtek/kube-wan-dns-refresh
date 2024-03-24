package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

type Config struct {
	ConfigFilePath string
	DryRun         bool
	Route53Records RecordSet `json:"route53records"`
}

type RecordSet struct {
	A []string `json:"A"`
}

func NewConfig() *Config {
	return &Config{}
}

// ParseFlags will read command line flags into the config object
func (c *Config) ParseFlags() error {
	// Implement the function to parse the command line flags
	cfgPath := flag.String("config", "config.json", "The name of the configuration file")
	dryRun := flag.Bool("dry-run", false, "If true, the program will not make any changes")

	flag.Parse()

	c.ConfigFilePath = *cfgPath
	c.DryRun = *dryRun

	return nil
}

// ReadConfig will load the record configuration from a file
func (c *Config) ReadConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// ToString will convert the config to a string and return it
func (c *Config) ToString() string {
	return fmt.Sprintf("%+v", *c)
}
