package services

import (
	"context"
	"database/sql"
	"testing"

	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/client/orm/clauses/order_clause"
	"github.com/beego/beego/v2/core/utils"
	"github.com/stretchr/testify/assert"
)

// Mocks para PushService
type mockPushOrmer struct {
	readFn       func(interface{}, ...string) error
	insertFn     func(interface{}) (int64, error)
	updateFn     func(interface{}, ...string) (int64, error)
	queryTableFn func(string) orm.QuerySeter
}

func (m *mockPushOrmer) Read(md interface{}, cols ...string) error {
	if m.readFn != nil {
		return m.readFn(md, cols...)
	}
	return orm.ErrNoRows
}

func (m *mockPushOrmer) Insert(md interface{}) (int64, error) {
	if m.insertFn != nil {
		return m.insertFn(md)
	}
	return 1, nil
}

func (m *mockPushOrmer) Update(md interface{}, cols ...string) (int64, error) {
	if m.updateFn != nil {
		return m.updateFn(md, cols...)
	}
	return 1, nil
}

func (m *mockPushOrmer) QueryTable(ptrStructOrTableName interface{}) orm.QuerySeter {
	if m.queryTableFn != nil {
		if tableName, ok := ptrStructOrTableName.(string); ok {
			return m.queryTableFn(tableName)
		}
	}
	return &mockPushQuerySeter{}
}

// Implementar métodos restantes de orm.Ormer
func (m *mockPushOrmer) ReadForUpdate(interface{}, ...string) error { return orm.ErrNoRows }
func (m *mockPushOrmer) ReadOrCreate(interface{}, string, ...string) (bool, int64, error) {
	return false, 0, nil
}
func (m *mockPushOrmer) InsertMulti(int, interface{}) (int64, error)  { return 0, nil }
func (m *mockPushOrmer) Delete(interface{}, ...string) (int64, error) { return 0, nil }
func (m *mockPushOrmer) LoadRelated(md interface{}, name string, args ...utils.KV) (int64, error) {
	return 0, nil
}
func (m *mockPushOrmer) QueryM2M(interface{}, string) orm.QueryM2Mer { return nil }
func (m *mockPushOrmer) Begin() (orm.TxOrmer, error)                 { return nil, nil }
func (m *mockPushOrmer) BeginWithCtx(context.Context) (orm.TxOrmer, error) {
	return nil, nil
}
func (m *mockPushOrmer) BeginWithOpts(opts *sql.TxOptions) (orm.TxOrmer, error) {
	return nil, nil
}
func (m *mockPushOrmer) BeginWithCtxAndOpts(ctx context.Context, opts *sql.TxOptions) (orm.TxOrmer, error) {
	return nil, nil
}
func (m *mockPushOrmer) DoTx(task func(ctx context.Context, txOrm orm.TxOrmer) error) error {
	return nil
}
func (m *mockPushOrmer) DoTxWithCtx(ctx context.Context, task func(ctx context.Context, txOrm orm.TxOrmer) error) error {
	return nil
}
func (m *mockPushOrmer) DoTxWithOpts(opts *sql.TxOptions, task func(ctx context.Context, txOrm orm.TxOrmer) error) error {
	return nil
}
func (m *mockPushOrmer) DoTxWithCtxAndOpts(ctx context.Context, opts *sql.TxOptions, task func(ctx context.Context, txOrm orm.TxOrmer) error) error {
	return nil
}
func (m *mockPushOrmer) Raw(string, ...interface{}) orm.RawSeter { return nil }
func (m *mockPushOrmer) Driver() orm.Driver                      { return nil }
func (m *mockPushOrmer) DBStats() *sql.DBStats                   { return nil }
func (m *mockPushOrmer) ReadWithCtx(context.Context, interface{}, ...string) error {
	return m.Read(nil)
}
func (m *mockPushOrmer) ReadForUpdateWithCtx(context.Context, interface{}, ...string) error {
	return orm.ErrNoRows
}
func (m *mockPushOrmer) ReadOrCreateWithCtx(context.Context, interface{}, string, ...string) (bool, int64, error) {
	return false, 0, nil
}
func (m *mockPushOrmer) InsertWithCtx(context.Context, interface{}) (int64, error) {
	return m.Insert(nil)
}
func (m *mockPushOrmer) InsertMultiWithCtx(context.Context, int, interface{}) (int64, error) {
	return 0, nil
}
func (m *mockPushOrmer) InsertOrUpdate(interface{}, ...string) (int64, error) { return 0, nil }
func (m *mockPushOrmer) InsertOrUpdateWithCtx(context.Context, interface{}, ...string) (int64, error) {
	return 0, nil
}
func (m *mockPushOrmer) UpdateWithCtx(context.Context, interface{}, ...string) (int64, error) {
	return m.Update(nil)
}
func (m *mockPushOrmer) DeleteWithCtx(context.Context, interface{}, ...string) (int64, error) {
	return 0, nil
}
func (m *mockPushOrmer) LoadRelatedWithCtx(ctx context.Context, md interface{}, name string, args ...utils.KV) (int64, error) {
	return 0, nil
}
func (m *mockPushOrmer) QueryM2MWithCtx(context.Context, interface{}, string) orm.QueryM2Mer {
	return nil
}
func (m *mockPushOrmer) QueryTableWithCtx(context.Context, interface{}) orm.QuerySeter {
	return &mockPushQuerySeter{}
}
func (m *mockPushOrmer) RawWithCtx(context.Context, string, ...interface{}) orm.RawSeter {
	return nil
}

