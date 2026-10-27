package main

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"metrics/middleware"
	"net/http"
	"net/http/pprof"
	_ "net/http/pprof"
)

type noGzipWriter struct {
	gin.ResponseWriter
}

func main() {
	go func() {
		log.Println("pprof is available at http://localhost:6060/debug/pprof")
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	r := gin.Default()

	r.Use(middleware.PrometheusMiddleware())

	r.GET("/ping", func(c *gin.Context) { c.String(http.StatusOK, "pong") })
	r.GET("/metrics/goroutine", textPprofHandler("goroutine"))
	r.GET("/metrics/mutex", textPprofHandler("mutex"))
	r.GET("/metrics/heap", textPprofHandler("heap"))

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	log.Fatal(r.Run(":8080"))
}

func (w *noGzipWriter) WriteHeader(code int) {
	w.Header().Set("Content-Encoding", "identity")
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.ResponseWriter.WriteHeader(code)
}

func (w *noGzipWriter) Write(data []byte) (int, error) {
	if w.Status() == 0 {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(data)
}

func textPprofHandler(profile string) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := c.Request.URL.Query()
		query.Set("debug", "1")
		c.Request.URL.RawQuery = query.Encode()
		noGzipW := &noGzipWriter{ResponseWriter: c.Writer}
		pprof.Handler(profile).ServeHTTP(noGzipW, c.Request)
	}
}
