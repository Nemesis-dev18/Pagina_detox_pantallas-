// Envía los formularios con data-ajax sin salir de la página:
// - Si el servidor responde con error (ej. "correo o contraseña incorrectos"),
//   el mensaje aparece debajo del formulario, en la misma página.
// - Si el servidor redirige (login/registro/restablecer correctos), se sigue la redirección.
// - Si responde OK con un texto (recuperar contraseña), se muestra como aviso.
// Estilos del mensaje (puedes moverlos a css/styles.css si prefieres)
const estilo = document.createElement('style');
estilo.textContent = `
.mensaje_form { margin: 12px 0 0; padding: 10px 14px; border-radius: 6px; font-size: 0.95rem; text-align: center; }
.mensaje_form[hidden] { display: none; }
.mensaje_form.error { background: #fde8e8; color: #9b1c1c; border: 1px solid #f5b5b5; }
.mensaje_form.ok { background: #e6f6ea; color: #1e6b34; border: 1px solid #a8dbb5; }`;
document.head.appendChild(estilo);

document.addEventListener('DOMContentLoaded', () => {
    document.querySelectorAll('form[data-ajax]').forEach(form => {
        const caja = form.parentElement.querySelector('.mensaje_form');
        const boton = form.querySelector('button[type="submit"]');

        function mostrar(texto, ok) {
            if (!caja) return;
            caja.textContent = texto;
            caja.classList.toggle('ok', ok);
            caja.classList.toggle('error', !ok);
            caja.hidden = false;
        }

        form.addEventListener('submit', async (e) => {
            e.preventDefault();
            if (caja) caja.hidden = true;
            if (boton) boton.disabled = true;
            try {
                const res = await fetch(form.action, {
                    method: 'POST',
                    body: new URLSearchParams(new FormData(form)),
                });
                if (res.redirected) {
                    window.location.href = res.url;
                    return;
                }
                const texto = (await res.text()).trim();
                if (res.ok) {
                    mostrar(texto || 'Listo.', true);
                    form.reset();
                } else {
                    mostrar(texto || 'Ocurrió un error, intenta de nuevo.', false);
                }
            } catch (err) {
                mostrar('No hay conexión con el servidor.', false);
            } finally {
                if (boton) boton.disabled = false;
            }
        });
    });
});