type mockPushQuerySeter struct {
	filterFn func(string, ...interface{}) orm.QuerySeter
	oneFn    func(interface{}, ...string) error
	allFn    func(interface{}, ...string) (int64, error)
}

func (m *mockPushQuerySeter) Filter(expr string, args ...interface{}) orm.QuerySeter {
	if m.filterFn != nil {
		return m.filterFn(expr, args...)
	}
	return m
}

func (m *mockPushQuerySeter) One(container interface{}, cols ...string) error {
	if m.oneFn != nil {
		return m.oneFn(container, cols...)
	}
	return orm.ErrNoRows
}

func (m *mockPushQuerySeter) All(container interface{}, cols ...string) (int64, error) {
	if m.allFn != nil {
		return m.allFn(container, cols...)
	}
	return 0, nil
}

// Implementar métodos restantes de orm.QuerySeter
func (m *mockPushQuerySeter) Limit(interface{}, ...interface{}) orm.QuerySeter { return m }
func (m *mockPushQuerySeter) Offset(interface{}) orm.QuerySeter                { return m }
func (m *mockPushQuerySeter) OrderBy(...string) orm.QuerySeter                 { return m }
func (m *mockPushQuerySeter) Distinct() orm.QuerySeter                         { return m }
func (m *mockPushQuerySeter) SetCond(*orm.Condition) orm.QuerySeter            { return m }
func (m *mockPushQuerySeter) GetCond() *orm.Condition                          { return nil }
func (m *mockPushQuerySeter) Exist() bool                                      { return false }
func (m *mockPushQuerySeter) Update(orm.Params) (int64, error)                 { return 0, nil }
func (m *mockPushQuerySeter) Delete() (int64, error)                           { return 0, nil }
func (m *mockPushQuerySeter) PrepareInsert() (orm.Inserter, error)             { return nil, nil }
func (m *mockPushQuerySeter) Values(*[]orm.Params, ...string) (int64, error) {
	return 0, nil
}
func (m *mockPushQuerySeter) ValuesList(*[]orm.ParamsList, ...string) (int64, error) {
	return 0, nil
}
func (m *mockPushQuerySeter) ValuesFlat(*orm.ParamsList, string) (int64, error) {
	return 0, nil
}
func (m *mockPushQuerySeter) RowsToMap(*orm.Params, string, string) (int64, error) {
	return 0, nil
}
func (m *mockPushQuerySeter) RowsToStruct(interface{}, string, string) (int64, error) {
	return 0, nil
}
func (m *mockPushQuerySeter) GroupBy(...string) orm.QuerySeter                   { return m }
func (m *mockPushQuerySeter) ForUpdate() orm.QuerySeter                          { return m }
func (m *mockPushQuerySeter) Aggregate(string) orm.QuerySeter                    { return m }
func (m *mockPushQuerySeter) FilterRaw(string, string) orm.QuerySeter            { return m }
func (m *mockPushQuerySeter) Exclude(string, ...interface{}) orm.QuerySeter      { return m }
func (m *mockPushQuerySeter) OrderClauses(...*order_clause.Order) orm.QuerySeter { return m }
func (m *mockPushQuerySeter) ForceIndex(...string) orm.QuerySeter                { return m }
func (m *mockPushQuerySeter) UseIndex(...string) orm.QuerySeter                  { return m }
func (m *mockPushQuerySeter) IgnoreIndex(...string) orm.QuerySeter               { return m }
func (m *mockPushQuerySeter) RelatedSel(...interface{}) orm.QuerySeter           { return m }
func (m *mockPushQuerySeter) Count() (int64, error)                              { return 0, nil }

// Métodos con contexto
func (m *mockPushQuerySeter) OneWithCtx(context.Context, interface{}, ...string) error {
	return m.One(nil)
}
func (m *mockPushQuerySeter) AllWithCtx(ctx context.Context, container interface{}, cols ...string) (int64, error) {
	return m.All(container, cols...)
}
func (m *mockPushQuerySeter) ValuesWithCtx(context.Context, *[]orm.Params, ...string) (int64, error) {
	return 0, nil
}
func (m *mockPushQuerySeter) ValuesListWithCtx(context.Context, *[]orm.ParamsList, ...string) (int64, error) {
	return 0, nil
}
func (m *mockPushQuerySeter) ValuesFlatWithCtx(context.Context, *orm.ParamsList, string) (int64, error) {
	return 0, nil
}
func (m *mockPushQuerySeter) CountWithCtx(ctx context.Context) (int64, error) {
	return m.Count()
}
func (m *mockPushQuerySeter) ExistWithCtx(context.Context) bool { return false }
func (m *mockPushQuerySeter) UpdateWithCtx(context.Context, orm.Params) (int64, error) {
	return 0, nil
}
func (m *mockPushQuerySeter) DeleteWithCtx(context.Context) (int64, error) { return 0, nil }
func (m *mockPushQuerySeter) PrepareInsertWithCtx(context.Context) (orm.Inserter, error) {
	return nil, nil
}

