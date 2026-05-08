// cmd/server/main.go
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/farmanexo/pharmacy-service/internal/application/commands"
	"github.com/farmanexo/pharmacy-service/internal/application/handlers"
	"github.com/farmanexo/pharmacy-service/internal/application/postprocessors"
	"github.com/farmanexo/pharmacy-service/internal/application/preprocessors"
	"github.com/farmanexo/pharmacy-service/internal/application/validators"
	"github.com/farmanexo/pharmacy-service/internal/infrastructure/cache"
	"github.com/farmanexo/pharmacy-service/internal/infrastructure/clients"
	"github.com/farmanexo/pharmacy-service/internal/infrastructure/messaging"
	"github.com/farmanexo/pharmacy-service/internal/infrastructure/persistence/postgres"
	"github.com/farmanexo/pharmacy-service/internal/infrastructure/security"
	"github.com/farmanexo/pharmacy-service/internal/infrastructure/storage"
	"github.com/farmanexo/pharmacy-service/internal/presentation/dto/responses"
	"github.com/farmanexo/pharmacy-service/internal/presentation/http/controllers"
	"github.com/farmanexo/pharmacy-service/internal/presentation/http/middlewares"
	"github.com/farmanexo/pharmacy-service/internal/presentation/http/routes"
	"github.com/farmanexo/pharmacy-service/pkg/config"
	"github.com/farmanexo/pharmacy-service/pkg/mediator"

	// Swagger docs
	_ "github.com/farmanexo/pharmacy-service/docs"

	"go.uber.org/zap"
	pgdriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// @title           FarmaNexo Pharmacy Service API
// @version         1.0
// @description     Servicio de gestión de farmacias para FarmaNexo - Microservicio con CQRS, Clean Architecture y PostGIS
// @termsOfService  https://farmanexo.pe/terms

// @contact.name    FarmaNexo API Support
// @contact.url     https://farmanexo.pe/support
// @contact.email   support@farmanexo.pe

// @license.name    Apache 2.0
// @license.url     http://www.apache.org/licenses/LICENSE-2.0.html

// @host            localhost:4004
// @BasePath        /api/v1

// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 JWT Authorization header using the Bearer scheme. Example: "Bearer {token}"

// @tag.name         Pharmacies
// @tag.description  Endpoints de gestión de farmacias

// @tag.name         Inventory
// @tag.description  Endpoints de inventario de farmacias

// @tag.name         Admin
// @tag.description  Endpoints de administración

// @tag.name         Health
// @tag.description  Endpoints de salud del servicio

