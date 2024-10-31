package gateway

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func main() {
	r := gin.Default()

	authService, _ := url.Parse("http://localhost:8082/")
	productService, _ := url.Parse("http://localhost:8081/")

	r.Any("/auth/*proxyPath", reverseProxy(authService))
	r.Any("/products/*proxyPath", reverseProxy(productService))

	log.Println("API Gateway запущен на порту 8080")
	r.Run(":8080")
}

// reverseProxy возвращает обработчик для проксирования запросов
func reverseProxy(target *url.URL) gin.HandlerFunc {
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
		http.Error(rw, "Сервис недоступен", http.StatusServiceUnavailable)
	}

	return func(c *gin.Context) {
		// Перенаправляем запрос на целевой сервис
		req := c.Request
		req.URL.Path = c.Param("proxyPath")
		req.Host = target.Host
		proxy.ServeHTTP(c.Writer, req)
	}
}