// Tests para ActualizarUltimaVista

func TestPushService_ActualizarUltimaVista_Success(t *testing.T) {
	service := NewPushService(&mockPushOrmer{
		readFn: func(md interface{}, cols ...string) error {
			if disp, ok := md.(*models.PushDispositivo); ok {
				disp.PkIdPushDispositivo = 1
				disp.Enabled = true
			}
			return nil
		},
		updateFn: func(md interface{}, cols ...string) (int64, error) {
			return 1, nil
		},
	})

	err := service.ActualizarUltimaVista(context.Background(), 1)
	assert.NoError(t, err)
}

func TestPushService_ActualizarUltimaVista_DispositivoNoEncontrado(t *testing.T) {
	service := NewPushService(&mockPushOrmer{
		readFn: func(md interface{}, cols ...string) error {
			return orm.ErrNoRows
		},
	})

	err := service.ActualizarUltimaVista(context.Background(), 999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "dispositivo no encontrado")
}

func TestPushService_ActualizarUltimaVista_ErrorLectura(t *testing.T) {
	service := NewPushService(&mockPushOrmer{
		readFn: func(md interface{}, cols ...string) error {
			return assert.AnError
		},
	})

	err := service.ActualizarUltimaVista(context.Background(), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error al buscar dispositivo")
}

func TestPushService_ActualizarUltimaVista_ErrorActualizacion(t *testing.T) {
	service := NewPushService(&mockPushOrmer{
		readFn: func(md interface{}, cols ...string) error {
			return nil
		},
		updateFn: func(md interface{}, cols ...string) (int64, error) {
			return 0, assert.AnError
		},
	})

	err := service.ActualizarUltimaVista(context.Background(), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error al actualizar última vista")
}

// Tests para ActualizarEstadoDispositivo

func TestPushService_ActualizarEstadoDispositivo_Habilitar(t *testing.T) {
	service := NewPushService(&mockPushOrmer{
		readFn: func(md interface{}, cols ...string) error {
			return nil
		},
		updateFn: func(md interface{}, cols ...string) (int64, error) {
			disp := md.(*models.PushDispositivo)
			assert.True(t, disp.Enabled)
			return 1, nil
		},
	})

	err := service.ActualizarEstadoDispositivo(context.Background(), 1, true)
	assert.NoError(t, err)
}

func TestPushService_ActualizarEstadoDispositivo_Deshabilitar(t *testing.T) {
	service := NewPushService(&mockPushOrmer{
		readFn: func(md interface{}, cols ...string) error {
			return nil
		},
		updateFn: func(md interface{}, cols ...string) (int64, error) {
			disp := md.(*models.PushDispositivo)
			assert.False(t, disp.Enabled)
			return 1, nil
		},
	})

	err := service.ActualizarEstadoDispositivo(context.Background(), 1, false)
	assert.NoError(t, err)
}

func TestPushService_ActualizarEstadoDispositivo_DispositivoNoEncontrado(t *testing.T) {
	service := NewPushService(&mockPushOrmer{
		readFn: func(md interface{}, cols ...string) error {
			return orm.ErrNoRows
		},
	})

	err := service.ActualizarEstadoDispositivo(context.Background(), 999, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "dispositivo no encontrado")
}

func TestPushService_ActualizarEstadoDispositivo_ErrorActualizacion(t *testing.T) {
	service := NewPushService(&mockPushOrmer{
		readFn: func(md interface{}, cols ...string) error {
			return nil
		},
		updateFn: func(md interface{}, cols ...string) (int64, error) {
			return 0, assert.AnError
		},
	})

	err := service.ActualizarEstadoDispositivo(context.Background(), 1, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error al actualizar estado del dispositivo")
}

// Tests para ActualizarTopicsDispositivo

func TestPushService_ActualizarTopicsDispositivo_Success(t *testing.T) {
	service := NewPushService(&mockPushOrmer{
		readFn: func(md interface{}, cols ...string) error {
			return nil
		},
		updateFn: func(md interface{}, cols ...string) (int64, error) {
			disp := md.(*models.PushDispositivo)
			assert.ElementsMatch(t, []string{"ofertas", "noticias"}, disp.SubscribedTopicsArray)
			return 1, nil
		},
	})

	err := service.ActualizarTopicsDispositivo(context.Background(), 1, []string{"ofertas", "noticias"})
	assert.NoError(t, err)
}

func TestPushService_ActualizarTopicsDispositivo_DispositivoNoEncontrado(t *testing.T) {
	service := NewPushService(&mockPushOrmer{
		readFn: func(md interface{}, cols ...string) error {
			return orm.ErrNoRows
		},
	})

	err := service.ActualizarTopicsDispositivo(context.Background(), 999, []string{"ofertas"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "dispositivo no encontrado")
}

func TestPushService_ActualizarTopicsDispositivo_ErrorActualizacion(t *testing.T) {
	service := NewPushService(&mockPushOrmer{
		readFn: func(md interface{}, cols ...string) error {
			return nil
		},
		updateFn: func(md interface{}, cols ...string) (int64, error) {
			return 0, assert.AnError
		},
	})

	err := service.ActualizarTopicsDispositivo(context.Background(), 1, []string{"ofertas"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error al actualizar topics del dispositivo")
}

// Tests para RegistrarEnvio

func TestPushService_RegistrarEnvio_Success(t *testing.T) {
	service := NewPushService(&mockPushOrmer{
		readFn: func(md interface{}, cols ...string) error {
			if disp, ok := md.(*models.PushDispositivo); ok {
				disp.PkIdPushDispositivo = 1
			}
			return nil
		},
		insertFn: func(md interface{}) (int64, error) {
			return 1, nil
		},
	})

	statusCode := 200
	req := &models.RegistrarEnvioRequest{
		PkIdPushDispositivo: 1,
		Proveedor:           models.ProveedorWebPush,
		Exito:               true,
		StatusCode:          &statusCode,
	}

	envio, err := service.RegistrarEnvio(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, envio)
	assert.True(t, envio.Exito)
}

func TestPushService_RegistrarEnvio_DispositivoNoEncontrado(t *testing.T) {
	service := NewPushService(&mockPushOrmer{
		readFn: func(md interface{}, cols ...string) error {
			return orm.ErrNoRows
		},
	})

	req := &models.RegistrarEnvioRequest{
		PkIdPushDispositivo: 999,
		Proveedor:           models.ProveedorWebPush,
	}

	envio, err := service.RegistrarEnvio(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, envio)
	assert.Contains(t, err.Error(), "dispositivo no encontrado")
}

func TestPushService_RegistrarEnvio_ProveedorInvalido(t *testing.T) {
	service := NewPushService(&mockPushOrmer{
		readFn: func(md interface{}, cols ...string) error {
			return nil
		},
	})

	req := &models.RegistrarEnvioRequest{
		PkIdPushDispositivo: 1,
		Proveedor:           "INVALIDO",
	}

	envio, err := service.RegistrarEnvio(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, envio)
	assert.Contains(t, err.Error(), "proveedor no válido")
}

func TestPushService_RegistrarEnvio_ErrorInsercion(t *testing.T) {
	service := NewPushService(&mockPushOrmer{
		readFn: func(md interface{}, cols ...string) error {
			return nil
		},
		insertFn: func(md interface{}) (int64, error) {
			return 0, assert.AnError
		},
	})

	req := &models.RegistrarEnvioRequest{
		PkIdPushDispositivo: 1,
		Proveedor:           models.ProveedorWebPush,
	}

	envio, err := service.RegistrarEnvio(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, envio)
	assert.Contains(t, err.Error(), "error al registrar envío")
}

// Tests para obtenerProveedor - Caso default

func TestPushService_ObtenerProveedor_Default(t *testing.T) {
	service := &PushService{}

	proveedor := service.obtenerProveedor("PLATAFORMA_INVALIDA")
	assert.Equal(t, models.ProveedorWebPush, proveedor) // Default
}

// Tests para obtenerDispositivosDestinatarios

func TestPushService_ObtenerDispositivosDestinatarios_Todos(t *testing.T) {
	service := NewPushService(&mockPushOrmer{
		queryTableFn: func(tableName string) orm.QuerySeter {
			return &mockPushQuerySeter{
				filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
					return &mockPushQuerySeter{
						allFn: func(container interface{}, cols ...string) (int64, error) {
							if disps, ok := container.(*[]models.PushDispositivo); ok {
								*disps = []models.PushDispositivo{
									{PkIdPushDispositivo: 1},
									{PkIdPushDispositivo: 2},
								}
							}
							return 2, nil
						},
					}
				},
			}
		},
	})

	destinatarios := &models.DestinatariosNotificacion{
		Tipo: models.DestinatarioTodos,
	}

	dispositivos, err := service.obtenerDispositivosDestinatarios(destinatarios)
	assert.NoError(t, err)
	assert.Len(t, dispositivos, 2)
}

func TestPushService_ObtenerDispositivosDestinatarios_Cliente_SinDocumento(t *testing.T) {
	service := NewPushService(&mockPushOrmer{})

	destinatarios := &models.DestinatariosNotificacion{
		Tipo:             models.DestinatarioCliente,
		DocumentoCliente: nil, // Falta documento
	}

	dispositivos, err := service.obtenerDispositivosDestinatarios(destinatarios)
	assert.Error(t, err)
	assert.Nil(t, dispositivos)
	assert.Contains(t, err.Error(), "documentoCliente es requerido")
}

func TestPushService_ObtenerDispositivosDestinatarios_Trabajador_SinDocumento(t *testing.T) {
	service := NewPushService(&mockPushOrmer{})

	destinatarios := &models.DestinatariosNotificacion{
		Tipo:                models.DestinatarioTrabajador,
		DocumentoTrabajador: nil, // Falta documento
	}

	dispositivos, err := service.obtenerDispositivosDestinatarios(destinatarios)
	assert.Error(t, err)
	assert.Nil(t, dispositivos)
	assert.Contains(t, err.Error(), "documentoTrabajador es requerido")
}

func TestPushService_ObtenerDispositivosDestinatarios_Topic_SinTopic(t *testing.T) {
	service := NewPushService(&mockPushOrmer{})

	destinatarios := &models.DestinatariosNotificacion{
		Tipo:  models.DestinatarioTopic,
		Topic: nil, // Falta topic
	}

	dispositivos, err := service.obtenerDispositivosDestinatarios(destinatarios)
	assert.Error(t, err)
	assert.Nil(t, dispositivos)
	assert.Contains(t, err.Error(), "topic es requerido")
}

func TestPushService_ObtenerDispositivosDestinatarios_TipoInvalido(t *testing.T) {
	service := NewPushService(&mockPushOrmer{})

	destinatarios := &models.DestinatariosNotificacion{
		Tipo: "INVALIDO",
	}

	dispositivos, err := service.obtenerDispositivosDestinatarios(destinatarios)
	assert.Error(t, err)
	assert.Nil(t, dispositivos)
	assert.Contains(t, err.Error(), "tipo de destinatario no válido")
}

// Tests para validarRemitente

func TestPushService_ValidarRemitente_SistemaValido(t *testing.T) {
	service := NewPushService(&mockPushOrmer{})

	remitente := &models.RemitenteNotificacion{
		Tipo: models.RemitenteSistema,
	}

	err := service.validarRemitente(remitente)
	assert.NoError(t, err)
}

func TestPushService_ValidarRemitente_TrabajadorSinDocumento(t *testing.T) {
	service := NewPushService(&mockPushOrmer{})

	remitente := &models.RemitenteNotificacion{
		Tipo:                models.RemitenteTrabajador,
		DocumentoTrabajador: nil, // Falta documento
	}

	err := service.validarRemitente(remitente)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "documentoTrabajador es requerido")
}

func TestPushService_ValidarRemitente_TrabajadorNoEncontrado(t *testing.T) {
	service := NewPushService(&mockPushOrmer{
		queryTableFn: func(tableName string) orm.QuerySeter {
			return &mockPushQuerySeter{
				filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
					return &mockPushQuerySeter{
						oneFn: func(container interface{}, cols ...string) error {
							return orm.ErrNoRows
						},
					}
				},
			}
		},
	})

	doc := int64(123)
	remitente := &models.RemitenteNotificacion{
		Tipo:                models.RemitenteTrabajador,
		DocumentoTrabajador: &doc,
	}

	err := service.validarRemitente(remitente)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "trabajador no encontrado")
}

func TestPushService_ValidarRemitente_TrabajadorValido(t *testing.T) {
	service := NewPushService(&mockPushOrmer{
		queryTableFn: func(tableName string) orm.QuerySeter {
			return &mockPushQuerySeter{
				filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
					return &mockPushQuerySeter{
						oneFn: func(container interface{}, cols ...string) error {
							// Trabajador encontrado
							return nil
						},
					}
				},
			}
		},
	})

	doc := int64(123)
	remitente := &models.RemitenteNotificacion{
		Tipo:                models.RemitenteTrabajador,
		DocumentoTrabajador: &doc,
	}

	err := service.validarRemitente(remitente)
	assert.NoError(t, err)
}

// Tests para enviarFCM - Casos de error básicos

func TestPushService_EnviarFCM_TokenVacio(t *testing.T) {
	service := &PushService{}

	fcmToken := ""
	dispositivo := &models.PushDispositivo{
		PkIdPushDispositivo: 1,
		Plataforma:          models.PlataformaAndroid,
		FcmToken:            &fcmToken,
	}

	notificacion := &models.ContenidoNotificacion{
		Titulo:  "Test",
		Mensaje: "Test message",
	}

	exito, statusCode, errorCode := service.enviarFCM(dispositivo, notificacion)
	assert.False(t, exito)
	assert.NotNil(t, statusCode)
	assert.NotNil(t, errorCode)
	assert.Equal(t, 400, *statusCode)
	assert.Equal(t, "FCM_TOKEN_VACIO", *errorCode)
}

func TestPushService_EnviarFCM_TokenNil(t *testing.T) {
	service := &PushService{}

	dispositivo := &models.PushDispositivo{
		PkIdPushDispositivo: 1,
		Plataforma:          models.PlataformaAndroid,
		FcmToken:            nil,
	}

	notificacion := &models.ContenidoNotificacion{
		Titulo:  "Test",
		Mensaje: "Test message",
	}

	exito, statusCode, errorCode := service.enviarFCM(dispositivo, notificacion)
	assert.False(t, exito)
	assert.NotNil(t, statusCode)
	assert.Equal(t, 400, *statusCode)
	assert.Equal(t, "FCM_TOKEN_VACIO", *errorCode)
}

// Tests para RegistrarDispositivo

func TestPushService_RegistrarDispositivo_ErrorValidacion(t *testing.T) {
	service := NewPushService(&mockPushOrmer{})

	// Request sin cliente ni trabajador
	req := &models.RegistrarDispositivoRequest{
		Plataforma: models.PlataformaWeb,
	}

	dispositivo, err := service.RegistrarDispositivo(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, dispositivo)
	assert.Contains(t, err.Error(), "exactamente uno")
}

func TestPushService_RegistrarDispositivo_NuevoDispositivo(t *testing.T) {
	service := NewPushService(&mockPushOrmer{
		queryTableFn: func(tableName string) orm.QuerySeter {
			return &mockPushQuerySeter{
				filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
					return &mockPushQuerySeter{
						oneFn: func(container interface{}, cols ...string) error {
							return orm.ErrNoRows // No existe
						},
					}
				},
			}
		},
		insertFn: func(md interface{}) (int64, error) {
			return 1, nil
		},
	})

	endpoint := "https://push.example.com"
	p256dh := "test_p256dh"
	auth := "test_auth"
	clienteId := int64(123)

	req := &models.RegistrarDispositivoRequest{
		Plataforma:         models.PlataformaWeb,
		Endpoint:           &endpoint,
		P256dh:             &p256dh,
		Auth:               &auth,
		PkDocumentoCliente: &clienteId,
	}

	dispositivo, err := service.RegistrarDispositivo(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, dispositivo)
	assert.Equal(t, models.PlataformaWeb, dispositivo.Plataforma)
}

