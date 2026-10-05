package trabajador

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"restaurante/internal/dberr"
	"restaurante/internal/httpx"
	"restaurante/logging"
	"restaurante/models"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
	"golang.org/x/crypto/bcrypt"
)

var newTrabajadorOrm = orm.NewOrm

// TrabajadorController gestiona /trabajadores. Todas sus rutas exigen token de
// administrador (filtro loginc.ValidateAdmin registrado en el router).
type TrabajadorController struct {
	web.Controller
}

var generateFromPassword = bcrypt.GenerateFromPassword

var hashPassword = func(password string) (string, error) {
	hashedPassword, err := generateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// maxPasswordBytes es el límite de bcrypt; más largo no se puede hashear.
const maxPasswordBytes = 72

func passwordProblem(p string) string {
	if p == "" {
		return "El campo 'password' no puede estar vacío"
	}
	if len(p) > maxPasswordBytes {
		return "El campo 'password' no puede superar 72 bytes"
	}
	return ""
}

func validateDates(fechaIngreso, fechaRetiro *time.Time) error {
	if fechaIngreso != nil && fechaRetiro != nil {
		if fechaRetiro.Before(*fechaIngreso) {
			return fmt.Errorf("la fecha de retiro no puede ser anterior a la fecha de ingreso")
		}
	}
	return nil
}

// fail registra el error y responde con el status indicado (code == status HTTP).
func (c *TrabajadorController) fail(status int, event, message string, err error, fields map[string]interface{}) {
	logging.LogControllerError(c.Ctx, event, err, fields)
	httpx.Fail(&c.Controller, status, message, err)
}

// loadHorarios devuelve los horarios de los trabajadores indicados agrupados por documento.
func loadHorarios(o orm.Ormer, docs []int64) (map[int64][]models.HorarioTrabajador, error) {
	out := make(map[int64][]models.HorarioTrabajador, len(docs))
	if len(docs) == 0 {
		return out, nil
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(docs)), ",")
	args := make([]interface{}, len(docs))
	for i, d := range docs {
		args[i] = d
	}
	var horarios []models.HorarioTrabajador
	query := "SELECT pk_documento_trabajador, dia, hora_inicio, hora_fin FROM horario_trabajador WHERE pk_documento_trabajador IN (" + placeholders + ") ORDER BY pk_documento_trabajador, dia, hora_inicio"
	if _, err := o.Raw(query, args...).QueryRows(&horarios); err != nil {
		return nil, err
	}
	for _, h := range horarios {
		// la columna pk_documento_trabajador es NOT NULL y siempre se selecciona
		doc := h.PK_DOCUMENTO_TRABAJADOR.PK_DOCUMENTO_TRABAJADOR
		out[doc] = append(out[doc], h)
	}
	return out, nil
}

// attachHorarios carga los horarios de un único trabajador (siempre una lista, nunca nil).
func attachHorarios(o orm.Ormer, t *models.Trabajador) error {
	m, err := loadHorarios(o, []int64{t.PK_DOCUMENTO_TRABAJADOR})
	if err != nil {
		return err
	}
	t.HORARIOS = httpx.List(m[t.PK_DOCUMENTO_TRABAJADOR])
	return nil
}

func parseBoolParam(c *TrabajadorController, name string) (bool, error) {
	raw := strings.TrimSpace(c.GetString(name))
	if raw == "" {
		return false, nil
	}
	return strconv.ParseBool(raw)
}

