// Edición del perfil de especialista: máximo dos cambios por mes (lo impone el servidor).
document.addEventListener('DOMContentLoaded', async () => {
    const estadoCuenta = document.getElementById('estado-cuenta');
    const ediciones = document.getElementById('ediciones');
    const form = document.getElementById('form-perfil');
    const aviso = document.getElementById('aviso');
    const especialidad = document.getElementById('especialidad');
    const descripcion = document.getElementById('descripcion');
    const guardar = document.getElementById('guardar');

    function mostrarAviso(texto, ok) {
        aviso.textContent = texto;
        aviso.classList.toggle('ok', ok);
        aviso.classList.toggle('error', !ok);
        aviso.hidden = false;
    }

    function marcar(el, texto) {
        const caja = el.parentElement.querySelector('.mensaje_campo');
        caja.textContent = texto;
        caja.hidden = !texto;
        el.setAttribute('aria-invalid', texto ? 'true' : 'false');
    }

    async function cargar() {
        let res;
        try {
            res = await fetch('/yo/especialista/perfil');
        } catch (err) {
            estadoCuenta.textContent = 'No hay conexión con el servidor.';
            return;
        }
        if (res.status === 401) {
            estadoCuenta.innerHTML = '';
            estadoCuenta.append('Inicia sesión para editar tu perfil: ');
            const a = document.createElement('a');
            a.href = 'inicio_sesion.html';
            a.textContent = 'iniciar sesión';
            estadoCuenta.append(a, '.');
            return;
        }
        if (!res.ok) {
            estadoCuenta.textContent = await res.text();
            return;
        }
        const datos = await res.json();
        especialidad.value = datos.especialidad;
        descripcion.value = datos.descripcion;
        form.hidden = false;
        ediciones.textContent = datos.ediciones_restantes > 0
            ? 'Te quedan ' + datos.ediciones_restantes + ' edición(es) este mes.'
            : 'Ya usaste tus 2 ediciones de este mes. Podrás editar de nuevo el próximo mes.';
        guardar.disabled = datos.ediciones_restantes === 0;
    }

    form.addEventListener('submit', async (e) => {
        e.preventDefault();
        aviso.hidden = true;

        const esp = especialidad.value.trim().replace(/\s+/g, ' ');
        const desc = descripcion.value.trim();
        let ok = true;
        if (esp.length < 3 || esp.length > 100) {
            marcar(especialidad, 'La especialidad debe tener de 3 a 100 caracteres.');
            ok = false;
        } else {
            marcar(especialidad, '');
        }
        if (desc.length > 600) {
            marcar(descripcion, 'La presentación no puede pasar de 600 caracteres.');
            ok = false;
        } else {
            marcar(descripcion, '');
        }
        if (!ok) return;

        guardar.disabled = true;
        try {
            const res = await fetch('/yo/especialista/perfil', {
                method: 'PUT',
                body: new URLSearchParams({ especialidad: esp, descripcion: desc }),
            });
            if (res.status === 204) {
                mostrarAviso('Cambios guardados.', true);
                await cargar();
            } else {
                mostrarAviso(await res.text(), false);
            }
        } catch (err) {
            mostrarAviso('No hay conexión con el servidor.', false);
        } finally {
            guardar.disabled = false;
        }
    });

    cargar();
});
