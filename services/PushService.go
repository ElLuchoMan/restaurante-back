//go:build !unit
// +build !unit

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"restaurante/models"

	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	"golang.org/x/oauth2"
	oauthgoogle "golang.org/x/oauth2/google"
)

type PushService struct {
	ormer orm.Ormer
}

var (
	webpushSendNotificationWithContextFn = func(ctx context.Context, payload []byte, subscription *webpush.Subscription, options *webpush.Options) (*http.Response, error) {
		return webpush.SendNotificationWithContext(ctx, payload, subscription, options)
	}
	newHTTPClientFn = func(timeout time.Duration) *http.Client {
		return &http.Client{Timeout: timeout}
	}
	httpDoFn = func(client *http.Client, req *http.Request) (*http.Response, error) {
		return client.Do(req)
	}
	httpNewRequestFn         = http.NewRequest
	findDefaultCredentialsFn = oauthgoogle.FindDefaultCredentials
	tokenSourceTokenFn       = func(ts oauth2.TokenSource) (*oauth2.Token, error) {
		if ts == nil {
			return nil, fmt.Errorf("token source nil")
		}
		return ts.Token()
	}
	// #nosec G101 -- Esta es la URL estándar pública del metadata server de GCP, no es una credencial
	metadataTokenURL = "http://169.254.169.254/computeMetadata/v1/instance/service-accounts/default/token"
	jsonMarshalFn    = json.Marshal
	jsonUnmarshalFn  = json.Unmarshal
	jsonNewDecoderFn = func(r io.Reader) *json.Decoder { return json.NewDecoder(r) }
)

var obtainFCMTokenWithADCFn = obtainFCMTokenWithADC

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

