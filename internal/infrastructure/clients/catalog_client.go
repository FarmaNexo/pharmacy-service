// Package clients — HTTP clients hacia otros microservicios.
//
// catalog_client.go: cliente para resolver una clave natural DIGEMID
// (source_product_code, concentration) → product_id local del catalog-service.
// Lo usa el SQS consumer de INVENTORY_DISCOVERED.
package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"go.uber.org/zap"
)

// CatalogClient define la interfaz mínima que el dominio espera del catálogo.
// Vive en infrastructure/clients (no en domain/services) porque es un detalle
// de transporte sin reglas de negocio: el handler de inventory consume esto
// directamente sin abstracción.
type CatalogClient interface {
	FindProductIDBySource(ctx context.Context, sourceProductCode int, concentration string) (string, error)
}

// CatalogClientImpl llama HTTP a catalog-service via API Gateway interno
// (CATALOG_SERVICE_URL apunta al ALB interno).
type CatalogClientImpl struct {
	baseURL    string
	httpClient *http.Client
	logger     *zap.Logger
}

func NewCatalogClient(baseURL string, logger *zap.Logger) *CatalogClientImpl {
	return &CatalogClientImpl{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		logger: logger,
	}
}

// catalogProductByCodeResponse — modelamos solo el campo que necesitamos.
type catalogProductByCodeResponse struct {
	Meta struct {
		Resultado bool `json:"resultado"`
	} `json:"meta"`
	Datos *struct {
		ID string `json:"id"`
	} `json:"datos"`
}

// ErrProductNotFound se retorna cuando catalog responde 404 o sin id.
// El consumer lo trata como "fail soft" para que SQS reintente (el producto
// quizás aún no fue UPSERTeado por catalog).
var ErrProductNotFound = fmt.Errorf("producto no encontrado en catalog")

// FindProductIDBySource llama GET /api/v1/products/by-source?code=N&concentration=...
// Retorna ErrProductNotFound si catalog no lo tiene aún (escenario esperado
// durante un export en curso). Otros errores (timeout, 5xx) se propagan.
func (c *CatalogClientImpl) FindProductIDBySource(ctx context.Context, sourceProductCode int, concentration string) (string, error) {
	endpoint := c.baseURL + "/api/v1/products/by-source"
	q := url.Values{}
	q.Set("code", strconv.Itoa(sourceProductCode))
	q.Set("concentration", concentration)
	endpoint = endpoint + "?" + q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("http call to catalog: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", ErrProductNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("catalog respondió status %d", resp.StatusCode)
	}

	var apiResp catalogProductByCodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	if !apiResp.Meta.Resultado || apiResp.Datos == nil || apiResp.Datos.ID == "" {
		return "", ErrProductNotFound
	}
	return apiResp.Datos.ID, nil
}

var _ CatalogClient = (*CatalogClientImpl)(nil)