func TestPushService_RegistrarDispositivo_ActualizarExistentePorFcmToken(t *testing.T) {
	fcmToken := "fcm_token_test"

	service := NewPushService(&mockPushOrmer{
		queryTableFn: func(tableName string) orm.QuerySeter {
			return &mockPushQuerySeter{
				filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
					return &mockPushQuerySeter{
						oneFn: func(container interface{}, cols ...string) error {
							if disp, ok := container.(*models.PushDispositivo); ok {
								disp.PkIdPushDispositivo = 1
								disp.FcmToken = &fcmToken
								disp.Enabled = false
							}
							return nil // Dispositivo encontrado
						},
					}
				},
			}
		},
		updateFn: func(md interface{}, cols ...string) (int64, error) {
			return 1, nil
		},
	})

	trabajadorId := int64(456)

	req := &models.RegistrarDispositivoRequest{
		Plataforma:            models.PlataformaAndroid,
		FcmToken:              &fcmToken,
		PkDocumentoTrabajador: &trabajadorId,
	}

	dispositivo, err := service.RegistrarDispositivo(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, dispositivo)
	assert.Equal(t, int64(1), dispositivo.PkIdPushDispositivo)
	assert.True(t, dispositivo.Enabled) // Debe reactivarse
}

func TestPushService_RegistrarDispositivo_ActualizarExistentePorEndpoint(t *testing.T) {
	endpoint := "https://push.example.com"

	service := NewPushService(&mockPushOrmer{
		queryTableFn: func(tableName string) orm.QuerySeter {
			return &mockPushQuerySeter{
				filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
					return &mockPushQuerySeter{
						oneFn: func(container interface{}, cols ...string) error {
							if disp, ok := container.(*models.PushDispositivo); ok {
								disp.PkIdPushDispositivo = 2
								disp.Endpoint = &endpoint
							}
							return nil // Dispositivo encontrado
						},
					}
				},
			}
		},
		updateFn: func(md interface{}, cols ...string) (int64, error) {
			return 1, nil
		},
	})

	p256dh := "test_p256dh"
	auth := "test_auth"
	clienteId := int64(123)

	req := &models.RegistrarDispositivoRequest{
		Plataforma:         models.PlataformaWeb,
		Endpoint:           &endpoint,
		P256dh:             &p256dh,
		Auth:               &auth,
		PkDocumentoCliente: &clienteId,
	}

	dispositivo, err := service.RegistrarDispositivo(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, dispositivo)
	assert.Equal(t, int64(2), dispositivo.PkIdPushDispositivo)
}

