# INFRAESTRUCTURA - Pharmacy Service

## Resumen

Este documento describe la infraestructura especifica utilizada por Pharmacy Service (puerto 4004).

---

## SERVICIOS REQUERIDOS

### PostgreSQL + PostGIS
- **Host:** localhost:5432 (local) / RDS endpoint (cloud)
- **Database:** `pharmacy_db`
- **User:** admin (local) / `${DB_USER}` (cloud)
- **Password:** admin (local) / `${DB_PASSWORD}` (cloud)
- **Schema:** `pharmacy`
- **Extensiones:** `postgis`, `gen_random_uuid()`
- **SSL:** disable (local) / require (produccion)

### Redis
- **Host:** localhost:6379 (local) / ElastiCache (cloud)
- **Password:** farmanexo2026 (local) / `${REDIS_PASSWORD}` (cloud)
- **DB:** 0 (compartido)
- **Pool size:** 10 (local) / 20 (produccion)
- **Max retries:** 3
- **Uso en este servicio:**
  - Cache de busquedas por geolocalizacion (15 min TTL)
  - Cache de detalles de farmacia
  - Cache de inventario por farmacia
  - Cache de farmacias por cadena

### LocalStack (Local) / AWS (Cloud)
- **Endpoint:** http://localhost:4566 (local)
- **Region:** us-east-1
- **Credenciales (local):** test/test (fake)

---

## RECURSOS AWS UTILIZADOS

### S3 Buckets

**Bucket:** `farmanexo-pharmacies`
**Uso:** Documentos de autorizacion y logos de farmacias

**Estructura:**
```
farmanexo-pharmacies/
  authorization-documents/
    {pharmacy_id}/
      autorizacion.pdf
  logos/
    {pharmacy_id}/
      logo.jpg
```

**Operaciones:**
- **Upload de documento:** Multipart, max 12 MB, solo PDF
- **Upload de logo:** Via actualizacion de farmacia
- **Delete:** Elimina archivo al reemplazar

**URLs generadas:**
- **Local:** `http://localhost:4566/farmanexo-pharmacies/{key}`
- **Cloud:** `https://farmanexo-pharmacies.s3.{region}.amazonaws.com/{key}`

### SQS Queues

**Cola que PUBLICA:**

**Cola:** `farmanexo-pharmacy-events`
- **URL (local):** `http://sqs.us-east-1.localhost.localstack.cloud:4566/000000000000/farmanexo-pharmacy-events`
- **URL (cloud):** `${SQS_PHARMACY_EVENTS_QUEUE_URL}`

**Eventos que genera:**

1. **PHARMACY_REGISTERED** - Cuando se registra una nueva farmacia
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

2. **PHARMACY_VERIFIED** - Cuando un admin verifica la farmacia
```json
{
  "event_type": "PHARMACY_VERIFIED",
  "pharmacy_id": "uuid",
  "timestamp": "2026-02-22T12:00:00Z",
  "metadata": {
    "source": "pharmacy-service",
    "version": "1.0"
  }
}
```

3. **INVENTORY_UPDATED** - Cuando se modifica el inventario
```json
{
  "event_type": "INVENTORY_UPDATED",
  "pharmacy_id": "uuid",
  "product_id": "uuid",
  "timestamp": "2026-02-22T12:00:00Z",
  "metadata": {
    "source": "pharmacy-service",
    "version": "1.0"
  }
}
```

4. **AUTHORIZATION_UPLOADED** - Cuando se sube documento de autorizacion
```json
{
  "event_type": "AUTHORIZATION_UPLOADED",
  "pharmacy_id": "uuid",
  "timestamp": "2026-02-22T12:00:00Z",
  "metadata": {
    "source": "pharmacy-service",
    "version": "1.0"
  }
}
```

**Cola que CONSUME (configurada, consumer no implementado):**

**Cola:** `farmanexo-catalog-events`
- **URL (local):** `http://sqs.us-east-1.localhost.localstack.cloud:4566/000000000000/farmanexo-catalog-events`
- **Evento planificado:** `PRODUCT_CREATED` -> Preparar entrada de inventario

**Patron de publicacion:** Fire-and-forget en goroutines.

---

## ESQUEMA DE BASE DE DATOS

### Extension PostGIS

```sql
CREATE EXTENSION IF NOT EXISTS postgis;
```

### Tabla: `pharmacy.pharmacies`

