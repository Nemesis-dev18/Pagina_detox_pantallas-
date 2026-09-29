// Muestra el botón "Cerrar sesión" cuando hay una sesión activa.
// Consulta /yo: si responde OK, hay sesión; si no, deja la página como está.
document.addEventListener('DOMContentLoaded', async () => {
    const cabecera = document.querySelector('.hero, .hero_nosotros');
    if (!cabecera) return;

    let alias;
    try {
        const res = await fetch('/yo');
        if (!res.ok) return;
        alias = (await res.json()).alias;
    } catch (err) {
        return; // sin conexión con el servidor: se queda como visitante
    }

    let zona = cabecera.querySelector('.hero-right');
    if (!zona) {
        zona = document.createElement('div');
        zona.className = 'hero-right';
        cabecera.appendChild(zona);
    }

    // Ya hay sesión: se quita el enlace "Iniciar Sesión"
    zona.querySelectorAll('a[href$="inicio_sesion.html"]').forEach(a => a.remove());

    // POST (no un simple enlace) para que nadie cierre tu sesión sin querer
    const form = document.createElement('form');
    form.action = '/logout';
    form.method = 'POST';
    form.className = 'form_salir';

    const boton = document.createElement('button');
    boton.type = 'submit';
    boton.className = 'btn btn_salir';
    boton.textContent = 'Cerrar sesión' + (alias ? ' (' + alias + ')' : '');

    form.appendChild(boton);
    zona.appendChild(form);
});