func TestPushService_RegistrarDispositivo_ErrorUpdate(t *testing.T) {
	fcmToken := "fcm_token_test"

	service := NewPushService(&mockPushOrmer{
		queryTableFn: func(tableName string) orm.QuerySeter {
			return &mockPushQuerySeter{
				filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
					return &mockPushQuerySeter{
						oneFn: func(container interface{}, cols ...string) error {
							if disp, ok := container.(*models.PushDispositivo); ok {
								disp.PkIdPushDispositivo = 1
								disp.FcmToken = &fcmToken
							}
							return nil
						},
					}
				},
			}
		},
		updateFn: func(md interface{}, cols ...string) (int64, error) {
			return 0, assert.AnError
		},
	})

	trabajadorId := int64(456)

	req := &models.RegistrarDispositivoRequest{
		Plataforma:            models.PlataformaAndroid,
		FcmToken:              &fcmToken,
		PkDocumentoTrabajador: &trabajadorId,
	}

	dispositivo, err := service.RegistrarDispositivo(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, dispositivo)
	assert.Contains(t, err.Error(), "error al actualizar dispositivo")
}

func TestPushService_RegistrarDispositivo_ErrorInsert(t *testing.T) {
	service := NewPushService(&mockPushOrmer{
		queryTableFn: func(tableName string) orm.QuerySeter {
			return &mockPushQuerySeter{
				filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
					return &mockPushQuerySeter{
						oneFn: func(container interface{}, cols ...string) error {
							return orm.ErrNoRows // No existe
						},
					}
				},
			}
		},
		insertFn: func(md interface{}) (int64, error) {
			return 0, assert.AnError
		},
	})

	endpoint := "https://push.example.com"
	p256dh := "test_p256dh"
	auth := "test_auth"
	clienteId := int64(123)

	req := &models.RegistrarDispositivoRequest{
		Plataforma:         models.PlataformaWeb,
		Endpoint:           &endpoint,
		P256dh:             &p256dh,
		Auth:               &auth,
		PkDocumentoCliente: &clienteId,
	}

	dispositivo, err := service.RegistrarDispositivo(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, dispositivo)
	assert.Contains(t, err.Error(), "error al registrar dispositivo")
}

