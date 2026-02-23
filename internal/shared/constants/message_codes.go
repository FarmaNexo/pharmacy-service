// internal/shared/constants/message_codes.go
package constants

type MessageCode string

const (
	// Success codes
	CodeSuccess        MessageCode = "SUCCESS_001"
	CodeCreatedSuccess MessageCode = "SUCCESS_002"
	CodeUpdatedSuccess MessageCode = "SUCCESS_003"
	CodeDeletedSuccess MessageCode = "SUCCESS_004"

	// Pharmacy domain codes
	CodePharmacyRetrieved   MessageCode = "PHA_001"
	CodePharmacyCreated     MessageCode = "PHA_002"
	CodePharmacyUpdated     MessageCode = "PHA_003"
	CodePharmacyDeleted     MessageCode = "PHA_004"
	CodePharmaciesListed    MessageCode = "PHA_005"
	CodePharmacyVerified    MessageCode = "PHA_006"
	CodeNearbyFound         MessageCode = "PHA_007"
	CodeInventoryListed     MessageCode = "PHA_008"
	CodeInventoryAdded      MessageCode = "PHA_009"
	CodeInventoryUpdated    MessageCode = "PHA_010"
	CodeInventoryRemoved    MessageCode = "PHA_011"
	CodeHoursRetrieved          MessageCode = "PHA_012"
	CodeHoursUpdated            MessageCode = "PHA_013"
	CodeAuthorizationUploaded   MessageCode = "PHA_014"
	CodeChainPharmaciesListed   MessageCode = "PHA_015"

	// Validation errors
	CodeValidationError    MessageCode = "VAL_001"
	CodeInvalidCoordinates MessageCode = "VAL_002"
	CodeInvalidRadius      MessageCode = "VAL_003"
	CodeInvalidFileType    MessageCode = "VAL_004"
	CodeFileTooLarge       MessageCode = "VAL_005"
	CodeRequiredField      MessageCode = "VAL_006"
	CodeInvalidDayOfWeek   MessageCode = "VAL_007"
	CodeInvalidTimeRange   MessageCode = "VAL_008"
	CodeInvalidPrice       MessageCode = "VAL_009"
	CodeInvalidStock       MessageCode = "VAL_010"

	// Authentication errors
	CodeUnauthorized MessageCode = "AUTH_ERR_001"
	CodeInvalidToken MessageCode = "AUTH_ERR_002"
	CodeTokenExpired MessageCode = "AUTH_ERR_003"
	CodeForbidden    MessageCode = "AUTH_ERR_005"
	CodeNotOwner     MessageCode = "AUTH_ERR_006"

	// Business errors
	CodePharmacyNotFound       MessageCode = "BUS_001"
	CodeInventoryNotFound      MessageCode = "BUS_002"
	CodeResourceNotFound       MessageCode = "BUS_003"
	CodeSlugAlreadyExists      MessageCode = "BUS_004"
	CodeInventoryAlreadyExists MessageCode = "BUS_005"
	CodePharmacyNotVerified    MessageCode = "BUS_006"

	// Rate limiting
	CodeRateLimitExceeded MessageCode = "RATE_001"

	// System errors
	CodeInternalError      MessageCode = "SYS_001"
	CodeDatabaseError      MessageCode = "SYS_002"
	CodeServiceUnavailable MessageCode = "SYS_003"
	CodeStorageError       MessageCode = "SYS_004"
	CodeCacheError         MessageCode = "SYS_005"
)

var MessageDescription = map[MessageCode]string{
	CodeSuccess:        "Operación exitosa",
	CodeCreatedSuccess: "Recurso creado exitosamente",
	CodeUpdatedSuccess: "Recurso actualizado exitosamente",
	CodeDeletedSuccess: "Recurso eliminado exitosamente",

	CodePharmacyRetrieved: "Farmacia obtenida exitosamente",
	CodePharmacyCreated:   "Farmacia registrada exitosamente",
	CodePharmacyUpdated:   "Farmacia actualizada exitosamente",
	CodePharmacyDeleted:   "Farmacia eliminada exitosamente",
	CodePharmaciesListed:  "Farmacias listadas exitosamente",
	CodePharmacyVerified:  "Farmacia verificada exitosamente",
	CodeNearbyFound:       "Farmacias cercanas encontradas",
	CodeInventoryListed:   "Inventario listado exitosamente",
	CodeInventoryAdded:    "Producto agregado al inventario",
	CodeInventoryUpdated:  "Inventario actualizado exitosamente",
	CodeInventoryRemoved:  "Producto removido del inventario",
	CodeHoursRetrieved:        "Horarios obtenidos exitosamente",
	CodeHoursUpdated:          "Horarios actualizados exitosamente",
	CodeAuthorizationUploaded: "Documento de autorización subido exitosamente",
	CodeChainPharmaciesListed: "Farmacias de la cadena listadas exitosamente",

	CodeValidationError:    "Error de validación",
	CodeInvalidCoordinates: "Coordenadas inválidas",
	CodeInvalidRadius:      "Radio de búsqueda inválido",
	CodeInvalidFileType:    "Tipo de archivo no permitido",
	CodeFileTooLarge:       "El archivo excede el tamaño máximo",
	CodeRequiredField:      "Campo requerido",
	CodeInvalidDayOfWeek:   "Día de la semana inválido (0-6)",
	CodeInvalidTimeRange:   "Rango de horario inválido",
	CodeInvalidPrice:       "Precio debe ser mayor a 0",
	CodeInvalidStock:       "Stock debe ser mayor o igual a 0",

	CodeUnauthorized: "No autorizado",
	CodeInvalidToken: "Token inválido",
	CodeTokenExpired: "Token expirado",
	CodeForbidden:    "No tiene permisos para esta acción",
	CodeNotOwner:     "No es el propietario de esta farmacia",

	CodePharmacyNotFound:       "Farmacia no encontrada",
	CodeInventoryNotFound:      "Producto no encontrado en inventario",
	CodeResourceNotFound:       "Recurso no encontrado",
	CodeSlugAlreadyExists:      "El slug ya existe",
	CodeInventoryAlreadyExists: "El producto ya existe en el inventario de esta farmacia",
	CodePharmacyNotVerified:    "La farmacia no está verificada",

	CodeRateLimitExceeded: "Demasiadas solicitudes",

	CodeInternalError:      "Error interno del servidor",
	CodeDatabaseError:      "Error de base de datos",
	CodeServiceUnavailable: "Servicio no disponible",
	CodeStorageError:       "Error en servicio de almacenamiento",
	CodeCacheError:         "Error en servicio de caché",
}

func GetDescription(code MessageCode) string {
	if desc, ok := MessageDescription[code]; ok {
		return desc
	}
	return "Descripción no disponible"
}