func main() {
	env := getEnvironment()
	cfg, err := config.LoadConfig(env)
	if err != nil {
		panic(fmt.Sprintf("Error cargando configuración: %v", err))
	}

	logger := initLogger(cfg)
	defer logger.Sync()

	logger.Info("Iniciando Pharmacy Service",
		zap.String("environment", cfg.Environment),
		zap.Int("port", cfg.Server.Port),
	)

	db := initDatabase(cfg, logger)

	logger.Info("Auto-migration deshabilitado - Usar migraciones manuales (PostGIS)")

	// ========================================
	// REPOSITORIOS
	// ========================================
	pharmacyRepo := postgres.NewPharmacyRepository(db, logger)
	hoursRepo := postgres.NewPharmacyHoursRepository(db, logger)
	inventoryRepo := postgres.NewInventoryRepository(db, logger)

	// ========================================
	// SERVICIOS
	// ========================================
	jwtService := security.NewJWTService(cfg.JWT.Secret, logger)

	eventPublisher, err := messaging.NewSQSEventPublisher(cfg.AWS, cfg.SQS, logger)
	if err != nil {
		logger.Fatal("Error inicializando SQS EventPublisher", zap.Error(err))
	}

	redisClient, err := cache.NewRedisClient(cfg.Redis, cfg.Environment, logger)
	if err != nil {
		logger.Fatal("Error inicializando Redis", zap.Error(err))
	}
	defer redisClient.Close()

	cacheService := cache.NewRedisCacheService(redisClient, logger)

	fileStorage, err := storage.NewS3FileStorage(cfg.AWS, logger)
	if err != nil {
		logger.Fatal("Error inicializando S3 FileStorage", zap.Error(err))
	}

	// ========================================
	// MEDIATOR
	// ========================================
	med := mediator.NewMediator()

	// ========================================
	// HANDLERS - Pharmacies (Queries)
	// ========================================
	listPharmaciesHandler := handlers.NewListPharmaciesHandler(pharmacyRepo, logger)
	mediator.RegisterHandler(med, listPharmaciesHandler)

	getPharmacyHandler := handlers.NewGetPharmacyHandler(pharmacyRepo, hoursRepo, cacheService, logger)
	mediator.RegisterHandler(med, getPharmacyHandler)

	getPharmacyBySlugHandler := handlers.NewGetPharmacyBySlugHandler(pharmacyRepo, hoursRepo, cacheService, logger)
	mediator.RegisterHandler(med, getPharmacyBySlugHandler)

	listInventoryByProductHandler := handlers.NewListInventoryByProductHandler(inventoryRepo, cacheService, logger)
	mediator.RegisterHandler(med, listInventoryByProductHandler)

	searchNearbyHandler := handlers.NewSearchNearbyPharmaciesHandler(pharmacyRepo, cacheService, logger)
	mediator.RegisterHandler(med, searchNearbyHandler)

	listByChainHandler := handlers.NewListPharmaciesByChainHandler(pharmacyRepo, cacheService, logger)
	mediator.RegisterHandler(med, listByChainHandler)

	// ========================================
	// HANDLERS - Pharmacies (Commands)
	// ========================================
	createPharmacyHandler := handlers.NewCreatePharmacyHandler(pharmacyRepo, eventPublisher, cacheService, logger)
	mediator.RegisterHandler(med, createPharmacyHandler)

	updatePharmacyHandler := handlers.NewUpdatePharmacyHandler(pharmacyRepo, cacheService, logger)
	mediator.RegisterHandler(med, updatePharmacyHandler)

	deletePharmacyHandler := handlers.NewDeletePharmacyHandler(pharmacyRepo, cacheService, logger)
	mediator.RegisterHandler(med, deletePharmacyHandler)

	verifyPharmacyHandler := handlers.NewVerifyPharmacyHandler(pharmacyRepo, eventPublisher, cacheService, logger)
	mediator.RegisterHandler(med, verifyPharmacyHandler)

	uploadAuthDocHandler := handlers.NewUploadAuthorizationDocumentHandler(
		pharmacyRepo, fileStorage, eventPublisher, cacheService, cfg.S3.PharmaciesBucket, logger,
	)
	mediator.RegisterHandler(med, uploadAuthDocHandler)

	// ========================================
	// HANDLERS - Inventory
	// ========================================
	listInventoryHandler := handlers.NewListPharmacyInventoryHandler(pharmacyRepo, inventoryRepo, cacheService, logger)
	mediator.RegisterHandler(med, listInventoryHandler)

	addInventoryHandler := handlers.NewAddInventoryItemHandler(pharmacyRepo, inventoryRepo, eventPublisher, cacheService, logger)
	mediator.RegisterHandler(med, addInventoryHandler)

	updateInventoryHandler := handlers.NewUpdateInventoryItemHandler(inventoryRepo, eventPublisher, cacheService, logger)
	mediator.RegisterHandler(med, updateInventoryHandler)

	removeInventoryHandler := handlers.NewRemoveInventoryItemHandler(inventoryRepo, cacheService, logger)
	mediator.RegisterHandler(med, removeInventoryHandler)

	// ========================================
	// HANDLERS - Hours
	// ========================================
	getHoursHandler := handlers.NewGetPharmacyHoursHandler(pharmacyRepo, hoursRepo, logger)
	mediator.RegisterHandler(med, getHoursHandler)

	updateHoursHandler := handlers.NewUpdatePharmacyHoursHandler(pharmacyRepo, hoursRepo, cacheService, logger)
	mediator.RegisterHandler(med, updateHoursHandler)

	// ========================================
	// HANDLERS - Scraper events (Tier 5)
	// No registrados en mediator: invocados directamente por el SQS consumer.
	// ========================================
	catalogClient := clients.NewCatalogClient(cfg.Services.CatalogService.BaseURL, logger)
	upsertPharmacyFromEventHandler := handlers.NewUpsertPharmacyFromEventHandler(pharmacyRepo, logger)
	upsertInventoryFromEventHandler := handlers.NewUpsertInventoryFromEventHandler(pharmacyRepo, inventoryRepo, catalogClient, logger)

	// ========================================
	// VALIDATORS
	// ========================================
	createPharmacyValidator := validators.NewCreatePharmacyValidator()
	mediator.RegisterValidator[commands.CreatePharmacyCommand, responses.PharmacyResponse](med, createPharmacyValidator)

	addInventoryValidator := validators.NewAddInventoryItemValidator()
	mediator.RegisterValidator[commands.AddInventoryItemCommand, responses.InventoryItemResponse](med, addInventoryValidator)

	// ========================================
	// PREPROCESSORS Y POSTPROCESSORS
	// ========================================
	sanitizePreProcessor := preprocessors.NewSanitizeInputPreProcessor(logger)
	med.RegisterPreProcessor(sanitizePreProcessor)

	auditPostProcessor := postprocessors.NewLogAuditPostProcessor(logger)
	med.RegisterPostProcessor(auditPostProcessor)

	logger.Info("Mediator configurado",
		zap.Int("handlers", 15),
		zap.Int("validators", 2),
		zap.Int("preprocessors", 1),
		zap.Int("postprocessors", 1),
	)

	// ========================================
	// MIDDLEWARES
	// ========================================
	authMiddleware := middlewares.NewAuthMiddleware(jwtService, logger)

	// ========================================
	// CONTROLADORES Y RUTAS
	// ========================================
	pharmacyController := controllers.NewPharmacyController(med, logger)
	router := routes.SetupRoutes(pharmacyController, authMiddleware)

	// ========================================
	// SQS Scraper Consumer (Tier 5)
	// Consume PHARMACY_DISCOVERED + INVENTORY_DISCOVERED de
	// farmanexo-{env}-scraper-events. Para INVENTORY hace lookup local
	// de pharmacy + lookup HTTP a catalog-service para resolver product_id.
	// ========================================
	scraperConsumer, err := messaging.NewSQSScraperConsumer(cfg.AWS, cfg.SQS, upsertPharmacyFromEventHandler, upsertInventoryFromEventHandler, logger)
	if err != nil {
		logger.Fatal("Error inicializando SQS ScraperConsumer", zap.Error(err))
	}
	scraperConsumerCtx, scraperConsumerCancel := context.WithCancel(context.Background())
	defer scraperConsumerCancel()
	scraperConsumer.Start(scraperConsumerCtx)

	// ========================================
	// SERVIDOR HTTP
	// ========================================
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	go func() {
		logger.Info("Servidor HTTP iniciado",
			zap.String("address", server.Addr),
			zap.String("swagger_url", fmt.Sprintf("http://localhost:%d/swagger/index.html", cfg.Server.Port)),
		)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Error iniciando servidor", zap.Error(err))
		}
	}()

	// ========================================
	// GRACEFUL SHUTDOWN
	// ========================================
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Iniciando graceful shutdown...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	scraperConsumer.Stop()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Error en shutdown", zap.Error(err))
	}

	logger.Info("Servidor detenido exitosamente")
}

