package cliente

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/mail"
	"strconv"
	"strings"

	"restaurante/internal/dberr"
	"restaurante/internal/httpx"
	"restaurante/logging"
	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
	"golang.org/x/crypto/bcrypt"
)

// ClienteController gestiona /clientes. POST es público (registro); el resto
// exige un access token válido (filtro loginc.ValidateToken en el router).
type ClienteController struct {
	web.Controller
}

const (
	// fieldsResumen es el único valor admitido para el parámetro `fields`.
	fieldsResumen = "nombre_completo_telefono"
	// maxLimit es el máximo de registros por página en GET /clientes.
	maxLimit = 100
	// maxPasswordBytes es el límite de bcrypt; más largo no se puede hashear.
	maxPasswordBytes = 72
)

var generateFromPassword = bcrypt.GenerateFromPassword

func hashPassword(password string) (string, error) {
	h, err := generateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(h), nil
}

func normalizeEmail(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// validEmail acepta únicamente direcciones simples (sin nombre ni <>).
func validEmail(s string) bool {
	a, err := mail.ParseAddress(s)
	return err == nil && a.Address == s
}

func passwordProblem(p string) string {
	if p == "" {
		return "El campo 'password' no puede estar vacío"
	}
	if len(p) > maxPasswordBytes {
		return "El campo 'password' no puede superar 72 bytes"
	}
	return ""
}

// fail registra el error y responde con el status indicado (code == status HTTP).
func (c *ClienteController) fail(status int, event, message string, err error, fields map[string]interface{}) {
	logging.LogControllerError(c.Ctx, event, err, fields)
	httpx.Fail(&c.Controller, status, message, err)
}

// writeError traduce errores de escritura: unicidad -> 409 (indicando el campo),
// resto -> 500.
func (c *ClienteController) writeError(op, message string, err error, doc int64) {
	fields := map[string]interface{}{"documento": doc}
	if dberr.IsUnique(err) {
		msg := "Ya existe un cliente con ese documento"
		switch {
		case dberr.Mentions(err, "correo"):
			msg = "El correo ya está registrado"
		case dberr.Mentions(err, "telefono"):
			msg = "El teléfono ya está registrado"
		}
		c.fail(http.StatusConflict, "clientes."+op+".unique_conflict", msg, err, fields)
		return
	}
	c.fail(http.StatusInternalServerError, "clientes."+op+".db_error", message, err, fields)
}

// @Title GetAll
// @Summary Listar clientes
// @Description Devuelve los clientes (nunca incluye la contraseña) ordenados por documento. Con `fields=nombre_completo_telefono` cada elemento es solo {documentoCliente, nombre_completo, telefono} (ver models.ClienteResumenResponse). Si se omite `limit` se devuelven todos los clientes; si se envía debe estar entre 1 y 100 (0 es inválido) y `offset` (>= 0, por defecto 0) permite paginar. Sin resultados responde 200 con `data: []`. Requiere token.
// @Tags clientes
// @Accept json
// @Produce json
// @Param   limit   query    int     false  "Cantidad de resultados por página (1-100). Si se omite, sin límite" minimum(1) maximum(100)
// @Param   offset  query    int     false  "Número de registros a omitir" minimum(0) default(0)
// @Param   fields  query    string  false  "Proyección reducida de cada cliente" Enums(nombre_completo_telefono)
// @Success 200 {object} models.ApiResponse{data=[]models.Cliente} "Lista de clientes (con fields=nombre_completo_telefono, data es []models.ClienteResumenResponse)"
// @Failure 400 {object} models.ApiResponse "Parámetro limit, offset o fields inválido"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /clientes [get]
func (c *ClienteController) GetAll() {
	fields := strings.TrimSpace(c.GetString("fields"))
	if fields != "" && fields != fieldsResumen {
		c.fail(http.StatusBadRequest, "clientes.getall.bad_fields", "Parámetro 'fields' inválido: solo se admite 'nombre_completo_telefono'", nil, map[string]interface{}{"fields": fields})
		return
	}

	limit := -1 // sin límite
	if raw := strings.TrimSpace(c.GetString("limit")); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 1 || v > maxLimit {
			c.fail(http.StatusBadRequest, "clientes.getall.bad_limit", "Parámetro 'limit' inválido: debe ser un entero entre 1 y 100", err, map[string]interface{}{"limit": raw})
			return
		}
		limit = v
	}
	offset := 0
	if raw := strings.TrimSpace(c.GetString("offset")); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 0 {
			c.fail(http.StatusBadRequest, "clientes.getall.bad_offset", "Parámetro 'offset' inválido: debe ser un entero mayor o igual a 0", err, map[string]interface{}{"offset": raw})
			return
		}
		offset = v
	}

	var clientes []models.Cliente
	if _, err := orm.NewOrm().QueryTable(new(models.Cliente)).OrderBy("PK_DOCUMENTO_CLIENTE").Limit(limit, offset).All(&clientes); err != nil {
		c.fail(http.StatusInternalServerError, "clientes.getall.db_error", "Error al obtener clientes", err, nil)
		return
	}

	if fields == fieldsResumen {
		resumen := make([]models.ClienteResumenResponse, 0, len(clientes))
		for _, cli := range clientes {
			resumen = append(resumen, models.ClienteResumenResponse{
				DocumentoCliente: cli.PK_DOCUMENTO_CLIENTE,
				NombreCompleto:   strings.TrimSpace(cli.NOMBRE + " " + cli.APELLIDO),
				Telefono:         cli.TELEFONO,
			})
		}
		httpx.Send(&c.Controller, http.StatusOK, "Clientes obtenidos", resumen)
		return
	}

	for i := range clientes {
		clientes[i].PASSWORD = ""
	}
	httpx.Send(&c.Controller, http.StatusOK, "Clientes obtenidos", httpx.List(clientes))
}

