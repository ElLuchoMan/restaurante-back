package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"restaurante/database"
	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
)

type OfertaService struct {
	ormer orm.Ormer
}

func NewOfertaService(ormer orm.Ormer) *OfertaService {
	return &OfertaService{ormer: ormer}
}

func (s *OfertaService) ObtenerOfertasActivas(ctx context.Context, restauranteId int64, fecha *time.Time, hora *time.Time, productoId *int64) ([]*models.OfertaActivaResponse, error) {
	// Asegurar zona Bogotá incluso si no fue inicializada en tests
	loc := database.BogotaZone
	if loc == nil {
		if l, err := time.LoadLocation("America/Bogota"); err == nil {
			loc = l
		} else {
			loc = time.FixedZone("UTC-5", -5*60*60)
		}
	}
	now := time.Now().In(loc)
	fechaConsulta := now
	horaConsulta := now

	if fecha != nil {
		fechaConsulta = *fecha
	}
	if hora != nil {
		horaConsulta = *hora
	}

	diaSemana := s.obtenerDiaSemanaEspanol(fechaConsulta.Weekday())

	qs := s.ormer.QueryTable("oferta").Filter("activo", true).Filter("pk_id_restaurante", restauranteId)

	qs = qs.Filter("fecha_inicio__lte", fechaConsulta).Filter("fecha_fin__gte", fechaConsulta)

	var ofertas []*models.Oferta
	_, err := qs.All(&ofertas)
	if err != nil {
		return nil, fmt.Errorf("error al obtener ofertas: %w", err)
	}

	var ofertasActivas []*models.OfertaActivaResponse

	for _, oferta := range ofertas {

		if len(oferta.DiasSemanaArray) > 0 {
			diaValido := false
			for _, dia := range oferta.DiasSemanaArray {
				if strings.TrimSpace(dia) == diaSemana {
					diaValido = true
					break
				}
			}
			if !diaValido {
				continue
			}
		}

		if oferta.HoraInicio != nil && oferta.HoraFin != nil {
			horaOfertaInicio := time.Date(0, 1, 1, oferta.HoraInicio.Hour(), oferta.HoraInicio.Minute(), oferta.HoraInicio.Second(), 0, time.UTC)
			horaOfertaFin := time.Date(0, 1, 1, oferta.HoraFin.Hour(), oferta.HoraFin.Minute(), oferta.HoraFin.Second(), 0, time.UTC)
			horaActual := time.Date(0, 1, 1, horaConsulta.Hour(), horaConsulta.Minute(), horaConsulta.Second(), 0, time.UTC)

			if horaActual.Before(horaOfertaInicio) || horaActual.After(horaOfertaFin) {
				continue
			}
		}

		productosIds, err := s.obtenerProductosOferta(oferta.PkIdOferta)
		if err != nil {
			return nil, fmt.Errorf("error al obtener productos de la oferta %d: %w", oferta.PkIdOferta, err)
		}

		if productoId != nil {
			productoEnOferta := false
			for _, pid := range productosIds {
				if pid == *productoId {
					productoEnOferta = true
					break
				}
			}
			if !productoEnOferta {
				continue
			}
		}

		ofertaActiva := &models.OfertaActivaResponse{
			OfertaId:       oferta.PkIdOferta,
			Titulo:         oferta.Titulo,
			TipoDescuento:  oferta.TipoDescuento,
			ValorDescuento: oferta.ValorDescuento,
			ProductosIds:   productosIds,
		}

		ofertasActivas = append(ofertasActivas, ofertaActiva)
	}

	return ofertasActivas, nil
}

