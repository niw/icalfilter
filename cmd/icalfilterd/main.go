package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/niw/icalfilter"
)

var (
	// Listening address.
	addr = flag.String("addr", "127.0.0.1", "Listening address")
	// Listening port.
	port = flag.Int("port", 3000, "Listening port")
	// Request timeout.
	timeout = flag.Int64("timeout", 5000, "Response timeout in milliseconds")
	// Cache duration.
	cachettl = flag.Int64("cachettl", 300, "Cache expiry duration in seconds")
)

// Default to filter a calendar by removing 3 months and older events.
const defaultMonths = 3

// Duration that is reserved for processing.
const processDuration = 500 * time.Millisecond

func writeErrorMessage(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func main() {
	flag.Parse()

	cacheDuration := time.Duration(*cachettl) * time.Second
	f := NewFetcher(cacheDuration, 10)

	fetchTimeout := time.Duration(*timeout) * time.Millisecond
	if fetchTimeout > processDuration {
		fetchTimeout = fetchTimeout - processDuration
	}

	http.HandleFunc("/filter", func(w http.ResponseWriter, r *http.Request) {
		url := r.FormValue("url")
		if url == "" {
			writeErrorMessage(w, http.StatusBadRequest, "url is required")
			return
		}

		months := defaultMonths
		monthsStr := r.FormValue("months")
		if monthsStr != "" {
			m, err := strconv.ParseInt(monthsStr, 10, 0)
			if err != nil {
				writeErrorMessage(w, http.StatusBadRequest, err.Error())
				return
			}
			if m < 0 {
				writeErrorMessage(w, http.StatusBadRequest, "months must be positive")
				return
			}
			months = int(m)
		}

		ctx := context.Background()
		if fetchTimeout > 0 {
			ctxWithTimeout, cancel := context.WithTimeout(ctx, fetchTimeout)
			ctx = ctxWithTimeout
			defer cancel()
		}

		body, statusCode, err := f.Fetch(ctx, url)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]interface{}{"message": err.Error(), "status": statusCode, "response": string(body)})
			return
		}

		c, err := icalfilter.Parse(string(body))
		if err != nil {
			writeErrorMessage(w, http.StatusInternalServerError, err.Error())
			return
		}
		defer c.Close()

		before := time.Now().AddDate(0, -months, 0)
		err = c.FilterBefore(before)
		if err != nil {
			writeErrorMessage(w, http.StatusInternalServerError, err.Error())
			return
		}

		w.Header().Set("Content-Type", "text/calendar; charset=UTF-8")
		fmt.Fprint(w, c.String())
	})

	http.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})

	err := http.ListenAndServe(fmt.Sprintf("%s:%d", *addr, *port), nil)
	if err != nil {
		log.Fatalln(err)
	}
}