```sql
CREATE TABLE pharmacy.pharmacies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_user_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    phone VARCHAR(20),
    email VARCHAR(255),
    website VARCHAR(500),
    logo_url VARCHAR(500),
    authorization_document_url VARCHAR(500),

    -- Direccion
    street VARCHAR(255) NOT NULL,
    city VARCHAR(100) NOT NULL,
    state VARCHAR(100),
    postal_code VARCHAR(20),
    country VARCHAR(100) NOT NULL DEFAULT 'Peru',

    -- Geolocalizacion (PostGIS)
    location GEOGRAPHY(POINT, 4326),

    -- Cadena
    chain_id VARCHAR(100),
    chain_name VARCHAR(255),

    -- Estado
    is_verified BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    is_24h BOOLEAN DEFAULT false,

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_pharmacies_location ON pharmacy.pharmacies USING GIST(location);
CREATE INDEX idx_pharmacies_owner_user_id ON pharmacy.pharmacies(owner_user_id);
CREATE UNIQUE INDEX idx_pharmacies_slug ON pharmacy.pharmacies(slug);
CREATE INDEX idx_pharmacies_is_active ON pharmacy.pharmacies(is_active);
CREATE INDEX idx_pharmacies_is_verified ON pharmacy.pharmacies(is_verified);
CREATE INDEX idx_pharmacies_authorization_doc ON pharmacy.pharmacies(id)
    WHERE authorization_document_url IS NOT NULL;
CREATE INDEX idx_pharmacies_chain_id ON pharmacy.pharmacies(chain_id)
    WHERE chain_id IS NOT NULL;
```

**Proposito:** Farmacias registradas. Incluye geolocalizacion PostGIS para busquedas por cercania.

**Geolocalizacion:**
- Tipo: `GEOGRAPHY(POINT, 4326)` - Modelo esferoidal (mas preciso que GEOMETRY)
- SRID: 4326 (WGS84 - coordenadas GPS estandar)
- Indice: GIST para busquedas espaciales eficientes
- Almacenamiento: `ST_MakePoint(longitude, latitude)::geography`

**Soft delete:** Campo `deleted_at` (NULL = activo)

### Tabla: `pharmacy.pharmacy_hours`

```sql
CREATE TABLE pharmacy.pharmacy_hours (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pharmacy_id UUID NOT NULL REFERENCES pharmacy.pharmacies(id) ON DELETE CASCADE,
    day_of_week INTEGER NOT NULL,  -- 0=Domingo, 1=Lunes, ..., 6=Sabado
    open_time TIME NOT NULL,
    close_time TIME NOT NULL,
    is_closed BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(pharmacy_id, day_of_week)
);

CREATE INDEX idx_pharmacy_hours_pharmacy_id ON pharmacy.pharmacy_hours(pharmacy_id);
```

**Proposito:** Horarios de atencion por dia de la semana. Un registro por dia, constraint UNIQUE previene duplicados.

### Tabla: `pharmacy.pharmacy_inventory`

```sql
CREATE TABLE pharmacy.pharmacy_inventory (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pharmacy_id UUID NOT NULL REFERENCES pharmacy.pharmacies(id) ON DELETE CASCADE,
    product_id UUID NOT NULL,
    stock INTEGER DEFAULT 0,
    price DECIMAL(10, 2) NOT NULL,
    is_available BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(pharmacy_id, product_id)
);

CREATE INDEX idx_pharmacy_inventory_pharmacy_id ON pharmacy.pharmacy_inventory(pharmacy_id);
CREATE INDEX idx_pharmacy_inventory_product_id ON pharmacy.pharmacy_inventory(product_id);
CREATE INDEX idx_pharmacy_inventory_is_available ON pharmacy.pharmacy_inventory(is_available);
```

**Proposito:** Inventario de productos por farmacia. `product_id` referencia a productos del Catalog Service (sin FK fisica, cross-service).

---

## POSTGIS - GEOLOCALIZACION

### Funciones SQL utilizadas

**Almacenar coordenadas:**
```sql
ST_MakePoint(longitude, latitude)::geography
-- Nota: PostGIS usa (x, y) = (longitude, latitude)
```

**Busqueda por radio:**
```sql
ST_DWithin(location, ST_MakePoint($1, $2)::geography, $3)
-- $1 = longitude, $2 = latitude, $3 = radio en metros
```

**Calcular distancia:**
```sql
ST_Distance(location, ST_MakePoint($1, $2)::geography) / 1000.0 AS distance_km
-- Retorna distancia en kilometros (ST_Distance devuelve metros)
```

**Extraer coordenadas:**
```sql
ST_Y(location::geometry) AS latitude
ST_X(location::geometry) AS longitude
```

