DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'pedido_delivery_requires_domicilio'
    ) THEN
        ALTER TABLE pedido
            ADD CONSTRAINT pedido_delivery_requires_domicilio
            CHECK (delivery = false OR pk_id_domicilio IS NOT NULL);
    END IF;
END;
$$;
