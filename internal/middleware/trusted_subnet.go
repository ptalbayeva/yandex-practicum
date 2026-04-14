package g

import (
	"net"
	"net/http"

	"github.com/yandex-practicum/shorten-url/internal/config"
)

// TrustedSubnetMiddleware проверка CIDR
func TrustedSubnetMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			realIP := r.Header.Get("X-Real-IP")
			ip := net.ParseIP(realIP)
			if ip == nil {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			_, ipNet, err := net.ParseCIDR(cfg.TrustedSubnet)
			if err != nil || !ipNet.Contains(ip) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
