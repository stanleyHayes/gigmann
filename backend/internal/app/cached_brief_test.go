package app_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xcreativs/gigmann/internal/app"
	"github.com/xcreativs/gigmann/internal/core/brief"
	"github.com/xcreativs/gigmann/internal/core/severity"
)

type fakeBrief struct {
	mu     sync.Mutex
	calls  int
	err    error
	called chan struct{}
}

func (f *fakeBrief) Generate(context.Context) (brief.Brief, error) {
	f.mu.Lock()
	f.calls++
	f.mu.Unlock()
	if f.called != nil {
		f.called <- struct{}{}
	}
	if f.err != nil {
		return brief.Brief{}, f.err
	}
	return brief.Brief{
		ID: "b", Date: time.Now(), Prose: "p", Model: "m",
		Items: []brief.Item{{Severity: severity.Critical, FacilityID: "f", Headline: "h"}},
	}, nil
}

func (f *fakeBrief) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls
}

func TestCachedBriefServesCacheWithinTTL(t *testing.T) {
	inner := &fakeBrief{}
	c := app.NewCachedBrief(inner, time.Hour)

	_, err := c.Generate(context.Background())
	require.NoError(t, err)
	_, err = c.Generate(context.Background())
	require.NoError(t, err)

	assert.Equal(t, 1, inner.count(), "second call within TTL must serve the cache")
}

func TestCachedBriefColdErrorPropagates(t *testing.T) {
	inner := &fakeBrief{err: errors.New("api down")}
	c := app.NewCachedBrief(inner, time.Hour)

	_, err := c.Generate(context.Background())
	require.Error(t, err)
}

func TestCachedBriefRefreshesWhenStale(t *testing.T) {
	inner := &fakeBrief{called: make(chan struct{}, 8)}
	c := app.NewCachedBrief(inner, time.Millisecond)

	_, err := c.Generate(context.Background()) // cold generate
	require.NoError(t, err)
	<-inner.called

	time.Sleep(5 * time.Millisecond) // let the entry go stale

	b, err := c.Generate(context.Background()) // serves cache, triggers bg refresh
	require.NoError(t, err)
	assert.Equal(t, "p", b.Prose)

	select {
	case <-inner.called: // background refresh fired
	case <-time.After(2 * time.Second):
		t.Fatal("expected a background refresh of the stale brief")
	}
}