func (s *OfertaService) ValidarReglasNegocioOferta(oferta *models.Oferta) error {

	switch oferta.TipoDescuento {
	case models.TipoDescuentoPorcentaje:
		if oferta.ValorDescuento < 1 || oferta.ValorDescuento > 100 {
			return fmt.Errorf("el porcentaje de descuento debe estar entre 1 y 100")
		}
	case models.TipoDescuentoMonto:
		if oferta.ValorDescuento < 0 {
			return fmt.Errorf("el monto de descuento debe ser mayor o igual a 0")
		}
	}

	if oferta.FechaFin.Before(oferta.FechaInicio) {
		return fmt.Errorf("la fecha de fin debe ser posterior a la fecha de inicio")
	}

	if (oferta.HoraInicio != nil && oferta.HoraFin == nil) || (oferta.HoraInicio == nil && oferta.HoraFin != nil) {
		return fmt.Errorf("si se especifica horario, debe incluir tanto hora de inicio como hora de fin")
	}

	if oferta.HoraInicio != nil && oferta.HoraFin != nil {
		horaInicio := time.Date(0, 1, 1, oferta.HoraInicio.Hour(), oferta.HoraInicio.Minute(), oferta.HoraInicio.Second(), 0, time.UTC)
		horaFin := time.Date(0, 1, 1, oferta.HoraFin.Hour(), oferta.HoraFin.Minute(), oferta.HoraFin.Second(), 0, time.UTC)

		if horaFin.Before(horaInicio) || horaFin.Equal(horaInicio) {
			return fmt.Errorf("la hora de fin debe ser posterior a la hora de inicio")
		}
	}

	if len(oferta.DiasSemana) > 0 {
		diasValidos := map[string]bool{
			string(models.DiaLunes):     true,
			string(models.DiaMartes):    true,
			string(models.DiaMiercoles): true,
			string(models.DiaJueves):    true,
			string(models.DiaViernes):   true,
			string(models.DiaSabado):    true,
			string(models.DiaDomingo):   true,
		}

		for _, dia := range oferta.DiasSemanaArray {
			if !diasValidos[strings.TrimSpace(dia)] {
				return fmt.Errorf("día de la semana inválido: %s", dia)
			}
		}
	}

	return nil
}

func (s *OfertaService) CalcularDescuentoOferta(oferta *models.Oferta, items []models.ValidarCuponItemRequest) (int64, error) {

	productosOferta, err := s.obtenerProductosOferta(oferta.PkIdOferta)
	if err != nil {
		return 0, fmt.Errorf("error al obtener productos de la oferta: %w", err)
	}

	var montoAplicable int64
	for _, item := range items {
		for _, prodId := range productosOferta {
			if item.ProductoId == prodId {
				montoAplicable += item.Precio * int64(item.Cantidad)
				break
			}
		}
	}

	if montoAplicable == 0 {
		return 0, nil
	}

	switch oferta.TipoDescuento {
	case models.TipoDescuentoPorcentaje:
		return (montoAplicable * oferta.ValorDescuento) / 100, nil
	case models.TipoDescuentoMonto:
		if oferta.ValorDescuento > montoAplicable {
			return montoAplicable, nil
		}
		return oferta.ValorDescuento, nil
	}

	return 0, nil
}

func (s *OfertaService) obtenerProductosOferta(ofertaId int64) ([]int64, error) {
	var ofertaProductos []*models.OfertaProducto
	_, err := s.ormer.QueryTable("oferta_producto").Filter("pk_id_oferta", ofertaId).All(&ofertaProductos)
	if err != nil {
		return nil, err
	}

	var productosIds []int64
	for _, op := range ofertaProductos {
		if op.PkIdProducto != nil {
			productosIds = append(productosIds, op.PkIdProducto.PK_ID_PRODUCTO)
		}
	}

	return productosIds, nil
}

func (s *OfertaService) obtenerDiaSemanaEspanol(weekday time.Weekday) string {
	switch weekday {
	case time.Monday:
		return string(models.DiaLunes)
	case time.Tuesday:
		return string(models.DiaMartes)
	case time.Wednesday:
		return string(models.DiaMiercoles)
	case time.Thursday:
		return string(models.DiaJueves)
	case time.Friday:
		return string(models.DiaViernes)
	case time.Saturday:
		return string(models.DiaSabado)
	case time.Sunday:
		return string(models.DiaDomingo)
	default:
		return ""
	}
}