// @Title GetById
// @Summary Obtener cliente por documento
// @Description Devuelve un cliente por su documento (nunca incluye la contraseña). Requiere token.
// @Tags clientes
// @Accept json
// @Produce json
// @Param   id     query    int     true        "Documento del cliente (entero positivo)"
// @Success 200 {object} models.ApiResponse{data=models.Cliente} "Cliente encontrado"
// @Failure 400 {object} models.ApiResponse "Parámetro 'id' inválido o ausente"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 404 {object} models.ApiResponse "Cliente no encontrado"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /clientes/search [get]
func (c *ClienteController) GetById() {
	id, err := httpx.PositiveInt64Param(&c.Controller, "id")
	if err != nil {
		c.fail(http.StatusBadRequest, "clientes.getbyid.bad_request", "El parámetro 'id' es inválido o está ausente", err, map[string]interface{}{"id": c.GetString("id")})
		return
	}

	cliente := models.Cliente{PK_DOCUMENTO_CLIENTE: id}
	if err := orm.NewOrm().Read(&cliente); err != nil {
		if errors.Is(err, orm.ErrNoRows) {
			httpx.Fail(&c.Controller, http.StatusNotFound, "Cliente no encontrado", nil)
			return
		}
		c.fail(http.StatusInternalServerError, "clientes.getbyid.db_error", "Error al consultar el cliente", err, map[string]interface{}{"id": id})
		return
	}

	cliente.PASSWORD = ""
	httpx.Send(&c.Controller, http.StatusOK, "Cliente encontrado", cliente)
}

