// Panel de cuentas. Solo administradores; el servidor vuelve a comprobar todo.
document.addEventListener('DOMContentLoaded', () => {
    const buscar = document.getElementById('buscar');
    const aviso = document.getElementById('aviso');
    const tbody = document.getElementById('lista-cuentas');
    let temporizador;

    function celda(texto) {
        const td = document.createElement('td');
        td.textContent = texto ?? '';
        return td;
    }

    function fila(c) {
        const tr = document.createElement('tr');
        const boton = document.createElement('button');
        boton.type = 'button';
        boton.className = 'btn';
        boton.textContent = 'Borrar';
        boton.addEventListener('click', () => borrar(c));

        const acc = document.createElement('td');
        acc.append(boton);
        tr.append(
            celda(c.alias),
            celda(c.correo),
            celda(c.tipo + (c.es_admin ? ' · admin' : '')),
            celda(new Date(c.creado_en).toLocaleDateString('es-CO')),
            acc
        );
        return tr;
    }

    async function cargar() {
        const res = await fetch('/admin/cuentas?q=' + encodeURIComponent(buscar.value.trim()));
        if (!res.ok) {
            aviso.textContent = res.status === 401 || res.status === 403
                ? 'Necesitas iniciar sesión como administrador.'
                : 'No se pudo cargar la lista.';
            tbody.replaceChildren();
            return;
        }
        const lista = (await res.json()) || [];
        if (lista.length === 0) {
            const td = celda('No hay cuentas con esa búsqueda.');
            td.colSpan = 5;
            const tr = document.createElement('tr');
            tr.append(td);
            tbody.replaceChildren(tr);
            return;
        }
        tbody.replaceChildren(...lista.map(fila));
    }

    async function borrar(c) {
        // Confirmación fuerte: hay que escribir el alias exacto
        const escrito = prompt(
            'Esto borra la cuenta "' + c.alias + '" y sus publicaciones y comentarios.\n\n' +
            'Para confirmar, escribe el alias exacto:'
        );
        if (escrito === null) return;
        if (escrito.trim() !== c.alias) {
            aviso.textContent = 'El alias no coincide. No se borró nada.';
            return;
        }

        const res = await fetch('/admin/cuentas/' + c.id, { method: 'DELETE' });
        if (res.status === 204) {
            aviso.textContent = 'Cuenta borrada.';
            cargar();
        } else {
            aviso.textContent = await res.text();
        }
    }

    buscar.addEventListener('input', () => {
        clearTimeout(temporizador);
        temporizador = setTimeout(cargar, 300);
    });

    cargar();
});
