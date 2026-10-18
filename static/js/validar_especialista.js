// Validación del formulario de especialistas en el navegador. Es una AYUDA:
// el servidor vuelve a validar todo (especialistas.go). Mismas reglas, mismos límites.
document.addEventListener('DOMContentLoaded', () => {
    const form = document.querySelector('form[action="/registro-especialista"]');
    if (!form) return;

    const reNombre = /^\p{L}[\p{L} .'-]{1,79}$/u;
    const reAlias = /^[\p{L}\p{N}][\p{L}\p{N} ._-]{2,29}$/u;
    const reDocumento = /^[A-Z0-9]{5,15}$/;
    const reRegistro = /^[A-Za-z0-9./ -]{3,30}$/;
    const reNIT = /^[0-9]{6,10}$/;
    const reCorreo = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

    const limpio = (s) => s.trim().replace(/\s+/g, ' ');
    const campo = (nombre) => form.elements.namedItem(nombre);

    function reglas() {
        const v = (n) => limpio(campo(n).value);
        const doc = campo('numero_documento').value.toUpperCase().replace(/[ .-]/g, '');
        const nit = campo('institucion_nit').value.replace(/[ .]/g, '').split('-')[0];
        const extras = [...form.querySelectorAll('input[name="otros_centros"]')].map(i => limpio(i.value)).filter(Boolean);

        return [
            [!reNombre.test(v('nombres')), 'nombres', 'Escribe tus nombres (solo letras, 2 a 80 caracteres).'],
            [!reNombre.test(v('primer_apellido')), 'primer_apellido', 'Escribe tu primer apellido (solo letras, 2 a 80 caracteres).'],
            [v('segundo_apellido') !== '' && !reNombre.test(v('segundo_apellido')), 'segundo_apellido', 'El segundo apellido solo puede llevar letras.'],
            [!reAlias.test(v('alias')), 'alias', 'El alias debe tener 3 a 30 caracteres (letras, números, espacio, punto o guion).'],
            [!reCorreo.test(v('correo')), 'correo', 'El correo electrónico no es válido.'],
            [campo('password').value.length < 8, 'password', 'La contraseña debe tener mínimo 8 caracteres.'],
            [campo('password').value !== campo('confirmar_password').value, 'confirmar_password', 'Las contraseñas no coinciden.'],
            [campo('tipo_documento').value === '', 'tipo_documento', 'Elige el tipo de documento.'],
            [!reDocumento.test(doc), 'numero_documento', 'El número de documento no es válido (solo letras y números).'],
            [campo('profesion').value === '', 'profesion', 'Elige una profesión de la lista.'],
            [!reRegistro.test(v('numero_registro')), 'numero_registro', 'El número de tarjeta o registro profesional no es válido.'],
            [v('especialidad').length < 3, 'especialidad', 'La especialidad debe tener 3 a 100 caracteres.'],
            [v('institucion').length < 3, 'institucion', 'La institución debe tener 3 a 150 caracteres.'],
            [nit !== '' && !reNIT.test(nit), 'institucion_nit', 'El NIT debe tener solo números, sin dígito de verificación.'],
            [extras.some(e => e.length < 3), 'otros_centros', 'Cada centro adicional debe tener 3 a 150 caracteres.'],
            [!campo('acepta_datos').checked, 'acepta_datos', 'Debes autorizar el tratamiento de tus datos.'],
            [!campo('acepta_perfil').checked, 'acepta_perfil', 'Debes aceptar que tu perfil sea público si te aprueban.'],
        ];
    }

    form.addEventListener('submit', (e) => {
        const errores = reglas().filter(r => r[0]);
        const caja = form.parentElement.querySelector('.mensaje_form');
        if (errores.length === 0) return; // pasa a formularios.js, que envía al servidor

        e.preventDefault();
        e.stopImmediatePropagation(); // no enviar: el servidor también lo rechazaría
        const primero = errores[0];
        if (caja) {
            caja.textContent = primero[2];
            caja.classList.add('error');
            caja.classList.remove('ok');
            caja.hidden = false;
        }
        const el = campo(primero[1]);
        if (el && el.focus) el.focus();
    }, true);
});