// @Title Create
// @Summary Registrar un cliente
// @Description Endpoint público (sin token) de registro de clientes. Obligatorios: documentoCliente (> 0), nombre, apellido, correo (válido; se guarda en minúsculas), telefono y password (máx. 72 bytes). Opcionales: direccion y observaciones. La respuesta nunca incluye la contraseña.
// @Tags clientes
// @Accept json
// @Produce json
// @Param   body  body   models.ClienteCreateRequest true  "Datos del cliente a crear"
// @Success 201 {object} models.ApiResponse{data=models.Cliente} "Cliente creado"
// @Failure 400 {object} models.ApiResponse "Solicitud inválida (JSON, campos obligatorios, correo o contraseña)"
// @Failure 409 {object} models.ApiResponse "Ya existe un cliente con ese documento, correo o teléfono"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Router /clientes [post]
func (c *ClienteController) Post() {
	var in models.ClienteCreateRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &in); err != nil {
		c.fail(http.StatusBadRequest, "clientes.post.bad_json", "Error al decodificar la solicitud", err, nil)
		return
	}

	bad := func(field, message string) {
		c.fail(http.StatusBadRequest, "clientes.post.validation_error", message, nil, map[string]interface{}{"field": field})
	}
	if in.DocumentoCliente <= 0 {
		bad("documentoCliente", "El campo 'documentoCliente' es obligatorio y debe ser un número positivo")
		return
	}
	nombre := strings.TrimSpace(in.Nombre)
	if nombre == "" {
		bad("nombre", "El campo 'nombre' es obligatorio")
		return
	}
	apellido := strings.TrimSpace(in.Apellido)
	if apellido == "" {
		bad("apellido", "El campo 'apellido' es obligatorio")
		return
	}
	correo := normalizeEmail(in.Correo)
	if correo == "" {
		bad("correo", "El campo 'correo' es obligatorio")
		return
	}
	if !validEmail(correo) {
		bad("correo", "El campo 'correo' no es un correo válido")
		return
	}
	telefono := strings.TrimSpace(in.Telefono)
	if telefono == "" {
		bad("telefono", "El campo 'telefono' es obligatorio")
		return
	}
	if msg := passwordProblem(in.Password); msg != "" {
		bad("password", msg)
		return
	}

	hashed, err := hashPassword(in.Password)
	if err != nil {
		c.fail(http.StatusInternalServerError, "clientes.post.hash_error", "Error al procesar la contraseña", err, nil)
		return
	}

	cliente := models.Cliente{
		PK_DOCUMENTO_CLIENTE: in.DocumentoCliente,
		NOMBRE:               nombre,
		APELLIDO:             apellido,
		CORREO:               correo,
		TELEFONO:             telefono,
		OBSERVACIONES:        in.Observaciones,
		PASSWORD:             hashed,
	}
	if in.Direccion != nil {
		cliente.DIRECCION = strings.TrimSpace(*in.Direccion)
	}

	if _, err := orm.NewOrm().Insert(&cliente); err != nil {
		c.writeError("post", "Error al crear el cliente", err, in.DocumentoCliente)
		return
	}

	cliente.PASSWORD = ""
	httpx.Send(&c.Controller, http.StatusCreated, "Cliente creado correctamente", cliente)
}

