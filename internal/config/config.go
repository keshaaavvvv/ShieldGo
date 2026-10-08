package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Route struct {
	Prefix string  `json:"prefix"`
	Rate   float64 `json:"rate"`
	Burst  float64 `json:"burst"`
}

type Config struct {
	Listen        string  `json:"listen"`
	Backend       string  `json:"backend"`
	Rate          float64 `json:"rate"`
	Burst         float64 `json:"burst"`
	GlobalRate    float64 `json:"global_rate"`
	GlobalBurst   float64 `json:"global_burst"`
	Strikes       int     `json:"strikes"`
	BanSeconds    int     `json:"ban_seconds"`
	MaxBanSeconds int     `json:"max_ban_seconds"`
	MaxBodyBytes  int64   `json:"max_body_bytes"`
	LogPath       string  `json:"log_path"`
	Routes        []Route `json:"routes"`
}

func Default() Config {
	return Config{
		Listen:        ":8080",
		Backend:       "http://192.168.26.129:80",
		Rate:          10,
		Burst:         20,
		GlobalRate:    150,
		GlobalBurst:   300,
		Strikes:       10,
		BanSeconds:    60,
		MaxBanSeconds: 86400,
		MaxBodyBytes:  1 << 20,
		LogPath:       "shield.log",
	}
}

func Load(path string) (Config, error) {
	c := Default()

	f, err := os.Open(path)
	if err != nil {
		return c, fmt.Errorf("open config: %w", err)
	}

	defer f.Close()

	dec := json.NewDecoder(f)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&c); err != nil {
		return c, fmt.Errorf("parse config: %w", err)
	}

	return c, nil
}

func (c Config) MatchRoute(path string) (Route, bool) {
	best := Route{}
	found := false

	for _, r := range c.Routes {
		if len(path) >= len(r.Prefix) &&
			path[:len(r.Prefix)] == r.Prefix {

			if !found || len(r.Prefix) > len(best.Prefix) {
				best = r
				found = true
			}
		}
	}

	return best, found
}
