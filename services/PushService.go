package services

import (
	"context"
	"fmt"
	"time"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
)

type PushService struct {
	ormer orm.Ormer
}

func NewPushService(ormer orm.Ormer) *PushService {
	return &PushService{ormer: ormer}
}

// ValidarRegistroDispositivo valida las reglas de negocio para registrar un dispositivo push
func (s *PushService) ValidarRegistroDispositivo(req *models.RegistrarDispositivoRequest) error {
	// Validar que exactamente uno de cliente o trabajador esté especificado
	if (req.PkDocumentoCliente == nil && req.PkDocumentoTrabajador == nil) ||
		(req.PkDocumentoCliente != nil && req.PkDocumentoTrabajador != nil) {
		return fmt.Errorf("debe especificar exactamente uno de cliente o trabajador")
	}

	// Validar coherencia entre plataforma y campos requeridos
	switch req.Plataforma {
	case models.PlataformaWeb:
		if req.Endpoint == nil || req.P256dh == nil || req.Auth == nil {
			return fmt.Errorf("para plataforma WEB se requieren endpoint, p256dh y auth")
		}
		if req.FcmToken != nil {
			return fmt.Errorf("para plataforma WEB no se debe especificar fcmToken")
		}
	case models.PlataformaAndroid, models.PlataformaIOS:
		if req.FcmToken == nil {
			return fmt.Errorf("para plataformas ANDROID/IOS se requiere fcmToken")
		}
		if req.Endpoint != nil || req.P256dh != nil || req.Auth != nil {
			return fmt.Errorf("para plataformas ANDROID/IOS no se deben especificar endpoint, p256dh o auth")
		}
	default:
		return fmt.Errorf("plataforma no válida: %s", req.Plataforma)
	}

	return nil
}

// RegistrarDispositivo registra un nuevo dispositivo push
func (s *PushService) RegistrarDispositivo(ctx context.Context, req *models.RegistrarDispositivoRequest) (*models.PushDispositivo, error) {
	// Validar reglas de negocio
	if err := s.ValidarRegistroDispositivo(req); err != nil {
		return nil, err
	}

	now := time.Now()
	dispositivo := &models.PushDispositivo{
		Plataforma:            req.Plataforma,
		Endpoint:              req.Endpoint,
		P256dh:                req.P256dh,
		Auth:                  req.Auth,
		FcmToken:              req.FcmToken,
		Enabled:               true,
		Locale:                req.Locale,
		TimeZone:              req.TimeZone,
		AppVersion:            req.AppVersion,
		UserAgent:             req.UserAgent,
		SubscribedTopicsArray: req.SubscribedTopics,
		CreatedAt:             now,
		LastSeenAt:            &now,
	}

	if req.PkDocumentoCliente != nil {
		dispositivo.PkDocumentoCliente = &models.Cliente{PK_DOCUMENTO_CLIENTE: *req.PkDocumentoCliente}
	}

	if req.PkDocumentoTrabajador != nil {
		dispositivo.PkDocumentoTrabajador = &models.Trabajador{PK_DOCUMENTO_TRABAJADOR: *req.PkDocumentoTrabajador}
	}

	_, err := s.ormer.Insert(dispositivo)
	if err != nil {
		return nil, fmt.Errorf("error al registrar dispositivo: %w", err)
	}

	return dispositivo, nil
}

// ActualizarUltimaVista actualiza el campo last_seen_at de un dispositivo
func (s *PushService) ActualizarUltimaVista(ctx context.Context, dispositivoId int64) error {
	dispositivo := &models.PushDispositivo{PkIdPushDispositivo: dispositivoId}
	err := s.ormer.Read(dispositivo)
	if err != nil {
		if err == orm.ErrNoRows {
			return fmt.Errorf("dispositivo no encontrado")
		}
		return fmt.Errorf("error al buscar dispositivo: %w", err)
	}

	now := time.Now()
	dispositivo.LastSeenAt = &now

	_, err = s.ormer.Update(dispositivo, "LastSeenAt")
	if err != nil {
		return fmt.Errorf("error al actualizar última vista: %w", err)
	}

	return nil
}