// @Title GetAll
// @Summary Listar trabajadores (solo administrador)
// @Description Devuelve los trabajadores (nunca incluye la contraseña), cada uno con su lista `horarios` (`[]` si no tiene). Por defecto excluye a los retirados. Las fechas se devuelven en formato DD-MM-YYYY; el filtro `fecha_ingreso` se envía en YYYY-MM-DD. Si no hay coincidencias responde 200 con `data: []`. Requiere token de un usuario con rol Administrador.
// @Tags trabajadores
// @Accept json
// @Produce json
// @Param   fecha_ingreso     query   string   false  "Filtrar por fecha exacta de ingreso (YYYY-MM-DD). También se acepta el alias fechaIngreso"
// @Param   rol               query   string   false  "Filtrar por rol" Enums(Administrador, Mesero, Cocinero, Domiciliario, Oficios_varios)
// @Param   incluir_retirados query   bool     false  "Incluir trabajadores retirados (true/false)" default(false)
// @Param   solo_retirados    query   bool     false  "Ver solo trabajadores retirados (true/false); tiene prioridad sobre incluir_retirados" default(false)
// @Success 200 {object} models.ApiResponse{data=[]models.TrabajadorResponse} "Lista de trabajadores (puede ser vacía)"
// @Failure 400 {object} models.ApiResponse "Parámetro inválido (fecha, rol o booleano)"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 403 {object} models.ApiResponse "El usuario no es administrador"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /trabajadores [get]
func (c *TrabajadorController) GetAll() {
	var trabajadores []models.Trabajador

	fechaIngreso := strings.TrimSpace(c.GetString("fecha_ingreso"))
	if fechaIngreso == "" {
		fechaIngreso = strings.TrimSpace(c.GetString("fechaIngreso"))
	}
	rol := strings.TrimSpace(c.GetString("rol"))
	incluirRetirados, err := parseBoolParam(c, "incluir_retirados")
	if err != nil {
		c.fail(http.StatusBadRequest, "trabajadores.getall.bad_incluir_retirados", "El parámetro 'incluir_retirados' debe ser true o false", err, nil)
		return
	}
	soloRetirados, err := parseBoolParam(c, "solo_retirados")
	if err != nil {
		c.fail(http.StatusBadRequest, "trabajadores.getall.bad_solo_retirados", "El parámetro 'solo_retirados' debe ser true o false", err, nil)
		return
	}
	if rol != "" && !models.RolTrabajador(rol).IsValid() {
		c.fail(http.StatusBadRequest, "trabajadores.getall.bad_rol", "Rol inválido para 'rol'", nil, map[string]interface{}{"rol": rol})
		return
	}

	var fechaFiltro time.Time
	if fechaIngreso != "" {
		var err error
		fechaFiltro, err = models.ParseDateToNoonUTC(fechaIngreso)
		if err != nil {
			c.fail(http.StatusBadRequest, "trabajadores.getall.bad_fecha_ingreso", "Formato de fecha inválido para 'fecha_ingreso', use YYYY-MM-DD", err, map[string]interface{}{"fecha_ingreso": fechaIngreso})
			return
		}
	}

	o := newTrabajadorOrm()
	query := o.QueryTable(new(models.Trabajador))
	if soloRetirados {
		query = query.Filter("FECHA_RETIRO__isnull", false)
	} else if !incluirRetirados {
		query = query.Filter("FECHA_RETIRO__isnull", true)
	}
	if fechaIngreso != "" {
		query = query.Filter("FECHA_INGRESO", fechaFiltro)
	}
	if rol != "" {
		query = query.Filter("ROL__exact", rol)
	}

	if _, err := query.All(&trabajadores); err != nil {
		c.fail(http.StatusInternalServerError, "trabajadores.getall.db_error", "Error al obtener trabajadores de la base de datos", err, map[string]interface{}{"rol": rol, "solo_retirados": soloRetirados, "incluir_retirados": incluirRetirados, "fecha_ingreso": fechaIngreso})
		return
	}

	docs := make([]int64, len(trabajadores))
	for i := range trabajadores {
		docs[i] = trabajadores[i].PK_DOCUMENTO_TRABAJADOR
	}
	horarios, err := loadHorarios(o, docs)
	if err != nil {
		c.fail(http.StatusInternalServerError, "trabajadores.getall.horarios_error", "Error al obtener los horarios de los trabajadores", err, nil)
		return
	}
	for i := range trabajadores {
		trabajadores[i].PASSWORD = ""
		trabajadores[i].HORARIOS = httpx.List(horarios[trabajadores[i].PK_DOCUMENTO_TRABAJADOR])
	}

	message := "Trabajadores obtenidos exitosamente"
	if len(trabajadores) == 0 {
		message = "No se encontraron trabajadores que coincidan con los filtros proporcionados"
	}
	httpx.Send(&c.Controller, http.StatusOK, message, httpx.List(trabajadores))
}

