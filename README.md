# Pharmacy Service

Microservicio de gestion de farmacias para FarmaNexo. Maneja registro de farmacias, inventarios, horarios de atencion, busqueda por geolocalizacion (PostGIS) y cadenas de farmacias.

## Inicio Rapido

### Prerequisitos
- Go 1.25+
- PostgreSQL 16 con PostGIS
- Redis 7
- LocalStack (desarrollo local)
- Docker & Docker Compose

### Instalacion
```bash
# Clonar repositorio
git clone <url>
cd services/pharmacy-service

# Instalar dependencias
go mod download

# Configurar ambiente local
cp configs/config.development.yaml configs/config.local.yaml
# Editar configs/config.local.yaml con tus credenciales

# Crear base de datos
docker exec -it farmanexo-postgres psql -U admin -c "CREATE DATABASE pharmacy_db;"

# Ejecutar migraciones (incluye PostGIS extension)
make migrate-up

# Ejecutar servicio
make dev
```

Swagger UI disponible en: http://localhost:4004/swagger/index.html

## Endpoints

### Publicos

**GET /api/v1/pharmacies** - Listar farmacias (paginado)
```bash
curl "http://localhost:4004/api/v1/pharmacies?page=1&limit=20"
```

**GET /api/v1/pharmacies/{id}** - Detalle de farmacia
```bash
curl http://localhost:4004/api/v1/pharmacies/{id}
```

Respuesta (200):
```json
{
  "meta": {
    "mensajes": [{ "codigo": "PH_001", "mensaje": "Farmacia obtenida exitosamente", "tipo": "exito" }],
    "idTransaccion": "...",
    "resultado": true,
    "timestamp": "20260222 103000"
  },
  "datos": {
    "id": "uuid",
    "name": "Farmacia San Pablo",
    "slug": "farmacia-san-pablo",
    "description": "Farmacia de confianza",
    "phone": "+51999888777",
    "email": "contacto@sanpablo.pe",
    "street": "Av. Arequipa 1234",
    "city": "Lima",
    "state": "Lima",
    "country": "Peru",
    "latitude": -12.0464,
    "longitude": -77.0428,
    "is_verified": true,
    "is_24h": false,
    "chain_id": "INKAFARMA",
    "chain_name": "Inkafarma"
  }
}
```

**POST /api/v1/pharmacies/nearby** - Buscar farmacias cercanas (geolocalizacion)
```bash
curl -X POST http://localhost:4004/api/v1/pharmacies/nearby \
  -H "Content-Type: application/json" \
  -d '{
    "latitude": -12.0464,
    "longitude": -77.0428,
    "radius_km": 5.0,
    "limit": 10
  }'
```

Respuesta incluye `distance_km` calculada con PostGIS.

**GET /api/v1/pharmacies/{id}/inventory** - Inventario de farmacia
```bash
curl http://localhost:4004/api/v1/pharmacies/{id}/inventory
```

**GET /api/v1/pharmacies/{id}/hours** - Horarios de atencion
```bash
curl http://localhost:4004/api/v1/pharmacies/{id}/hours
```

**GET /api/v1/chains/{chainId}/pharmacies** - Farmacias por cadena
```bash
curl "http://localhost:4004/api/v1/chains/INKAFARMA/pharmacies?page=1&limit=20"
```

### Owner (requieren JWT + rol pharmacy_owner o admin)

**POST /api/v1/pharmacies** - Registrar farmacia
```bash
curl -X POST http://localhost:4004/api/v1/pharmacies \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Mi Farmacia",
    "description": "Farmacia de barrio",
    "phone": "+51999888777",
    "email": "info@mifarmacia.pe",
    "street": "Jr. Union 456",
    "city": "Lima",
    "state": "Lima",
    "postal_code": "15001",
    "country": "Peru",
    "latitude": -12.0464,
    "longitude": -77.0428,
    "is_24h": false
  }'
```

**PUT /api/v1/pharmacies/{id}** - Actualizar farmacia
```bash
curl -X PUT http://localhost:4004/api/v1/pharmacies/{id} \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{"name": "Mi Farmacia Actualizada", "phone": "+51999777666"}'
```

