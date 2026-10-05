package reservacontacto

import (
	"errors"
	"net/http"

	"restaurante/internal/httpx"
	"restaurante/logging"
	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
)

// ReservaContactoController expone los contactos de reserva (invitados o
// clientes registrados). Las respuestas siempre llevan el envoltorio
// models.ApiResponse y el status HTTP coincide con su `code`.
type ReservaContactoController struct{ web.Controller }

// optionalPositive lee un query param opcional entero positivo: ok=false si no vino.
func (c *ReservaContactoController) optionalPositive(key string) (v int64, ok bool, err error) {
	if c.GetString(key) == "" {
		return 0, false, nil
	}
	v, err = httpx.PositiveInt64Param(&c.Controller, key)
	return v, err == nil, err
}

// @Title GetAll
// @Summary Listar contactos de reserva
// @Description Devuelve los contactos de reserva, opcionalmente filtrados por documento de invitado y/o de cliente registrado. `documentoCliente` se responde como objeto `{documentoCliente}` (solo el documento; nunca datos del cliente ni contraseña). Sin resultados: 200 con `data: []`. Solo personal (trabajadores/administrador): contiene datos personales.
// @Tags reserva_contacto
// @Accept json
// @Produce json
// @Param documento_contacto query int false "Documento del contacto invitado (entero positivo)"
// @Param documento_cliente query int false "Documento del cliente registrado (entero positivo)"
// @Success 200 {object} models.ApiResponse{data=[]models.ReservaContactoResponse} "Lista de contactos (puede ser [])"
// @Failure 400 {object} models.ApiResponse "documento_contacto o documento_cliente inválidos"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 403 {object} models.ApiResponse "El token no es de un trabajador"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /reserva_contacto [get]
func (c *ReservaContactoController) GetAll() {
	qs := orm.NewOrm().QueryTable(new(models.ReservaContacto))
	v, ok, err := c.optionalPositive("documento_contacto")
	if err != nil {
		httpx.Fail(&c.Controller, http.StatusBadRequest, "El parámetro 'documento_contacto' es inválido", err)
		return
	}
	if ok {
		qs = qs.Filter("DocumentoContacto", v)
	}
	v, ok, err = c.optionalPositive("documento_cliente")
	if err != nil {
		httpx.Fail(&c.Controller, http.StatusBadRequest, "El parámetro 'documento_cliente' es inválido", err)
		return
	}
	if ok {
		qs = qs.Filter("PKDocumentoCliente", v)
	}
	var list []models.ReservaContacto
	if _, err := qs.All(&list); err != nil {
		logging.LogControllerError(c.Ctx, "reserva_contacto.getall.db_error", err, nil)
		httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error al obtener contactos", err)
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, "Contactos obtenidos", httpx.List(list))
}

// @Title GetById
// @Summary Obtener contacto por ID
// @Description Devuelve un contacto de reserva por su ID. Solo personal (trabajadores/administrador): contiene datos personales.
// @Tags reserva_contacto
// @Accept json
// @Produce json
// @Param id query int true "ID del contacto (entero positivo)"
// @Success 200 {object} models.ApiResponse{data=models.ReservaContactoResponse} "Contacto encontrado"
// @Failure 400 {object} models.ApiResponse "id ausente o inválido"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 403 {object} models.ApiResponse "El token no es de un trabajador"
// @Failure 404 {object} models.ApiResponse "Contacto no encontrado"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /reserva_contacto/search [get]
func (c *ReservaContactoController) GetById() {
	id, err := httpx.PositiveInt64Param(&c.Controller, "id")
	if err != nil {
		httpx.Fail(&c.Controller, http.StatusBadRequest, "El parámetro 'id' es inválido o está ausente", err)
		return
	}
	var row models.ReservaContacto
	if err := orm.NewOrm().QueryTable(new(models.ReservaContacto)).Filter("PKIDContacto", id).One(&row); err != nil {
		if errors.Is(err, orm.ErrNoRows) {
			httpx.Fail(&c.Controller, http.StatusNotFound, "Contacto no encontrado", err)
			return
		}
		logging.LogControllerError(c.Ctx, "reserva_contacto.getbyid.db_error", err, map[string]interface{}{"id": id})
		httpx.Fail(&c.Controller, http.StatusInternalServerError, "Error al obtener el contacto", err)
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, "Contacto encontrado", row)
}
