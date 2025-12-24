package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/seuuser/focus-company-integration-service/internal/config"
	"github.com/seuuser/focus-company-integration-service/internal/focus"
	"github.com/seuuser/focus-company-integration-service/internal/handler"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func RegisterRoutes(r *chi.Mux, cfg config.Config) {
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CorsAllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"X-Total-Count", "Rate-Limit-Limit", "Rate-Limit-Remaining", "Rate-Limit-Reset"},
		AllowCredentials: true,
		MaxAge:           300, // 5 minutes
	}))

	r.Get("/health", handler.Health)

	focusClient := focus.NewClient(cfg.FocusURL, cfg.FocusToken)
	empresas := handler.NewEmpresasHandler(focusClient)

	r.Route("/v2/empresas", func(r chi.Router) {
		r.Post("/", empresas.CreateEmpresa)
		r.Get("/", empresas.ListEmpresas)

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", empresas.GetEmpresa)
			r.Put("/", empresas.UpdateEmpresa)
			r.Delete("/", empresas.DeleteEmpresa)
		})
	})

	r.Get("/swagger/*", httpSwagger.WrapHandler)
}


