      (function () {
            var contenedor = document.querySelector('.contenedor_menu');
            var boton = document.getElementById('boton_menu');
            if (!contenedor || !boton) return;

            function fijarEstado(abierto) {
                contenedor.classList.toggle('abierto', abierto);
                boton.setAttribute('aria-expanded', String(abierto));
            }

            boton.addEventListener('click', function () {
                fijarEstado(!contenedor.classList.contains('abierto'));
            });

            // Se cierra al hacer clic fuera del menú
            document.addEventListener('click', function (e) {
                if (!contenedor.contains(e.target)) fijarEstado(false);
            });

            // Se cierra con la tecla Escape
            document.addEventListener('keydown', function (e) {
                if (e.key === 'Escape') {
                    fijarEstado(false);
                    boton.focus();
                }
            });
        })();