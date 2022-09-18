package preview

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

const (
	MaxWidth  = 4096
	MaxHeight = 4096
)

type Image struct {
	Path   string
	Width  int
	Height int
}

var ErrBadImageRequestURL = errors.New("bad image request url")

var urlRexExp = regexp.MustCompile(`^\/fill\/(\d+)\/(\d+)/((?:[\/a-z\d\-\._])+\.jpe?g)$`)

func NewFromURL(uri string) (Image, error) {
	parts := urlRexExp.FindStringSubmatch(uri)
	if len(parts) != 3+1 {
		return Image{}, ErrBadImageRequestURL
	}

	w, _ := strconv.Atoi(parts[1])
	if w < 1 || w > MaxWidth {
		return Image{}, ErrBadImageRequestURL
	}

	h, _ := strconv.Atoi(parts[2])
	if h < 1 || h > MaxHeight {
		return Image{}, ErrBadImageRequestURL
	}

	return Image{
		Path:   parts[3],
		Width:  w,
		Height: h,
	}, nil
}

func (i Image) Key() string {
	hash := sha256.New()
	hash.Write([]byte(i.Path))
	h := hash.Sum(nil)
	str := fmt.Sprintf("%dx%d-%x", i.Width, i.Height, h)
	return str
}
