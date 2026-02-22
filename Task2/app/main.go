package main

import (
	"fmt"
	"math"
	"net/http"
	"os"
	"sync/atomic"
)

var requestCount uint64

func main() {
	hostname, _ := os.Hostname()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddUint64(&requestCount, 1)
		
		// Создаём CPU нагрузку для демонстрации масштабирования
		result := 0.0
		for i := 0; i < 50000; i++ {
			result += math.Sqrt(float64(i))
		}
		
		fmt.Fprintf(w, "Pod ID: %s\n", hostname)
	})

	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		count := atomic.LoadUint64(&requestCount)
		fmt.Fprintf(w, "# HELP http_requests_total Total number of HTTP requests\n")
		fmt.Fprintf(w, "# TYPE http_requests_total counter\n")
		fmt.Fprintf(w, "http_requests_total %d\n", count)
	})

	fmt.Println("Server starting on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}