**POST /api/v1/pharmacies/{id}/inventory** - Agregar producto al inventario
```bash
curl -X POST http://localhost:4004/api/v1/pharmacies/{id}/inventory \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{
    "product_id": "uuid",
    "stock": 100,
    "price": 25.50,
    "is_available": true
  }'
```

**PUT /api/v1/pharmacies/{id}/inventory/{productId}** - Actualizar inventario
```bash
curl -X PUT http://localhost:4004/api/v1/pharmacies/{id}/inventory/{productId} \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{"stock": 50, "price": 22.00}'
```

**DELETE /api/v1/pharmacies/{id}/inventory/{productId}** - Eliminar del inventario
```bash
curl -X DELETE http://localhost:4004/api/v1/pharmacies/{id}/inventory/{productId} \
  -H "Authorization: Bearer {token}"
```

**PUT /api/v1/pharmacies/{id}/hours** - Actualizar horarios
```bash
curl -X PUT http://localhost:4004/api/v1/pharmacies/{id}/hours \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{
    "hours": [
      { "day_of_week": 1, "open_time": "08:00", "close_time": "21:00", "is_closed": false },
      { "day_of_week": 0, "open_time": "09:00", "close_time": "13:00", "is_closed": false }
    ]
  }'
```

**PUT /api/v1/pharmacies/{id}/authorization-document** - Subir documento de autorizacion
```bash
curl -X PUT http://localhost:4004/api/v1/pharmacies/{id}/authorization-document \
  -H "Authorization: Bearer {token}" \
  -F "document=@autorizacion.pdf"
```

- Tamano maximo: 12 MB
- Formato: PDF

### Admin (requieren JWT + rol admin)

**PUT /api/v1/pharmacies/{id}/verify** - Verificar farmacia
```bash
curl -X PUT http://localhost:4004/api/v1/pharmacies/{id}/verify \
  -H "Authorization: Bearer {token}"
```

**DELETE /api/v1/pharmacies/{id}** - Eliminar farmacia (soft delete)
```bash
curl -X DELETE http://localhost:4004/api/v1/pharmacies/{id} \
  -H "Authorization: Bearer {token}"
```

### Health Check

**GET /health** - Estado del servicio
```bash
curl http://localhost:4004/health
```

## Arquitectura

- **Puerto:** 4004
- **Base de datos:** `pharmacy_db`
- **Schema:** `pharmacy`
- **Patron:** Clean Architecture + CQRS + MediatR

### Capas

```
internal/
  domain/           Entidades (Pharmacy, PharmacyHours, PharmacyInventory), interfaces
  application/      Commands (8), Queries (11), Handlers, Validators, Pre/Post processors
  infrastructure/   PostgreSQL + PostGIS (GORM), Redis, S3, SQS, JWT
  presentation/     Controllers, Middlewares, Routes, DTOs
  shared/           ApiResponse[T], Constants, Errors
pkg/
  mediator/         Mediator CQRS generico con pipeline
  config/           Carga de configuracion por ambiente (Viper)
```

### Flujo de un request

```
HTTP Request
  -> Chi Router
    -> [Middlewares: RequestID, RealIP, Logger, Recoverer, CORS, CorrelationID]
    -> [AuthMiddleware + RequireOwner/RequireAdmin (si es protegido)]
    -> Controller
      -> Mediator.Send(Command/Query)
        -> Validator
        -> PreProcessor (SanitizeInput)
        -> Handler (+ PostGIS para geo queries)
        -> PostProcessor (LogAudit)
      <- ApiResponse[T]
    <- JSON Response
```

## Configuracion

### Variables de Entorno

