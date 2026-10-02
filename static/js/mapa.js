// Datos reales tomados de Google Maps (nombre, dirección y teléfono pueden
// cambiar con el tiempo; por eso el aviso en la página pide confirmar antes
// de ir). Se dejaron fuera lugares con reseñas que alertaban de maltrato o
// cobros irregulares, aunque aparecían en la búsqueda.
const LUGARES = [
    {
        nombre: 'Colectivo Aquí y Ahora | Salud Mental',
        categoria: 'ayuda',
        etiqueta: 'Centro de ayuda',
        direccion: 'Ak 9 #115-06 Of. 703, Bogotá',
        telefono: '+57 313 4249593',
        horario: 'Lun-Sáb 8:00 a.m. - 5:00 p.m.',
        lat: 4.6932577, lng: -74.0345322
    },
    {
        nombre: 'Psicosanando',
        categoria: 'psicologia',
        etiqueta: 'Psicología',
        direccion: 'Calle 125 #21a-50, Edificio Nova125, Usaquén, Bogotá',
        telefono: '+57 315 8025453',
        horario: 'Lun-Vie 7:00 a.m. - 8:00 p.m., Sáb 7:00 a.m. - 3:00 p.m.',
        lat: 4.7056387, lng: -74.0520778
    },
    {
        nombre: 'Cuerpo Arte y Palabra',
        categoria: 'psicologia',
        etiqueta: 'Psicología',
        direccion: 'Cra 20 #137-21, Usaquén, Bogotá',
        telefono: '+57 321 9465830',
        horario: 'Abierto 24 horas',
        lat: 4.7221888, lng: -74.048301
    },
    {
        nombre: 'Psicologa.co - Sede Chicó',
        categoria: 'psicologia',
        etiqueta: 'Psicología',
        direccion: 'Cra. 16 #80-77, Bogotá',
        telefono: '+57 314 8131016',
        horario: 'Lun-Vie 8:00 a.m. - 8:00 p.m., Sáb 8:00 a.m. - 3:00 p.m.',
        lat: 4.66757, lng: -74.0573533
    },
    {
        nombre: 'Los Mejores Psicólogos de Bogotá',
        categoria: 'psicologia',
        etiqueta: 'Psicología',
        direccion: 'Cra. 18a #39-11, Teusaquillo, Bogotá',
        telefono: '+57 320 2271548',
        horario: 'Lun-Vie 6:30 a.m. - 9:00 p.m., Sáb 7:00 a.m. - 6:00 p.m.',
        lat: 4.6275176, lng: -74.0725109
    },
    {
        nombre: 'Fundación Función Futuro · Salud Mental y Adicciones',
        categoria: 'rehabilitacion',
        etiqueta: 'Rehabilitación',
        direccion: 'Kr 13 #102-28, Bogotá',
        telefono: '+57 316 5362141',
        horario: 'Abierto 24 horas',
        lat: 4.685597, lng: -74.0444639
    },
    {
        nombre: 'Fundación Evoluciona',
        categoria: 'rehabilitacion',
        etiqueta: 'Rehabilitación',
        direccion: 'Cra. 82a #80-99, Engativá, Bogotá',
        telefono: '+57 314 3252595',
        horario: 'Abierto 24 horas',
        lat: 4.7022776, lng: -74.0976252
    },
    {
        nombre: 'Biblioteca Pública Virgilio Barco',
        categoria: 'desconexion',
        etiqueta: 'Lugar de desconexión',
        direccion: 'Av. La Esmeralda #57-60, Teusaquillo, Bogotá',
        telefono: '+57 601 5803010',
        horario: 'Mar-Sáb 8:00 a.m. - 7:00 p.m., Dom 10:00 a.m. - 5:00 p.m.',
        lat: 4.6571588, lng: -74.088429
    },
    {
        nombre: 'Biblioteca Julio Mario Santo Domingo',
        categoria: 'desconexion',
        etiqueta: 'Lugar de desconexión',
        direccion: 'Cl. 170 #67-51, Bogotá',
        telefono: '+57 601 5803080',
        horario: 'Mar-Sáb 8:00 a.m. - 7:00 p.m., Dom 10:00 a.m. - 5:00 p.m.',
        lat: 4.7565838, lng: -74.0626967
    },
    {
        nombre: 'Parque Metropolitano Bosque San Carlos',
        categoria: 'desconexion',
        etiqueta: 'Lugar de desconexión',
        direccion: 'Entre Cra. 13 y 13A con Cl. 31A y 31F Sur, Bogotá',
        telefono: null,
        horario: 'Todos los días 6:00 a.m. - 6:00 p.m.',
        lat: 4.5726952, lng: -74.1055157
    },
    {
        nombre: 'Biblioteca Pública El Parque',
        categoria: 'desconexion',
        etiqueta: 'Lugar de desconexión',
        direccion: 'Cra. 5 #36-21, Bogotá',
        telefono: '+57 601 5803010 ext. 6211',
        horario: 'Mar-Sáb 8:00 a.m. - 5:00 p.m., Dom 9:00 a.m. - 4:00 p.m.',
        lat: 4.6221402, lng: -74.063519
    }
];