// @Title Update
// @Summary Actualizar un cliente
// @Description Actualización parcial con merge: los campos ausentes se CONSERVAN. `observaciones` es el único campo anulable (null lo limpia); null en cualquier otro campo responde 400, igual que nombre, apellido, correo (inválido), telefono o password vacíos. Una `password` nueva se guarda hasheada y nunca se devuelve. Requiere token.
// @Tags clientes
// @Accept json
// @Produce json
// @Param   id    query    int  true   "Documento del cliente (entero positivo)"
// @Param   body  body   models.ClienteUpdateRequest true  "Campos a modificar (todos opcionales)"
// @Success 200 {object} models.ApiResponse{data=models.Cliente} "Cliente actualizado"
// @Failure 400 {object} models.ApiResponse "Solicitud inválida (id, JSON, null en campo no anulable, correo, teléfono o contraseña)"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 404 {object} models.ApiResponse "Cliente no encontrado"
// @Failure 409 {object} models.ApiResponse "Correo o teléfono ya registrados por otro cliente"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /clientes [put]
func (c *ClienteController) Put() {
	id, err := httpx.PositiveInt64Param(&c.Controller, "id")
	if err != nil {
		c.fail(http.StatusBadRequest, "clientes.put.bad_request", "El parámetro 'id' es inválido o está ausente", err, map[string]interface{}{"id": c.GetString("id")})
		return
	}

	o := orm.NewOrm()
	cliente := models.Cliente{PK_DOCUMENTO_CLIENTE: id}
	if err := o.Read(&cliente); err != nil {
		if errors.Is(err, orm.ErrNoRows) {
			httpx.Fail(&c.Controller, http.StatusNotFound, "Cliente no encontrado", nil)
			return
		}
		c.fail(http.StatusInternalServerError, "clientes.put.db_error", "Error al buscar el cliente", err, map[string]interface{}{"id": id})
		return
	}

	var in models.ClienteUpdateRequest
	if err := httpx.DecodeMerge(c.Ctx.Input.RequestBody, &in, "observaciones"); err != nil {
		c.fail(http.StatusBadRequest, "clientes.put.bad_json", "Error al decodificar la solicitud", err, map[string]interface{}{"id": id})
		return
	}

	bad := func(message string) {
		c.fail(http.StatusBadRequest, "clientes.put.validation_error", message, nil, map[string]interface{}{"id": id})
	}
	if in.Nombre != nil {
		v := strings.TrimSpace(*in.Nombre)
		if v == "" {
			bad("El campo 'nombre' no puede estar vacío")
			return
		}
		cliente.NOMBRE = v
	}
	if in.Apellido != nil {
		v := strings.TrimSpace(*in.Apellido)
		if v == "" {
			bad("El campo 'apellido' no puede estar vacío")
			return
		}
		cliente.APELLIDO = v
	}
	if in.Correo != nil {
		v := normalizeEmail(*in.Correo)
		if !validEmail(v) {
			bad("El campo 'correo' no es un correo válido")
			return
		}
		cliente.CORREO = v
	}
	if in.Telefono != nil {
		v := strings.TrimSpace(*in.Telefono)
		if v == "" {
			bad("El campo 'telefono' no puede estar vacío")
			return
		}
		cliente.TELEFONO = v
	}
	if in.Direccion != nil {
		cliente.DIRECCION = strings.TrimSpace(*in.Direccion)
	}
	if httpx.IsNull(c.Ctx.Input.RequestBody, "observaciones") {
		cliente.OBSERVACIONES = nil
	} else if in.Observaciones != nil {
		cliente.OBSERVACIONES = in.Observaciones
	}
	if in.Password != nil {
		if msg := passwordProblem(*in.Password); msg != "" {
			bad(msg)
			return
		}
		hashed, err := hashPassword(*in.Password)
		if err != nil {
			c.fail(http.StatusInternalServerError, "clientes.put.hash_error", "Error al procesar la contraseña", err, map[string]interface{}{"id": id})
			return
		}
		cliente.PASSWORD = hashed
	}

	if _, err := o.Update(&cliente); err != nil {
		c.writeError("put", "Error al actualizar el cliente", err, id)
		return
	}

	cliente.PASSWORD = ""
	httpx.Send(&c.Controller, http.StatusOK, "Cliente actualizado", cliente)
}

// @Title Delete
// @Summary Eliminar un cliente
// @Description Elimina un cliente. Si tiene registros asociados (pedidos, reservas, etc.) responde 409. Requiere token.
// @Tags clientes
// @Accept json
// @Produce json
// @Param   id     query    int     true        "Documento del cliente (entero positivo)"
// @Success 200 {object} models.ApiResponse "Cliente eliminado"
// @Failure 400 {object} models.ApiResponse "Parámetro 'id' inválido o ausente"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 404 {object} models.ApiResponse "Cliente no encontrado"
// @Failure 409 {object} models.ApiResponse "El cliente tiene registros asociados"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /clientes [delete]
func (c *ClienteController) Delete() {
	id, err := httpx.PositiveInt64Param(&c.Controller, "id")
	if err != nil {
		c.fail(http.StatusBadRequest, "clientes.delete.bad_request", "El parámetro 'id' es inválido o está ausente", err, map[string]interface{}{"id": c.GetString("id")})
		return
	}

	// DELETE directo (no Ormer.Delete): el ORM borraría en cascada los pedidos y
	// demás registros del cliente; así lo decide la FK de la BD (409).
	res, err := orm.NewOrm().Raw("DELETE FROM cliente WHERE pk_documento_cliente = ?", id).Exec()
	if err != nil {
		if dberr.IsForeignKey(err) {
			c.fail(http.StatusConflict, "clientes.delete.fk_conflict", "No se puede eliminar el cliente porque tiene registros asociados", err, map[string]interface{}{"id": id})
			return
		}
		c.fail(http.StatusInternalServerError, "clientes.delete.db_error", "Error al eliminar el cliente", err, map[string]interface{}{"id": id})
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		httpx.Fail(&c.Controller, http.StatusNotFound, "Cliente no encontrado", nil)
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, "Cliente eliminado", nil)
}