// @Title GetById
// @Summary Obtener trabajador por documento (solo administrador)
// @Description Devuelve un trabajador (incluso retirado) por su documento, con su lista `horarios`. Nunca incluye la contraseña. Fechas en DD-MM-YYYY. Requiere rol Administrador.
// @Tags trabajadores
// @Accept json
// @Produce json
// @Param   id     query    int     true        "Documento del trabajador (entero positivo)"
// @Success 200 {object} models.ApiResponse{data=models.TrabajadorResponse} "Trabajador encontrado"
// @Failure 400 {object} models.ApiResponse "Parámetro 'id' inválido o ausente"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 403 {object} models.ApiResponse "El usuario no es administrador"
// @Failure 404 {object} models.ApiResponse "Trabajador no encontrado"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /trabajadores/search [get]
func (c *TrabajadorController) GetById() {
	id, err := httpx.PositiveInt64Param(&c.Controller, "id")
	if err != nil {
		c.fail(http.StatusBadRequest, "trabajadores.getbyid.bad_request", "El parámetro 'id' es inválido o está ausente", err, map[string]interface{}{"id": c.GetString("id")})
		return
	}

	o := newTrabajadorOrm()
	trabajador := models.Trabajador{PK_DOCUMENTO_TRABAJADOR: id}
	if err := o.Read(&trabajador); err != nil {
		if errors.Is(err, orm.ErrNoRows) {
			httpx.Fail(&c.Controller, http.StatusNotFound, "Trabajador no encontrado", nil)
			return
		}
		c.fail(http.StatusInternalServerError, "trabajadores.getbyid.db_error", "Error al buscar el trabajador", err, map[string]interface{}{"id": id})
		return
	}

	trabajador.PASSWORD = ""
	if err := attachHorarios(o, &trabajador); err != nil {
		c.fail(http.StatusInternalServerError, "trabajadores.getbyid.horarios_error", "Error al obtener los horarios del trabajador", err, map[string]interface{}{"id": id})
		return
	}
	httpx.Send(&c.Controller, http.StatusOK, "Trabajador encontrado", trabajador)
}