// ActualizarEstadoDispositivo actualiza el estado enabled de un dispositivo
func (s *PushService) ActualizarEstadoDispositivo(ctx context.Context, dispositivoId int64, enabled bool) error {
	dispositivo := &models.PushDispositivo{PkIdPushDispositivo: dispositivoId}
	err := s.ormer.Read(dispositivo)
	if err != nil {
		if err == orm.ErrNoRows {
			return fmt.Errorf("dispositivo no encontrado")
		}
		return fmt.Errorf("error al buscar dispositivo: %w", err)
	}

	dispositivo.Enabled = enabled

	_, err = s.ormer.Update(dispositivo, "Enabled")
	if err != nil {
		return fmt.Errorf("error al actualizar estado del dispositivo: %w", err)
	}

	return nil
}

// ActualizarTopicsDispositivo actualiza los topics suscritos de un dispositivo
func (s *PushService) ActualizarTopicsDispositivo(ctx context.Context, dispositivoId int64, topics []string) error {
	dispositivo := &models.PushDispositivo{PkIdPushDispositivo: dispositivoId}
	err := s.ormer.Read(dispositivo)
	if err != nil {
		if err == orm.ErrNoRows {
			return fmt.Errorf("dispositivo no encontrado")
		}
		return fmt.Errorf("error al buscar dispositivo: %w", err)
	}

	dispositivo.SubscribedTopicsArray = topics

	_, err = s.ormer.Update(dispositivo, "SubscribedTopicsArray")
	if err != nil {
		return fmt.Errorf("error al actualizar topics del dispositivo: %w", err)
	}

	return nil
}

// RegistrarEnvio registra el resultado de un envío push
func (s *PushService) RegistrarEnvio(ctx context.Context, req *models.RegistrarEnvioRequest) (*models.PushEnvio, error) {
	// Verificar que el dispositivo existe
	dispositivo := &models.PushDispositivo{PkIdPushDispositivo: req.PkIdPushDispositivo}
	err := s.ormer.Read(dispositivo)
	if err != nil {
		if err == orm.ErrNoRows {
			return nil, fmt.Errorf("dispositivo no encontrado")
		}
		return nil, fmt.Errorf("error al buscar dispositivo: %w", err)
	}

	// Validar proveedor
	if !req.Proveedor.IsValid() {
		return nil, fmt.Errorf("proveedor no válido: %s", req.Proveedor)
	}

	envio := &models.PushEnvio{
		PkIdPushDispositivo: dispositivo,
		Proveedor:           req.Proveedor,
		DataObj:             req.Data,
		Exito:               req.Exito,
		StatusCode:          req.StatusCode,
		ErrorCode:           req.ErrorCode,
		SentAt:              time.Now(),
	}

	_, err = s.ormer.Insert(envio)
	if err != nil {
		return nil, fmt.Errorf("error al registrar envío: %w", err)
	}

	return envio, nil
}

// EnviarNotificacion envía una notificación push a los destinatarios especificados
func (s *PushService) EnviarNotificacion(req *models.EnviarNotificacionRequest) (*models.EnviarNotificacionResponse, error) {
	// Validar el remitente
	if err := s.validarRemitente(&req.Remitente); err != nil {
		return nil, fmt.Errorf("remitente inválido: %w", err)
	}

	// Obtener dispositivos destinatarios
	dispositivos, err := s.obtenerDispositivosDestinatarios(&req.Destinatarios)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo dispositivos: %w", err)
	}

	if len(dispositivos) == 0 {
		return &models.EnviarNotificacionResponse{
			TotalDispositivos: 0,
			EnviosExitosos:    0,
			EnviosFallidos:    0,
			DetalleEnvios:     []models.DetalleEnvioNotificacion{},
			ResumenDestinatarios: models.ResumenDestinatarios{
				TipoDestinatario: string(req.Destinatarios.Tipo),
			},
		}, nil
	}

	// Enviar notificaciones a cada dispositivo
	var detalleEnvios []models.DetalleEnvioNotificacion
	enviosExitosos := 0
	enviosFallidos := 0

	for _, dispositivo := range dispositivos {
		detalle := s.enviarNotificacionDispositivo(&dispositivo, &req.Notificacion)
		detalleEnvios = append(detalleEnvios, detalle)

		// Registrar el envío en la base de datos
		s.registrarEnvioNotificacion(&dispositivo, &req.Notificacion, detalle.Exito, detalle.StatusCode, detalle.ErrorCode)

		if detalle.Exito {
			enviosExitosos++
		} else {
			enviosFallidos++
		}
	}

	// Crear resumen de destinatarios
	resumen := s.crearResumenDestinatarios(&req.Destinatarios, dispositivos)

	return &models.EnviarNotificacionResponse{
		TotalDispositivos:    len(dispositivos),
		EnviosExitosos:       enviosExitosos,
		EnviosFallidos:       enviosFallidos,
		DetalleEnvios:        detalleEnvios,
		ResumenDestinatarios: resumen,
	}, nil
}

