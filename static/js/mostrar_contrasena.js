// Agrega un botón "ojo" a cada campo de contraseña para poder verla u ocultarla.
// No hay que tocar el HTML de los formularios: basta con incluir este script.
//   <script src="../static/js/mostrar_password.js" defer></script>
(function () {
    const OJO = '<svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/></svg>';
    const OJO_TACHADO = '<svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M17.9 17.9A10.9 10.9 0 0 1 12 20c-7 0-11-8-11-8a19.8 19.8 0 0 1 5.1-5.9"/><path d="M9.9 4.2A10.7 10.7 0 0 1 12 4c7 0 11 8 11 8a19.9 19.9 0 0 1-3.2 4.2"/><path d="M14.1 14.1a3 3 0 1 1-4.2-4.2"/><line x1="1" y1="1" x2="23" y2="23"/></svg>';

    // Estilos del botón (puedes moverlos a css/styles.css si prefieres)
    const estilo = document.createElement('style');
    estilo.textContent = `
    input.con_ver_password { padding-right: 44px !important; }
    .ver_password { position: absolute; margin: 0; padding: 0; display: flex; align-items: center;
        justify-content: center; background: transparent; border: 0; border-radius: 6px; color: #666;
        cursor: pointer; z-index: 1; }
    .ver_password:hover { color: #111; }
    .ver_password:focus-visible { outline: 2px solid currentColor; outline-offset: 1px; }`;
    document.head.appendChild(estilo);

    function agregarBoton(input) {
        if (input.dataset.verPassword) return;           // ya tiene botón
        input.dataset.verPassword = '1';

        // El campo NO se mueve ni se envuelve: el botón se coloca encima, a su derecha.
        const padre = input.parentNode;
        if (getComputedStyle(padre).position === 'static') padre.style.position = 'relative';
        input.classList.add('con_ver_password');

        const boton = document.createElement('button');
        boton.type = 'button';                           // nunca envía el formulario
        boton.className = 'ver_password';
        boton.innerHTML = OJO;                           // solo SVG fijo de arriba
        input.insertAdjacentElement('afterend', boton);

        function actualizar(visible) {
            input.type = visible ? 'text' : 'password';
            boton.innerHTML = visible ? OJO_TACHADO : OJO;
            const texto = visible ? 'Ocultar contraseña' : 'Mostrar contraseña';
            boton.setAttribute('aria-label', texto);
            boton.title = texto;
            boton.setAttribute('aria-pressed', String(visible));
        }
        actualizar(false);

        boton.addEventListener('click', () => {
            actualizar(input.type === 'password');
            input.focus();
        });

        // Si el formulario se limpia (formularios.js hace form.reset()), vuelve a ocultarse
        if (input.form) input.form.addEventListener('reset', () => actualizar(false));

        // Coloca el botón sobre el borde derecho del campo, centrado, sin importar el CSS
        function colocar() {
            const lado = Math.min(34, Math.max(20, input.offsetHeight - 4)); // nunca más alto que el campo
            boton.style.width = lado + 'px';
            boton.style.height = lado + 'px';
            boton.style.left = (input.offsetLeft + input.offsetWidth - lado - 6) + 'px';
            boton.style.top = (input.offsetTop + (input.offsetHeight - lado) / 2) + 'px';
        }
        colocar();
        window.addEventListener('resize', colocar);
        window.addEventListener('load', colocar);        // por si las fuentes cambian el tamaño
        if (window.ResizeObserver) {
            const obs = new ResizeObserver(colocar);
            obs.observe(input);
            obs.observe(padre);
        }
    }

    function iniciar() {
        document.querySelectorAll('input[type="password"]').forEach(agregarBoton);
    }

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', iniciar);
    } else {
        iniciar();
    }
})();