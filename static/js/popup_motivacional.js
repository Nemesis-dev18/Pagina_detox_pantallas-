// Aviso motivacional en la esquina inferior derecha. Aparece cada 20 segundos y se
// oculta a los 5 segundos. No bloquea la página. Los mensajes son de apoyo general:
// no diagnostican ni reemplazan a un profesional.
(function () {
    const MENSAJES = {
        inicio: [
            '¡Eh, tú! Tu celular no manda en tu día: tú decides cuántos minutos le regalas. Hoy, un poquito menos.',
            'Mide, respira y elige. Cada minuto que recuperas es tuyo para lo que de verdad importa.',
            'Tu atención es tu superpoder. Protégela como si fuera el último pedazo de pastel de la nevera.',
            'Pausa de tres respiraciones antes de volver a la pantalla. Tu cuerpo te lo agradece.',
            'No necesitas hacerlo perfecto. Solo necesitas hacerlo un poco mejor que ayer.',
        ],
        comunidad: [
            'Aquí no estás solo: cada publicación es alguien que también se hizo la misma pregunta.',
            'Pregunta sin miedo y comparte sin mostrar tu cara. Todos estamos aprendiendo algo.',
            'Responder con amabilidad también es un acto de valentía. Gracias por cuidar este espacio.',
            'Si lo que lees te pesa, date permiso de cerrar la pantalla y volver cuando estés listo.',
            'Tu experiencia puede ser justo lo que alguien necesita leer hoy. Compártela si quieres.',
        ],
        especialistas: [
            'Pedir ayuda no es rendirse: es elegir al equipo adecuado para tu historia.',
            'Un especialista no te juzga; te acompaña. Dar el primer paso ya es avanzar.',
            'Si estás en crisis, busca ayuda inmediata en tu zona. Hablar con alguien hoy cuenta mucho.',
            'Revisar un perfil no te compromete con nada. Tómate el tiempo que necesites para decidir.',
        ],
        mapa: [
            'Tu entorno también influye en tu tiempo de pantalla. Observa dónde pasas el día y ajústalo poco a poco.',
            'Un lugar tranquilo, una charla con alguien o un paseo corto también recargan la mente.',
            'Explorar sin prisa es un buen antídoto contra el desplazamiento infinito.',
        ],
        test: [
            'Responder con honestidad es el primer paso para cambiar lo que quieres cambiar.',
            'No hay respuestas buenas ni malas: es un espejo para que decidas tú.',
            'Tómate tu tiempo. Un resultado no define quién eres; solo te muestra un punto de partida.',
        ],
        nosotros: [
            'R-HOPE nació para acompañar, no para juzgar. Cada persona llega a su propio ritmo.',
            'El objetivo es preventivo: que tengas más control sobre tu vida digital, sin presiones.',
        ],
        perfil: [
            'Tu perfil es tu carta de presentación. Escríbelo con calma y con la verdad de tu experiencia.',
            'Cada cambio pensado vale más que diez ediciones apuradas. Tú decides cuándo es suficiente.',
        ],
    };
    const GENERAL = [
        'Respira. Estás haciendo algo bueno por ti al entrar aquí.',
        'Pequeños pasos también cuentan. Hoy basta con uno.',
    ];

    const DURACION_VISIBLE = 5000;   // ms que se ve el aviso
    const CADA = 20000;              // ms entre avisos

    const script = document.currentScript || document.querySelector('script[data-pagina]');
    const pagina = (script && script.dataset.pagina) || 'general';
    const lista = MENSAJES[pagina] || GENERAL;

    const aviso = document.createElement('aside');
    aviso.className = 'popup_motivacional';
    aviso.setAttribute('role', 'status');
    aviso.setAttribute('aria-live', 'polite');
    aviso.hidden = true;

    const titulo = document.createElement('p');
    titulo.className = 'popup_titulo';
    titulo.textContent = 'Un recordatorio para ti';
    const texto = document.createElement('p');
    texto.className = 'popup_texto';
    const cerrar = document.createElement('button');
    cerrar.type = 'button';
    cerrar.className = 'popup_cerrar';
    cerrar.setAttribute('aria-label', 'Cerrar recordatorio');
    cerrar.textContent = '×';
    aviso.append(titulo, texto, cerrar);
    document.body.append(aviso);

    let temporizador;
    function ocultar() {
        clearTimeout(temporizador);
        aviso.hidden = true;
    }
    function mostrar() {
        if (document.hidden) return; // pestaña en segundo plano: no se muestra
        texto.textContent = lista[Math.floor(Math.random() * lista.length)];
        aviso.classList.remove('entrar');
        void aviso.offsetWidth; // reinicia la animación
        aviso.classList.add('entrar');
        aviso.hidden = false;
        clearTimeout(temporizador);
        temporizador = setTimeout(ocultar, DURACION_VISIBLE);
    }

    cerrar.addEventListener('click', ocultar);
    setInterval(mostrar, CADA);
})();
