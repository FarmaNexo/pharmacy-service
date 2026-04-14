// internal/presentation/http/controllers/pharmacy_controller.go
package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/farmanexo/pharmacy-service/internal/application/commands"
	"github.com/farmanexo/pharmacy-service/internal/application/queries"
	"github.com/farmanexo/pharmacy-service/internal/presentation/dto/requests"
	"github.com/farmanexo/pharmacy-service/internal/presentation/dto/responses"
	"github.com/farmanexo/pharmacy-service/internal/presentation/http/middlewares"
	"github.com/farmanexo/pharmacy-service/internal/shared/common"
	"github.com/farmanexo/pharmacy-service/pkg/mediator"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type PharmacyController struct {
	mediator *mediator.Mediator
	logger   *zap.Logger
}

func NewPharmacyController(med *mediator.Mediator, logger *zap.Logger) *PharmacyController {
	return &PharmacyController{mediator: med, logger: logger}
}

func (c *PharmacyController) respondJSON(w http.ResponseWriter, response interface{}) {
	statusCode := http.StatusOK

	if resp, ok := response.(interface{ GetHttpStatus() *int }); ok {
		if httpStatus := resp.GetHttpStatus(); httpStatus != nil {
			statusCode = *httpStatus
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		c.logger.Error("Error codificando respuesta JSON", zap.Error(err))
	}
}

// HealthCheck godoc
// @Summary      Health check del servicio
// @Description  Retorna el estado del servicio
// @Tags         Health
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "Servicio saludable"
// @Router       /health [get]
func (c *PharmacyController) HealthCheck(w http.ResponseWriter, r *http.Request) {
	type HealthResponse struct {
		Status  string `json:"status" example:"healthy"`
		Service string `json:"service" example:"pharmacy-service"`
		Version string `json:"version" example:"1.0.0"`
	}

	health := HealthResponse{
		Status:  "healthy",
		Service: "pharmacy-service",
		Version: "1.0.0",
	}

	c.respondJSON(w, common.OkResponse(health))
}

// ========================================
// FARMACIAS - ENDPOINTS PÚBLICOS
// ========================================

// ListPharmacies godoc
// @Summary      Listar farmacias
// @Description  Retorna lista paginada de farmacias activas
// @Tags         Pharmacies
// @Produce      json
// @Param        page   query    int  false  "Página"           default(1)
// @Param        limit  query    int  false  "Límite por página" default(20)
// @Success      200  {object}  common.ApiResponse[responses.PharmacyListResponse]
// @Failure      500  {object}  common.ApiResponse[responses.PharmacyListResponse]
// @Router       /api/v1/pharmacies [get]
func (c *PharmacyController) ListPharmacies(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	query := queries.ListPharmaciesQuery{Page: page, Limit: limit}

	response, _ := mediator.Send[queries.ListPharmaciesQuery, responses.PharmacyListResponse](r.Context(), c.mediator, query)
	c.respondJSON(w, response)
}

// GetPharmacy godoc
// @Summary      Obtener farmacia por ID
// @Description  Retorna el detalle de una farmacia
// @Tags         Pharmacies
// @Produce      json
// @Param        id   path     string  true  "Pharmacy ID"
// @Success      200  {object}  common.ApiResponse[responses.PharmacyResponse]
// @Failure      404  {object}  common.ApiResponse[responses.PharmacyResponse]
// @Router       /api/v1/pharmacies/{id} [get]
func (c *PharmacyController) GetPharmacy(w http.ResponseWriter, r *http.Request) {
	pharmacyID := chi.URLParam(r, "id")
	query := queries.GetPharmacyQuery{ID: pharmacyID}

	response, _ := mediator.Send[queries.GetPharmacyQuery, responses.PharmacyResponse](r.Context(), c.mediator, query)
	c.respondJSON(w, response)
}

// SearchNearbyPharmacies godoc
// @Summary      Farmacias cercanas
// @Description  Busca farmacias cercanas usando geolocalización (PostGIS)
// @Tags         Pharmacies
// @Accept       json
// @Produce      json
// @Param        body  body     requests.NearbyPharmaciesRequest  true  "Coordenadas y radio"
// @Success      200  {object}  common.ApiResponse[responses.NearbyPharmaciesResponse]
// @Failure      400  {object}  common.ApiResponse[responses.NearbyPharmaciesResponse]
// @Router       /api/v1/pharmacies/nearby [post]
func (c *PharmacyController) SearchNearbyPharmacies(w http.ResponseWriter, r *http.Request) {
	var req requests.NearbyPharmaciesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.respondJSON(w, common.BadRequestResponse[responses.NearbyPharmaciesResponse]("VAL_001", "Body inválido"))
		return
	}

	query := queries.SearchNearbyPharmaciesQuery{
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		RadiusKm:  req.RadiusKm,
		Limit:     req.Limit,
	}

	response, _ := mediator.Send[queries.SearchNearbyPharmaciesQuery, responses.NearbyPharmaciesResponse](r.Context(), c.mediator, query)
	c.respondJSON(w, response)
}

