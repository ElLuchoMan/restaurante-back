-- +migrate Up
CREATE TYPE estado_domicilio AS ENUM ('pendiente','en_camino','entregado','cancelado');
CREATE TYPE estado_reserva AS ENUM ('pendiente','confirmada','cancelada');
CREATE TYPE tipo_nomina AS ENUM ('mensual','semanal');
CREATE TYPE dia_semana AS ENUM ('lunes','martes','miercoles','jueves','viernes','sabado','domingo');

CREATE TABLE domicilio (
    id SERIAL PRIMARY KEY,
    pedido_id INT,
    direccion TEXT NOT NULL,
    estado estado_domicilio NOT NULL DEFAULT 'pendiente',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE TABLE nomina (
    id SERIAL PRIMARY KEY,
    trabajador_id INT NOT NULL,
    fecha DATE NOT NULL,
    tipo tipo_nomina NOT NULL,
    monto NUMERIC(10,2) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE TABLE control_nomina (
    id SERIAL PRIMARY KEY,
    nomina_id INT REFERENCES nomina(id),
    aprobado BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE TABLE precio_producto_hist (
    id SERIAL PRIMARY KEY,
    producto_id INT NOT NULL,
    precio NUMERIC(10,2) NOT NULL,
    vigente_desde TIMESTAMP NOT NULL DEFAULT NOW(),
    vigente_hasta TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE TABLE detalle_pedido (
    id SERIAL PRIMARY KEY,
    pedido_id INT NOT NULL,
    producto_id INT NOT NULL,
    cantidad INT NOT NULL,
    precio NUMERIC(10,2) NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE OR REPLACE FUNCTION set_precio_detalle_pedido() RETURNS TRIGGER AS $$
BEGIN
    SELECT p.precio INTO NEW.precio
    FROM precio_producto_hist p
    WHERE p.producto_id = NEW.producto_id
      AND p.vigente_hasta IS NULL
    ORDER BY p.vigente_desde DESC
    LIMIT 1;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_precio_detalle_pedido
BEFORE INSERT ON detalle_pedido
FOR EACH ROW EXECUTE PROCEDURE set_precio_detalle_pedido();

CREATE TABLE reserva (
    id SERIAL PRIMARY KEY,
    cliente_id INT NOT NULL,
    fecha_reserva TIMESTAMP NOT NULL,
    estado estado_reserva NOT NULL DEFAULT 'pendiente',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE TABLE reserva_contacto (
    id SERIAL PRIMARY KEY,
    reserva_id INT REFERENCES reserva(id),
    nombre TEXT NOT NULL,
    telefono TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE TABLE restaurante_dia (
    id SERIAL PRIMARY KEY,
    dia dia_semana NOT NULL,
    abierto BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE TABLE horario_trabajador (
    id SERIAL PRIMARY KEY,
    trabajador_id INT NOT NULL,
    dia dia_semana NOT NULL,
    hora_inicio TIME NOT NULL,
    hora_fin TIME NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- +migrate Down
DROP TRIGGER IF EXISTS set_precio_detalle_pedido ON detalle_pedido;
DROP FUNCTION IF EXISTS set_precio_detalle_pedido;
DROP TABLE IF EXISTS horario_trabajador;
DROP TABLE IF EXISTS restaurante_dia;
DROP TABLE IF EXISTS reserva_contacto;
DROP TABLE IF EXISTS reserva;
DROP TABLE IF EXISTS detalle_pedido;
DROP TABLE IF EXISTS precio_producto_hist;
DROP TABLE IF EXISTS control_nomina;
DROP TABLE IF EXISTS nomina;
DROP TABLE IF EXISTS domicilio;
DROP TYPE IF EXISTS dia_semana;
DROP TYPE IF EXISTS tipo_nomina;
DROP TYPE IF EXISTS estado_reserva;
DROP TYPE IF EXISTS estado_domicilio;
