-- Seed initial data for new tables

INSERT INTO restaurante_dia (dia) VALUES
  ('lunes'),
  ('martes'),
  ('miercoles'),
  ('jueves'),
  ('viernes'),
  ('sabado'),
  ('domingo');

INSERT INTO precio_producto_hist (producto_id, precio) VALUES
  (1, 10.00),
  (2, 15.50);

INSERT INTO nomina (trabajador_id, fecha, tipo, monto) VALUES
  (1, CURRENT_DATE, 'mensual', 1000.00);

INSERT INTO control_nomina (nomina_id, aprobado) VALUES
  (1, true);

INSERT INTO reserva (cliente_id, fecha_reserva, estado) VALUES
  (1, NOW() + INTERVAL '1 day', 'pendiente');

INSERT INTO reserva_contacto (reserva_id, nombre, telefono) VALUES
  (1, 'Contacto', '0000000000');

INSERT INTO domicilio (pedido_id, direccion, estado) VALUES
  (1, 'Calle 123', 'pendiente');

INSERT INTO detalle_pedido (pedido_id, producto_id, cantidad) VALUES
  (1, 1, 2);

INSERT INTO horario_trabajador (trabajador_id, dia, hora_inicio, hora_fin) VALUES
  (1, 'lunes', '09:00', '17:00');