// GetPharmacyInventory godoc
// @Summary      Inventario de farmacia
// @Description  Retorna el inventario de una farmacia
// @Tags         Pharmacies
// @Produce      json
// @Param        id   path     string  true  "Pharmacy ID"
// @Success      200  {object}  common.ApiResponse[responses.InventoryListResponse]
// @Failure      404  {object}  common.ApiResponse[responses.InventoryListResponse]
// @Router       /api/v1/pharmacies/{id}/inventory [get]
func (c *PharmacyController) GetPharmacyInventory(w http.ResponseWriter, r *http.Request) {
	pharmacyID := chi.URLParam(r, "id")
	query := queries.ListPharmacyInventoryQuery{PharmacyID: pharmacyID}

	response, _ := mediator.Send[queries.ListPharmacyInventoryQuery, responses.InventoryListResponse](r.Context(), c.mediator, query)
	c.respondJSON(w, response)
}

// GetPharmacyHours godoc
// @Summary      Horarios de farmacia
// @Description  Retorna los horarios de una farmacia
// @Tags         Pharmacies
// @Produce      json
// @Param        id   path     string  true  "Pharmacy ID"
// @Success      200  {object}  common.ApiResponse[responses.HoursListResponse]
// @Failure      404  {object}  common.ApiResponse[responses.HoursListResponse]
// @Router       /api/v1/pharmacies/{id}/hours [get]
func (c *PharmacyController) GetPharmacyHours(w http.ResponseWriter, r *http.Request) {
	pharmacyID := chi.URLParam(r, "id")
	query := queries.GetPharmacyHoursQuery{PharmacyID: pharmacyID}

	response, _ := mediator.Send[queries.GetPharmacyHoursQuery, responses.HoursListResponse](r.Context(), c.mediator, query)
	c.respondJSON(w, response)
}

// ========================================
// FARMACIAS - ENDPOINTS OWNER
// ========================================

// CreatePharmacy godoc
// @Summary      Registrar farmacia
// @Description  Registra una nueva farmacia (requiere rol pharmacy_owner)
// @Tags         Pharmacies
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body     requests.CreatePharmacyRequest  true  "Datos de la farmacia"
// @Success      201  {object}  common.ApiResponse[responses.PharmacyResponse]
// @Failure      400  {object}  common.ApiResponse[responses.PharmacyResponse]
// @Failure      401  {object}  common.ApiResponse[responses.PharmacyResponse]
// @Router       /api/v1/pharmacies [post]
func (c *PharmacyController) CreatePharmacy(w http.ResponseWriter, r *http.Request) {
	var req requests.CreatePharmacyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.respondJSON(w, common.BadRequestResponse[responses.PharmacyResponse]("VAL_001", "Body inválido"))
		return
	}

	userID, _ := middlewares.GetUserIDFromContext(r.Context())

	cmd := commands.CreatePharmacyCommand{
		OwnerUserID: userID,
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		Phone:       req.Phone,
		Email:       req.Email,
		Website:     req.Website,
		Street:      req.Street,
		City:        req.City,
		State:       req.State,
		PostalCode:  req.PostalCode,
		Country:     req.Country,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		Is24h:       req.Is24h,
		ChainID:     req.ChainID,
		ChainName:   req.ChainName,
	}

	response, _ := mediator.Send[commands.CreatePharmacyCommand, responses.PharmacyResponse](r.Context(), c.mediator, cmd)
	c.respondJSON(w, response)
}