// @Title Create
// @Summary Crear un trabajador (solo administrador)
// @Description Crea un trabajador. Solo un administrador puede hacerlo (y por tanto crear otro administrador). Obligatorios: documentoTrabajador, nombre, apellido, rol, fechaIngreso (YYYY-MM-DD), sueldo (>= 0) y password (máx. 72 bytes). Opcionales: nuevo (booleano; si se omite queda en false), telefono (cadena vacía equivale a no informado), restauranteId y fechaNacimiento (YYYY-MM-DD). La respuesta (fechas en DD-MM-YYYY) nunca incluye la contraseña y trae `horarios: []`.
// @Tags trabajadores
// @Accept json
// @Produce json
// @Param   body  body   models.TrabajadorCreateRequest true  "Datos del trabajador a crear (fechas YYYY-MM-DD)"
// @Success 201 {object} models.ApiResponse{data=models.TrabajadorResponse} "Trabajador creado"
// @Failure 400 {object} models.ApiResponse "Solicitud inválida (JSON, campos obligatorios, rol, fechas, password o restauranteId inexistente)"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 403 {object} models.ApiResponse "El usuario no es administrador"
// @Failure 409 {object} models.ApiResponse "Ya existe un trabajador con ese documento o teléfono"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /trabajadores [post]
func (c *TrabajadorController) Post() {
	body := c.Ctx.Input.RequestBody

	var in models.TrabajadorCreateRequest
	if err := json.Unmarshal(body, &in); err != nil {
		c.fail(http.StatusBadRequest, "trabajadores.post.bad_json", "Error al decodificar la solicitud", err, nil)
		return
	}
	present, _ := httpx.Present(body)

	bad := func(field, message string) {
		c.fail(http.StatusBadRequest, "trabajadores.post.validation_error", message, nil, map[string]interface{}{"field": field})
	}

	if in.DocumentoTrabajador <= 0 {
		bad("documentoTrabajador", "El campo 'documentoTrabajador' es obligatorio y debe ser un número positivo")
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
	rol := models.RolTrabajador(strings.TrimSpace(in.Rol))
	if rol == "" {
		bad("rol", "El campo 'rol' es obligatorio")
		return
	}
	if !rol.IsValid() {
		bad("rol", "Rol inválido: use Administrador, Mesero, Cocinero, Domiciliario u Oficios_varios")
		return
	}
	if strings.TrimSpace(in.FechaIngreso) == "" {
		bad("fechaIngreso", "El campo 'fechaIngreso' es obligatorio")
		return
	}
	fechaIngreso, err := models.ParseDateToNoonUTC(in.FechaIngreso)
	if err != nil {
		c.fail(http.StatusBadRequest, "trabajadores.post.bad_fecha_ingreso", "Formato de fecha inválido para 'fechaIngreso', use YYYY-MM-DD", err, map[string]interface{}{"fechaIngreso": in.FechaIngreso})
		return
	}
	if !present["sueldo"] {
		bad("sueldo", "El campo 'sueldo' es obligatorio")
		return
	}
	if in.Sueldo < 0 {
		bad("sueldo", "El campo 'sueldo' no puede ser negativo")
		return
	}
	if msg := passwordProblem(in.Password); msg != "" {
		bad("password", msg)
		return
	}

	trabajador := models.Trabajador{
		PK_DOCUMENTO_TRABAJADOR: in.DocumentoTrabajador,
		NOMBRE:                  nombre,
		APELLIDO:                apellido,
		ROL:                     rol,
		FECHA_INGRESO:           fechaIngreso,
		SUELDO:                  in.Sueldo,
	}
	if in.Nuevo != nil {
		trabajador.NUEVO = *in.Nuevo
	}
	if in.Telefono != nil {
		if tel := strings.TrimSpace(*in.Telefono); tel != "" {
			trabajador.TELEFONO = &tel
		}
	}
	if in.RestauranteId != nil {
		if *in.RestauranteId <= 0 {
			bad("restauranteId", "El campo 'restauranteId' debe ser un número positivo")
			return
		}
		trabajador.PK_ID_RESTAURANTE = &models.Restaurante{PK_ID_RESTAURANTE: *in.RestauranteId}
	}
	if in.FechaNacimiento != nil && strings.TrimSpace(*in.FechaNacimiento) != "" {
		parsed, err := models.ParseDateToNoonUTC(*in.FechaNacimiento)
		if err != nil {
			c.fail(http.StatusBadRequest, "trabajadores.post.bad_fecha_nacimiento", "Formato de fecha inválido para 'fechaNacimiento', use YYYY-MM-DD", err, map[string]interface{}{"fechaNacimiento": *in.FechaNacimiento})
			return
		}
		trabajador.FECHA_NACIMIENTO = &parsed
	}

	hashed, err := hashPassword(in.Password)
	if err != nil {
		c.fail(http.StatusInternalServerError, "trabajadores.post.hash_error", "Error al procesar la contraseña", err, nil)
		return
	}
	trabajador.PASSWORD = hashed

	if _, err := newTrabajadorOrm().Insert(&trabajador); err != nil {
		c.writeError("post", "Error al crear el trabajador", err, trabajador.PK_DOCUMENTO_TRABAJADOR)
		return
	}

	trabajador.PASSWORD = ""
	trabajador.HORARIOS = []models.HorarioTrabajador{}
	httpx.Send(&c.Controller, http.StatusCreated, "Trabajador creado correctamente", trabajador)
}

// writeError traduce errores de escritura: unicidad -> 409, FK -> 400, resto -> 500.
func (c *TrabajadorController) writeError(op, message string, err error, doc int64) {
	fields := map[string]interface{}{"documento": doc}
	switch {
	case dberr.IsUnique(err):
		msg := "Ya existe un trabajador con ese documento"
		if dberr.Mentions(err, "telefono") {
			msg = "El teléfono ya está registrado por otro trabajador"
		}
		c.fail(http.StatusConflict, "trabajadores."+op+".unique_conflict", msg, err, fields)
	case dberr.IsForeignKey(err):
		c.fail(http.StatusBadRequest, "trabajadores."+op+".fk_error", "El 'restauranteId' indicado no existe", err, fields)
	default:
		c.fail(http.StatusInternalServerError, "trabajadores."+op+".db_error", message, err, fields)
	}
}

// @Title Update
// @Summary Actualizar un trabajador (solo administrador)
// @Description Actualización parcial con merge: los campos ausentes se CONSERVAN. Un campo anulable enviado como null se limpia (telefono, fechaNacimiento, fechaRetiro y restauranteId); null en cualquier otro campo responde 400, igual que las cadenas vacías en nombre, apellido o password. Fechas en YYYY-MM-DD en la petición y DD-MM-YYYY en la respuesta. Una `password` nueva se guarda hasheada y nunca se devuelve. Un teléfono vacío equivale a null. Requiere rol Administrador (un no administrador no puede modificar roles).
// @Tags trabajadores
// @Accept json
// @Produce json
// @Param   id    query    int  true   "Documento del trabajador (entero positivo)"
// @Param   body  body   models.TrabajadorUpdateRequest true  "Campos a modificar (todos opcionales)"
// @Success 200 {object} models.ApiResponse{data=models.TrabajadorResponse} "Trabajador actualizado"
// @Failure 400 {object} models.ApiResponse "Solicitud inválida (id, JSON, null en campo no anulable, rol, fechas, password, restauranteId inexistente)"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 403 {object} models.ApiResponse "El usuario no es administrador"
// @Failure 404 {object} models.ApiResponse "Trabajador no encontrado"
// @Failure 409 {object} models.ApiResponse "El teléfono ya está registrado por otro trabajador"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /trabajadores [put]
func (c *TrabajadorController) Put() {
	id, err := httpx.PositiveInt64Param(&c.Controller, "id")
	if err != nil {
		c.fail(http.StatusBadRequest, "trabajadores.put.bad_request", "ID inválido o ausente", err, map[string]interface{}{"id": c.GetString("id")})
		return
	}

	o := newTrabajadorOrm()
	trabajador := models.Trabajador{PK_DOCUMENTO_TRABAJADOR: id}
	if err := o.Read(&trabajador); err != nil {
		if errors.Is(err, orm.ErrNoRows) {
			httpx.Fail(&c.Controller, http.StatusNotFound, "Trabajador no encontrado", nil)
			return
		}
		c.fail(http.StatusInternalServerError, "trabajadores.put.db_read_error", "Error al buscar el trabajador", err, map[string]interface{}{"id": id})
		return
	}
	if err := attachHorarios(o, &trabajador); err != nil {
		c.fail(http.StatusInternalServerError, "trabajadores.put.horarios_error", "Error al obtener los horarios del trabajador", err, map[string]interface{}{"id": id})
		return
	}

	body := c.Ctx.Input.RequestBody
	var in models.TrabajadorUpdateRequest
	if err := httpx.DecodeMerge(body, &in, "telefono", "fechaNacimiento", "fechaRetiro", "restauranteId"); err != nil {
		c.fail(http.StatusBadRequest, "trabajadores.put.bad_json", "Error al decodificar los datos", err, map[string]interface{}{"id": id})
		return
	}

	bad := func(message string, err error) {
		c.fail(http.StatusBadRequest, "trabajadores.put.validation_error", message, err, map[string]interface{}{"id": id})
	}

	if in.Nombre != nil {
		v := strings.TrimSpace(*in.Nombre)
		if v == "" {
			bad("El campo 'nombre' no puede estar vacío", nil)
			return
		}
		trabajador.NOMBRE = v
	}
	if in.Apellido != nil {
		v := strings.TrimSpace(*in.Apellido)
		if v == "" {
			bad("El campo 'apellido' no puede estar vacío", nil)
			return
		}
		trabajador.APELLIDO = v
	}
	if in.Rol != nil {
		rol := models.RolTrabajador(strings.TrimSpace(*in.Rol))
		if !rol.IsValid() {
			bad("Rol inválido: use Administrador, Mesero, Cocinero, Domiciliario u Oficios_varios", nil)
			return
		}
		trabajador.ROL = rol
	}
	if in.Sueldo != nil {
		if *in.Sueldo < 0 {
			bad("El campo 'sueldo' no puede ser negativo", nil)
			return
		}
		trabajador.SUELDO = *in.Sueldo
	}
	if in.Nuevo != nil {
		trabajador.NUEVO = *in.Nuevo
	}

	if httpx.IsNull(body, "telefono") {
		trabajador.TELEFONO = nil
	} else if in.Telefono != nil {
		if tel := strings.TrimSpace(*in.Telefono); tel == "" {
			trabajador.TELEFONO = nil
		} else {
			trabajador.TELEFONO = &tel
		}
	}

	if in.FechaIngreso != nil {
		parsed, err := models.ParseDateToNoonUTC(*in.FechaIngreso)
		if err != nil {
			bad("Formato de fecha inválido para 'fechaIngreso', use YYYY-MM-DD", err)
			return
		}
		trabajador.FECHA_INGRESO = parsed
	}
	if httpx.IsNull(body, "fechaRetiro") {
		trabajador.FECHA_RETIRO = nil
	} else if in.FechaRetiro != nil {
		parsed, err := models.ParseDateToNoonUTC(*in.FechaRetiro)
		if err != nil {
			bad("Formato de fecha inválido para 'fechaRetiro', use YYYY-MM-DD", err)
			return
		}
		trabajador.FECHA_RETIRO = &parsed
	}
	if httpx.IsNull(body, "fechaNacimiento") {
		trabajador.FECHA_NACIMIENTO = nil
	} else if in.FechaNacimiento != nil {
		parsed, err := models.ParseDateToNoonUTC(*in.FechaNacimiento)
		if err != nil {
			bad("Formato de fecha inválido para 'fechaNacimiento', use YYYY-MM-DD", err)
			return
		}
		trabajador.FECHA_NACIMIENTO = &parsed
	}
	if httpx.IsNull(body, "restauranteId") {
		trabajador.PK_ID_RESTAURANTE = nil
	} else if in.RestauranteId != nil {
		if *in.RestauranteId <= 0 {
			bad("El campo 'restauranteId' debe ser un número positivo", nil)
			return
		}
		trabajador.PK_ID_RESTAURANTE = &models.Restaurante{PK_ID_RESTAURANTE: *in.RestauranteId}
	}
	if in.Password != nil {
		if msg := passwordProblem(*in.Password); msg != "" {
			bad(msg, nil)
			return
		}
		hashed, err := hashPassword(*in.Password)
		if err != nil {
			c.fail(http.StatusInternalServerError, "trabajadores.put.hash_error", "Error al procesar la contraseña", err, map[string]interface{}{"id": id})
			return
		}
		trabajador.PASSWORD = hashed
	}

	if err := validateDates(&trabajador.FECHA_INGRESO, trabajador.FECHA_RETIRO); err != nil {
		bad(err.Error(), nil)
		return
	}

	if _, err := o.Update(&trabajador); err != nil {
		c.writeError("put", "Error al actualizar el trabajador", err, id)
		return
	}

	trabajador.PASSWORD = ""
	httpx.Send(&c.Controller, http.StatusOK, "Trabajador actualizado correctamente", trabajador)
}

// @Title Delete
// @Summary Retirar un trabajador (solo administrador)
// @Description Baja lógica: no borra el registro, fija `fechaRetiro` a la fecha actual (UTC) y el trabajador deja de aparecer en el listado por defecto. Si ya estaba retirado responde 409 para no sobrescribir su fecha de retiro (use PUT con fechaRetiro para corregirla). Devuelve el trabajador actualizado (fechas DD-MM-YYYY, sin contraseña). Requiere rol Administrador.
// @Tags trabajadores
// @Accept json
// @Produce json
// @Param   id     query    int     true        "Documento del trabajador (entero positivo)"
// @Success 200 {object} models.ApiResponse{data=models.TrabajadorResponse} "Trabajador retirado"
// @Failure 400 {object} models.ApiResponse "Parámetro 'id' inválido o ausente"
// @Failure 401 {object} models.ApiResponse "Token ausente o inválido"
// @Failure 403 {object} models.ApiResponse "El usuario no es administrador"
// @Failure 404 {object} models.ApiResponse "Trabajador no encontrado"
// @Failure 409 {object} models.ApiResponse "El trabajador ya estaba retirado"
// @Failure 500 {object} models.ApiResponse "Error en la base de datos"
// @Security BearerAuth
// @Router /trabajadores [delete]
func (c *TrabajadorController) Delete() {
	id, err := httpx.PositiveInt64Param(&c.Controller, "id")
	if err != nil {
		c.fail(http.StatusBadRequest, "trabajadores.delete.bad_request", "El parámetro 'id' es inválido o está ausente", err, map[string]interface{}{"id": c.GetString("id")})
		return
	}

	o := newTrabajadorOrm()
	trabajador := models.Trabajador{PK_DOCUMENTO_TRABAJADOR: id}
	if err := o.Read(&trabajador); err != nil {
		if errors.Is(err, orm.ErrNoRows) {
			httpx.Fail(&c.Controller, http.StatusNotFound, "Trabajador no encontrado", nil)
			return
		}
		c.fail(http.StatusInternalServerError, "trabajadores.delete.db_read_error", "Error al buscar el trabajador", err, map[string]interface{}{"id": id})
		return
	}
	if trabajador.FECHA_RETIRO != nil {
		httpx.Fail(&c.Controller, http.StatusConflict, "El trabajador ya está retirado", nil)
		return
	}
	if err := attachHorarios(o, &trabajador); err != nil {
		c.fail(http.StatusInternalServerError, "trabajadores.delete.horarios_error", "Error al obtener los horarios del trabajador", err, map[string]interface{}{"id": id})
		return
	}

	nowUTC := time.Now().UTC()
	fechaRetiro := time.Date(nowUTC.Year(), nowUTC.Month(), nowUTC.Day(), 12, 0, 0, 0, time.UTC)
	trabajador.FECHA_RETIRO = &fechaRetiro

	if _, err := o.Update(&trabajador, "FECHA_RETIRO"); err != nil {
		c.fail(http.StatusInternalServerError, "trabajadores.delete.update_error", "Error al actualizar la fecha de retiro del trabajador", err, map[string]interface{}{"id": id})
		return
	}
	trabajador.PASSWORD = ""
	httpx.Send(&c.Controller, http.StatusOK, "Fecha de retiro del trabajador actualizada correctamente", trabajador)
}
