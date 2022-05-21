package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/iamcathal/galahad/statsmonitoring"

	"github.com/gorilla/mux"
)

// UptimeResponse is the standard response
// for any service's /status endpoint
type UptimeResponse struct {
	Status string        `json:"status"`
	Uptime time.Duration `json:"uptime"`
}

var (
	ApplicationStartUpTime time.Time
)

func DisallowFileBrowsing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = filepath.Clean(r.URL.Path)
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/static") {
			next.ServeHTTP(w, r)
			return
		}
		http.NotFound(w, r)
	})
}

func Status(w http.ResponseWriter, r *http.Request) {
	req := UptimeResponse{
		Uptime: time.Since(ApplicationStartUpTime),
		Status: "operational",
	}
	jsonObj, err := json.MarshalIndent(req, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, string(jsonObj))
}

func Metrics(w http.ResponseWriter, r *http.Request) {
	jsonObj, err := json.MarshalIndent(statsmonitoring.GetMetrics(), "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, string(jsonObj))
}

func Home(w http.ResponseWriter, r *http.Request) {
	jsonObj, err := json.MarshalIndent(statsmonitoring.GetMetrics(), "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, string(jsonObj))
}

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%v %+v\n", time.Now().Format(time.RFC3339), r)
		next.ServeHTTP(w, r)
	})
}

func main() {
	port := "2944"
	ApplicationStartUpTime = time.Now()

	r := mux.NewRouter()
	r.HandleFunc("/home", Home).Methods("GET")
	r.HandleFunc("/status", Status).Methods("POST")
	r.HandleFunc("/metrics", Metrics).Methods("POST")

	r.Handle("/static", http.NotFoundHandler())
	fs := http.FileServer(http.Dir("./static"))
	r.PathPrefix("/").Handler(DisallowFileBrowsing(fs))
	r.Use(logMiddleware)

	go statsmonitoring.CollectAndShipStats()

	srv := &http.Server{
		Handler:      r,
		Addr:         ":" + fmt.Sprint(2944),
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}
	fmt.Println("serving requests on :" + port)
	log.Fatal(srv.ListenAndServe())
}