// Tests para obtenerDispositivosDestinatarios - Casos faltantes

func TestPushService_ObtenerDispositivosDestinatarios_Cliente(t *testing.T) {
	clienteId := int64(123)

	service := NewPushService(&mockPushOrmer{
		queryTableFn: func(tableName string) orm.QuerySeter {
			return &mockPushQuerySeter{
				filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
					return &mockPushQuerySeter{
						allFn: func(container interface{}, cols ...string) (int64, error) {
							if disps, ok := container.(*[]models.PushDispositivo); ok {
								*disps = []models.PushDispositivo{
									{PkIdPushDispositivo: 1},
								}
							}
							return 1, nil
						},
					}
				},
			}
		},
	})

	destinatarios := &models.DestinatariosNotificacion{
		Tipo:             models.DestinatarioCliente,
		DocumentoCliente: &clienteId,
	}

	dispositivos, err := service.obtenerDispositivosDestinatarios(destinatarios)
	assert.NoError(t, err)
	assert.Len(t, dispositivos, 1)
}

func TestPushService_ObtenerDispositivosDestinatarios_Clientes(t *testing.T) {
	service := NewPushService(&mockPushOrmer{
		queryTableFn: func(tableName string) orm.QuerySeter {
			return &mockPushQuerySeter{
				filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
					return &mockPushQuerySeter{
						allFn: func(container interface{}, cols ...string) (int64, error) {
							if disps, ok := container.(*[]models.PushDispositivo); ok {
								*disps = []models.PushDispositivo{
									{PkIdPushDispositivo: 1},
									{PkIdPushDispositivo: 2},
								}
							}
							return 2, nil
						},
					}
				},
			}
		},
	})

	destinatarios := &models.DestinatariosNotificacion{
		Tipo: models.DestinatarioClientes,
	}

	dispositivos, err := service.obtenerDispositivosDestinatarios(destinatarios)
	assert.NoError(t, err)
	assert.Len(t, dispositivos, 2)
}

