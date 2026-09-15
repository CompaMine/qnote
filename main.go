package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

type App struct {
	store *Store
}

func main() {
	port := envOr("PORT", "8443")
	publicHost := os.Getenv("PUBLIC_HOST")
	certFile := envOr("TLS_CERT", "/data/cert.pem")
	keyFile := envOr("TLS_KEY", "/data/key.pem")

	if _, err := os.Stat(certFile); err != nil {
		log.Fatalf("TLS certificate not found at %s — run entrypoint to generate it", certFile)
	}
	if _, err := os.Stat(keyFile); err != nil {
		log.Fatalf("TLS key not found at %s", keyFile)
	}

	app := &App{store: NewStore()}
	limiter := newRateLimiter(20, time.Minute) // 20 creates per IP per minute

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", app.handleIndex)
	mux.HandleFunc("GET /note/{id}", app.handleNotePage)
	mux.HandleFunc("POST /api/notes", limiter.middleware(app.handleCreateNote))
	mux.HandleFunc("GET /api/notes/{id}", app.handleGetNote)
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	handler := securityHeaders(mux)

	addr := ":" + port
	server := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}

	host := publicHost
	if host == "" {
		host = detectIP()
	}

	fmt.Printf("\n✔ Сервис запущен: https://%s:%s\n", host, port)
	fmt.Println("⚠ Используется самоподписанный сертификат — браузер покажет предупреждение.")
	fmt.Println("   Записки хранятся только в памяти; перезапуск контейнера очищает все данные.\n")

	log.Printf("listening on %s", addr)
	if err := server.ListenAndServeTLS(certFile, keyFile); err != nil {
		log.Fatal(err)
	}
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; script-src 'self'; style-src 'self'; img-src 'none'; connect-src 'self'; base-uri 'none'; form-action 'self'; frame-ancestors 'none'")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Robots-Tag", "noindex, nofollow")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

func detectIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "127.0.0.1"
	}
	defer conn.Close()
	addr := conn.LocalAddr().(*net.UDPAddr)
	return addr.IP.String()
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
