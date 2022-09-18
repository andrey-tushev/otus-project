package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/andrey-tushev/otus-go/project/internal/cache"
	"github.com/andrey-tushev/otus-go/project/internal/logger"
	"github.com/andrey-tushev/otus-go/project/internal/proxy"
)

var (
	targetURL string
	port      int
	maxFiles  int
)

const cacheDir = "cache"

func init() {
	flag.IntVar(&maxFiles, "max-files", 10, "Maximum files in cache")
	flag.StringVar(&targetURL, "target-url", "http://localhost:8082/", "Target URL")
	flag.IntVar(&port, "port", 8081, "Server port")
}

func main() {
	ret := retMain()
	os.Exit(ret)
}

func retMain() int {
	var ret int

	flag.Parse()

	log := logger.New(logger.LevelInfo)

	log.Info(fmt.Sprintf("Target URL: %s", targetURL))
	log.Info(fmt.Sprintf("Max files in cache: %d", maxFiles))
	log.Info(fmt.Sprintf("Listening port: %d", port))

	log.Info("Proxy started")
	defer log.Info("Proxy finished")

	cache := cache.New(cacheDir, maxFiles)
	cache.Clear()

	proxyServer := proxy.New(log, cache, targetURL)

	// Останавливалка серверов по сигналу
	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	go func() {
		<-ctx.Done() // ждем сигнала
		log.Info("got terminating signal")

		// На остановку выделяем не более 3 секунд
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		// Останавливаем web-сервер
		go func() {
			log.Info("terminating web-server")
			if err := proxyServer.Stop(ctx); err != nil {
				log.Error("failed to stop web-server: " + err.Error())
			}
		}()
	}()

	listenURL := fmt.Sprintf("http://%s:%d/", "localhost", port)
	log.Info("starting web-server on " + listenURL)
	if err := proxyServer.Start(ctx, "", port); err != nil {
		log.Error("failed to start http-server: " + err.Error())
		ret = 1
		cancel()
	}

	return ret
}