func TestPushService_ObtenerDispositivosDestinatarios_Trabajador(t *testing.T) {
	trabajadorId := int64(456)

	service := NewPushService(&mockPushOrmer{
		queryTableFn: func(tableName string) orm.QuerySeter {
			return &mockPushQuerySeter{
				filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
					return &mockPushQuerySeter{
						allFn: func(container interface{}, cols ...string) (int64, error) {
							if disps, ok := container.(*[]models.PushDispositivo); ok {
								*disps = []models.PushDispositivo{
									{PkIdPushDispositivo: 1},
								}
							}
							return 1, nil
						},
					}
				},
			}
		},
	})

	destinatarios := &models.DestinatariosNotificacion{
		Tipo:                models.DestinatarioTrabajador,
		DocumentoTrabajador: &trabajadorId,
	}

	dispositivos, err := service.obtenerDispositivosDestinatarios(destinatarios)
	assert.NoError(t, err)
	assert.Len(t, dispositivos, 1)
}

func TestPushService_ObtenerDispositivosDestinatarios_Trabajadores(t *testing.T) {
	service := NewPushService(&mockPushOrmer{
		queryTableFn: func(tableName string) orm.QuerySeter {
			return &mockPushQuerySeter{
				filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
					return &mockPushQuerySeter{
						allFn: func(container interface{}, cols ...string) (int64, error) {
							if disps, ok := container.(*[]models.PushDispositivo); ok {
								*disps = []models.PushDispositivo{
									{PkIdPushDispositivo: 1},
								}
							}
							return 1, nil
						},
					}
				},
			}
		},
	})

	destinatarios := &models.DestinatariosNotificacion{
		Tipo: models.DestinatarioTrabajadores,
	}

	dispositivos, err := service.obtenerDispositivosDestinatarios(destinatarios)
	assert.NoError(t, err)
	assert.Len(t, dispositivos, 1)
}

func TestPushService_ObtenerDispositivosDestinatarios_Topic(t *testing.T) {
	topic := "ofertas"

	service := NewPushService(&mockPushOrmer{
		queryTableFn: func(tableName string) orm.QuerySeter {
			return &mockPushQuerySeter{
				filterFn: func(expr string, args ...interface{}) orm.QuerySeter {
					return &mockPushQuerySeter{
						allFn: func(container interface{}, cols ...string) (int64, error) {
							if disps, ok := container.(*[]models.PushDispositivo); ok {
								*disps = []models.PushDispositivo{
									{PkIdPushDispositivo: 1},
								}
							}
							return 1, nil
						},
					}
				},
			}
		},
	})

	destinatarios := &models.DestinatariosNotificacion{
		Tipo:  models.DestinatarioTopic,
		Topic: &topic,
	}

	dispositivos, err := service.obtenerDispositivosDestinatarios(destinatarios)
	assert.NoError(t, err)
	assert.Len(t, dispositivos, 1)
}

// Tests para crearResumenDestinatarios

