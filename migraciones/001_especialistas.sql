-- Migración 001: especialistas verificados y acompañamiento.
-- Solo AGREGA (columna con valor por defecto, tablas nuevas). No borra ni altera datos existentes.
-- Es idempotente: se puede ejecutar más de una vez sin efecto.

ALTER TABLE usuarios ADD COLUMN IF NOT EXISTS tipo text NOT NULL DEFAULT 'usuario';

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'usuarios_tipo_chk') THEN
    ALTER TABLE usuarios ADD CONSTRAINT usuarios_tipo_chk CHECK (tipo IN ('usuario', 'especialista'));
  END IF;
END $$;

CREATE TABLE IF NOT EXISTS perfiles_especialistas (
    usuario_id          uuid PRIMARY KEY REFERENCES usuarios(id) ON DELETE CASCADE,
    nombres             text NOT NULL,
    primer_apellido     text NOT NULL,
    segundo_apellido    text,
    profesion           text NOT NULL,
    especialidad        text NOT NULL,
    institucion         text NOT NULL,
    institucion_nit     text,
    tipo_documento      text NOT NULL CHECK (tipo_documento IN ('CC', 'CE', 'PPT')),
    -- El documento se guarda CIFRADO (AES-256-GCM); la huella (HMAC) solo sirve para evitar duplicados.
    documento_cifrado   bytea NOT NULL,
    documento_huella    text NOT NULL UNIQUE,
    numero_registro     text NOT NULL,
    descripcion         text,
    -- estado general y verificaciones por separado (credencial / institución)
    estado              text NOT NULL DEFAULT 'pendiente'
                        CHECK (estado IN ('pendiente', 'aprobado', 'rechazado', 'suspendido')),
    credencial_estado   text NOT NULL DEFAULT 'pendiente'
                        CHECK (credencial_estado IN ('pendiente', 'verificada', 'rechazada')),
    institucion_estado  text NOT NULL DEFAULT 'pendiente'
                        CHECK (institucion_estado IN ('pendiente', 'verificada', 'rechazada')),
    notas_revision      text,
    motivo              text,
    evidencia_institucion jsonb,
    revisado_por        uuid REFERENCES usuarios(id) ON DELETE SET NULL,
    revisado_en         timestamptz,
    -- Constancia de las autorizaciones (Ley 1581 de 2012)
    acepta_datos_en     timestamptz NOT NULL,
    acepta_perfil_en    timestamptz NOT NULL,
    creado_en           timestamptz NOT NULL DEFAULT now(),
    -- La base misma impide "aprobado" sin las dos verificaciones y sin revisión
    CONSTRAINT perfiles_aprobado_chk CHECK (
        estado <> 'aprobado'
        OR (credencial_estado = 'verificada' AND institucion_estado = 'verificada' AND revisado_en IS NOT NULL)
    )
);
CREATE INDEX IF NOT EXISTS perfiles_estado_idx ON perfiles_especialistas (estado);

CREATE TABLE IF NOT EXISTS acompanamientos (
    id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    solicitante_id   uuid NOT NULL REFERENCES usuarios(id) ON DELETE CASCADE,
    especialista_id  uuid NOT NULL REFERENCES perfiles_especialistas(usuario_id) ON DELETE CASCADE,
    mensaje          text NOT NULL,
    estado           text NOT NULL DEFAULT 'pendiente'
                     CHECK (estado IN ('pendiente', 'aceptada', 'rechazada', 'cancelada', 'finalizada')),
    autorizado_en    timestamptz NOT NULL,
    creado_en        timestamptz NOT NULL DEFAULT now(),
    respondido_en    timestamptz,
    CHECK (solicitante_id <> especialista_id)
);
-- Una sola solicitud activa por pareja usuario-especialista
CREATE UNIQUE INDEX IF NOT EXISTS acompanamientos_activo_idx
    ON acompanamientos (solicitante_id, especialista_id) WHERE estado IN ('pendiente', 'aceptada');
CREATE INDEX IF NOT EXISTS acompanamientos_especialista_idx ON acompanamientos (especialista_id, estado);
CREATE INDEX IF NOT EXISTS acompanamientos_solicitante_idx ON acompanamientos (solicitante_id);

CREATE TABLE IF NOT EXISTS acompanamiento_mensajes (
    id                 uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    acompanamiento_id  uuid NOT NULL REFERENCES acompanamientos(id) ON DELETE CASCADE,
    autor_id           uuid NOT NULL REFERENCES usuarios(id) ON DELETE CASCADE,
    contenido          text NOT NULL,
    creado_en          timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS acompanamiento_mensajes_idx ON acompanamiento_mensajes (acompanamiento_id, creado_en);
