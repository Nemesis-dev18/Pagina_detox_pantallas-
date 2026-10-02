// Panel de palabras excluidas. Solo funciona con una sesión de administrador.
document.addEventListener('DOMContentLoaded', () => {
    const form = document.getElementById('form-palabra');
    const aviso = document.getElementById('aviso');
    const listaPalabras = document.getElementById('lista-palabras');
    const listaPatrones = document.getElementById('lista-patrones');
    const listaComentarios = document.getElementById('lista-comentarios');

    function celda(texto, clase) {
        const td = document.createElement('td');
        td.textContent = texto; // textContent: nunca se interpreta como HTML
        if (clase) td.className = clase;
        return td;
    }

    function fila(...celdas) {
        const tr = document.createElement('tr');
        tr.append(...celdas);
        return tr;
    }

    function vacio(tbody, columnas, texto) {
        const td = celda(texto);
        td.colSpan = columnas;
        tbody.replaceChildren(fila(td));
    }

    async function cargarPalabras() {
        const res = await fetch('/admin/palabras');
        if (!res.ok) {
            aviso.textContent = res.status === 401 || res.status === 403
                ? 'Necesitas iniciar sesión como administrador.'
                : 'No se pudo cargar la lista.';
            return false;
        }
        const palabras = await res.json();
        if (palabras.length === 0) return vacio(listaPalabras, 3, 'Aún no hay palabras.') || true;

        listaPalabras.replaceChildren(...palabras.map(p => {
            const btn = document.createElement('button');
            btn.type = 'button';
            btn.className = 'btn';
            btn.textContent = 'Quitar';
            btn.addEventListener('click', async () => {
                const r = await fetch('/admin/palabras/' + p.id, { method: 'DELETE' });
                if (r.ok) { cargarTodo(); } else { aviso.textContent = await r.text(); }
            });
            const td = document.createElement('td');
            td.append(btn);
            return fila(celda(p.palabra), celda(new Date(p.creado_en).toLocaleDateString()), td);
        }));
        return true;
    }

    async function cargarCoincidencias() {
        const res = await fetch('/admin/coincidencias');
        if (!res.ok) return;
        const datos = await res.json();

        if (datos.length === 0) {
            vacio(listaPatrones, 3, 'Todavía no hay coincidencias.');
            vacio(listaComentarios, 3, 'Todavía no hay coincidencias.');
            return;
        }

        // Agrupa por (palabra, forma escrita) y cuenta
        const grupos = new Map();
        for (const d of datos) {
            const forma = d.fragmento.toLowerCase();
            const clave = d.palabra + '\u0000' + forma;
            const g = grupos.get(clave) || { palabra: d.palabra, forma, veces: 0 };
            g.veces++;
            grupos.set(clave, g);
        }
        const ordenados = [...grupos.values()].sort((a, b) => b.veces - a.veces);
        listaPatrones.replaceChildren(...ordenados.map(g =>
            fila(
                celda(g.palabra),
                celda(g.forma, g.forma !== g.palabra ? 'disfrazada' : ''),
                celda(String(g.veces))
            )));

        // Un comentario puede tener varias coincidencias: se muestra una vez
        const vistos = new Set();
        const recientes = datos.filter(d => !vistos.has(d.comentario_id) && vistos.add(d.comentario_id)).slice(0, 30);
        listaComentarios.replaceChildren(...recientes.map(d =>
            fila(celda(new Date(d.creado_en).toLocaleString()), celda(d.alias), celda(d.contenido))));
    }

    async function cargarTodo() {
        if (await cargarPalabras()) cargarCoincidencias();
    }

    form.addEventListener('submit', async (e) => {
        e.preventDefault();
        aviso.textContent = '';
        const res = await fetch('/admin/palabras', {
            method: 'POST',
            body: new URLSearchParams(new FormData(form)),
        });
        if (res.ok) {
            form.reset();
            aviso.textContent = 'Palabra agregada.';
            cargarTodo();
        } else {
            aviso.textContent = (await res.text()).trim();
        }
    });

    cargarTodo();
});
