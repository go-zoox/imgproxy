package main

import (
	"time"

	"github.com/go-zoox/cli"
	"github.com/go-zoox/core-utils/cast"
	"github.com/go-zoox/imgproxy"

	"github.com/go-zoox/imgproxy/server"
)

// //go:embed static/*
// var static embed.FS

func main() {
	app := cli.NewSingleProgram(&cli.SingleProgramConfig{
		Name:        "imgproxy",
		Usage:       "The Server of imgproxy",
		Description: "The Server of imgproxy",
		Version:     imgproxy.Version,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "port",
				Value:   "8080",
				Usage:   "The port to listen on",
				Aliases: []string{"p"},
				EnvVars: []string{"PORT"},
			},
			&cli.BoolFlag{
				Name:    "enable-gzip",
				Usage:   "Enable gzip compression",
				EnvVars: []string{"ENABLE_GZIP"},
			},
			&cli.StringFlag{
				Name:    "cache-dir",
				Usage:   "The cache directory, if not set, the cache will be disabled",
				EnvVars: []string{"CACHE_DIR"},
				Value:   "/tmp/cache/go-zoox/imgproxy",
			},
			&cli.Int64Flag{
				Name:    "cache-max-age",
				Usage:   "The cache max age, if not set, the default value is 1 year",
				EnvVars: []string{"CACHE_MAX_AGE"},
				Value:   1 * 365 * 24 * 60 * 60,
			},
		},
	})

	app.Command(func(c *cli.Context) error {
		var cfg server.Config
		cfg.Port = cast.ToInt64(c.String("port"))
		cfg.EnableGzip = c.Bool("enable-gzip")

		cfg.CacheDir = c.String("cache-dir")
		cfg.CacheMaxAge = time.Duration(c.Int64("cache-max-age")) * time.Second

		return server.Run(&cfg)
	})

	app.Run()
}