// validarRemitente valida que el remitente sea válido
func (s *PushService) validarRemitente(remitente *models.RemitenteNotificacion) error {
	if remitente.Tipo == models.RemitenteTrabajador {
		if remitente.DocumentoTrabajador == nil {
			return fmt.Errorf("documentoTrabajador es requerido para remitente TRABAJADOR")
		}

		// Verificar que el trabajador existe
		trabajador := &models.Trabajador{}
		err := s.ormer.QueryTable("trabajador").Filter("pk_documento_trabajador", *remitente.DocumentoTrabajador).One(trabajador)
		if err != nil {
			return fmt.Errorf("trabajador no encontrado")
		}
	}
	return nil
}

// obtenerDispositivosDestinatarios obtiene los dispositivos según el tipo de destinatario
func (s *PushService) obtenerDispositivosDestinatarios(destinatarios *models.DestinatariosNotificacion) ([]models.PushDispositivo, error) {
	qs := s.ormer.QueryTable("push_dispositivo").Filter("enabled", true)

	switch destinatarios.Tipo {
	case models.DestinatarioTodos:
		// Todos los dispositivos activos
		break
	case models.DestinatarioCliente:
		if destinatarios.DocumentoCliente == nil {
			return nil, fmt.Errorf("documentoCliente es requerido para destinatario CLIENTE")
		}
		qs = qs.Filter("pk_documento_cliente", *destinatarios.DocumentoCliente)
	case models.DestinatarioTrabajador:
		if destinatarios.DocumentoTrabajador == nil {
			return nil, fmt.Errorf("documentoTrabajador es requerido para destinatario TRABAJADOR")
		}
		qs = qs.Filter("pk_documento_trabajador", *destinatarios.DocumentoTrabajador)
	case models.DestinatarioTopic:
		if destinatarios.Topic == nil {
			return nil, fmt.Errorf("topic es requerido para destinatario TOPIC")
		}
		// Filtrar dispositivos que tengan el topic suscrito
		qs = qs.Filter("subscribed_topics__contains", *destinatarios.Topic)
	default:
		return nil, fmt.Errorf("tipo de destinatario no válido: %s", destinatarios.Tipo)
	}

	var dispositivos []models.PushDispositivo
	_, err := qs.All(&dispositivos)
	return dispositivos, err
}

// enviarNotificacionDispositivo envía la notificación a un dispositivo específico
func (s *PushService) enviarNotificacionDispositivo(dispositivo *models.PushDispositivo, notificacion *models.ContenidoNotificacion) models.DetalleEnvioNotificacion {
	var documentoCliente *int64
	var documentoTrabajador *int64

	if dispositivo.PkDocumentoCliente != nil {
		documentoCliente = &dispositivo.PkDocumentoCliente.PK_DOCUMENTO_CLIENTE
	}
	if dispositivo.PkDocumentoTrabajador != nil {
		documentoTrabajador = &dispositivo.PkDocumentoTrabajador.PK_DOCUMENTO_TRABAJADOR
	}

	detalle := models.DetalleEnvioNotificacion{
		PushDispositivoId:   dispositivo.PkIdPushDispositivo,
		Plataforma:          string(dispositivo.Plataforma),
		DocumentoCliente:    documentoCliente,
		DocumentoTrabajador: documentoTrabajador,
	}

	// Aquí iría la lógica real de envío según la plataforma
	switch dispositivo.Plataforma {
	case models.PlataformaWeb:
		// Envío Web Push
		detalle.Exito, detalle.StatusCode, detalle.ErrorCode = s.enviarWebPush(dispositivo, notificacion)
	case models.PlataformaAndroid, models.PlataformaIOS:
		// Envío FCM
		detalle.Exito, detalle.StatusCode, detalle.ErrorCode = s.enviarFCM(dispositivo, notificacion)
	default:
		detalle.Exito = false
		statusCode := 400
		errorCode := "PLATAFORMA_NO_SOPORTADA"
		detalle.StatusCode = &statusCode
		detalle.ErrorCode = &errorCode
	}

	return detalle
}

