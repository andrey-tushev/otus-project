package proxy

// nolint:gci
import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image/jpeg"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/disintegration/imaging"

	"github.com/andrey-tushev/otus-go/project/internal/preview"
)

type Server struct {
	logger       Logger
	httpServer   *http.Server
	cache        Cache
	targetPrefix string
}

type Logger interface {
	Info(msg string)
	Error(msg string)
}

type Cache interface {
	Get(img preview.Image) *preview.Container
	Set(img preview.Image, container *preview.Container)
}

const jpegContentType = "image/jpeg"

var (
	ErrBadContentType = errors.New("bad content type")
	ErrResizeError    = errors.New("resize error")
)

func New(logger Logger, previewCache Cache, targetPrefix string) *Server {
	return &Server{
		logger:       logger,
		cache:        previewCache,
		targetPrefix: targetPrefix,
	}
}

func (s *Server) Start(ctx context.Context, host string, port int) error {
	s.logger.Info("web-server start")

	s.httpServer = &http.Server{
		Addr:         net.JoinHostPort(host, strconv.Itoa(port)),
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
		Handler:      s,
	}

	err := s.httpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.httpServer.Shutdown(ctx)
	s.logger.Info("web-server stopped")

	return nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.logger.Info("Request URI " + r.RequestURI)

	// Получаем параметры требуемой превьюшки
	requestedPreview, err := preview.NewFromURL(r.RequestURI)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	// Попробуем найти в кэше и вернуть из кэша
	cached := s.cache.Get(requestedPreview)
	if cached != nil {
		for name, value := range cached.Headers {
			w.Header().Set(name, value)
		}
		w.Header().Set("X-Proxy", "proxy-resizer")
		w.Header().Set("X-Cached", "yes")
		io.Copy(w, bytes.NewReader(cached.Body))

		s.logger.Info("taken from cache")
		return
	}

	// Запросим оригинальную картинку из целевого сервера
	targetURL := s.targetPrefix + requestedPreview.Path
	s.logger.Info("target: " + targetURL)

	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	client := &http.Client{Transport: tr}

	targetReq, err := http.NewRequestWithContext(r.Context(), http.MethodGet, targetURL, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		s.logger.Error(err.Error())
		return
	}
	targetResp, err := client.Do(targetReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		s.logger.Error(err.Error())
		return
	}
	defer targetResp.Body.Close()

	// Убедимся что целевой сервер вернул ожидаемый ответ
	if targetResp.StatusCode == 404 {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	if targetResp.StatusCode < 200 || targetResp.StatusCode > 299 {
		text := fmt.Sprintf("bad target server response %d", targetResp.StatusCode)
		s.logger.Error(text)
		http.Error(w, text, http.StatusBadGateway)
		return
	}
	if targetResp.Header.Get("Content-Type") != jpegContentType {
		http.Error(w, ErrBadContentType.Error(), http.StatusBadGateway)
		return
	}

	container := preview.NewContainer()

	// Ресайзим
	container.Body, err = resize(targetResp.Body, requestedPreview.Width, requestedPreview.Height)
	if err != nil {
		s.logger.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	// Копируем заголовки целевого сервера и корректируем Content-Length
	for name := range targetResp.Header {
		container.SetHeader(name, targetResp.Header.Get(name))
	}
	container.SetHeader("Content-Length", strconv.Itoa(len(container.Body)))

	// Записываем уменьшенную картику в ответ и в кэш
	for name, value := range container.Headers {
		w.Header().Set(name, value)
	}
	w.Header().Set("X-Proxy", "proxy-resizer")
	w.Header().Set("X-Cached", "no")
	w.WriteHeader(targetResp.StatusCode)
	io.Copy(w, bytes.NewReader(container.Body))

	s.cache.Set(requestedPreview, container)
}

func resize(content io.ReadCloser, width, height int) ([]byte, error) {
	targetImage, err := imaging.Decode(content)
	if err != nil {
		return nil, err
	}
	resizedImage := imaging.Fit(targetImage, width, height, imaging.Lanczos)
	buf := bytes.Buffer{}
	err = jpeg.Encode(&buf, resizedImage, nil)
	if err != nil {
		return nil, ErrResizeError
	}

	return buf.Bytes(), nil
}