func TestPushService_CrearResumenDestinatarios_Clientes(t *testing.T) {
	service := &PushService{}

	destinatarios := &models.DestinatariosNotificacion{
		Tipo: models.DestinatarioClientes,
	}

	cliente1 := models.Cliente{PK_DOCUMENTO_CLIENTE: 123}
	cliente2 := models.Cliente{PK_DOCUMENTO_CLIENTE: 456}

	dispositivos := []models.PushDispositivo{
		{PkIdPushDispositivo: 1, PkDocumentoCliente: &cliente1},
		{PkIdPushDispositivo: 2, PkDocumentoCliente: &cliente2},
		{PkIdPushDispositivo: 3, PkDocumentoCliente: &cliente1}, // Duplicado
	}

	resumen := service.crearResumenDestinatarios(destinatarios, dispositivos)
	assert.Equal(t, string(models.DestinatarioClientes), resumen.TipoDestinatario)
	assert.Len(t, resumen.ClientesNotificados, 2) // Sin duplicados
}

func TestPushService_CrearResumenDestinatarios_Trabajadores(t *testing.T) {
	service := &PushService{}

	destinatarios := &models.DestinatariosNotificacion{
		Tipo: models.DestinatarioTrabajadores,
	}

	trabajador1 := models.Trabajador{PK_DOCUMENTO_TRABAJADOR: 789}
	trabajador2 := models.Trabajador{PK_DOCUMENTO_TRABAJADOR: 101}

	dispositivos := []models.PushDispositivo{
		{PkIdPushDispositivo: 1, PkDocumentoTrabajador: &trabajador1},
		{PkIdPushDispositivo: 2, PkDocumentoTrabajador: &trabajador2},
	}

	resumen := service.crearResumenDestinatarios(destinatarios, dispositivos)
	assert.Equal(t, string(models.DestinatarioTrabajadores), resumen.TipoDestinatario)
	assert.Len(t, resumen.TrabajadoresNotificados, 2)
}

func TestPushService_CrearResumenDestinatarios_Topic(t *testing.T) {
	service := &PushService{}

	topic := "ofertas"
	destinatarios := &models.DestinatariosNotificacion{
		Tipo:  models.DestinatarioTopic,
		Topic: &topic,
	}

	dispositivos := []models.PushDispositivo{
		{PkIdPushDispositivo: 1},
		{PkIdPushDispositivo: 2},
	}

	resumen := service.crearResumenDestinatarios(destinatarios, dispositivos)
	assert.Equal(t, string(models.DestinatarioTopic), resumen.TipoDestinatario)
	assert.Len(t, resumen.TopicsNotificados, 1)
	assert.Equal(t, "ofertas", resumen.TopicsNotificados[0])
}

// Tests para registrarEnvioNotificacion

func TestPushService_RegistrarEnvioNotificacion_Success(t *testing.T) {
	service := NewPushService(&mockPushOrmer{
		insertFn: func(md interface{}) (int64, error) {
			envio := md.(*models.PushEnvio)
			assert.NotNil(t, envio)
			assert.True(t, envio.Exito)
			return 1, nil
		},
	})

	dispositivo := &models.PushDispositivo{
		PkIdPushDispositivo: 1,
		Plataforma:          models.PlataformaWeb,
	}

	notificacion := &models.ContenidoNotificacion{
		Titulo:  "Test",
		Mensaje: "Test message",
	}

	statusCode := 200
	service.registrarEnvioNotificacion(dispositivo, notificacion, true, &statusCode, nil)

	// No hay assert porque la función no retorna error, pero valida que no panic
}

func TestPushService_RegistrarEnvioNotificacion_ConDatos(t *testing.T) {
	service := NewPushService(&mockPushOrmer{
		insertFn: func(md interface{}) (int64, error) {
			envio := md.(*models.PushEnvio)
			assert.NotNil(t, envio.DataObj)
			return 1, nil
		},
	})

	dispositivo := &models.PushDispositivo{
		PkIdPushDispositivo: 1,
		Plataforma:          models.PlataformaAndroid,
	}

	notificacion := &models.ContenidoNotificacion{
		Titulo:  "Oferta",
		Mensaje: "Nueva oferta",
		Datos:   []byte(`{"tipo":"OFERTA"}`),
	}

	statusCode := 200
	service.registrarEnvioNotificacion(dispositivo, notificacion, true, &statusCode, nil)
}

// Tests para enviarNotificacionDispositivo

func TestPushService_EnviarNotificacionDispositivo_PlataformaNoSoportada(t *testing.T) {
	service := &PushService{ormer: &mockPushOrmer{}}

	dispositivo := &models.PushDispositivo{
		PkIdPushDispositivo: 1,
		Plataforma:          "PLATAFORMA_INVALIDA",
	}

	notificacion := &models.ContenidoNotificacion{
		Titulo:  "Test",
		Mensaje: "Test message",
	}

	detalle := service.enviarNotificacionDispositivo(dispositivo, notificacion)
	assert.False(t, detalle.Exito)
	assert.NotNil(t, detalle.StatusCode)
	assert.NotNil(t, detalle.ErrorCode)
	assert.Equal(t, 400, *detalle.StatusCode)
	assert.Equal(t, "PLATAFORMA_NO_SOPORTADA", *detalle.ErrorCode)
}
