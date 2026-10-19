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
            'Elige un momento del día para dejar el teléfono a un lado y estar presente.',
            'Cada vez que haces una pausa, recuperas espacio para elegir.',
            'No se trata de contar cada minuto, sino de decidir qué quieres hacer con él.',
            'Prueba a dejar el celular fuera de tu alcance durante una actividad que disfrutes.',
            'Tu atención merece descansos, igual que tu cuerpo.',
            'Una meta pequeña y posible vale más que un cambio enorme que no puedes sostener.',
            'Cuando desbloquees el celular por costumbre, detente un segundo y pregúntate qué necesitas.',
            'Celebra las veces que eliges conscientemente, aunque sean pocas.',
        ],
        comunidad: [
            'Aquí no estás solo: cada publicación es alguien que también se hizo la misma pregunta.',
            'Pregunta sin miedo y comparte sin mostrar tu cara. Todos estamos aprendiendo algo.',
            'Responder con amabilidad también es un acto de valentía. Gracias por cuidar este espacio.',
            'Si lo que lees te pesa, date permiso de cerrar la pantalla y volver cuando estés listo.',
            'Tu experiencia puede ser justo lo que alguien necesita leer hoy. Compártela si quieres.',
            'Puedes participar a tu ritmo: leer también es una forma válida de estar aquí.',
            'No hace falta tener la frase perfecta para apoyar; escuchar con respeto ya importa.',
            'Una pregunta sincera puede abrir una conversación que alguien necesitaba.',
            'Comparte solo aquello que te sientas cómodo dejando en este espacio.',
            'No estás obligado a resolver los problemas de otros para acompañarlos.',
            'Las diferencias pueden convivir con el respeto y la curiosidad.',
        ],
        especialistas: [
            'Pedir ayuda no es rendirse: es elegir al equipo adecuado para tu historia.',
            'Un especialista no te juzga; te acompaña. Dar el primer paso ya es avanzar.',
            'Si estás en crisis, busca ayuda inmediata en tu zona. Hablar con alguien hoy cuenta mucho.',
            'Revisar un perfil no te compromete con nada. Tómate el tiempo que necesites para decidir.',
            'Pedir información no te obliga a iniciar un proceso.',
            'Puedes conversar con más de un profesional hasta encontrar una opción adecuada para ti.',
            'Está bien preguntar cómo se protege tu privacidad antes de compartir información.',
            'Una primera consulta también puede servir para conocer tus opciones.',
            'Avanzar con apoyo sigue siendo avanzar por ti mismo.',
            'Si algo no te queda claro, pedir que te lo expliquen es parte de una atención informada.',
        ],
        mapa: [
            'Tu entorno también influye en tu tiempo de pantalla. Observa dónde pasas el día y ajústalo poco a poco.',
            'Un lugar tranquilo, una charla con alguien o un paseo corto también recargan la mente.',
            'Explorar sin prisa es un buen antídoto contra el desplazamiento infinito.',
            'Un paseo corto puede ser una pausa suficiente para cambiar de ritmo.',
            'Descubre un sitio nuevo cerca de casa, aunque solo tengas unos minutos.',
            'Compartir un plan sencillo con alguien puede ayudarte a disfrutar el momento.',
            'Elige un espacio donde te sientas cómodo para descansar un rato.',
            'Una actividad fuera de pantalla no tiene que ser productiva para ser valiosa.',
            'Cada lugar nuevo puede inspirar una rutina diferente.',
        ],
        test: [
            'Responder con honestidad es el primer paso para cambiar lo que quieres cambiar.',
            'No hay respuestas buenas ni malas: es un espejo para que decidas tú.',
            'Tómate tu tiempo. Un resultado no define quién eres; solo te muestra un punto de partida.',
            'Contesta según lo que suele ocurrir, no según un día perfecto o uno especialmente difícil.',
            'No necesitas apresurarte; tómate el tiempo que te ayude a responder con honestidad.',
            'Lo que descubras aquí puede ayudarte a decidir qué quieres explorar después.',
            'Una respuesta sincera te da un punto de partida más útil que una respuesta ideal.',
            'Puedes revisar tus hábitos sin convertirlos en una etiqueta sobre quién eres.',
            'Lo importante es conocerte un poco mejor, no obtener una calificación.',
        ],
        nosotros: [
            'R-HOPE nació para acompañar, no para juzgar. Cada persona llega a su propio ritmo.',
            'El objetivo es preventivo: que tengas más control sobre tu vida digital, sin presiones.',
            'La prevención también empieza con información clara y decisiones propias.',
            'Cada persona puede encontrar un equilibrio distinto con la tecnología.',
            'Construir un espacio respetuoso requiere cuidado y participación de todos.',
            'Queremos que este proyecto acompañe tus preguntas, no que decida por ti.',
            'Tu privacidad y tu ritmo son parte del bienestar digital.',
            'La tecnología puede sumar a tu vida cuando tú eliges cómo usarla.',
        ],
        perfil: [
            'Tu perfil es tu carta de presentación. Escríbelo con calma y con la verdad de tu experiencia.',
            'Cada cambio pensado vale más que diez ediciones apuradas. Tú decides cuándo es suficiente.',
            'Elige con calma qué detalles te ayudan a presentarte y cuáles prefieres reservar.',
            'Un perfil claro y auténtico puede ser breve.',
            'Tu experiencia tiene valor, incluso si estás dando tus primeros pasos.',
            'No tienes que responder a todo para conectar con otras personas.',
            'Puedes destacar lo que te interesa ahora, sin definirte para siempre.',
            'Una revisión rápida puede ayudarte a cuidar tu privacidad antes de guardar los cambios.',
        ],
    };
    const GENERAL = [
        'Respira. Estás haciendo algo bueno por ti al entrar aquí.',
        'Pequeños pasos también cuentan. Hoy basta con uno.',
        'Una pausa pequeña puede abrir espacio para algo que disfrutas.',
        'No tienes que hacerlo todo hoy; elige un siguiente paso posible.',
        'Tu ritmo también merece respeto.',
        'Cada vez que vuelves a intentarlo, estás practicando una nueva elección.',
        'Permítete descansar sin tener que ganártelo.',
        'Puedes cambiar de plan cuando descubras qué te funciona mejor.',
        'Un avance discreto también cuenta, aunque nadie más lo note.',
        'Pregúntate qué necesitas ahora: moverte, descansar, conversar o simplemente respirar.',
        'No hace falta que todo salga bien para que el día todavía tenga buenos momentos.',
        'Busca una actividad que te dé curiosidad y regálale un rato.',
        'Puedes pedir compañía cuando hacer un cambio por tu cuenta se sienta difícil.',
        'Recuerda que una notificación no decide por ti cuándo atenderla.',
    ];

    const DURACION_VISIBLE = 8000;   // ms que se ve el aviso
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
    titulo.append('Un recordatorio para ti ');
    const atribucion = document.createElement('span');
    atribucion.className = 'popup_atribucion';
    atribucion.textContent = '(att: Hey Ya!, "Stand")';
    titulo.append(atribucion);
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
    document.addEventListener('visibilitychange', () => {
        if (!document.hidden && aviso.hidden) mostrar();
    });
    mostrar();
    setInterval(mostrar, CADA);
})();
