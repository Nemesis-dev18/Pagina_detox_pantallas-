--
-- PostgreSQL database dump
--

\restrict f401KLa26ItdlhA8bFm7Q8H7l6y2bkVD8DlnsS2558jdPTuMuou0ih39F8x20uX

-- Dumped from database version 17.11 (Debian 17.11-1.pgdg13+2)
-- Dumped by pg_dump version 17.11 (Debian 17.11-1.pgdg13+2)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: pgcrypto; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public;


--
-- Name: EXTENSION pgcrypto; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION pgcrypto IS 'cryptographic functions';


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: bloqueos_comentarios; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.bloqueos_comentarios (
    usuario_id uuid NOT NULL,
    hasta timestamp with time zone NOT NULL
);


--
-- Name: coincidencias_censura; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.coincidencias_censura (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    comentario_id uuid NOT NULL,
    palabra text NOT NULL,
    fragmento text NOT NULL,
    creado_en timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: comentarios; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.comentarios (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    publicacion_id uuid NOT NULL,
    autor_id uuid NOT NULL,
    contenido text NOT NULL,
    creado_en timestamp without time zone DEFAULT now() NOT NULL,
    editado_en timestamp without time zone
);


--
-- Name: palabras_excluidas; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.palabras_excluidas (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    palabra text NOT NULL,
    creado_en timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: palabras_prohibidas; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.palabras_prohibidas (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    palabra text NOT NULL,
    creado_en timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: publicaciones; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.publicaciones (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    autor_id uuid NOT NULL,
    titulo text NOT NULL,
    contenido text NOT NULL,
    categoria text,
    creado_en timestamp without time zone DEFAULT now() NOT NULL,
    editado_en timestamp without time zone
);


--
-- Name: recuperaciones; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.recuperaciones (
    token_hash text NOT NULL,
    usuario_id uuid NOT NULL,
    expira_en timestamp without time zone NOT NULL,
    usado boolean DEFAULT false NOT NULL
);


--
-- Name: respuestas_cuestionario; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.respuestas_cuestionario (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    usuario_id uuid NOT NULL,
    pregunta text NOT NULL,
    respuesta text NOT NULL,
    creado_en timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: sesiones; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sesiones (
    token_hash text NOT NULL,
    usuario_id uuid NOT NULL,
    creado_en timestamp without time zone DEFAULT now() NOT NULL,
    expira_en timestamp without time zone NOT NULL
);


--
-- Name: usuarios; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.usuarios (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    alias text NOT NULL,
    correo text NOT NULL,
    password_hash text NOT NULL,
    creado_en timestamp without time zone DEFAULT now() NOT NULL,
    es_admin boolean DEFAULT false NOT NULL
);


--
-- Name: bloqueos_comentarios bloqueos_comentarios_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.bloqueos_comentarios
    ADD CONSTRAINT bloqueos_comentarios_pkey PRIMARY KEY (usuario_id);


--
-- Name: coincidencias_censura coincidencias_censura_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.coincidencias_censura
    ADD CONSTRAINT coincidencias_censura_pkey PRIMARY KEY (id);


--
-- Name: comentarios comentarios_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.comentarios
    ADD CONSTRAINT comentarios_pkey PRIMARY KEY (id);


--
-- Name: palabras_excluidas palabras_excluidas_palabra_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.palabras_excluidas
    ADD CONSTRAINT palabras_excluidas_palabra_key UNIQUE (palabra);


--
-- Name: palabras_excluidas palabras_excluidas_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.palabras_excluidas
    ADD CONSTRAINT palabras_excluidas_pkey PRIMARY KEY (id);


--
-- Name: palabras_prohibidas palabras_prohibidas_palabra_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.palabras_prohibidas
    ADD CONSTRAINT palabras_prohibidas_palabra_key UNIQUE (palabra);


--
-- Name: palabras_prohibidas palabras_prohibidas_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.palabras_prohibidas
    ADD CONSTRAINT palabras_prohibidas_pkey PRIMARY KEY (id);


--
-- Name: publicaciones publicaciones_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.publicaciones
    ADD CONSTRAINT publicaciones_pkey PRIMARY KEY (id);


--
-- Name: recuperaciones recuperaciones_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.recuperaciones
    ADD CONSTRAINT recuperaciones_pkey PRIMARY KEY (token_hash);


--
-- Name: respuestas_cuestionario respuestas_cuestionario_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.respuestas_cuestionario
    ADD CONSTRAINT respuestas_cuestionario_pkey PRIMARY KEY (id);


--
-- Name: sesiones sesiones_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sesiones
    ADD CONSTRAINT sesiones_pkey PRIMARY KEY (token_hash);


--
-- Name: usuarios usuarios_correo_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.usuarios
    ADD CONSTRAINT usuarios_correo_key UNIQUE (correo);


--
-- Name: usuarios usuarios_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.usuarios
    ADD CONSTRAINT usuarios_pkey PRIMARY KEY (id);


--
-- Name: comentarios_autor_fecha; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX comentarios_autor_fecha ON public.comentarios USING btree (autor_id, creado_en DESC);


--
-- Name: idx_coincidencias_comentario; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_coincidencias_comentario ON public.coincidencias_censura USING btree (comentario_id);


--
-- Name: idx_comentarios_autor_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_comentarios_autor_id ON public.comentarios USING btree (autor_id);


--
-- Name: idx_comentarios_publicacion_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_comentarios_publicacion_id ON public.comentarios USING btree (publicacion_id);


--
-- Name: idx_publicaciones_autor_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_publicaciones_autor_id ON public.publicaciones USING btree (autor_id);


--
-- Name: idx_recuperaciones_usuario_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_recuperaciones_usuario_id ON public.recuperaciones USING btree (usuario_id);


--
-- Name: idx_sesiones_usuario_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sesiones_usuario_id ON public.sesiones USING btree (usuario_id);


--
-- Name: usuarios_alias_lower_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX usuarios_alias_lower_idx ON public.usuarios USING btree (lower(alias));


--
-- Name: bloqueos_comentarios bloqueos_comentarios_usuario_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.bloqueos_comentarios
    ADD CONSTRAINT bloqueos_comentarios_usuario_id_fkey FOREIGN KEY (usuario_id) REFERENCES public.usuarios(id) ON DELETE CASCADE;


--
-- Name: coincidencias_censura coincidencias_censura_comentario_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.coincidencias_censura
    ADD CONSTRAINT coincidencias_censura_comentario_id_fkey FOREIGN KEY (comentario_id) REFERENCES public.comentarios(id) ON DELETE CASCADE;


--
-- Name: comentarios comentarios_autor_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.comentarios
    ADD CONSTRAINT comentarios_autor_id_fkey FOREIGN KEY (autor_id) REFERENCES public.usuarios(id);


--
-- Name: comentarios comentarios_publicacion_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.comentarios
    ADD CONSTRAINT comentarios_publicacion_id_fkey FOREIGN KEY (publicacion_id) REFERENCES public.publicaciones(id) ON DELETE CASCADE;


--
-- Name: publicaciones publicaciones_autor_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.publicaciones
    ADD CONSTRAINT publicaciones_autor_id_fkey FOREIGN KEY (autor_id) REFERENCES public.usuarios(id);


--
-- Name: recuperaciones recuperaciones_usuario_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.recuperaciones
    ADD CONSTRAINT recuperaciones_usuario_id_fkey FOREIGN KEY (usuario_id) REFERENCES public.usuarios(id) ON DELETE CASCADE;


--
-- Name: respuestas_cuestionario respuestas_cuestionario_usuario_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.respuestas_cuestionario
    ADD CONSTRAINT respuestas_cuestionario_usuario_id_fkey FOREIGN KEY (usuario_id) REFERENCES public.usuarios(id);


--
-- Name: sesiones sesiones_usuario_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sesiones
    ADD CONSTRAINT sesiones_usuario_id_fkey FOREIGN KEY (usuario_id) REFERENCES public.usuarios(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

\unrestrict f401KLa26ItdlhA8bFm7Q8H7l6y2bkVD8DlnsS2558jdPTuMuou0ih39F8x20uX

