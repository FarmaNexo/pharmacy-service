// internal/presentation/http/routes/routes.go
package routes

import (
	"net/http"

	"github.com/farmanexo/pharmacy-service/internal/presentation/http/controllers"
	"github.com/farmanexo/pharmacy-service/internal/presentation/http/middlewares"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"
)

func SetupRoutes(
	pharmacyController *controllers.PharmacyController,
	authMiddleware *middlewares.AuthMiddleware,
) *chi.Mux {
	r := chi.NewRouter()

	// ========================================
	// MIDDLEWARES GLOBALES
	// ========================================

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "https://farmanexo.pe"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Use(middlewares.CorrelationID)

	// ========================================
	// SWAGGER DOCUMENTATION
	// ========================================

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:4004/swagger/doc.json"),
	))

	// ========================================
	// HEALTH CHECK
	// ========================================

	r.Get("/health", pharmacyController.HealthCheck)
	r.Get("/", pharmacyController.HealthCheck)

	// ========================================
	// API ROUTES - VERSION 1
	// ========================================

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/pharmacies", func(r chi.Router) {
			// ========================================
			// ENDPOINTS PÚBLICOS
			// ========================================
			r.Get("/", pharmacyController.ListPharmacies)
			r.Post("/nearby", pharmacyController.SearchNearbyPharmacies)
			r.Get("/{id}", pharmacyController.GetPharmacy)
			r.Get("/{id}/inventory", pharmacyController.GetPharmacyInventory)
			r.Get("/{id}/hours", pharmacyController.GetPharmacyHours)

			// ========================================
			// ENDPOINTS OWNER (pharmacy_owner o admin)
			// ========================================
			r.Group(func(r chi.Router) {
				r.Use(authMiddleware.RequireAuth)
				r.Use(authMiddleware.RequireOwner)

				r.Post("/", pharmacyController.CreatePharmacy)
				r.Put("/{id}", pharmacyController.UpdatePharmacy)

				// Inventario
				r.Post("/{id}/inventory", pharmacyController.AddInventoryItem)
				r.Put("/{id}/inventory/{productId}", pharmacyController.UpdateInventoryItem)
				r.Delete("/{id}/inventory/{productId}", pharmacyController.RemoveInventoryItem)

				// Horarios
				r.Put("/{id}/hours", pharmacyController.UpdatePharmacyHours)

				// Documento de autorización
				r.Put("/{id}/authorization-document", pharmacyController.UploadAuthorizationDocument)
			})

			// ========================================
			// ENDPOINTS ADMIN
			// ========================================
			r.Group(func(r chi.Router) {
				r.Use(authMiddleware.RequireAuth)
				r.Use(authMiddleware.RequireAdmin)

				r.Put("/{id}/verify", pharmacyController.VerifyPharmacy)
				r.Delete("/{id}", pharmacyController.DeletePharmacy)
			})
		})

		// ========================================
		// CADENAS DE FARMACIAS
		// ========================================
		r.Route("/chains", func(r chi.Router) {
			r.Get("/{chainId}/pharmacies", pharmacyController.ListPharmaciesByChain)
		})
	})

	// ========================================
	// API ROUTES - VERSION 2 (Futuro)
	// ========================================

	r.Route("/api/v2", func(r chi.Router) {
		r.Get("/status", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("API v2 - Próximamente"))
		})
	})

	return r
}
