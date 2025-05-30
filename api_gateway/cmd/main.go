package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
)

// proxy возвращает gin.HandlerFunc, проксирующий запрос к target
func proxy(target *url.URL) gin.HandlerFunc {
	return func(c *gin.Context) {
		director := func(req *http.Request) {
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.URL.Path = c.Param("proxyPath")
			if raw := c.Request.URL.RawQuery; raw != "" {
				req.URL.RawQuery = raw
			}
			req.Header = c.Request.Header
		}
		p := &httputil.ReverseProxy{Director: director}
		p.ServeHTTP(c.Writer, c.Request)
	}
}

func main() {
	r := gin.Default()

	// Список сервисов и их базовые URL
	services := map[string]string{
		"books":         "http://book_service:50051",
		"users":         "http://user_service:50052",
		"libraries":     "http://user_library_service:50053",
		"exchange":      "http://exchange_service:50054",
		"orders":        "http://order_service:50055",
		"notifications": "http://notification_service:50056",
	}

	// Для каждого сервиса заводим маршрут вида /<service>/*proxyPath
	for prefix, addr := range services {
		target, err := url.Parse(addr)
		if err != nil {
			log.Fatalf("invalid URL for %s: %v", prefix, err)
		}
		group := r.Group("/" + prefix)
		group.Any("/*proxyPath", proxy(target))
	}

	log.Println("🚀 API Gateway running on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("failed to run API Gateway: %v", err)
	}
}
