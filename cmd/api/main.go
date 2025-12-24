package main

import (
	"log"
	"net/http"

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

	// sobrescreve info do Swagger
	docs.SwaggerInfo.Host = "localhost:" + cfg.Port
	docs.SwaggerInfo.BasePath = "/"

	r := chi.NewRouter()
	server.RegisterRoutes(r, cfg)

	log.Printf("listening on :%s …", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}


