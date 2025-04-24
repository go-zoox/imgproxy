package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/go-zoox/fs"
	"github.com/go-zoox/safe"

	"github.com/go-zoox/crypto/md5"
	"github.com/go-zoox/debug"
	"github.com/go-zoox/fetch"
	"github.com/go-zoox/logger"
	"github.com/go-zoox/zoox"
	defaults "github.com/go-zoox/zoox/defaults"
	"github.com/go-zoox/zoox/middleware"

	"github.com/h2non/bimg"
)

const DEFAULT_CACHE_DIR = "/tmp/cache/go-zoox/imgproxy"
const DEFAULT_CACHE_MAX_AGE = 31536000 * time.Second

var DEFAULT_ALLOW_CONVERTS = map[string]bool{
	"jpg":  true,
	"jpeg": true,
	"png":  true,
}

// Config is the configuration of the server.
type Config struct {
	Port int64 `yaml:"port"`
	//
	EnableGzip bool
	//
	CacheMaxAge time.Duration
	//
	CacheDir string
}

// Run starts the server.
func Run(cfg *Config) error {
	if cfg.Port == 0 {
		cfg.Port = 8080
	}
	if cfg.CacheMaxAge == 0 {
		cfg.CacheMaxAge = DEFAULT_CACHE_MAX_AGE
	}
	if cfg.CacheDir == "" {
		cfg.CacheDir = DEFAULT_CACHE_DIR
	}

	if debug.IsDebugMode() {
		j, _ := json.MarshalIndent(cfg, "", "  ")
		logger.Infof("%s", string(j))
	}

	if ok := fs.IsExist(cfg.CacheDir); !ok {
		if err := fs.Mkdir(cfg.CacheDir, 0755); err != nil {
			return err
		}
	}

	app := defaults.Default()

	app.Cron().AddJob("cleanup_image", "0 4 1 * *", func() error {
		if ok := fs.IsExist(cfg.CacheDir); !ok {
			return nil
		}

		logger.Infof("[cron] cleanup image cache[%s]", cfg.CacheDir)
		if err := fs.RemoveDir(cfg.CacheDir); err != nil {
			return err
		}

		if err := fs.Mkdir(cfg.CacheDir, 0755); err != nil {
			return err
		}

		return nil
	})

	app.Use(middleware.Gzip())
	// app.Use(middleware.CacheControl(&middleware.CacheControlConfig{
	// 	MaxAge: 1 * 356 * 24 * time.Hour,
	// }))

	app.Get("/process", func(ctx *zoox.Context) {
		url := ctx.Query().Get("url").String()
		if url == "" {
			ctx.JSON(400, zoox.H{
				"message": "url is required",
			})
			return
		}

		cacheKey := md5.Md5(ctx.Request.URL.String())
		cacheFilePath := fmt.Sprintf("%s/%s", cfg.CacheDir, cacheKey)
		if ok := fs.IsExist(cacheFilePath); ok {
			ctx.Logger.Infof("process(%s) => load from cache(path: %s)", url, cacheFilePath)

			imgType := "png"
			if bytes, err := fs.ReadFile(cacheFilePath); err != nil {
				ctx.JSON(http.StatusInternalServerError, zoox.H{
					"message": fmt.Sprintf("failed to read cache file: %s", err.Error()),
				})
				return
			} else {
				imgType = bimg.NewImage(bytes).Type()
			}

			f, err := os.OpenFile(cacheFilePath, os.O_RDONLY, 0644)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, zoox.H{
					"message": fmt.Sprintf("failed to open cache file: %s", err.Error()),
				})
				return
			}

			// ctx.SetCacheControlWithMaxAge(cfg.CacheMaxAge)
			ctx.SetCacheControl(fmt.Sprintf("public, max-age=%d", cfg.CacheMaxAge/time.Second))
			ctx.SetContentType(fmt.Sprintf("image/%s", imgType))

			if _, err := io.Copy(ctx.Writer, f); err != nil {
				ctx.JSON(http.StatusInternalServerError, zoox.H{
					"message": fmt.Sprintf("failed to copy cache file: %s", err.Error()),
				})
				return
			}
			return
		}

		ctx.Logger.Infof("process(%s) => load from request", url)

		options := bimg.Options{}
		if width := ctx.Query().Get("w").Int(); width > 0 {
			options.Width = width
		}
		if height := ctx.Query().Get("h").Int(); height > 0 {
			options.Height = height
		}
		if quality := ctx.Query().Get("q").Int(); quality > 0 {
			options.Quality = quality
		}
		if rotate := ctx.Query().Get("r").Int(); rotate > 0 {
			options.Rotate = bimg.Angle(rotate)
		}
		if crop := ctx.Query().Get("c").Bool(); crop {
			options.Crop = crop
		}

		response, err := fetch.Get(url)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, zoox.H{
				"message": fmt.Sprintf("failed to fetch image: %s", err.Error()),
			})
			return
		}

		// process image
		img := bimg.NewImage(response.Body)
		imgType := img.Type()
		bytes, err := img.Process(options)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, zoox.H{
				"message": fmt.Sprintf("failed to process image: %s", err.Error()),
			})
			return
		}

		err = safe.Do(func() error {
			if _, ok := DEFAULT_ALLOW_CONVERTS[imgType]; !ok {
				return nil
			}

			// convert to webp
			bytes2, err2 := bimg.NewImage(bytes).Convert(bimg.WEBP)
			if err2 != nil {
				return err2
			}

			imgType = "webp"
			bytes = bytes2
			return nil
		})
		if err != nil {
			ctx.Logger.Warnf("failed to convert to webp: %s", err.Error())
		}

		ctx.SetCacheControl(fmt.Sprintf("public, max-age=%d", cfg.CacheMaxAge/time.Second))
		ctx.SetContentType(fmt.Sprintf("image/%s", imgType))
		if err := fs.WriteFile(cacheFilePath, bytes); err != nil {
			ctx.JSON(http.StatusInternalServerError, zoox.H{
				"message": fmt.Sprintf("failed to write cache file: %s", err.Error()),
			})
			return
		}

		ctx.Write(bytes)
	})

	return app.Run(fmt.Sprintf(":%d", cfg.Port))
}
