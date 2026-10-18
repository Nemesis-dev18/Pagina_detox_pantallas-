-- Migración 002: centros (instituciones) de cada especialista.
-- Solo AGREGA una tabla nueva y copia los datos que ya existen en perfiles_especialistas.
-- No borra ni modifica columnas existentes.

CREATE TABLE IF NOT EXISTS especialista_centros (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    usuario_id    uuid NOT NULL REFERENCES perfiles_especialistas(usuario_id) ON DELETE CASCADE,
    institucion   text NOT NULL CHECK (char_length(institucion) BETWEEN 3 AND 150),
    es_principal  boolean NOT NULL DEFAULT false,
    creado_en     timestamptz NOT NULL DEFAULT now(),
    UNIQUE (usuario_id, institucion)
);
CREATE INDEX IF NOT EXISTS especialista_centros_usuario_idx ON especialista_centros (usuario_id);

-- Copia la institución principal de quienes ya están registrados (idempotente)
INSERT INTO especialista_centros (usuario_id, institucion, es_principal)
SELECT pe.usuario_id, pe.institucion, true
FROM perfiles_especialistas pe
ON CONFLICT (usuario_id, institucion) DO NOTHING;
