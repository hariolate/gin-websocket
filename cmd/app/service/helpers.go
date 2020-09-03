package service

import (
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"io/ioutil"
	"log"
)

func NoError(err error) {
	if err != nil {
		log.Panicln(err)
	}
}

func MustReadFile(path string) []byte {
	data, err := ioutil.ReadFile(path)
	NoError(err)
	return data
}

func MustUnmarshal(data []byte, o interface{}) {
	NoError(json.Unmarshal(data, o))
}

func MustParseRedisURL(url string) *redis.Options {
	opt, err := redis.ParseURL(url)
	NoError(err)
	return opt
}

func ServeAddrFromConfig(c *Config) string {
	addr, port := c.ServeOn.Addr, c.ServeOn.Port

	if len(addr) == 0 && len(port) == 0 {
		return ""
	}
	if len(addr) == 0 {
		return ":" + port
	}

	if len(port) == 0 {
		return addr + ":0"
	}

	return addr + ":" + port
}
