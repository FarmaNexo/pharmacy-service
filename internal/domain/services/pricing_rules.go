// Package services — reglas de negocio del dominio pharmacy.
//
// Este archivo concentra las constantes y predicados relacionados con la
// evaluación de precios de inventario. Vive en `domain/` (no en
// `infrastructure/`) porque es regla de NEGOCIO — qué se considera "caro"
// para el usuario final — y no debe acoplarse al motor de persistencia.
//
// La capa de persistencia (repository) provee los datos brutos
// (`DistrictAvgPrice`, `DistrictPharmacyCount`); el dominio aplica el
// criterio. Si negocio quiere ajustar el umbral, se cambia aquí en una
// línea y queda cubierto por todos los consumidores aguas arriba.
package services

// OverpriceThresholdMultiplier — multiplicador sobre el promedio del distrito
// a partir del cual se considera que un precio está "elevado" (HU-016).
// Valor 1.30 = +30% sobre el promedio del distrito.
const OverpriceThresholdMultiplier = 1.30

// MinDistrictPharmaciesForAverage — mínimo de farmacias en el distrito que
// deben ofrecer el producto para que el promedio sea estadísticamente
// representativo y se pueda emitir el flag "is_overpriced". Bajo este
// umbral, el promedio podría ser ruido (ej. 1 sola farmacia ⇒ promedio ==
// price ⇒ flag spurious). Valor 3 = al menos 2 farmacias adicionales.
const MinDistrictPharmaciesForAverage = 3

// IsOverpriced retorna true cuando el precio supera el umbral configurado
// sobre el promedio del distrito y hay suficientes farmacias en el
// distrito como para que el promedio sea representativo.
//
// Casos en los que SIEMPRE retorna false (graceful degradation):
//   - El producto no tiene promedio calculado en el distrito (avg nil o 0).
//   - El distrito tiene menos de MinDistrictPharmaciesForAverage farmacias
//     ofreciendo el producto.
//   - El precio es <= 0 (datos corruptos).
//
// El consumidor puede usar el `*float64` devuelto por el repository en
// `DistrictAvgPrice` para mostrar contexto al usuario aun cuando este
// predicado retorne false (ej. "promedio del distrito: S/ 5.00").
func IsOverpriced(price float64, districtAvg *float64, districtPharmacyCount int) bool {
	if price <= 0 {
		return false
	}
	if districtAvg == nil || *districtAvg <= 0 {
		return false
	}
	if districtPharmacyCount < MinDistrictPharmaciesForAverage {
		return false
	}
	return price > *districtAvg*OverpriceThresholdMultiplier
}

// OverpricePercentage retorna el porcentaje exacto por encima del promedio
// del distrito (ej. 0.45 == 45% más caro que el promedio). Útil para que
// el frontend muestre un valor concreto en el badge ("+45% sobre promedio").
//
// Devuelve nil si no hay promedio o si el dato no es relevante para mostrar
// (price <= avg). Devolver puntero permite distinguir "no aplica" de "0%".
func OverpricePercentage(price float64, districtAvg *float64) *float64 {
	if districtAvg == nil || *districtAvg <= 0 || price <= *districtAvg {
		return nil
	}
	pct := (price - *districtAvg) / *districtAvg
	return &pct
}
