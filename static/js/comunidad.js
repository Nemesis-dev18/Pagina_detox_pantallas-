  const lista = document.getElementById('lista');
    const vacio = document.getElementById('vacio');
    const buscador = document.getElementById('buscador');
    const btnCrear = document.getElementById('btn-crear');
    const btnCerrar = document.getElementById('btn-cerrar');
    const modal = document.getElementById('modal-publicar');
    const formulario = document.getElementById('form-publicar');
    const pideLogin = document.getElementById('pide-login');
    const miAlias = document.getElementById('mi-alias');
    const mensaje = document.getElementById('mensaje');
    const toast = document.getElementById('toast');
    const filtros = document.querySelectorAll('.filtro');

    const ETIQUETAS = { experiencia: 'Experiencia', pregunta: 'Pregunta', consejo: 'Consejo' };
    const ICONO_COMENTAR = '<svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M21 12a8 8 0 0 1-11.6 7.1L4 20l1-4.6A8 8 0 1 1 21 12z"/></svg>';
    const ICONO_COMPARTIR = '<svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M4 12v7a1 1 0 0 0 1 1h14a1 1 0 0 0 1-1v-7"/><path d="M16 6l-4-4-4 4"/><path d="M12 2v14"/></svg>';
    const ICONO_ELIMINAR = '<svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><polyline points="3 6 5 6 21 6"/><path d="M19 6l-1 14a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2L5 6"/><path d="M10 11v6"/><path d="M14 11v6"/><path d="M9 6V4a2 2 0 0 1 2-2h2a2 2 0 0 1 2 2v2"/></svg>';

    const ICONO_EDITAR = '<svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M12 20h9"/><path d="M16.5 3.5a2.1 2.1 0 0 1 3 3L7 19l-4 1 1-4 12.5-12.5z"/></svg>';

    // Estilos del modo edición (puedes moverlos a css/styles.css si prefieres)
    const estiloEdicion = document.createElement('style');
    estiloEdicion.textContent = `
    .editor { display: flex; flex-direction: column; gap: 8px; margin: 8px 0; }
    .editor input, .editor select, .editor textarea { width: 100%; box-sizing: border-box; padding: 8px 10px; font: inherit; border: 1px solid #ccc; border-radius: 6px; }
    .editor .editor_botones { display: flex; gap: 8px; }
    .editor .btn_cancelar { padding: 8px 14px; border: 1px solid #ccc; border-radius: 6px; background: #fff; cursor: pointer; font: inherit; }
    .link_editar { background: none; border: 0; padding: 2px 0; color: inherit; opacity: .7; text-decoration: underline; cursor: pointer; font: inherit; font-size: .85rem; }
    .link_editar:hover { opacity: 1; }`;
    document.head.appendChild(estiloEdicion);

    let haySesion = false;
    let soyAdmin = false;
    let filtroCategoria = '';
    let elementos = [];          // { id, art, categoria, texto, alternar }
    let hashRevisado = false;
    let temporizadorAviso;

    /* ---------- Utilidades ---------- */
    function crearElemento(etiqueta, clase, texto) {
        const el = document.createElement(etiqueta);
        if (clase) el.className = clase;
        if (texto !== undefined) el.textContent = texto;
        return el;
    }

    function crearBotonIcono(etiqueta, svg) {
        const boton = crearElemento('button', 'icono_btn');
        boton.type = 'button';
        boton.setAttribute('aria-label', etiqueta);
        boton.title = etiqueta;
        boton.innerHTML = svg;   // solo SVG fijo de arriba, nunca texto de usuarios
        return boton;
    }

    function formatearFecha(iso) {
        return new Date(iso).toLocaleString('es-CO');
    }

    function hace(iso) {
        const seg = Math.max(0, Math.floor((Date.now() - new Date(iso).getTime()) / 1000));
        if (seg < 60) return 'Hace un momento';
        const min = Math.floor(seg / 60);
        if (min < 60) return 'Hace ' + min + (min === 1 ? ' minuto' : ' minutos');
        const h = Math.floor(min / 60);
        if (h < 24) return 'Hace ' + h + (h === 1 ? ' hora' : ' horas');
        const d = Math.floor(h / 24);
        if (d < 7) return 'Hace ' + d + (d === 1 ? ' día' : ' días');
        return new Date(iso).toLocaleDateString('es-CO');
    }

    function normalizar(texto) {
        return texto.toLowerCase().normalize('NFD').replace(/[\u0300-\u036f]/g, '');
    }

    function avisar(texto) {
        toast.textContent = texto;
        toast.classList.add('visible');
        clearTimeout(temporizadorAviso);
        temporizadorAviso = setTimeout(() => toast.classList.remove('visible'), 2200);
    }

    /* ---------- Sesión ---------- */
    async function comprobarSesion() {
        try {
            const res = await fetch('/yo');
            if (res.ok) {
                const datos = await res.json();
                miAlias.textContent = datos.alias;
                soyAdmin = datos.es_admin;
                pideLogin.hidden = true;
                haySesion = true;
            }
        } catch (err) {
            // sin conexión con el servidor: se queda como visitante
        }
    }

    /* ---------- Comentarios ---------- */
    async function cargarComentarios(publicacionId, contenedor) {
        const res = await fetch('/publicaciones/' + publicacionId + '/comentarios');
        const comentarios = (await res.json()) || [];

        contenedor.replaceChildren();
        if (comentarios.length === 0) {
            contenedor.append(crearElemento('p', 'meta', 'Aún no hay respuestas.'));
            return;
        }
        for (const c of comentarios) {
            const item = crearElemento('div', 'comentario');
            const meta = crearElemento('p', 'meta',
                c.alias + ' · ' + formatearFecha(c.creado_en) + (c.editado_en ? ' · editado' : ''));
            const cuerpo = crearElemento('p', 'cuerpo', c.contenido);
            item.append(meta, cuerpo);

            // El autor del comentario (o un admin) puede editarlo
            if (haySesion && (c.alias === miAlias.textContent || soyAdmin)) {
                const btnEditar = crearElemento('button', 'link_editar', 'Editar');
                btnEditar.type = 'button';
                btnEditar.addEventListener('click', () => {
                    editarComentario(publicacionId, c, item, cuerpo, btnEditar, contenedor);
                });
                item.append(btnEditar);
            }
            contenedor.append(item);
        }
    }

    function editarComentario(publicacionId, c, item, cuerpo, btnEditar, contenedor) {
        const form = crearElemento('form', 'editor');
        const caja = document.createElement('textarea');
        caja.name = 'contenido';
        caja.rows = 3;
        caja.maxLength = 2000;
        caja.required = true;
        caja.value = c.contenido;

        const botones = crearElemento('div', 'editor_botones');
        const guardar = crearElemento('button', 'boton_rojo', 'Guardar');
        guardar.type = 'submit';
        const cancelar = crearElemento('button', 'btn_cancelar', 'Cancelar');
        cancelar.type = 'button';
        botones.append(guardar, cancelar);
        const aviso = crearElemento('p', 'meta');
        form.append(caja, botones, aviso);

        cuerpo.hidden = true;
        btnEditar.hidden = true;
        item.append(form);
        caja.focus();

        cancelar.addEventListener('click', () => {
            form.remove();
            cuerpo.hidden = false;
            btnEditar.hidden = false;
        });

        form.addEventListener('submit', async (e) => {
            e.preventDefault();
            guardar.disabled = true;
            try {
                const res = await fetch('/publicaciones/' + publicacionId + '/comentarios/' + c.id, {
                    method: 'PUT',
                    body: new URLSearchParams(new FormData(form))
                });
                if (res.ok) {
                    cargarComentarios(publicacionId, contenedor);
                    avisar('Comentario actualizado');
                } else {
                    aviso.textContent = await res.text();
                }
            } catch (err) {
                aviso.textContent = 'No hay conexión con el servidor.';
            } finally {
                guardar.disabled = false;
            }
        });
    }

    function crearBloqueRespuestas(publicacionId) {
        const bloque = crearElemento('div', 'respuestas');

        const listaComentarios = crearElemento('div');
        bloque.append(listaComentarios);

        if (haySesion) {
            const form = document.createElement('form');
            const caja = document.createElement('textarea');
            caja.name = 'contenido';
            caja.rows = 3;
            caja.maxLength = 2000;
            caja.required = true;
            caja.placeholder = 'Escribe tu respuesta...';
            const boton = crearElemento('button', 'boton_rojo', 'Responder');
            boton.type = 'submit';
            const aviso = crearElemento('p', 'meta');
            form.append(caja, boton, aviso);

            form.addEventListener('submit', async (e) => {
                e.preventDefault();
                const res = await fetch('/publicaciones/' + publicacionId + '/comentarios', {
                    method: 'POST',
                    body: new URLSearchParams(new FormData(form))
                });
                if (res.ok) {
                    form.reset();
                    aviso.textContent = '';
                    cargarComentarios(publicacionId, listaComentarios);
                } else {
                    aviso.textContent = await res.text();
                }
            });
            bloque.append(form);
        } else {
            bloque.append(crearElemento('p', 'meta', 'Inicia sesión para responder.'));
        }
        return bloque;
    }

    /* ---------- Publicaciones ---------- */
 function crearPublicacion(p) {
    const art = crearElemento('article', 'publicacion');

    art.id = 'p-' + p.id;

    const cabecera = crearElemento('div', 'pub_cabecera');
    const avatar = crearElemento('div', 'avatar', (p.alias || '?').charAt(0).toUpperCase());
    avatar.setAttribute('aria-hidden', 'true');

    const info = crearElemento('div', 'pub_info');
    const titulo = crearElemento('button', 'pub_titulo', p.titulo);
    titulo.type = 'button';
    titulo.setAttribute('aria-expanded', 'false');
    titulo.setAttribute('aria-controls', 'detalle-' + p.id);
    const meta = crearElemento('p', 'meta');
    meta.title = formatearFecha(p.creado_en);
    info.append(titulo, meta);

    const acciones = crearElemento('div', 'acciones');
    const btnResponder = crearBotonIcono('Ver respuestas', ICONO_COMENTAR);
    btnResponder.setAttribute('aria-expanded', 'false');
    btnResponder.setAttribute('aria-controls', 'detalle-' + p.id);
    const btnCompartir = crearBotonIcono('Copiar enlace', ICONO_COMPARTIR);
    acciones.append(btnResponder, btnCompartir);

    cabecera.append(avatar, info, acciones);

    const detalle = crearElemento('div', 'pub_detalle');
    detalle.id = 'detalle-' + p.id;
    detalle.hidden = true;
    const bloque = crearBloqueRespuestas(p.id);
    const listaComentarios = bloque.firstChild;
    const cuerpoPost = crearElemento('p', 'cuerpo', p.contenido);
    detalle.append(cuerpoPost, bloque);

    const entrada = { id: p.id, art: art, categoria: p.categoria, texto: '', alternar: null };
    function pintarMeta() {
        meta.textContent = p.alias + ' · ' + (ETIQUETAS[p.categoria] || p.categoria) + ' · ' + hace(p.creado_en)
            + (p.editado_en ? ' · editado' : '');
        entrada.categoria = p.categoria;
        entrada.texto = normalizar(p.titulo + ' ' + p.contenido + ' ' + p.alias);
    }
    pintarMeta();

    // El autor (o un admin) puede editar la publicación
    if (haySesion && (p.alias === miAlias.textContent || soyAdmin)) {
        const btnEditar = crearBotonIcono('Editar', ICONO_EDITAR);
        btnEditar.addEventListener('click', () => {
            if (detalle.querySelector('.editor_post')) return;   // ya está editando
            alternar(true);

            const form = crearElemento('form', 'editor editor_post');
            const campoTitulo = document.createElement('input');
            campoTitulo.name = 'titulo';
            campoTitulo.maxLength = 120;
            campoTitulo.required = true;
            campoTitulo.value = p.titulo;
            const campoCategoria = document.createElement('select');
            campoCategoria.name = 'categoria';
            for (const clave in ETIQUETAS) {
                const op = crearElemento('option', '', ETIQUETAS[clave]);
                op.value = clave;
                op.selected = clave === p.categoria;
                campoCategoria.append(op);
            }
            const campoContenido = document.createElement('textarea');
            campoContenido.name = 'contenido';
            campoContenido.rows = 6;
            campoContenido.maxLength = 5000;
            campoContenido.required = true;
            campoContenido.value = p.contenido;

            const botones = crearElemento('div', 'editor_botones');
            const guardar = crearElemento('button', 'boton_rojo', 'Guardar cambios');
            guardar.type = 'submit';
            const cancelar = crearElemento('button', 'btn_cancelar', 'Cancelar');
            cancelar.type = 'button';
            botones.append(guardar, cancelar);
            const aviso = crearElemento('p', 'meta');
            form.append(campoTitulo, campoCategoria, campoContenido, botones, aviso);

            cuerpoPost.hidden = true;
            detalle.insertBefore(form, cuerpoPost);
            campoTitulo.focus();

            cancelar.addEventListener('click', () => {
                form.remove();
                cuerpoPost.hidden = false;
            });

            form.addEventListener('submit', async (e) => {
                e.preventDefault();
                guardar.disabled = true;
                try {
                    const res = await fetch('/publicaciones/' + p.id, {
                        method: 'PUT',
                        body: new URLSearchParams(new FormData(form))
                    });
                    if (res.ok) {
                        p.titulo = campoTitulo.value.trim();
                        p.categoria = campoCategoria.value;
                        p.contenido = campoContenido.value.trim();
                        titulo.textContent = p.titulo;
                        cuerpoPost.textContent = p.contenido;
                        pintarMeta();
                        form.remove();
                        cuerpoPost.hidden = false;
                        aplicarFiltros();
                        avisar('Publicación actualizada');
                    } else {
                        aviso.textContent = await res.text();
                    }
                } catch (err) {
                    aviso.textContent = 'No hay conexión con el servidor.';
                } finally {
                    guardar.disabled = false;
                }
            });
        });
        acciones.append(btnEditar);
    }

    // El autor (o un admin) puede borrar
    if (haySesion && (p.alias === miAlias.textContent || soyAdmin)) {
        const btnBorrar = crearBotonIcono('Eliminar', ICONO_ELIMINAR);
        btnBorrar.addEventListener('click', async () => {
            if (!confirm('¿Eliminar esta publicación?')) return;
            const res = await fetch('/publicaciones/' + p.id, { method: 'DELETE' });
            if (res.ok) {
                cargarPublicaciones();
            } else {
                alert(await res.text());
            }
        });
        acciones.append(btnBorrar);
    }

    function alternar(forzar) {
        const abrir = forzar === undefined ? detalle.hidden : forzar;
        detalle.hidden = !abrir;
        titulo.setAttribute('aria-expanded', String(abrir));
        btnResponder.setAttribute('aria-expanded', String(abrir));
        btnResponder.title = abrir ? 'Ocultar respuestas' : 'Ver respuestas';
        if (abrir) cargarComentarios(p.id, listaComentarios);
    }

    titulo.addEventListener('click', () => alternar());
    btnResponder.addEventListener('click', () => alternar());
    btnCompartir.addEventListener('click', async () => {
        const enlace = location.href.split('#')[0] + '#p-' + p.id;
        try {
            await navigator.clipboard.writeText(enlace);
            avisar('Enlace copiado');
        } catch (err) {
            window.prompt('Copia este enlace:', enlace);
        }
    });
    art.append(cabecera, detalle);
    entrada.alternar = alternar;
    return entrada;
}

    function aplicarFiltros() {
    const consulta = normalizar(buscador.value.trim());
    let visibles = 0;
    for (const e of elementos) {
        const coincide = (!filtroCategoria || e.categoria === filtroCategoria)
            && (!consulta || e.texto.includes(consulta));
        e.art.hidden = !coincide;
        if (coincide) visibles++;
    }
    lista.hidden = visibles === 0;
    if (elementos.length === 0) {
        vacio.textContent = 'Aún no hay publicaciones. ¡Sé la primera persona!';
        vacio.hidden = false;
    } else if (visibles === 0) {
        vacio.textContent = 'No encontramos publicaciones con ese filtro.';
        vacio.hidden = false;
    } else {
        vacio.hidden = true;
    }
}