// RegistrarDispositivo registra un nuevo dispositivo push o actualiza uno existente
func (s *PushService) RegistrarDispositivo(ctx context.Context, req *models.RegistrarDispositivoRequest) (*models.PushDispositivo, error) {
	// Validar reglas de negocio
	if err := s.ValidarRegistroDispositivo(req); err != nil {
		return nil, err
	}

	now := time.Now()

	// Intentar encontrar dispositivo existente por fcm_token o endpoint
	var existingDevice *models.PushDispositivo
	if req.FcmToken != nil && *req.FcmToken != "" {
		// Buscar por FCM token
		device := &models.PushDispositivo{}
		err := s.ormer.QueryTable("push_dispositivo").Filter("fcm_token", *req.FcmToken).One(device)
		if err == nil {
			existingDevice = device
		}
	} else if req.Endpoint != nil && *req.Endpoint != "" {
		// Buscar por endpoint (para WEB push)
		device := &models.PushDispositivo{}
		err := s.ormer.QueryTable("push_dispositivo").Filter("endpoint", *req.Endpoint).One(device)
		if err == nil {
			existingDevice = device
		}
	}

	if existingDevice != nil {
		// Actualizar dispositivo existente
		existingDevice.Plataforma = req.Plataforma
		existingDevice.Endpoint = req.Endpoint
		existingDevice.P256dh = req.P256dh
		existingDevice.Auth = req.Auth
		existingDevice.FcmToken = req.FcmToken
		existingDevice.Enabled = true // Reactivar si estaba deshabilitado
		existingDevice.Locale = req.Locale
		existingDevice.TimeZone = req.TimeZone
		existingDevice.AppVersion = req.AppVersion
		existingDevice.UserAgent = req.UserAgent
		existingDevice.SubscribedTopicsArray = req.SubscribedTopics
		existingDevice.LastSeenAt = &now

		// Actualizar relaciones de usuario si cambiaron
		if req.PkDocumentoCliente != nil {
			existingDevice.PkDocumentoCliente = &models.Cliente{PK_DOCUMENTO_CLIENTE: *req.PkDocumentoCliente}
		} else {
			existingDevice.PkDocumentoCliente = nil
		}

		if req.PkDocumentoTrabajador != nil {
			existingDevice.PkDocumentoTrabajador = &models.Trabajador{PK_DOCUMENTO_TRABAJADOR: *req.PkDocumentoTrabajador}
		} else {
			existingDevice.PkDocumentoTrabajador = nil
		}

		// Forzar serialización antes del update
		existingDevice.BeforeUpdate()

		_, err := s.ormer.Update(existingDevice)
		if err != nil {
			return nil, fmt.Errorf("error al actualizar dispositivo: %w", err)
		}

		return existingDevice, nil
	}

	// Crear nuevo dispositivo
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

	// Forzar serialización antes del insert
	dispositivo.BeforeInsert()

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
	case models.DestinatarioClientes:
		// Todos los dispositivos asociados a clientes (excluye trabajadores)
		qs = qs.Filter("pk_documento_cliente__isnull", false)
		qs = qs.Filter("pk_documento_trabajador__isnull", true)
	case models.DestinatarioTrabajador:
		if destinatarios.DocumentoTrabajador == nil {
			return nil, fmt.Errorf("documentoTrabajador es requerido para destinatario TRABAJADOR")
		}
		qs = qs.Filter("pk_documento_trabajador", *destinatarios.DocumentoTrabajador)
	case models.DestinatarioTrabajadores:
		// Todos los dispositivos asociados a trabajadores (excluye clientes)
		qs = qs.Filter("pk_documento_trabajador__isnull", false)
		qs = qs.Filter("pk_documento_cliente__isnull", true)
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

// enviarWebPush envía una notificación Web Push a un dispositivo usando el protocolo Web Push
func (s *PushService) enviarWebPush(dispositivo *models.PushDispositivo, notificacion *models.ContenidoNotificacion) (bool, *int, *string) {
	// Validar que el dispositivo tenga las credenciales necesarias
	if dispositivo.Endpoint == nil || *dispositivo.Endpoint == "" {
		logs.Error("[Web Push] Dispositivo %d sin endpoint", dispositivo.PkIdPushDispositivo)
		status := 400
		code := "ENDPOINT_VACIO"
		return false, &status, &code
	}
	if dispositivo.P256dh == nil || *dispositivo.P256dh == "" {
		logs.Error("[Web Push] Dispositivo %d sin p256dh", dispositivo.PkIdPushDispositivo)
		status := 400
		code := "P256DH_VACIO"
		return false, &status, &code
	}
	if dispositivo.Auth == nil || *dispositivo.Auth == "" {
		logs.Error("[Web Push] Dispositivo %d sin auth", dispositivo.PkIdPushDispositivo)
		status := 400
		code := "AUTH_VACIO"
		return false, &status, &code
	}

	// Obtener configuración VAPID
	vapidPublicKey := os.Getenv("VAPID_PUBLIC_KEY")
	vapidPrivateKey := os.Getenv("VAPID_PRIVATE_KEY")
	vapidSubject := os.Getenv("VAPID_SUBJECT")

	// Fallback a configuración de Beego
	if vapidPublicKey == "" {
		if v, err := web.AppConfig.String("vapid_public_key"); err == nil && strings.TrimSpace(v) != "" {
			vapidPublicKey = v
		}
	}
	if vapidPrivateKey == "" {
		if v, err := web.AppConfig.String("vapid_private_key"); err == nil && strings.TrimSpace(v) != "" {
			vapidPrivateKey = v
		}
	}
	if vapidSubject == "" {
		if v, err := web.AppConfig.String("vapid_subject"); err == nil && strings.TrimSpace(v) != "" {
			vapidSubject = v
		}
	}

	// Validar que estén configuradas las claves VAPID
	if vapidPublicKey == "" || vapidPrivateKey == "" {
		logs.Error("[Web Push] Claves VAPID no configuradas (VAPID_PUBLIC_KEY, VAPID_PRIVATE_KEY)")
		status := 500
		code := "CONFIG_VAPID_FALTA"
		return false, &status, &code
	}
	if vapidSubject == "" {
		logs.Warn("[Web Push] VAPID_SUBJECT no configurado, usando default")
		vapidSubject = "mailto:admin@restaurante.com"
	}

	// Construir el payload de la notificación según Web Push estándar
	payload := map[string]interface{}{
		"notification": map[string]interface{}{
			"title": notificacion.Titulo,
			"body":  notificacion.Mensaje,
			"icon":  "/icons/web-app-manifest-192x192.png",
			"badge": "/icons/web-app-manifest-192x192.png",
		},
	}

	// Agregar datos adicionales si existen
	if len(notificacion.Datos) > 0 {
		var datosMap map[string]interface{}
		if err := jsonUnmarshalFn(notificacion.Datos, &datosMap); err == nil {
			if notif, ok := payload["notification"].(map[string]interface{}); ok {
				notif["data"] = datosMap
			}
		}
	}

	payloadBytes, err := jsonMarshalFn(payload)
	if err != nil {
		logs.Error("[Web Push] Error al serializar payload: %v", err)
		status := 500
		code := "ERROR_SERIALIZAR_PAYLOAD"
		return false, &status, &code
	}

	// Crear suscripción Web Push
	subscription := &webpush.Subscription{
		Endpoint: *dispositivo.Endpoint,
		Keys: webpush.Keys{
			P256dh: *dispositivo.P256dh,
			Auth:   *dispositivo.Auth,
		},
	}

	// Configurar opciones VAPID
	options := &webpush.Options{
		Subscriber:      vapidSubject,
		VAPIDPublicKey:  vapidPublicKey,
		VAPIDPrivateKey: vapidPrivateKey,
		TTL:             30 * 24 * 60 * 60, // 30 días
	}

	// Enviar la notificación con timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := webpushSendNotificationWithContextFn(ctx, payloadBytes, subscription, options)
	if err != nil {
		logs.Error("[Web Push] Error al enviar notificación al dispositivo %d: %v", dispositivo.PkIdPushDispositivo, err)
		status := 500
		code := "ERROR_ENVIO_WEB_PUSH"
		return false, &status, &code
	}
	defer func() {
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()

	statusCode := resp.StatusCode

	// Manejar códigos de estado específicos
	switch statusCode {
	case 201, 200:
		// Éxito
		logs.Info("[Web Push] Notificación enviada exitosamente al dispositivo %d (status: %d)", dispositivo.PkIdPushDispositivo, statusCode)
		return true, &statusCode, nil

	case 410: // Gone - el dispositivo fue desregistrado
		logs.Warn("[Web Push] Dispositivo %d desregistrado (410 Gone), deshabilitando...", dispositivo.PkIdPushDispositivo)
		// Desactivar el dispositivo
		if err := s.ActualizarEstadoDispositivo(context.Background(), dispositivo.PkIdPushDispositivo, false); err != nil {
			logs.Error("[Web Push] Error al desactivar dispositivo %d: %v", dispositivo.PkIdPushDispositivo, err)
		}
		code := "DISPOSITIVO_DESREGISTRADO"
		return false, &statusCode, &code

	case 404: // Not Found - el endpoint no existe o es inválido
		logs.Warn("[Web Push] Endpoint inválido para dispositivo %d (404 Not Found), deshabilitando...", dispositivo.PkIdPushDispositivo)
		// Desactivar el dispositivo
		if err := s.ActualizarEstadoDispositivo(context.Background(), dispositivo.PkIdPushDispositivo, false); err != nil {
			logs.Error("[Web Push] Error al desactivar dispositivo %d: %v", dispositivo.PkIdPushDispositivo, err)
		}
		code := "ENDPOINT_INVALIDO"
		return false, &statusCode, &code

	case 401: // Unauthorized - error de autenticación VAPID
		logs.Error("[Web Push] Error de autenticación VAPID (401 Unauthorized) - verificar claves VAPID")
		code := "ERROR_AUTH_VAPID"
		return false, &statusCode, &code

	case 400: // Bad Request
		// Intentar leer el cuerpo de la respuesta para más detalles
		bodyBytes, _ := io.ReadAll(resp.Body)
		logs.Error("[Web Push] Bad Request (400) para dispositivo %d: %s", dispositivo.PkIdPushDispositivo, string(bodyBytes))
		code := "BAD_REQUEST"
		return false, &statusCode, &code

	case 413: // Payload Too Large
		logs.Error("[Web Push] Payload demasiado grande (413) para dispositivo %d", dispositivo.PkIdPushDispositivo)
		code := "PAYLOAD_TOO_LARGE"
		return false, &statusCode, &code

	case 429: // Too Many Requests
		logs.Warn("[Web Push] Demasiadas solicitudes (429) para dispositivo %d - rate limit", dispositivo.PkIdPushDispositivo)
		code := "RATE_LIMIT"
		return false, &statusCode, &code

	default:
		// Otros errores
		bodyBytes, _ := io.ReadAll(resp.Body)
		logs.Error("[Web Push] Error desconocido (status %d) para dispositivo %d: %s", statusCode, dispositivo.PkIdPushDispositivo, string(bodyBytes))
		code := fmt.Sprintf("ERROR_HTTP_%d", statusCode)
		return false, &statusCode, &code
	}
}

// enviarFCM simula el envío de FCM (aquí iría la integración real)
func (s *PushService) enviarFCM(dispositivo *models.PushDispositivo, notificacion *models.ContenidoNotificacion) (bool, *int, *string) {
	if dispositivo.FcmToken == nil || strings.TrimSpace(*dispositivo.FcmToken) == "" {
		status := 400
		code := "FCM_TOKEN_VACIO"
		return false, &status, &code
	}

	projectId := os.Getenv("FIREBASE_PROJECT_ID")
	if projectId == "" {
		if v, err := web.AppConfig.String("firebase_project_id"); err == nil && strings.TrimSpace(v) != "" {
			projectId = v
		}
	}
	if projectId == "" {
		status := 500
		code := "CONFIG_FIREBASE_FALTA_PROJECT_ID"
		return false, &status, &code
	}

	// Construir payload HTTP v1
	body := map[string]interface{}{
		"message": map[string]interface{}{
			"token": *dispositivo.FcmToken,
			"notification": map[string]string{
				"title": notificacion.Titulo,
				"body":  notificacion.Mensaje,
			},
			"data": func() map[string]string {
				if len(notificacion.Datos) == 0 {
					return nil
				}
				// Intentar convertir datos arbitrarios a mapa de strings
				var tmp map[string]interface{}
				_ = jsonUnmarshalFn(notificacion.Datos, &tmp)
				out := map[string]string{}
				for k, v := range tmp {
					out[k] = fmt.Sprint(v)
				}
				return out
			}(),
			"android": map[string]interface{}{
				"priority": "high",
				"notification": map[string]string{
					"channel_id":   "default",
					"click_action": "FLUTTER_NOTIFICATION_CLICK",
				},
			},
		},
	}

	// Obtener bearer token del metadata server o credenciales (requiere que el entorno ya tenga auth lista)
	// Para simplicidad: usar Application Default Credentials vía metadata server (GCE) o ADC local
	// Aquí solo llamamos al endpoint y dejamos al entorno manejar Authorization (por proxy o inyectado)
	// Si no hay Authorization, retornamos error claro.

	// Permitir inyección de Authorization via env FCM_BEARER_TOKEN (fallback)
	bearer := os.Getenv("FCM_BEARER_TOKEN")
	if bearer == "" {
		// Fallback: intentar obtener token con ADC (Application Default Credentials)
		// Requiere que el entorno tenga GOOGLE_APPLICATION_CREDENTIALS o identidad GCE/GKE con permisos
		token, adcErr := obtainFCMTokenWithADCFn()
		if adcErr != nil || strings.TrimSpace(token) == "" {
			status := 500
			code := "AUTH_FCM_NO_CONFIGURADO"
			return false, &status, &code
		}
		bearer = token
	}

	url := fmt.Sprintf("https://fcm.googleapis.com/v1/projects/%s/messages:send", projectId)

	reqBody, _ := jsonMarshalFn(body)
	httpReq, _ := httpNewRequestFn(http.MethodPost, url, strings.NewReader(string(reqBody)))
	httpReq.Header.Set("Content-Type", "application/json; charset=utf-8")
	httpReq.Header.Set("Authorization", "Bearer "+bearer)

	client := newHTTPClientFn(10 * time.Second)
	resp, err := httpDoFn(client, httpReq)
	if err != nil {
		status := 500
		code := "FCM_HTTP_ERROR"
		return false, &status, &code
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	status := resp.StatusCode
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return true, &status, nil
	}

	// Intentar extraer error code de respuesta FCM
	var respErr struct {
		Error struct {
			Status  string `json:"status"`
			Message string `json:"message"`
		} `json:"error"`
	}
	_ = jsonNewDecoderFn(resp.Body).Decode(&respErr)
	errCode := respErr.Error.Status
	if errCode == "" {
		errCode = "FCM_ERROR"
	}
	return false, &status, &errCode
}

// obtainFCMTokenWithADC intenta obtener un token OAuth2 usando ADC sin añadir dependencias externas.
// Para mantener el repo liviano, llamamos directamente al endpoint de token de gcloud si está disponible
// vía env GOOGLE_APPLICATION_CREDENTIALS no es trivial sin librerías; por eso este fallback primero intenta
// usar el metadata server (entornos GCE/GKE). Si no, retorna error.
func obtainFCMTokenWithADC() (string, error) {
	// 1) Intentar ADC local/SA (gcloud, GOOGLE_APPLICATION_CREDENTIALS, Workload Identity, etc.)
	ctx := context.Background()
	const scope = "https://www.googleapis.com/auth/firebase.messaging"
	creds, err := findDefaultCredentialsFn(ctx, scope)
	if err == nil && creds != nil && creds.TokenSource != nil {
		tok, tErr := tokenSourceTokenFn(creds.TokenSource)
		if tErr == nil && tok != nil && strings.TrimSpace(tok.AccessToken) != "" {
			return tok.AccessToken, nil
		}
	}

	// 2) Fallback: metadata server (GCE/GKE)
	client := newHTTPClientFn(2 * time.Second)
	req, err := httpNewRequestFn(http.MethodGet, metadataTokenURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Metadata-Flavor", "Google")
	resp, err := httpDoFn(client, req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("metadata token status %d", resp.StatusCode)
	}
	var payload struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := jsonNewDecoderFn(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	if strings.TrimSpace(payload.AccessToken) == "" {
		return "", fmt.Errorf("metadata token vacío")
	}
	return payload.AccessToken, nil
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