const iconoRojo = L.divIcon({
    className: 'marcador_rojo',
    iconSize: [16, 16],
});

const mapa = L.map('mapa_leaflet', { scrollWheelZoom: false }).setView([4.6686, -74.0708], 12);

// OpenStreetMap estándar: gratis, sin clave. El mapa se oscurece con CSS
// (ver la regla .leaflet-tile-pane en el <style> de mapa.html).
L.tileLayer('https://tile.openstreetmap.org/{z}/{x}/{y}.png', {
    attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
    maxZoom: 19,
}).addTo(mapa);

const listaEl = document.getElementById('lista_lugares');
const buscadorEl = document.getElementById('buscador_mapa');
const chips = document.querySelectorAll('.chip');

let categoriaActual = '';
const marcadores = new Map();

function normalizar(texto) {
    return texto.toLowerCase().normalize('NFD').replace(/[\u0300-\u036f]/g, '');
}

function crearPopup(lugar) {
    const div = document.createElement('div');
    div.className = 'popup_lugar';

    const h3 = document.createElement('h3');
    h3.textContent = lugar.nombre;
    div.appendChild(h3);

    const dir = document.createElement('p');
    dir.textContent = lugar.direccion;
    div.appendChild(dir);

    const hor = document.createElement('p');
    hor.textContent = lugar.horario;
    div.appendChild(hor);

    if (lugar.telefono) {
        const tel = document.createElement('p');
        const a = document.createElement('a');
        a.href = 'tel:' + lugar.telefono.replace(/\s+/g, '');
        a.textContent = lugar.telefono;
        tel.appendChild(a);
        div.appendChild(tel);
    }

    return div;
}

function crearTarjeta(lugar) {
    const art = document.createElement('article');
    art.className = 'tarjeta_lugar';
    art.tabIndex = 0;

    const tag = document.createElement('span');
    tag.className = 'categoria_tag';
    tag.textContent = lugar.etiqueta;

    const h3 = document.createElement('h3');
    h3.textContent = lugar.nombre;

    const dir = document.createElement('p');
    dir.textContent = lugar.direccion;

    const hor = document.createElement('p');
    hor.textContent = lugar.horario;

    art.append(tag, h3, dir, hor);

    if (lugar.telefono) {
        const tel = document.createElement('p');
        tel.textContent = lugar.telefono;
        art.append(tel);
    }

    function abrir() {
        mapa.flyTo([lugar.lat, lugar.lng], 15, { duration: 0.6 });
        const m = marcadores.get(lugar.nombre);
        if (m) m.openPopup();
    }
    art.addEventListener('click', abrir);
    art.addEventListener('keydown', (e) => {
        if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); abrir(); }
    });

    return art;
}

for (const lugar of LUGARES) {
    const m = L.marker([lugar.lat, lugar.lng], { icon: iconoRojo }).addTo(mapa);
    m.bindPopup(crearPopup(lugar));
    marcadores.set(lugar.nombre, m);
}

function aplicarFiltro() {
    const consulta = normalizar(buscadorEl.value.trim());
    listaEl.replaceChildren();
    let algunoVisible = false;

    for (const lugar of LUGARES) {
        const coincideCategoria = !categoriaActual || lugar.categoria === categoriaActual;
        const coincideTexto = !consulta || normalizar(lugar.nombre + ' ' + lugar.direccion).includes(consulta);
        const visible = coincideCategoria && coincideTexto;

        const m = marcadores.get(lugar.nombre);
        if (visible) {
            if (!mapa.hasLayer(m)) m.addTo(mapa);
            listaEl.appendChild(crearTarjeta(lugar));
            algunoVisible = true;
        } else if (mapa.hasLayer(m)) {
            mapa.removeLayer(m);
        }
    }

    if (!algunoVisible) {
        const vacio = document.createElement('p');
        vacio.textContent = 'No encontramos lugares con ese filtro.';
        listaEl.appendChild(vacio);
    }
}

chips.forEach((chip) => {
    chip.addEventListener('click', () => {
        categoriaActual = chip.dataset.categoria;
        chips.forEach((c) => c.classList.toggle('activo', c === chip));
        aplicarFiltro();
    });
});

buscadorEl.addEventListener('input', aplicarFiltro);

aplicarFiltro();