// UpdatePharmacy godoc
// @Summary      Actualizar farmacia
// @Description  Actualiza una farmacia existente (requiere ser owner)
// @Tags         Pharmacies
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path     string                          true  "Pharmacy ID"
// @Param        body  body     requests.UpdatePharmacyRequest  true  "Datos de la farmacia"
// @Success      200  {object}  common.ApiResponse[responses.PharmacyResponse]
// @Failure      404  {object}  common.ApiResponse[responses.PharmacyResponse]
// @Router       /api/v1/pharmacies/{id} [put]
func (c *PharmacyController) UpdatePharmacy(w http.ResponseWriter, r *http.Request) {
	pharmacyID := chi.URLParam(r, "id")

	var req requests.UpdatePharmacyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.respondJSON(w, common.BadRequestResponse[responses.PharmacyResponse]("VAL_001", "Body inválido"))
		return
	}

	cmd := commands.UpdatePharmacyCommand{
		ID:          pharmacyID,
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		Phone:       req.Phone,
		Email:       req.Email,
		Website:     req.Website,
		Street:      req.Street,
		City:        req.City,
		State:       req.State,
		PostalCode:  req.PostalCode,
		Country:     req.Country,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		Is24h:       req.Is24h,
		IsActive:    req.IsActive,
		ChainID:     req.ChainID,
		ChainName:   req.ChainName,
	}

	response, _ := mediator.Send[commands.UpdatePharmacyCommand, responses.PharmacyResponse](r.Context(), c.mediator, cmd)
	c.respondJSON(w, response)
}

// AddInventoryItem godoc
// @Summary      Agregar producto al inventario
// @Description  Agrega un producto al inventario de la farmacia (requiere ser owner)
// @Tags         Inventory
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path     string                          true  "Pharmacy ID"
// @Param        body  body     requests.AddInventoryItemRequest true  "Datos del inventario"
// @Success      201  {object}  common.ApiResponse[responses.InventoryItemResponse]
// @Failure      400  {object}  common.ApiResponse[responses.InventoryItemResponse]
// @Failure      409  {object}  common.ApiResponse[responses.InventoryItemResponse]
// @Router       /api/v1/pharmacies/{id}/inventory [post]
func (c *PharmacyController) AddInventoryItem(w http.ResponseWriter, r *http.Request) {
	pharmacyID := chi.URLParam(r, "id")

	var req requests.AddInventoryItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.respondJSON(w, common.BadRequestResponse[responses.InventoryItemResponse]("VAL_001", "Body inválido"))
		return
	}

	cmd := commands.AddInventoryItemCommand{
		PharmacyID:  pharmacyID,
		ProductID:   req.ProductID,
		Stock:       req.Stock,
		Price:       req.Price,
		IsAvailable: req.IsAvailable,
	}

	response, _ := mediator.Send[commands.AddInventoryItemCommand, responses.InventoryItemResponse](r.Context(), c.mediator, cmd)
	c.respondJSON(w, response)
}

// UpdateInventoryItem godoc
// @Summary      Actualizar inventario
// @Description  Actualiza stock/precio de un producto en inventario (requiere ser owner)
// @Tags         Inventory
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id         path     string                              true  "Pharmacy ID"
// @Param        productId  path     string                              true  "Product ID"
// @Param        body       body     requests.UpdateInventoryItemRequest  true  "Datos del inventario"
// @Success      200  {object}  common.ApiResponse[responses.InventoryItemResponse]
// @Failure      404  {object}  common.ApiResponse[responses.InventoryItemResponse]
// @Router       /api/v1/pharmacies/{id}/inventory/{productId} [put]
func (c *PharmacyController) UpdateInventoryItem(w http.ResponseWriter, r *http.Request) {
	pharmacyID := chi.URLParam(r, "id")
	productID := chi.URLParam(r, "productId")

	var req requests.UpdateInventoryItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.respondJSON(w, common.BadRequestResponse[responses.InventoryItemResponse]("VAL_001", "Body inválido"))
		return
	}

	cmd := commands.UpdateInventoryItemCommand{
		PharmacyID:  pharmacyID,
		ProductID:   productID,
		Stock:       req.Stock,
		Price:       req.Price,
		IsAvailable: req.IsAvailable,
	}

	response, _ := mediator.Send[commands.UpdateInventoryItemCommand, responses.InventoryItemResponse](r.Context(), c.mediator, cmd)
	c.respondJSON(w, response)
}

