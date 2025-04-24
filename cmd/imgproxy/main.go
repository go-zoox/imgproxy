package main

import (
	"fmt"

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
		},
	})

	app.Command(func(c *cli.Context) error {
		var cfg server.Config
		cfg.Port = cast.ToInt64(c.String("port"))
		cfg.EnableGzip = c.Bool("enable-gzip")

		fmt.Println("imgproxy serve on port: ", cfg.Port)
		return server.Run(&cfg)
	})

	app.Run()
}