### Query ejemplo (FindNearby)
```sql
SELECT id, name, slug, phone, street, city,
       ST_Y(location::geometry) AS latitude,
       ST_X(location::geometry) AS longitude,
       ST_Distance(location, ST_MakePoint($1, $2)::geography) / 1000.0 AS distance_km
FROM pharmacy.pharmacies
WHERE ST_DWithin(location, ST_MakePoint($1, $2)::geography, $3)
  AND is_active = true
  AND is_verified = true
  AND deleted_at IS NULL
ORDER BY distance_km ASC
LIMIT $4;
```

### Validaciones
- Latitud: -90 a 90
- Longitud: -180 a 180
- Radio: 0.1 a 50 km (default: 20 km si valor invalido)

---

## CONFIGURACION POR AMBIENTE

### Local (config.local.yaml)
```yaml
environment: local
server:
  port: 4004
  read_timeout: 15s
  write_timeout: 15s
database:
  host: localhost
  user: admin
  password: admin
  db_name: pharmacy_db
  schema: pharmacy
  sslmode: disable
  max_open_conns: 25
jwt:
  secret: "dev-super-secret-key-change-in-production-min-32-chars"
redis:
  host: localhost
  password: farmanexo2026
  pool_size: 10
aws:
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

### Development (config.development.yaml)
```yaml
environment: development
database:
  host: ${DB_HOST}
  user: ${DB_USER}
  password: ${DB_PASSWORD}
  sslmode: require
jwt:
  secret: ${JWT_SECRET}
aws:
  endpoint: ""  # AWS real
log:
  level: info
  encoding: json
```

### Production (config.production.yaml)
```yaml
environment: production
database:
  max_open_conns: 50
  max_idle_conns: 15
  sslmode: require
redis:
  pool_size: 20
log:
  level: error
  encoding: json
```

---

## SECRETS Y CREDENCIALES

### Secrets Manager (Produccion)
- `farmanexo/auth/jwt-secret` - JWT validation key (mismo que Auth Service)
- `farmanexo/database/password` - Database password

### Variables de Entorno (Ambientes desplegados)

| Variable | Descripcion |
|---|---|
| `ENV` | local, development, qa, uat, production |
| `DB_HOST` | Host PostgreSQL |
| `DB_USER` | Usuario PostgreSQL |
| `DB_PASSWORD` | Password PostgreSQL |
| `JWT_SECRET` | JWT secret (mismo que Auth Service) |
| `REDIS_HOST` | Host Redis |
| `REDIS_PASSWORD` | Password Redis |
| `AWS_REGION` | Region AWS |
| `SQS_PHARMACY_EVENTS_QUEUE_URL` | URL cola para publicar |
| `SQS_CATALOG_EVENTS_QUEUE_URL` | URL cola para consumir |

---

## CACHE REDIS - PATRONES

### Keys Utilizadas

| Pattern | TTL | Uso |
|---------|-----|-----|
| `cache:pharmacy:nearby:{lat:.4f}:{lng:.4f}:{radius:.1f}:{limit}` | 15 min | Busquedas por geolocalizacion |
| `cache:pharmacy:{id}:details` | Invalidado en update | Detalle de farmacia |
| `cache:pharmacy:{id}:inventory` | Invalidado en cambios | Inventario de farmacia |
| `cache:pharmacy:chain:{chain_id}:page:{p}:limit:{l}` | 15 min | Farmacias por cadena (paginado) |

### Invalidacion de Cache
- **Actualizar farmacia:** Invalida `cache:pharmacy:{id}:details`
- **Cambiar inventario:** Invalida `cache:pharmacy:{id}:inventory`
- **Busquedas nearby:** Auto-expiran por TTL (15 min)
- **Cadenas:** Auto-expiran por TTL

### Implementacion
- Servicio: `RedisCacheService`
- Metodos: `Get()`, `Set()`, `Delete()`, `DeleteByPattern()`
- Invalidacion asincrona en operaciones de escritura

---

## EVENTOS SQS

### Flujo de Eventos

```
[Catalog Service] --PRODUCT_CREATED--> [farmanexo-catalog-events] --> [Pharmacy Service] (planificado)
[Pharmacy Service] --PHARMACY_REGISTERED--> [farmanexo-pharmacy-events] --> [Consumidores futuros]
[Pharmacy Service] --INVENTORY_UPDATED----> [farmanexo-pharmacy-events] --> [Price Service] (futuro)
[Pharmacy Service] --PHARMACY_VERIFIED----> [farmanexo-pharmacy-events] --> [Notificaciones] (futuro)
```

### Procesamiento
- Los eventos se publican de forma asincrona en goroutines
- Fire-and-forget: no bloquea el flujo principal
- `INVENTORY_UPDATED` es consumido por Price Service para tracking de precios (futuro)

---

## DESPLIEGUE

### Checklist Pre-Deploy
- [ ] Migraciones aplicadas (incluye PostGIS extension)
- [ ] Variables de entorno configuradas
- [ ] JWT_SECRET identico al de Auth Service
- [ ] S3 bucket `farmanexo-pharmacies` creado
- [ ] SQS queues creadas (`farmanexo-pharmacy-events`, `farmanexo-catalog-events`)
- [ ] Redis accesible
- [ ] PostgreSQL accesible con PostGIS y database `pharmacy_db`

### Comandos de Deploy
```bash
# Build
make build

