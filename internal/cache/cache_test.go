package cache

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/andrey-tushev/otus-go/project/internal/preview"
)

func TestCacheRace(t *testing.T) {
	cache := New("../../cache", 1000)

	p0 := preview.Image{Path: "/test-0.jpg", Width: 100, Height: 100}
	p1 := preview.Image{Path: "/test-1.jpg", Width: 100, Height: 100}
	p2 := preview.Image{Path: "/test-2.jpg", Width: 100, Height: 100}

	c1 := preview.Container{Body: []byte("TEST-1")}
	c2 := preview.Container{Body: []byte("TEST-2")}

	cache.Set(p1, &c1)
	cache.Set(p2, &c2)
	require.Nil(t, cache.Get(p0))
	require.NotNil(t, cache.Get(p1))
	require.EqualValues(t, "TEST-1", cache.Get(p1).Body)
	require.NotNil(t, cache.Get(p2))
	require.EqualValues(t, "TEST-2", cache.Get(p2).Body)
}
