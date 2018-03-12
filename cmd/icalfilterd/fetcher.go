package main

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

type result struct {
	body       []byte
	statusCode int
	err        error
}

type entry struct {
	res    result
	expiry time.Time
}

type Fetcher struct {
	cache *sync.Map
	ttl   time.Duration
}

func fetch(url string) result {
	resp, err := http.Get(url)
	if err != nil {
		return result{nil, 0, err}
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return result{nil, 0, err}
	}

	if resp.StatusCode != http.StatusOK {
		err = errors.New("request has not succeeded")
	}

	return result{body, resp.StatusCode, err}
}

func (f *Fetcher) fetch(url string) result {
	i, ok := f.cache.Load(url)
	if ok {
		e := i.(entry)
		if time.Now().Before(e.expiry) {
			// Cache hit
			return e.res
		}
	}
	// Cache miss
	res := fetch(url)
	if res.err == nil {
		expiry := time.Now().Add(f.ttl)
		f.cache.Store(url, entry{res, expiry})
	}
	return res
}

func (f *Fetcher) Fetch(ctx context.Context, url string) ([]byte, int, error) {
	// Give one buffer size to allow channel is writable even after the context is done.
	cresult := make(chan result, 1)

	go func() {
		cresult <- f.fetch(url)
	}()

	select {
	case <-ctx.Done():
		return nil, 0, ctx.Err()
	case res := <-cresult:
		return res.body, res.statusCode, res.err
	}
}

func NewFetcher(d time.Duration) *Fetcher {
	return &Fetcher{&sync.Map{}, d}
}