// enviarWebPush simula el envío de Web Push (aquí iría la integración real)
func (s *PushService) enviarWebPush(dispositivo *models.PushDispositivo, notificacion *models.ContenidoNotificacion) (bool, *int, *string) {
	// TODO: Implementar envío real de Web Push usando dispositivo.Endpoint, P256dh, Auth
	// Por ahora simulamos éxito
	statusCode := 200
	return true, &statusCode, nil
}

// enviarFCM simula el envío de FCM (aquí iría la integración real)
func (s *PushService) enviarFCM(dispositivo *models.PushDispositivo, notificacion *models.ContenidoNotificacion) (bool, *int, *string) {
	// TODO: Implementar envío real de FCM usando dispositivo.FcmToken
	// Por ahora simulamos éxito
	statusCode := 200
	return true, &statusCode, nil
}

// registrarEnvioNotificacion registra el envío en la base de datos
func (s *PushService) registrarEnvioNotificacion(dispositivo *models.PushDispositivo, notificacion *models.ContenidoNotificacion, exito bool, statusCode *int, errorCode *string) {
	envio := &models.PushEnvio{
		PkIdPushDispositivo: dispositivo,
		Proveedor:           s.obtenerProveedor(dispositivo.Plataforma),
		DataObj:             notificacion.Datos,
		Exito:               exito,
		StatusCode:          statusCode,
		ErrorCode:           errorCode,
		SentAt:              time.Now(),
	}

	_, _ = s.ormer.Insert(envio)
}

// obtenerProveedor obtiene el proveedor según la plataforma
func (s *PushService) obtenerProveedor(plataforma models.PlataformaNotificacion) models.ProveedorPush {
	switch plataforma {
	case models.PlataformaWeb:
		return models.ProveedorWebPush
	case models.PlataformaAndroid, models.PlataformaIOS:
		return models.ProveedorFCM
	default:
		return models.ProveedorWebPush
	}
}

// crearResumenDestinatarios crea el resumen de destinatarios notificados
func (s *PushService) crearResumenDestinatarios(destinatarios *models.DestinatariosNotificacion, dispositivos []models.PushDispositivo) models.ResumenDestinatarios {
	resumen := models.ResumenDestinatarios{
		TipoDestinatario: string(destinatarios.Tipo),
	}

	clientesMap := make(map[int64]bool)
	trabajadoresMap := make(map[int64]bool)
	topicsMap := make(map[string]bool)

	for _, dispositivo := range dispositivos {
		if dispositivo.PkDocumentoCliente != nil {
			clientesMap[dispositivo.PkDocumentoCliente.PK_DOCUMENTO_CLIENTE] = true
		}
		if dispositivo.PkDocumentoTrabajador != nil {
			trabajadoresMap[dispositivo.PkDocumentoTrabajador.PK_DOCUMENTO_TRABAJADOR] = true
		}
		// Agregar topics si es el caso
		if destinatarios.Tipo == models.DestinatarioTopic && destinatarios.Topic != nil {
			topicsMap[*destinatarios.Topic] = true
		}
	}

	// Convertir maps a slices
	for clienteId := range clientesMap {
		resumen.ClientesNotificados = append(resumen.ClientesNotificados, clienteId)
	}
	for trabajadorId := range trabajadoresMap {
		resumen.TrabajadoresNotificados = append(resumen.TrabajadoresNotificados, trabajadorId)
	}
	for topic := range topicsMap {
		resumen.TopicsNotificados = append(resumen.TopicsNotificados, topic)
	}

	return resumen
}