function abrirDesdeHash() {
    if (hashRevisado) return;
    hashRevisado = true;
    if (!location.hash.startsWith('#p-')) return;
    const id = location.hash.slice(3);
    const e = elementos.find((x) => x.id === id);
    if (e) {
        e.alternar(true);
        e.art.scrollIntoView({ block: 'center' });
    }
}

async function cargarPublicaciones() {
    let publicaciones = [];
    try {
        const res = await fetch('/publicaciones');
        publicaciones = (await res.json()) || [];
    } catch (err) {
        vacio.textContent = 'No se pudieron cargar las publicaciones.';
        vacio.hidden = false;
        lista.hidden = true;
        return;
    }

    lista.replaceChildren();
    elementos = publicaciones.map(crearPublicacion);
    for (const e of elementos) lista.append(e.art);
    aplicarFiltros();
    abrirDesdeHash();

}
    /* ---------- Filtros, búsqueda y ventana de crear ---------- */
    for (const boton of filtros) {
        boton.addEventListener('click', () => {
            filtroCategoria = boton.dataset.categoria;
            for (const b of filtros) {
                const activo = b === boton;
                b.classList.toggle('activo', activo);
                b.setAttribute('aria-pressed', String(activo));
            }
            aplicarFiltros();
        });
    }

    buscador.addEventListener('input', aplicarFiltros);

    btnCrear.addEventListener('click', () => {
        if (!haySesion) {
            window.location.href = 'inicio_sesion.html';
            return;
        }
        mensaje.textContent = '';
        modal.showModal();
        document.getElementById('titulo').focus();
    });

    btnCerrar.addEventListener('click', () => modal.close());
    modal.addEventListener('click', (e) => {
        if (e.target === modal) modal.close();   // clic en el fondo oscuro
    });

    formulario.addEventListener('submit', async (e) => {
        e.preventDefault();
        const res = await fetch('/publicaciones', {
            method: 'POST',
            body: new URLSearchParams(new FormData(formulario))
        });
        if (res.ok) {
            formulario.reset();
            mensaje.textContent = '';
            modal.close();
            cargarPublicaciones();
        } else {
            mensaje.textContent = await res.text();
        }
    });

    // Primero sabemos si hay sesión, después dibujamos las publicaciones
    (async function iniciar() {
        await comprobarSesion();
        await cargarPublicaciones();
    })();