// Directorio de especialistas APROBADOS, con el mismo formato de tarjeta que Comunidad.
// El servidor filtra (GET /especialistas); aquí solo se pinta. Todo va con textContent.
(function () {
    const lista = document.getElementById('lista');
    const vacio = document.getElementById('vacio');
    const buscador = document.getElementById('buscar-especialista');
    const filtros = document.querySelectorAll('.filtro');
    let profesion = '';
    let temporizador;

    function elemento(etiqueta, clase, texto) {
        const el = document.createElement(etiqueta);
        if (clase) el.className = clase;
        if (texto !== undefined) el.textContent = texto;
        return el;
    }

    // Fila con el mismo esquema de una publicación: avatar con inicial, nombre y datos
    function fila(e) {
        const art = elemento('article', 'publicacion');

        const cabecera = elemento('div', 'pub_cabecera');
        const avatar = elemento('div', 'avatar', (e.nombre || '?').charAt(0).toUpperCase());
        avatar.setAttribute('aria-hidden', 'true');

        const info = elemento('div', 'pub_info');
        const nombre = elemento('p', 'pub_titulo', e.nombre);
        const insignia = elemento('span', 'insignia', '✓ Profesional verificado');
        const meta = elemento('p', 'meta');
        meta.textContent = e.profesion + ' · ' + e.especialidad;
        const institucion = elemento('p', 'meta', 'Institución: ' + e.institucion);
        const alias = elemento('p', 'meta', 'En la comunidad: ' + e.alias);
        info.append(nombre, insignia, meta, institucion, alias);

        cabecera.append(avatar, info);
        art.append(cabecera);
        return art;
    }

    async function cargar() {
        const params = new URLSearchParams();
        if (profesion) params.set('profesion', profesion);
        const q = buscador.value.trim();
        if (q) params.set('q', q);

        try {
            const res = await fetch('/especialistas?' + params.toString());
            if (!res.ok) {
                vacio.textContent = await res.text();
                vacio.hidden = false;
                lista.replaceChildren();
                return;
            }
            const datos = (await res.json()) || [];
            lista.replaceChildren(...datos.map(fila));
            vacio.textContent = 'No hay especialistas aprobados con esos filtros.';
            vacio.hidden = datos.length > 0;
        } catch (err) {
            vacio.textContent = 'No hay conexión con el servidor.';
            vacio.hidden = false;
        }
    }

    filtros.forEach(boton => {
        boton.addEventListener('click', () => {
            profesion = boton.dataset.profesion;
            filtros.forEach(b => {
                const activo = b === boton;
                b.classList.toggle('activo', activo);
                b.setAttribute('aria-pressed', String(activo));
            });
            cargar();
        });
    });

    buscador.addEventListener('input', () => {
        clearTimeout(temporizador);
        temporizador = setTimeout(cargar, 300);
    });

    cargar();
})();
