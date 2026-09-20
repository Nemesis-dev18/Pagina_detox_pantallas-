const recursos = [
    {
        nombre: "Centro de Rehabilitación",
        localidad: "Kennedy",
        servicio: "Rehabilitación",
        detalle: "Psicología y orientación",
        coordenadas: [4.625, -74.161]
    },
    {
        nombre: "Centro de Orientación Psicológica",
        localidad: "Chapinero",
        servicio: "Psicología",
        detalle: "Orientación y acompañamiento",
        coordenadas: [4.648, -74.063]
    },
    {
        nombre: "Espacio de Desconexión",
        localidad: "Teusaquillo",
        servicio: "Desconexión",
        detalle: "Actividades de bienestar digital",
        coordenadas: [4.637, -74.089]
    }
];

const mapa = L.map("mapa-real", {
    scrollWheelZoom: false
}).setView([4.65, -74.10], 11);

L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
    attribution: "&copy; OpenStreetMap contributors",
    maxZoom: 19
}).addTo(mapa);

const marcadores = recursos.map(recurso => {
    const marcador = L.marker(recurso.coordenadas).addTo(mapa);
    marcador.bindPopup(`
        <strong>${recurso.nombre}</strong><br>
        ${recurso.localidad} · ${recurso.servicio}<br>
        <small>${recurso.detalle}</small>
    `);
    marcador.recurso = recurso;
    return marcador;
});

const grupo = L.featureGroup(marcadores);
mapa.fitBounds(grupo.getBounds().pad(0.25));

const botonesFiltro = document.querySelectorAll(".filtro");
const campoBusqueda = document.querySelector(".mapa-busqueda input");

function actualizarMarcadores() {
    const texto = campoBusqueda.value.trim().toLowerCase();
    const filtroActivo = document.querySelector(".filtro.activo")?.textContent.trim();

    marcadores.forEach(marcador => {
        const recurso = marcador.recurso;
        const coincideTexto = [recurso.nombre, recurso.localidad, recurso.servicio]
            .some(valor => valor.toLowerCase().includes(texto));
        const coincideFiltro = filtroActivo === "Todos" || recurso.servicio === filtroActivo;
        const visible = coincideTexto && coincideFiltro;

        if (visible) {
            marcador.addTo(mapa);
        } else {
            mapa.removeLayer(marcador);
        }
    });
}

botonesFiltro.forEach(boton => {
    boton.addEventListener("click", () => {
        botonesFiltro.forEach(elemento => elemento.classList.remove("activo"));
        boton.classList.add("activo");
        actualizarMarcadores();
    });
});

campoBusqueda.addEventListener("input", actualizarMarcadores);
