package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/seuuser/focus-company-integration-service/docs"
	"github.com/seuuser/focus-company-integration-service/internal/config"
	"github.com/seuuser/focus-company-integration-service/internal/server"
	"github.com/seuuser/focus-company-integration-service/internal/supabase"
)

// @title           Focus Company Integration Service
// @version         0.1.0
// @description     Microservice para integração de cadastro/consulta/edição/remoção de empresas na Focus NFe (v2).
// @host            localhost:8082
// @BasePath        /
func main() {
	cfg := config.Load()
	supabase.InitClient()

	// sobrescreve info do Swagger (PRD deve setar SWAGGER_HOST e (opcionalmente) SWAGGER_SCHEMES)
	// Ex:
	// - SWAGGER_HOST=api-focus-company-integration.carteiracontabil.com
	// - SWAGGER_SCHEMES=https
	swaggerHost := strings.TrimSpace(os.Getenv("SWAGGER_HOST"))
	if swaggerHost == "" {
		swaggerHost = strings.TrimSpace(os.Getenv("PUBLIC_HOST"))
	}
	if swaggerHost == "" {
		swaggerHost = "localhost:" + cfg.Port
	}

	docs.SwaggerInfo.Host = swaggerHost
	docs.SwaggerInfo.BasePath = "/"

	if schemes := strings.TrimSpace(os.Getenv("SWAGGER_SCHEMES")); schemes != "" {
		parts := strings.Split(schemes, ",")
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			if s := strings.TrimSpace(p); s != "" {
				out = append(out, s)
			}
		}
		if len(out) > 0 {
			docs.SwaggerInfo.Schemes = out
		}
	}

	r := chi.NewRouter()
	server.RegisterRoutes(r, cfg)

	log.Printf("listening on :%s …", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}