```yaml
# configs/config.local.yaml
environment: local

server:
  host: 0.0.0.0
  port: 4004
  read_timeout: 15s
  write_timeout: 15s
  idle_timeout: 60s

database:
  host: localhost
  port: 5432
  user: admin
  password: admin
  db_name: pharmacy_db
  schema: pharmacy
  sslmode: disable
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 5m

jwt:
  secret: "dev-super-secret-key-change-in-production-min-32-chars"
  access_token_duration: 15m
  issuer: "farmanexo-pharmacy-service"

redis:
  host: localhost
  port: 6379
  password: farmanexo2026
  db: 0
  max_retries: 3
  pool_size: 10

aws:
  region: us-east-1
  endpoint: "http://localhost:4566"

sqs:
  pharmacy_events_queue_url: "http://sqs.us-east-1.localhost.localstack.cloud:4566/000000000000/farmanexo-pharmacy-events"
  catalog_events_queue_url: "http://sqs.us-east-1.localhost.localstack.cloud:4566/000000000000/farmanexo-catalog-events"

s3:
  pharmacies_bucket: "farmanexo-pharmacies"

log:
  level: debug
  encoding: console
```

### Variables de entorno requeridas (ambientes desplegados)

| Variable | Descripcion |
|---|---|
| `DB_HOST` | Host de PostgreSQL |
| `DB_USER` | Usuario de PostgreSQL |
| `DB_PASSWORD` | Password de PostgreSQL |
| `JWT_SECRET` | Secret para validar JWT (mismo que Auth Service) |
| `REDIS_HOST` | Host de Redis |
| `REDIS_PASSWORD` | Password de Redis |
| `AWS_REGION` | Region de AWS |
| `SQS_PHARMACY_EVENTS_QUEUE_URL` | URL cola para publicar eventos |
| `SQS_CATALOG_EVENTS_QUEUE_URL` | URL cola para consumir eventos de Catalog |

## Infraestructura

### PostgreSQL + PostGIS
- **Database:** `pharmacy_db`
- **Schema:** `pharmacy`
- **Extension:** PostGIS (busquedas geoespaciales)
- **Tablas:** `pharmacies`, `pharmacy_hours`, `pharmacy_inventory`
- **Soft delete:** Farmacias usan campo `deleted_at`

### Redis
- **Uso:** Cache de busquedas y detalles
- **Keys:**
  - `cache:pharmacy:nearby:{lat}:{lng}:{radius}:{limit}` - TTL: 15 min
  - `cache:pharmacy:{id}:details` - Invalidado en update
  - `cache:pharmacy:{id}:inventory` - Invalidado en cambios de inventario
  - `cache:pharmacy:chain:{chain_id}:page:{p}:limit:{l}` - Cache paginado por cadena

### S3
- **Bucket:** `farmanexo-pharmacies`
- **Paths:**
  - `authorization-documents/{pharmacy_id}/{filename}` - Documentos de autorizacion (PDF)
  - `logos/{pharmacy_id}/{filename}` - Logos de farmacia

### SQS
- **Publica en:** `farmanexo-pharmacy-events`
- **Eventos:** `PHARMACY_REGISTERED`, `PHARMACY_VERIFIED`, `INVENTORY_UPDATED`, `AUTHORIZATION_UPLOADED`
- **Consume:** `farmanexo-catalog-events` (configurado pero consumer no implementado aun)

## Geolocalizacion (PostGIS)

### Funciones usadas
- `ST_MakePoint(lng, lat)::geography` - Crear punto geografico
- `ST_DWithin(location, point, radius_meters)` - Busqueda dentro de radio
- `ST_Distance(location, point) / 1000.0` - Distancia en kilometros
- `ST_Y(location::geometry)` / `ST_X(location::geometry)` - Extraer lat/lng

### Validaciones
- Latitud: -90 a 90
- Longitud: -180 a 180
- Radio: 0.1 a 50 km (default: 20 km)

### Indice
- GIST index en columna `location` para busquedas eficientes

## Eventos

### Publica

| Evento | Cola | Trigger |
|---|---|---|
| `PHARMACY_REGISTERED` | `farmanexo-pharmacy-events` | Registro de nueva farmacia |
| `PHARMACY_VERIFIED` | `farmanexo-pharmacy-events` | Admin verifica farmacia |
| `INVENTORY_UPDATED` | `farmanexo-pharmacy-events` | Cambio en inventario |
| `AUTHORIZATION_UPLOADED` | `farmanexo-pharmacy-events` | Upload de documento |

### Consume (planificado)