// RemoveInventoryItem godoc
// @Summary      Remover producto del inventario
// @Description  Remueve un producto del inventario (requiere ser owner)
// @Tags         Inventory
// @Produce      json
// @Security     BearerAuth
// @Param        id         path     string  true  "Pharmacy ID"
// @Param        productId  path     string  true  "Product ID"
// @Success      200  {object}  common.ApiResponse[responses.EmptyResponse]
// @Failure      404  {object}  common.ApiResponse[responses.EmptyResponse]
// @Router       /api/v1/pharmacies/{id}/inventory/{productId} [delete]
func (c *PharmacyController) RemoveInventoryItem(w http.ResponseWriter, r *http.Request) {
	pharmacyID := chi.URLParam(r, "id")
	productID := chi.URLParam(r, "productId")

	cmd := commands.RemoveInventoryItemCommand{
		PharmacyID: pharmacyID,
		ProductID:  productID,
	}

	response, _ := mediator.Send[commands.RemoveInventoryItemCommand, responses.EmptyResponse](r.Context(), c.mediator, cmd)
	c.respondJSON(w, response)
}

// UpdatePharmacyHours godoc
// @Summary      Actualizar horarios
// @Description  Actualiza los horarios de la farmacia (requiere ser owner)
// @Tags         Pharmacies
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path     string                              true  "Pharmacy ID"
// @Param        body  body     requests.UpdatePharmacyHoursRequest  true  "Horarios"
// @Success      200  {object}  common.ApiResponse[responses.HoursListResponse]
// @Failure      400  {object}  common.ApiResponse[responses.HoursListResponse]
// @Router       /api/v1/pharmacies/{id}/hours [put]
func (c *PharmacyController) UpdatePharmacyHours(w http.ResponseWriter, r *http.Request) {
	pharmacyID := chi.URLParam(r, "id")

	var req requests.UpdatePharmacyHoursRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.respondJSON(w, common.BadRequestResponse[responses.HoursListResponse]("VAL_001", "Body inválido"))
		return
	}

	hoursEntries := make([]commands.HoursEntry, len(req.Hours))
	for i, h := range req.Hours {
		hoursEntries[i] = commands.HoursEntry{
			DayOfWeek: h.DayOfWeek,
			OpenTime:  h.OpenTime,
			CloseTime: h.CloseTime,
			IsClosed:  h.IsClosed,
		}
	}

	cmd := commands.UpdatePharmacyHoursCommand{
		PharmacyID: pharmacyID,
		Hours:      hoursEntries,
	}

	response, _ := mediator.Send[commands.UpdatePharmacyHoursCommand, responses.HoursListResponse](r.Context(), c.mediator, cmd)
	c.respondJSON(w, response)
}

// ========================================
// FARMACIAS - ENDPOINTS ADMIN
// ========================================

// VerifyPharmacy godoc
// @Summary      Verificar farmacia
// @Description  Marca una farmacia como verificada (requiere rol admin)
// @Tags         Admin
// @Produce      json
// @Security     BearerAuth
// @Param        id   path     string  true  "Pharmacy ID"
// @Success      200  {object}  common.ApiResponse[responses.PharmacyResponse]
// @Failure      404  {object}  common.ApiResponse[responses.PharmacyResponse]
// @Router       /api/v1/pharmacies/{id}/verify [put]
func (c *PharmacyController) VerifyPharmacy(w http.ResponseWriter, r *http.Request) {
	pharmacyID := chi.URLParam(r, "id")
	cmd := commands.VerifyPharmacyCommand{ID: pharmacyID}

	response, _ := mediator.Send[commands.VerifyPharmacyCommand, responses.PharmacyResponse](r.Context(), c.mediator, cmd)
	c.respondJSON(w, response)
}

