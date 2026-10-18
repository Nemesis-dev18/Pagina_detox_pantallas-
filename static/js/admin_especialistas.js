// Panel de revisión de especialistas. Solo muestra y envía; el servidor decide
// (permisos, orden de estados, y que solo se apruebe con verificaciones confirmadas).
document.addEventListener('DOMContentLoaded', () => {
    const aviso = document.getElementById('aviso');
    const filtro = document.getElementById('filtro-estado');
    const tbody = document.getElementById('lista-solicitudes');
    const detalle = document.getElementById('detalle');
    const credencial = document.getElementById('credencial');
    const institucion = document.getElementById('institucion');
    const notas = document.getElementById('notas');
    const motivo = document.getElementById('motivo');
    const evidencia = document.getElementById('evidencia');
    let actual = null; // id de la solicitud abierta

    const td = (texto) => {
        const c = document.createElement('td');
        c.textContent = texto ?? '';
        return c;
    };

    async function cargarLista() {
        const res = await fetch('/admin/especialistas?estado=' + encodeURIComponent(filtro.value));
        if (!res.ok) {
            aviso.textContent = res.status === 401 || res.status === 403
                ? 'Necesitas iniciar sesión como administrador.'
                : 'No se pudo cargar la lista.';
            tbody.replaceChildren();
            return;
        }
        const lista = (await res.json()) || [];
        if (lista.length === 0) {
            const vacio = td('No hay solicitudes en este estado.');
            vacio.colSpan = 7;
            const tr = document.createElement('tr');
            tr.append(vacio);
            tbody.replaceChildren(tr);
            return;
        }
        tbody.replaceChildren(...lista.map(s => {
            const tr = document.createElement('tr');
            const boton = document.createElement('button');
            boton.type = 'button';
            boton.className = 'btn';
            boton.textContent = 'Revisar';
            boton.addEventListener('click', () => abrir(s.id));
            const acc = document.createElement('td');
            acc.append(boton);
            tr.append(td(s.nombre), td(s.profesion), td(s.institucion), td(s.estado),
                td(s.credencial_estado), td(s.institucion_estado), acc);
            return tr;
        }));
    }

    async function abrir(id) {
        const res = await fetch('/admin/especialistas/' + id);
        if (!res.ok) { aviso.textContent = await res.text(); return; }
        const s = await res.json();
        actual = id;

        document.getElementById('detalle-nombre').textContent = s.nombre + ' · ' + s.alias;
        const datos = document.getElementById('detalle-datos');
        const filas = [
            ['Correo', s.correo], ['Documento', s.tipo_documento + ' ' + s.documento],
            ['Registro profesional', s.numero_registro], ['Profesión', s.profesion],
            ['Especialidad', s.especialidad], ['Institución', s.institucion],
            ['NIT institución', s.institucion_nit], ['Presentación', s.descripcion],
            ['Estado', s.estado], ['Credencial', s.credencial_estado], ['Institución (verif.)', s.institucion_estado],
            ['Motivo', s.motivo]
        ];
        datos.replaceChildren(...filas.flatMap(([k, v]) => {
            const dt = document.createElement('dt'); dt.textContent = k;
            const dd = document.createElement('dd'); dd.textContent = v || '—';
            return [dt, dd];
        }));

        document.getElementById('enlace-rethus').href = s.url_rethus;
        evidencia.textContent = s.evidencia_institucion
            ? JSON.stringify(s.evidencia_institucion, null, 2)
            : 'Aún no se ha consultado.';
        credencial.value = s.credencial_estado;
        institucion.value = s.institucion_estado;
        notas.value = s.notas_revision || '';
        motivo.value = '';
        detalle.hidden = false;
        detalle.scrollIntoView({ behavior: 'smooth' });
    }

    async function enviar(decision) {
        if (!actual) return;
        const cuerpo = new URLSearchParams({
            credencial: credencial.value, institucion: institucion.value,
            notas: notas.value, motivo: motivo.value, decision: decision
        });
        const res = await fetch('/admin/especialistas/' + actual + '/revisar', { method: 'POST', body: cuerpo });
        if (res.ok) {
            aviso.textContent = 'Guardado.';
            await cargarLista();
            await abrir(actual);
        } else {
            aviso.textContent = await res.text();
        }
    }

    document.getElementById('btn-consultar').addEventListener('click', async () => {
        if (!actual) return;
        const res = await fetch('/admin/especialistas/' + actual + '/consultar-institucion', { method: 'POST' });
        if (res.ok) {
            evidencia.textContent = JSON.stringify(await res.json(), null, 2);
        } else {
            aviso.textContent = await res.text();
        }
    });
    document.getElementById('btn-guardar').addEventListener('click', () => enviar(''));
    document.getElementById('btn-aprobar').addEventListener('click', () => enviar('aprobar'));
    document.getElementById('btn-rechazar').addEventListener('click', () => enviar('rechazar'));
    document.getElementById('btn-suspender').addEventListener('click', () => enviar('suspender'));
    document.getElementById('btn-reabrir').addEventListener('click', () => enviar('reabrir'));
    filtro.addEventListener('change', cargarLista);

    cargarLista();
});
