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
    },
    {
        nombre: "Punto de Orientación Familiar",
        localidad: "Suba",
        servicio: "Psicología",
        detalle: "Escucha inicial y orientación para familias",
        coordenadas: [4.747, -74.083]
    },
    {
        nombre: "Centro de Acompañamiento Comunitario",
        localidad: "Bosa",
        servicio: "Rehabilitación",
        detalle: "Acompañamiento para crear hábitos saludables",
        coordenadas: [4.617, -74.186]
    },
    {
        nombre: "Punto de Bienestar Digital",
        localidad: "Usaquén",
        servicio: "Desconexión",
        detalle: "Actividades de descanso y desconexión tecnológica",
        coordenadas: [4.704, -74.033]
    },
    {
        nombre: "Orientación Juvenil",
        localidad: "Engativá",
        servicio: "Psicología",
        detalle: "Orientación para jóvenes y cuidadores",
        coordenadas: [4.704, -74.112]
    },
    {
        nombre: "Red de Apoyo Local",
        localidad: "San Cristóbal",
        servicio: "Rehabilitación",
        detalle: "Información y acompañamiento comunitario",
        coordenadas: [4.568, -74.091]
    },
    {
        nombre: "Zona de Pausa Digital",
        localidad: "Barrios Unidos",
        servicio: "Desconexión",
        detalle: "Espacio para actividades de bienestar y pausa",
        coordenadas: [4.669, -74.083]
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

const estadoUbicacion = document.querySelector("#estado-ubicacion");

async function ubicarPorIp() {
    try {
        const respuesta = await fetch("https://ipwho.is/", {
            headers: { Accept: "application/json" }
        });
        if (!respuesta.ok) throw new Error("No se pudo consultar la ubicación");

        const datos = await respuesta.json();
        if (datos.success === false) throw new Error("El servicio no devolvió ubicación");
        const latitud = Number(datos.latitude);
        const longitud = Number(datos.longitude);

        if (!Number.isFinite(latitud) || !Number.isFinite(longitud)) {
            throw new Error("La respuesta no contiene coordenadas válidas");
        }

        mapa.setView([latitud, longitud], 12);
        const ubicacion = L.marker([latitud, longitud])
            .addTo(mapa)
            .bindPopup("Tu ubicación aproximada según tu IP");
        ubicacion.openPopup();

        const ciudad = datos.city || "tu zona";
        estadoUbicacion.textContent = `Ubicación aproximada detectada: ${ciudad}.`;
    } catch (error) {
        estadoUbicacion.textContent = "No se pudo detectar tu IP; mostrando Bogotá como ubicación inicial.";
    }
}

ubicarPorIp();

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