| Evento | Cola | Accion |
|---|---|---|
| `PRODUCT_CREATED` | `farmanexo-catalog-events` | Preparar entrada de inventario (no implementado) |

Formato:
```json
{
  "event_type": "PHARMACY_REGISTERED",
  "pharmacy_id": "uuid",
  "user_id": "uuid",
  "timestamp": "2026-02-22T12:00:00Z",
  "metadata": {
    "source": "pharmacy-service",
    "version": "1.0"
  }
}
```

## Testing
```bash
# Unit tests
make test

# Tests con coverage
make test-coverage

# Generar mocks
make gen-mocks
```

## Comandos Utiles
```bash
# Desarrollo
make dev              # Ejecutar en modo desarrollo
make build            # Compilar binario a bin/pharmacy-service
make swagger          # Generar documentacion Swagger

# Base de datos
make migrate-up       # Aplicar migraciones pendientes
make migrate-down     # Revertir ultima migracion
make migrate-create NAME=nombre  # Crear nueva migracion

# Calidad
make lint             # Ejecutar golangci-lint
make format           # Formatear codigo con goimports

# Docker
make docker-build     # Construir imagen Docker
make docker-run       # Ejecutar container
```

## Dependencias

### Principales
- `github.com/go-chi/chi/v5` - HTTP router
- `gorm.io/gorm` - ORM
- `github.com/redis/go-redis/v9` - Redis client
- `github.com/aws/aws-sdk-go-v2` - AWS SDK (S3, SQS)
- `github.com/golang-jwt/jwt/v5` - JWT validation
- `go.uber.org/zap` - Structured logging
- `github.com/spf13/viper` - Configuracion
- `github.com/swaggo/swag` - Swagger
- `golang.org/x/text` - Generacion de slugs

### Completas
Ver `go.mod`

## Documentacion Adicional

- [CLAUDE.md](./CLAUDE.md) - Contexto para Claude AI
- [INFRASTRUCTURE.md](./INFRASTRUCTURE.md) - Detalle de infraestructura
- [Swagger UI](http://localhost:4004/swagger/index.html) - API docs interactiva

## Estructura de Directorios

```
pharmacy-service/
  cmd/server/main.go                    Punto de entrada con DI
  configs/                              YAML por ambiente (5 archivos)
  migrations/                           SQL (golang-migrate, schema: pharmacy + PostGIS)
  internal/
    application/
      commands/                         CreatePharmacy, UpdatePharmacy, VerifyPharmacy,
                                        AddInventoryItem, UpdateInventoryItem,
                                        RemoveInventoryItem, UpdateHours,
                                        UploadAuthorizationDocument
      queries/                          ListPharmacies, GetPharmacy, SearchNearby,
                                        GetInventory, GetHours, ListByChain, etc.
      handlers/                         Handler por cada command/query (19 total)
      validators/                       CreatePharmacy, AddInventoryItem
      preprocessors/                    SanitizeInput
      postprocessors/                   LogAudit
    domain/
      entities/                         Pharmacy, PharmacyHours, PharmacyInventory
      events/                           Eventos de farmacia (4 tipos)
      repositories/                     PharmacyRepository, InventoryRepository, HoursRepository
      services/                         CacheService, EventPublisher, FileStorage
    infrastructure/
      persistence/postgres/             Repositorios GORM + PostGIS queries
      cache/                            Redis cache service
      messaging/                        SQS event publisher
      storage/                          S3 file storage
      security/                         JWT service (validacion)
    presentation/
      dto/requests/                     CreatePharmacyRequest, NearbyRequest, etc.
      dto/responses/                    PharmacyResponse, InventoryResponse, etc.
      http/controllers/                 PharmacyController
      http/middlewares/                 AuthMiddleware, RequireOwner, RequireAdmin, CorrelationID
      http/routes/                      Configuracion de rutas Chi
    shared/
      common/                           ApiResponse[T], response factories
      constants/                        Codigos HTTP, message codes
      errors/                           Domain errors
  pkg/
    config/                             Carga de configuracion (Viper)
    mediator/                           Mediator CQRS generico
  docs/                                 Swagger generado
```
