// Directorio de especialistas APROBADOS. El servidor filtra (GET /especialistas);
// aquí solo se pinta. Todo se escribe con textContent, nunca como HTML.
(function () {
    const lista = document.getElementById('lista');
    const estado = document.getElementById('estado');
    const filtroProfesion = document.getElementById('filtro-profesion');
    const buscador = document.getElementById('buscar-especialista');
    let temporizador;

    function tarjeta(e) {
        const li = document.createElement('li');
        const titulo = document.createElement('h3');
        titulo.textContent = e.nombre;

        const insignia = document.createElement('span');
        insignia.className = 'insignia';
        insignia.textContent = '✓ Profesional verificado';

        const profesion = document.createElement('p');
        profesion.textContent = e.profesion + ' · ' + e.especialidad;

        const institucion = document.createElement('p');
        institucion.textContent = 'Institución: ' + e.institucion;

        const alias = document.createElement('p');
        alias.textContent = 'En la comunidad: ' + e.alias;

        li.append(titulo, insignia, profesion, institucion, alias);
        return li;
    }

    async function cargar() {
        const params = new URLSearchParams();
        if (filtroProfesion.value) params.set('profesion', filtroProfesion.value);
        const q = buscador.value.trim();
        if (q) params.set('q', q);

        estado.textContent = 'Cargando…';
        try {
            const res = await fetch('/especialistas?' + params.toString());
            if (!res.ok) {
                estado.textContent = await res.text();
                lista.replaceChildren();
                return;
            }
            const datos = (await res.json()) || [];
            lista.replaceChildren(...datos.map(tarjeta));
            estado.textContent = datos.length === 0
                ? 'No hay especialistas aprobados con esos filtros.'
                : '';
        } catch (err) {
            estado.textContent = 'No hay conexión con el servidor.';
        }
    }

    filtroProfesion.addEventListener('change', cargar);
    buscador.addEventListener('input', () => {
        clearTimeout(temporizador);
        temporizador = setTimeout(cargar, 300);
    });
    cargar();
})();
