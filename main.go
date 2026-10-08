package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"

	"shieldgo/internal/ban"
	"shieldgo/internal/config"
	"shieldgo/internal/limiter"
	"shieldgo/internal/logging"
)

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

type routeLimiters struct {
	mu   sync.Mutex
	byID map[string]*limiter.Limiter
}

func newRouteLimiters() *routeLimiters {
	return &routeLimiters{
		byID: make(map[string]*limiter.Limiter),
	}
}

func (rl *routeLimiters) get(prefix string, rate, burst float64) *limiter.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	l, ok := rl.byID[prefix]
	if !ok {
		l = limiter.New(rate, burst)
		rl.byID[prefix] = l
	}

	return l
}

func main() {
	configPath := flag.String(
		"config",
		"",
		"path to config.json",
	)

	listen := flag.String(
		"listen",
		":8080",
		"address to listen on",
	)

	backend := flag.String(
		"backend",
		"http://127.0.0.1:3000",
		"backend URL to protect",
	)

	rate := flag.Float64(
		"rate",
		10,
		"allowed requests per second per IP",
	)

	burst := flag.Float64(
		"burst",
		20,
		"burst size per IP",
	)

	strikes := flag.Int(
		"strikes",
		20,
		"rate-limit violations before a ban",
	)

	banFor := flag.Duration(
		"ban",
		time.Minute,
		"first ban duration",
	)

	maxBody := flag.Int64(
		"max-body",
		1<<20,
		"maximum request body size",
	)

	flag.Parse()

	cfg := config.Default()

	if *configPath != "" {
		loaded, err := config.Load(*configPath)
		if err != nil {
			log.Fatalf("config: %v", err)
		}

		cfg = loaded
	} else {
		cfg.Listen = *listen
		cfg.Backend = *backend
		cfg.Rate = *rate
		cfg.Burst = *burst
		cfg.Strikes = *strikes
		cfg.BanSeconds = int(banFor.Seconds())
		cfg.MaxBodyBytes = *maxBody
	}

	target, err := url.Parse(cfg.Backend)
	if err != nil {
		log.Fatalf("bad backend URL: %v", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	perIP := limiter.New(cfg.Rate, cfg.Burst)
	global := limiter.New(cfg.GlobalRate, cfg.GlobalBurst)

	routes := newRouteLimiters()

	bans := ban.New(
		cfg.Strikes,
		time.Duration(cfg.BanSeconds)*time.Second,
		time.Duration(cfg.MaxBanSeconds)*time.Second,
	)

	lg, err := logging.Open(cfg.LogPath)
	if err != nil {
		log.Fatalf("log file: %v", err)
	}

	defer lg.Close()

	go func() {
		for range time.Tick(time.Minute) {
			perIP.Cleanup(10 * time.Minute)
			global.Cleanup(10 * time.Minute)
			bans.Cleanup(time.Hour)
		}
	}()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		path := r.URL.Path

		// Check whether IP is already banned.
		if bans.IsBanned(ip) {
			lg.Log(
				ip,
				path,
				"blocked_existing_ban",
				http.StatusForbidden,
			)

			http.Error(
				w,
				"banned",
				http.StatusForbidden,
			)

			return
		}

		// Global rate limiter.
		if !global.Allow("global") {
			lg.Log(
				ip,
				path,
				"rate_limited_global",
				http.StatusTooManyRequests,
			)

			w.Header().Set("Retry-After", "1")

			http.Error(
				w,
				"service busy",
				http.StatusTooManyRequests,
			)

			return
		}

		// Route-specific limiter.
		if route, ok := cfg.MatchRoute(path); ok {
			rl := routes.get(
				route.Prefix,
				route.Rate,
				route.Burst,
			)

			if !rl.Allow(ip) {
				banned := bans.Violation(ip)

				action := "rate_limited_route"

				if banned {
					action = "banned"

					log.Printf(
						"BAN ip=%s route=%s",
						ip,
						route.Prefix,
					)
				}

				lg.Log(
					ip,
					path,
					action,
					http.StatusTooManyRequests,
				)

				w.Header().Set("Retry-After", "1")

				http.Error(
					w,
					"too many requests",
					http.StatusTooManyRequests,
				)

				return
			}
		}

		// Per-IP limiter.
		if !perIP.Allow(ip) {
			banned := bans.Violation(ip)

			action := "rate_limited"

			if banned {
				action = "banned"

				log.Printf(
					"BAN ip=%s",
					ip,
				)
			}

			lg.Log(
				ip,
				path,
				action,
				http.StatusTooManyRequests,
			)

			w.Header().Set("Retry-After", "1")

			http.Error(
				w,
				"too many requests",
				http.StatusTooManyRequests,
			)

			return
		}

		// Limit request body size.
		r.Body = http.MaxBytesReader(
			w,
			r.Body,
			cfg.MaxBodyBytes,
		)

		lg.Log(
			ip,
			path,
			"allow",
			http.StatusOK,
		)

		// Forward request to Metasploitable.
		proxy.ServeHTTP(w, r)
	})

	// HTTP server protections against slow HTTP attacks.
	srv := &http.Server{
		Addr:    cfg.Listen,
		Handler: handler,

		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,

		MaxHeaderBytes: 16 << 10,
	}

	log.Printf(
		"ShieldGo listening on %s -> %s",
		cfg.Listen,
		cfg.Backend,
	)

	log.Fatal(
		srv.ListenAndServe(),
	)
}