# Migraciones (incluye CREATE EXTENSION postgis)
make migrate-up ENV=production

# Docker
make docker-build
make docker-run
```

---

## TESTING LOCAL

### 1. Levantar Infraestructura
```bash
cd FarmaNexo/Helpers
./start-local.sh --full
./init-localstack-resources.sh
```

### 2. Verificar Servicios
```bash
# PostgreSQL + PostGIS
docker exec -it farmanexo-postgres psql -U admin -d pharmacy_db -c "SELECT PostGIS_Version();"

# Redis
docker exec -it farmanexo-redis redis-cli -a farmanexo2026

# LocalStack S3
aws --endpoint-url=http://localhost:4566 s3 ls s3://farmanexo-pharmacies/

# LocalStack SQS
aws --endpoint-url=http://localhost:4566 sqs list-queues
```

### 3. Crear Base de Datos
```bash
docker exec -it farmanexo-postgres psql -U admin -c "CREATE DATABASE pharmacy_db;"
```

### 4. Ejecutar Migraciones
```bash
cd services/pharmacy-service
make migrate-up
```

### 5. Ejecutar Servicio
```bash
make dev
```

---

## MONITOREO

### Metricas Importantes
- Tasa de busquedas nearby (geo queries/min)
- Cache hit rate en busquedas geoespaciales
- Total de farmacias activas/verificadas
- Latencia de geo queries (PostGIS)
- Tamano de inventarios por farmacia

### Logs
- **Formato:** Console (local/dev), JSON (produccion)
- **Logger:** Zap structured logging
- **Campos contextuales:** pharmacy_id, owner_user_id, product_id, correlation_id, distance_km

### Alertas Recomendadas
- Rate de errores 5xx > 1%
- Latencia de geo queries p99 > 2s
- S3 upload failures
- Redis no disponible (afecta cache de nearby)
- PostGIS extension no disponible

---

## TROUBLESHOOTING

### Problema: "PostGIS extension not available"
**Sintoma:** Error en migraciones o queries geoespaciales
**Causa:** PostgreSQL no tiene PostGIS instalado
**Solucion:**
```bash
docker exec -it farmanexo-postgres psql -U admin -d pharmacy_db -c "CREATE EXTENSION IF NOT EXISTS postgis;"
```

### Problema: "Nearby search returns empty"
**Sintoma:** POST /pharmacies/nearby no retorna resultados
**Causa:** No hay farmacias verificadas y activas en el radio
**Solucion:** Verificar que existan farmacias con `is_active=true`, `is_verified=true` y `location` no NULL. Ampliar el radio de busqueda.

### Problema: "Authorization document upload failed"
**Sintoma:** Error al subir PDF
**Causa:** S3 bucket no existe o archivo excede 12 MB
**Solucion:**
```bash
aws --endpoint-url=http://localhost:4566 s3 mb s3://farmanexo-pharmacies
```

### Problema: "Inventory item already exists"
**Sintoma:** Error 409 al agregar producto al inventario
**Causa:** Constraint UNIQUE(pharmacy_id, product_id) violado
**Solucion:** Usar PUT para actualizar en lugar de POST para crear.

### Problema: "Invalid coordinates"
**Sintoma:** Error al registrar farmacia con coordenadas
**Causa:** Latitud/longitud fuera de rango
**Solucion:** Latitud: -90 a 90, Longitud: -180 a 180. Verificar que no esten invertidas.

---

## Referencias

- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [PostGIS Documentation](https://postgis.net/documentation/)
- [Redis Documentation](https://redis.io/documentation)
- [AWS S3 Documentation](https://docs.aws.amazon.com/s3/)
- [AWS SQS Documentation](https://docs.aws.amazon.com/sqs/)
- [LocalStack Documentation](https://docs.localstack.cloud/)
- [GORM Documentation](https://gorm.io/docs/)

---

Ultima actualizacion: 2026-02-22
