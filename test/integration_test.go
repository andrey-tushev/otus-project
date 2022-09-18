package test

// nolint:gci
import (
	"context"
	"fmt"
	"image/jpeg"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/andrey-tushev/otus-go/project/internal/cache"
	"github.com/andrey-tushev/otus-go/project/internal/logger"
	"github.com/andrey-tushev/otus-go/project/internal/proxy"
)

const (
	proxyHost = "localhost"
	proxyPort = 8081

	//targetURL    = "http://localhost:8082/"
	targetURL    = "https://raw.githubusercontent.com/andrey-tushev/otus-go/project/project/images/www/"
	badTargetURL = "http://localhost:6666/"

	cacheDir = "cache"
)

func TestBadTargetServer(t *testing.T) {
	log := logger.New(logger.LevelInfo)

	cache := cache.New(cacheDir, 10)
	cache.Clear()
	defer cache.Clear()

	ctx := context.Background()

	proxyServer := proxy.New(log, cache, badTargetURL+"fill")
	go proxyServer.Start(ctx, proxyHost, proxyPort)
	defer proxyServer.Stop(ctx)

	proxyPref := fmt.Sprintf("http://%s:%d/fill", proxyHost, proxyPort)

	// nolint:noctx
	resp, err := http.Get(proxyPref + "/100/100/cat-5.jpg")
	require.NoError(t, err)
	require.Equal(t, 502, resp.StatusCode)
	resp.Body.Close()
}

func TestProxyResponses(t *testing.T) {
	log := logger.New(logger.LevelInfo)

	cache := cache.New(cacheDir, 10)
	cache.Clear()
	defer cache.Clear()

	ctx := context.Background()

	proxyServer := proxy.New(log, cache, targetURL)
	go proxyServer.Start(ctx, "localhost", proxyPort)
	defer proxyServer.Stop(ctx)

	proxyPref := fmt.Sprintf("http://%s:%d/fill", proxyHost, proxyPort)

	// Корректный запрос
	{
		// nolint:noctx
		resp, err := http.Get(proxyPref + "/100/100/cat-1.jpg")
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode)
		require.Equal(t, "image/jpeg", resp.Header.Get("Content-Type"))
		resp.Body.Close()
	}

	// Такой картинки нет
	{
		// nolint:noctx
		resp, err := http.Get(proxyPref + "/100/100/cat-x.jpg")
		require.NoError(t, err)
		require.Equal(t, 404, resp.StatusCode)
		resp.Body.Close()
	}

	// Кривая картинка
	{
		// nolint:noctx
		resp, err := http.Get(proxyPref + "/100/100/bad.jpg")
		require.NoError(t, err)
		require.Equal(t, 502, resp.StatusCode)
		resp.Body.Close()
	}

	// Не картинка
	{
		// nolint:noctx
		resp, err := http.Get(proxyPref + "/100/100/text.txt")
		require.NoError(t, err)
		require.Equal(t, 502, resp.StatusCode)
		resp.Body.Close()
	}
}

func TestProxyResize(t *testing.T) {
	log := logger.New(logger.LevelInfo)

	cache := cache.New(cacheDir, 10)
	cache.Clear()
	defer cache.Clear()

	ctx := context.Background()

	proxyServer := proxy.New(log, cache, targetURL)
	go proxyServer.Start(ctx, "localhost", proxyPort)
	defer proxyServer.Stop(ctx)

	proxyPref := fmt.Sprintf("http://%s:%d/fill", proxyHost, proxyPort)

	// Запросим первый раз
	{
		// nolint:noctx
		resp, err := http.Get(proxyPref + "/100/100/cat-1.jpg")
		require.NoError(t, err)
		require.Equal(t, "no", resp.Header.Get("X-Cached"))
		resp.Body.Close()
	}

	// Запросим второй раз
	{
		// nolint:noctx
		resp, err := http.Get(proxyPref + "/100/100/cat-1.jpg")
		require.NoError(t, err)
		require.Equal(t, "yes", resp.Header.Get("X-Cached"))
		resp.Body.Close()
	}

	// Запросим первый раз
	{
		// nolint:noctx
		resp, err := http.Get(proxyPref + "/100/100/cat-2.jpg")
		require.NoError(t, err)
		require.Equal(t, "no", resp.Header.Get("X-Cached"))
		resp.Body.Close()
	}

	// Сожмем по горизонтали
	{
		// nolint: noctx
		resp, err := http.Get(proxyPref + "/100/2000/cat-1.jpg")
		require.NoError(t, err)

		preview, err := jpeg.Decode(resp.Body)
		require.NoError(t, err)
		require.Equal(t, 100, preview.Bounds().Dx())

		resp.Body.Close()
	}

	// Сожмем по вертикали
	{
		// nolint: noctx
		resp, err := http.Get(proxyPref + "/2000/100/cat-1.jpg")
		require.NoError(t, err)

		preview, err := jpeg.Decode(resp.Body)
		require.NoError(t, err)
		require.Equal(t, 100, preview.Bounds().Dy())

		resp.Body.Close()
	}

	// Сожмем по обоим сторонам
	{
		// nolint:noctx
		resp, err := http.Get(proxyPref + "/100/100/cat-1.jpg")
		require.NoError(t, err)

		preview, err := jpeg.Decode(resp.Body)
		require.NoError(t, err)
		require.LessOrEqual(t, preview.Bounds().Dy(), 100)
		require.GreaterOrEqual(t, preview.Bounds().Dy(), 10)
		require.LessOrEqual(t, preview.Bounds().Dy(), 100)
		require.GreaterOrEqual(t, preview.Bounds().Dy(), 10)

		resp.Body.Close()
	}

	// Сжатие не требуется
	{
		// nolint:noctx
		resp, err := http.Get(proxyPref + "/2000/2000/cat-1.jpg")
		require.NoError(t, err)

		preview, err := jpeg.Decode(resp.Body)
		require.NoError(t, err)
		require.Equal(t, 681, preview.Bounds().Dx())
		require.Equal(t, 1024, preview.Bounds().Dy())

		resp.Body.Close()
	}
}
