-- Migración 003: registro de ediciones del perfil de especialista (límite mensual).
-- Solo AGREGA una tabla nueva. No modifica datos existentes.

CREATE TABLE IF NOT EXISTS perfil_ediciones (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    usuario_id   uuid NOT NULL REFERENCES perfiles_especialistas(usuario_id) ON DELETE CASCADE,
    creado_en    timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS perfil_ediciones_usuario_fecha_idx ON perfil_ediciones (usuario_id, creado_en);