// DeletePharmacy godoc
// @Summary      Eliminar farmacia
// @Description  Elimina una farmacia (soft delete, requiere rol admin)
// @Tags         Admin
// @Produce      json
// @Security     BearerAuth
// @Param        id   path     string  true  "Pharmacy ID"
// @Success      200  {object}  common.ApiResponse[responses.EmptyResponse]
// @Failure      404  {object}  common.ApiResponse[responses.EmptyResponse]
// @Router       /api/v1/pharmacies/{id} [delete]
func (c *PharmacyController) DeletePharmacy(w http.ResponseWriter, r *http.Request) {
	pharmacyID := chi.URLParam(r, "id")
	cmd := commands.DeletePharmacyCommand{ID: pharmacyID}

	response, _ := mediator.Send[commands.DeletePharmacyCommand, responses.EmptyResponse](r.Context(), c.mediator, cmd)
	c.respondJSON(w, response)
}

// ========================================
// DOCUMENTO DE AUTORIZACIÓN
// ========================================

// UploadAuthorizationDocument godoc
// @Summary      Subir documento de autorización sanitaria
// @Description  Sube la Resolución de Autorización Sanitaria de la farmacia (PDF, max 10MB, requiere ser owner)
// @Tags         Pharmacies
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        id        path     string  true  "Pharmacy ID"
// @Param        document  formData file    true  "Archivo PDF de autorización"
// @Success      200  {object}  common.ApiResponse[responses.AuthorizationDocumentResponse]
// @Failure      400  {object}  common.ApiResponse[responses.AuthorizationDocumentResponse]
// @Failure      404  {object}  common.ApiResponse[responses.AuthorizationDocumentResponse]
// @Router       /api/v1/pharmacies/{id}/authorization-document [put]
func (c *PharmacyController) UploadAuthorizationDocument(w http.ResponseWriter, r *http.Request) {
	pharmacyID := chi.URLParam(r, "id")

	if err := r.ParseMultipartForm(12 << 20); err != nil { // 12MB max
		c.respondJSON(w, common.BadRequestResponse[responses.AuthorizationDocumentResponse]("VAL_001", "Error procesando formulario multipart"))
		return
	}

	file, header, err := r.FormFile("document")
	if err != nil {
		c.respondJSON(w, common.BadRequestResponse[responses.AuthorizationDocumentResponse]("VAL_006", "El documento es requerido"))
		return
	}
	defer file.Close()

	cmd := commands.UploadAuthorizationDocumentCommand{
		PharmacyID:  pharmacyID,
		Reader:      file,
		Filename:    header.Filename,
		ContentType: header.Header.Get("Content-Type"),
		Size:        header.Size,
	}

	response, _ := mediator.Send[commands.UploadAuthorizationDocumentCommand, responses.AuthorizationDocumentResponse](r.Context(), c.mediator, cmd)
	c.respondJSON(w, response)
}

// ========================================
// CADENAS DE FARMACIAS
// ========================================

// ListPharmaciesByChain godoc
// @Summary      Farmacias por cadena
// @Description  Retorna lista paginada de farmacias de una cadena específica
// @Tags         Pharmacies
// @Produce      json
// @Param        chainId  path     string  true   "Chain ID"
// @Param        page     query    int     false  "Página"           default(1)
// @Param        limit    query    int     false  "Límite por página" default(20)
// @Success      200  {object}  common.ApiResponse[responses.PharmacyListResponse]
// @Failure      400  {object}  common.ApiResponse[responses.PharmacyListResponse]
// @Router       /api/v1/pharmacies/chains/{chainId} [get]
func (c *PharmacyController) ListPharmaciesByChain(w http.ResponseWriter, r *http.Request) {
	chainID := chi.URLParam(r, "chainId")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	query := queries.ListPharmaciesByChainQuery{
		ChainID: chainID,
		Page:    page,
		Limit:   limit,
	}

	response, _ := mediator.Send[queries.ListPharmaciesByChainQuery, responses.PharmacyListResponse](r.Context(), c.mediator, query)
	c.respondJSON(w, response)
}
