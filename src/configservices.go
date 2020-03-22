/**
 * Created by IntelliJ IDEA.
 * User: kernel
 * Mail: kernelman79@gmail.com
 * Date: 2017/8/22
 * Time: 01:38
 */

package services

import (
	"github.com/dlintw/goconf"
	"log"
)

type ConfigServices struct {
	cfg   *goconf.ConfigFile
	class string
}

var ConfigService = new(ConfigServices)

func (c *ConfigServices) Get(configName string) *ConfigServices {
	configPath := Path.CurrentConfig(false) + configName + ".ini"
	get, err := goconf.ReadConfigFile(configPath)
	if err != nil {
		log.Println(err)
	}
	c.cfg = get

	return c
}

func (c *ConfigServices) Classify(className string) *ConfigServices {
	c.class = className
	return c
}

func (c *ConfigServices) Key(key string) string {
	value, err := c.cfg.GetString(c.class, key)
	if err != nil {
		log.Println(value)
	}

	return value
}
