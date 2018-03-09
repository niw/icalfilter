package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/niw/icalfilter"
)

var (
	addr = flag.String("addr", "127.0.0.1", "Listening address")
	port = flag.Int("port", 3000, "Listening port")
)

// Default to filter a calendar by removing 3 months and older events.
const defaultMonths = 3

func writeErrorMessage(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func main() {
	// Set flag values from compatible envrionment variables.
	flag.VisitAll(func(f *flag.Flag) {
		if s := os.Getenv(strings.ToUpper(f.Name)); s != "" {
			f.Value.Set(s)
		}
	})
	flag.Parse()

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

		resp, err := http.Get(url)
		if err != nil {
			writeErrorMessage(w, http.StatusInternalServerError, err.Error())
			return
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			writeErrorMessage(w, http.StatusInternalServerError, err.Error())
			return
		}

		if resp.StatusCode != http.StatusOK {
			writeJSON(w, http.StatusInternalServerError, map[string]interface{}{"message": "Fail to GET url.", "status": resp.StatusCode, "response": string(body)})
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
