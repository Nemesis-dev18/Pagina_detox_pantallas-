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
    // Algunos nombres se repiten (otros_centros): se toma el primero, nunca una lista
    const campo = (nombre) => {
        const el = form.elements.namedItem(nombre);
        return el && el.tagName === undefined ? el[0] : el;
    };

    // Contenedor visual de un campo; crea el mensaje de error si hace falta
    function mensajeDe(el) {
        const caja = el.closest('.rx-campo') || el.parentElement;
        let msg = caja.querySelector('.rx-msg');
        if (!msg) {
            msg = document.createElement('p');
            msg.className = 'rx-msg';
            msg.setAttribute('role', 'alert');
            caja.append(msg);
        }
        return { caja, msg };
    }

    function marcar(el, texto) {
        const { caja, msg } = mensajeDe(el);
        msg.textContent = texto;
        msg.hidden = false;
        caja.classList.add('rx-con-error');
        el.setAttribute('aria-invalid', 'true');
    }

    function limpiarMarca(el) {
        const { caja, msg } = mensajeDe(el);
        msg.hidden = true;
        caja.classList.remove('rx-con-error');
        el.removeAttribute('aria-invalid');
    }

    // Obligatorios vacíos: lo primero que se revisa
    function vacios() {
        return [...form.querySelectorAll('[required]')].filter(el => {
            if (el.type === 'checkbox') return !el.checked;
            return el.value.trim() === '';
        });
    }

    function reglas() {
        const v = (n) => limpio(campo(n).value);
        const doc = campo('numero_documento').value.toUpperCase().replace(/[ .-]/g, '');
        const nit = campo('institucion_nit').value.replace(/[ .]/g, '').split('-')[0];
        const extras = [...form.querySelectorAll('input[name="otros_centros"]')].map(i => limpio(i.value)).filter(Boolean);

        return [
            [!reNombre.test(v('nombres')), 'nombres', 'Escribe tus nombres: solo letras, de 2 a 80 caracteres.'],
            [!reNombre.test(v('primer_apellido')), 'primer_apellido', 'Escribe tu primer apellido: solo letras, de 2 a 80 caracteres.'],
            [v('segundo_apellido') !== '' && !reNombre.test(v('segundo_apellido')), 'segundo_apellido', 'El segundo apellido solo puede llevar letras.'],
            [!reAlias.test(v('alias')), 'alias', 'De 3 a 30 caracteres: letras, números, espacio, punto o guion.'],
            [!reCorreo.test(v('correo')), 'correo', 'Escribe un correo electrónico válido.'],
            [campo('password').value.length < 8, 'password', 'La contraseña debe tener al menos 8 caracteres.'],
            [campo('password').value !== campo('confirmar_password').value, 'confirmar_password', 'Las contraseñas no coinciden.'],
            [!reDocumento.test(doc), 'numero_documento', 'Solo letras y números, de 5 a 15 caracteres.'],
            [!reRegistro.test(v('numero_registro')), 'numero_registro', 'Solo letras, números, puntos, guiones o barras, de 3 a 30 caracteres.'],
            [v('especialidad').length < 3, 'especialidad', 'La especialidad debe tener de 3 a 100 caracteres.'],
            [v('institucion').length < 3, 'institucion', 'La institución debe tener de 3 a 150 caracteres.'],
            [nit !== '' && !reNIT.test(nit), 'institucion_nit', 'Solo números, sin dígito de verificación (6 a 10 cifras).'],
            [extras.some(e => e.length < 3), 'otros_centros', 'Cada centro adicional debe tener de 3 a 150 caracteres.'],
        ];
    }

    // Quitar la marca al corregir: se revisa mientras la persona escribe
    form.querySelectorAll('input, select, textarea').forEach(el => {
        el.addEventListener('input', () => { if (el.getAttribute('aria-invalid')) limpiarMarca(el); });
        el.addEventListener('change', () => { if (el.getAttribute('aria-invalid')) limpiarMarca(el); });
    });

    form.addEventListener('submit', (e) => {
        form.querySelectorAll('[aria-invalid]').forEach(limpiarMarca);

        const faltan = vacios();
        faltan.forEach(el => marcar(el, el.type === 'checkbox'
            ? 'Debes marcar esta casilla para continuar.'
            : 'Este campo es obligatorio.'));

        const errores = reglas().filter(r => r[0] && campo(r[1]).value.trim() !== '');
        errores.forEach(([, nombre, texto]) => marcar(campo(nombre), texto));

        const todos = [...faltan.map(el => el.name), ...errores.map(r => r[1])];
        if (todos.length === 0) return; // pasa a formularios.js, que envía al servidor

        e.preventDefault();
        e.stopImmediatePropagation();
        const primero = campo(todos[0]);
        if (primero && primero.focus) primero.focus();

        const caja = form.parentElement.querySelector('.mensaje_form');
        if (caja) {
            caja.textContent = 'Revisa los campos marcados en rojo (' + todos.length + ').';
            caja.classList.add('error');
            caja.classList.remove('ok');
            caja.hidden = false;
        }
    }, true);
});
