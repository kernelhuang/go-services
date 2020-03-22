/**
 * Created by Kernel.Huang.
 * User: kernel@live.com
 * Date: 2019/8/21
 * Time: 15:41
 */

package services

import (
	"github.com/Unknwon/goconfig"
)

type config struct {
	key     string
	name    string
	section string
}

var Configure = new(config)

// Get config filename
func (c *config) Name(name string) *config {
	c.name = name
	return c
}

// Get section by key
func (c *config) Section(key string) *config {
	c.section = key
	return c
}

// Get the value of the configure file by key
func (c *config) Get(key string) string {
	if c.name == "" {
		c.name = "config"
	}

	getValue, err := getConfigFile(c.name).GetValue(c.section, key)
	if err != nil {
		Logs.Error("Not found the key: " + key)
	}

	return getValue
}

// Load config file
func getConfigFile(name string) *goconfig.ConfigFile {
	configDir := "../config/"
	cfg, err := goconfig.LoadConfigFile(configDir + name + ".ini")
	if err != nil {
		Logs.Error(configDir + "conf.ini not found!")
	}

	return cfg
}
