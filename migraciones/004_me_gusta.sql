-- Migración 004: "me gusta" en publicaciones y comentarios.
-- Solo AGREGA una tabla nueva. No modifica datos existentes.

CREATE TABLE IF NOT EXISTS me_gusta (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    usuario_id      uuid NOT NULL REFERENCES usuarios(id) ON DELETE CASCADE,
    publicacion_id  uuid REFERENCES publicaciones(id) ON DELETE CASCADE,
    comentario_id   uuid REFERENCES comentarios(id) ON DELETE CASCADE,
    creado_en       timestamptz NOT NULL DEFAULT now(),
    -- Cada "me gusta" apunta a UNA sola cosa: una publicación o un comentario
    CONSTRAINT me_gusta_un_destino CHECK (num_nonnulls(publicacion_id, comentario_id) = 1)
);

-- Un usuario da como máximo un "me gusta" por publicación y por comentario
CREATE UNIQUE INDEX IF NOT EXISTS me_gusta_usuario_publicacion_idx
    ON me_gusta (usuario_id, publicacion_id) WHERE publicacion_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS me_gusta_usuario_comentario_idx
    ON me_gusta (usuario_id, comentario_id) WHERE comentario_id IS NOT NULL;
