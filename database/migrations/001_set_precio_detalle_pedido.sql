-- Sets the price of detalle_pedido rows based on the current product price.
-- This function and trigger ensure the precio field is populated automatically
-- when inserting or updating records in the detalle_pedido table.

CREATE OR REPLACE FUNCTION set_precio_detalle_pedido()
RETURNS TRIGGER AS $$
BEGIN
    NEW.precio := (
        SELECT precio FROM producto WHERE pk_id_producto = NEW.pk_id_producto
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS tr_set_precio_detalle_pedido ON detalle_pedido;

CREATE TRIGGER tr_set_precio_detalle_pedido
BEFORE INSERT OR UPDATE ON detalle_pedido
FOR EACH ROW
EXECUTE FUNCTION set_precio_detalle_pedido();
