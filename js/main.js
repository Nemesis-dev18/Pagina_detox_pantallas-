
// Menú desplegable de OPCIONES.
// Para usarlo en otra página: mismo HTML del menú + <script src="js/menu.js" defer></script>
document.addEventListener('DOMContentLoaded', () => {
    const contenedor = document.querySelector('.contenedor_menu');
    const boton = document.getElementById('boton_menu');
    if (!contenedor || !boton) return;
 
    function estaAbierto() {
        return contenedor.classList.contains('abierto');
    }
 
    function fijarEstado(abierto) {
        contenedor.classList.toggle('abierto', abierto);
        boton.setAttribute('aria-expanded', String(abierto));
    }
 
    boton.addEventListener('click', () => fijarEstado(!estaAbierto()));
 
    // Se cierra al hacer clic fuera del menú
    document.addEventListener('click', (e) => {
        if (estaAbierto() && !contenedor.contains(e.target)) fijarEstado(false);
    });
 
    // Se cierra con la tecla Escape
    document.addEventListener('keydown', (e) => {
        if (e.key === 'Escape' && estaAbierto()) {
            fijarEstado(false);
            boton.focus();
        }
    });
});
 