func getEnvironment() string {
	env := os.Getenv("ENV")
	if env == "" {
		env = "local"
	}
	return env
}

func initLogger(cfg *config.Config) *zap.Logger {
	var logger *zap.Logger
	var err error

	if cfg.IsProduction() || cfg.IsUAT() {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}

	if err != nil {
		panic(fmt.Sprintf("Error inicializando logger: %v", err))
	}

	return logger
}

func initDatabase(cfg *config.Config, logger *zap.Logger) *gorm.DB {
	gormLogLevel := gormlogger.Silent
	if cfg.IsDevelopment() {
		gormLogLevel = gormlogger.Info
	}

	gormLogger := gormlogger.Default.LogMode(gormLogLevel)

	db, err := gorm.Open(pgdriver.Open(cfg.Database.GetDSN()), &gorm.Config{
		Logger: gormLogger,
	})

	if err != nil {
		logger.Fatal("Error conectando a PostgreSQL",
			zap.Error(err),
			zap.String("host", cfg.Database.Host),
			zap.Int("port", cfg.Database.Port),
		)
	}

	sqlDB, err := db.DB()
	if err != nil {
		logger.Fatal("Error obteniendo SQL DB", zap.Error(err))
	}

	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	logger.Info("Conexión a PostgreSQL establecida",
		zap.String("host", cfg.Database.Host),
		zap.String("database", cfg.Database.DBName),
	)

	return db
}
