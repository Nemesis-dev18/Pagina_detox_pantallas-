// Aviso motivacional en la esquina inferior derecha. Aparece cada 20 segundos y se
// oculta a los 7 segundos. No bloquea la página. Los mensajes son de apoyo general:
// no diagnostican ni reemplazan a un profesional.
const MENSAJES = {
    inicio: [
        '¡Eh, tú! Tu celular no manda en tu día: tú decides cuántos minutos le regalas. Hoy, un poquito menos.',
        'Mide, respira y elige. Cada minuto que recuperas es tuyo para lo que de verdad importa.',
        'Tu atención es tu superpoder. Protégela como si fuera el último pedazo de pastel de la nevera.',
        'Pausa de tres respiraciones antes de volver a la pantalla. Tu cuerpo te lo agradece.',
        'No necesitas hacerlo perfecto. Solo necesitas hacerlo un poco mejor que ayer.',
        'Tu vida pasa también fuera de la pantalla. Date la oportunidad de descubrirlo.',
        'No tienes que abandonar tu celular; aprende a usarlo sin que controle tu tiempo.',
        'Cada pausa consciente es una oportunidad para decidir qué quieres hacer después.',
        '¿Cuánto tiempo llevas deslizando? No te juzgues, simplemente haz una pausa y compruébalo.',
        'Hoy puedes cambiar una pequeña costumbre. Los cambios grandes también empiezan así.',
        'Tu tiempo vale mucho. Procura reservar una parte para las cosas que te hacen bien.',
        'Desconectarte un rato no significa perderte algo. También puedes ganar tranquilidad.',
        'No necesitas responder a todo inmediatamente. Tus momentos de descanso también importan.',
        'Antes de abrir otra aplicación, pregúntate: ¿realmente quiero entrar o lo hago por costumbre?',
        'Una notificación puede esperar. Algunas cosas importantes de tu vida merecen toda tu atención.',
        'Descansar de las pantallas también es una forma de cuidarte.',
        'No compitas contra el reloj. Empieza con una meta pequeña y realista.',
        'Tu progreso no se mide por un día perfecto, sino por lo que aprendes de cada intento.',
        'Regálate unos minutos para hacer algo que no necesite batería ni conexión.',
        'Hoy también cuenta cada vez que eliges conscientemente cómo usar tu tiempo.'
    ],

    comunidad: [
        'Aquí no estás solo: cada publicación es alguien que también se hizo la misma pregunta.',
        'Pregunta sin miedo y comparte sin mostrar tu cara. Todos estamos aprendiendo algo.',
        'Responder con amabilidad también es un acto de valentía. Gracias por cuidar este espacio.',
        'Si lo que lees te pesa, date permiso de cerrar la pantalla y volver cuando estés listo.',
        'Tu experiencia puede ser justo lo que alguien necesita leer hoy. Compártela si quieres.',
        'No necesitas tener todas las respuestas para participar. Una pregunta también puede ayudar.',
        'Detrás de cada publicación hay una persona con su propia historia. Tratémosla con respeto.',
        'Escuchar una experiencia diferente puede ayudarte a mirar tus propios hábitos de otra manera.',
        'No tienes que contar nada que te haga sentir incómodo. Tú decides qué compartir.',
        'Una respuesta amable puede hacer que alguien se sienta escuchado.',
        'Comparte tus avances, incluso los pequeños. Lo importante es reconocer tu propio proceso.',
        'No compares tu progreso con el de otras personas. Cada historia tiene su propio ritmo.',
        'Aquí podemos aprender juntos sin convertir las dificultades en motivo de burla.',
        'Antes de responder, piensa si tus palabras pueden orientar, acompañar o simplemente escuchar.',
        'Pedir opiniones está bien, pero recuerda que puedes decidir qué consejos te sirven.',
        'Está bien tomarte un descanso de la comunidad cuando necesites tiempo para ti.',
        'No tienes que estar de acuerdo con todos para mantener una conversación respetuosa.',
        'Una comunidad sana se construye entre todos, con respeto, empatía y responsabilidad.'
    ],

    especialistas: [
        'Pedir ayuda no es rendirse: es elegir al equipo adecuado para tu historia.',
        'Un especialista puede ayudarte a comprender tus hábitos y explorar nuevas estrategias.',
        'Si estás en crisis, busca ayuda inmediata en tu zona. Hablar con alguien hoy cuenta mucho.',
        'Revisar un perfil no te compromete con nada. Tómate el tiempo que necesites para decidir.',
        'Buscar orientación también puede ser una forma de conocerte mejor.',
        'No necesitas esperar a que todo sea difícil para buscar apoyo profesional.',
        'Elegir un especialista con credenciales verificadas ayuda a tomar una decisión informada.',
        'Puedes preparar tus preguntas antes de una consulta. No necesitas tener todo claro desde el principio.',
        'Hablar de tus dificultades puede resultar complicado. Puedes comenzar por aquello que te resulte más fácil explicar.',
        'La orientación profesional debe respetar tu privacidad y tu dignidad.',
        'Pregunta por la experiencia y las credenciales del profesional antes de iniciar un proceso.',
        'Pedir una segunda opinión puede ser razonable cuando necesitas comprender mejor tus opciones.',
        'Cada proceso es diferente. Un profesional puede ayudarte a establecer objetivos adecuados para ti.',
        'No tienes que resolverlo todo en una sola conversación. Avanzar paso a paso también cuenta.',
        'Buscar acompañamiento no te hace menos capaz. Pedir apoyo y tomar decisiones pueden ir de la mano.',
        'Si no te sientes cómodo con un profesional, puedes expresar tus inquietudes y explorar otras opciones.',
        'Informarte sobre las opciones disponibles es un buen primer paso para cuidar tu bienestar.',
        'Ante una emergencia, no dependas únicamente de un foro o de los mensajes de esta plataforma.'
    ],

    mapa: [
        'Tu entorno también influye en tu tiempo de pantalla. Observa dónde pasas el día y ajústalo poco a poco.',
        'Un lugar tranquilo, una charla con alguien o un paseo corto también recargan la mente.',
        'Explorar sin prisa es un buen antídoto contra el desplazamiento infinito.',
        'A veces, cambiar de espacio ayuda a romper una costumbre automática.',
        'Busca actividades cercanas que te permitan disfrutar del día sin depender del celular.',
        'Un parque, una biblioteca o un espacio de estudio pueden convertirse en lugares para hacer una pausa.',
        'Explora tu entorno con curiosidad. Hay actividades interesantes que quizá todavía no conoces.',
        'Puedes organizar un pequeño plan con alguien y dejar el celular guardado durante un rato.',
        'No necesitas hacer un gran viaje para cambiar de ambiente. A veces basta con salir un momento.',
        'Descubrir nuevos lugares puede ayudarte a crear rutinas diferentes.',
        'El mejor lugar para desconectarte es aquel en el que te sientas cómodo y seguro.',
        'Planea una actividad que disfrutes y reserva un tiempo para vivirla sin interrupciones digitales.',
        'Aprovecha los espacios tranquilos para leer, pensar, conversar o simplemente descansar.',
        'Tu barrio también puede ofrecer oportunidades para aprender, compartir y moverte un poco más.',
        'Explorar el mundo real no significa rechazar la tecnología, sino encontrar un equilibrio.'
    ],

    test: [
        'Responder con honestidad es el primer paso para cambiar lo que quieres cambiar.',
        'No hay respuestas buenas ni malas: es un espejo para que decidas tú.',
        'Tómate tu tiempo. Un resultado no define quién eres; solo te muestra un punto de partida.',
        'Este cuestionario busca ayudarte a reflexionar, no a juzgarte.',
        'No respondas pensando en lo que deberías hacer, sino en lo que realmente ocurre.',
        'Tus hábitos pueden cambiar con el tiempo. Una evaluación representa un momento, no toda tu historia.',
        'Reconocer una costumbre es una oportunidad para decidir qué hacer con ella.',
        'Responde con calma. No tienes que demostrarle nada a nadie.',
        'Un resultado puede servirte para plantear preguntas y descubrir qué hábitos quieres mejorar.',
        'No conviertas una puntuación en una etiqueta personal. Tú eres mucho más que un resultado.',
        'Si alguna pregunta te incomoda, puedes hacer una pausa antes de continuar.',
        'Identificar lo que te cuesta también puede ayudarte a reconocer tus fortalezas.',
        'Piensa en tus hábitos habituales, no solamente en lo que hiciste durante un día excepcional.',
        'El objetivo no es conseguir una puntuación perfecta, sino comprenderte mejor.',
        'Usa tus resultados como una guía para reflexionar, no como un diagnóstico clínico.'
    ],

    nosotros: [
        'R-HOPE nació para acompañar, no para juzgar. Cada persona llega a su propio ritmo.',
        'El objetivo es preventivo: que tengas más control sobre tu vida digital, sin presiones.',
        'Creemos que la tecnología puede ser útil sin ocupar todos nuestros momentos.',
        'R-HOPE busca promover hábitos digitales conscientes y espacios de apoyo respetuosos.',
        'Nuestra comunidad se construye con empatía, privacidad y responsabilidad.',
        'Cada persona tiene necesidades diferentes. Por eso, promovemos decisiones informadas y personales.',
        'No buscamos que abandones la tecnología, sino que encuentres una relación más equilibrada con ella.',
        'Un espacio seguro comienza con el respeto a los demás y el cuidado de la información personal.',
        'La prevención también consiste en reconocer las señales y conocer las opciones de apoyo disponibles.',
        'Queremos que la información te ayude a reflexionar, no que te haga sentir culpable.',
        'Promover el bienestar digital implica escuchar, aprender y mejorar continuamente.',
        'R-HOPE es un punto de partida para explorar hábitos más conscientes, a tu propio ritmo.',
        'Cada pequeño cambio puede ser una oportunidad para construir una rutina más equilibrada.'
    ],

    perfil: [
        'Tu perfil es tu carta de presentación. Escríbelo con calma y con la verdad de tu experiencia.',
        'Cada cambio pensado vale más que diez ediciones apuradas. Tú decides cuándo es suficiente.',
        'Comparte solamente la información con la que te sientas cómodo.',
        'No necesitas contar toda tu historia para presentarte ante la comunidad.',
        'Una descripción sencilla puede ser suficiente para que otras personas conozcan tus intereses.',
        'Puedes actualizar tu perfil cuando tus intereses o tus objetivos cambien.',
        'Tu perfil debe representarte, no cumplir las expectativas de otras personas.',
        'Protege tus datos personales y evita publicar información privada que no sea necesaria.',
        'No hay una forma perfecta de presentarte. Busca una descripción que se sienta auténtica.',
        'Tus intereses pueden ayudar a conectar con personas que comparten objetivos similares.',
        'Piensa qué quieres compartir y qué prefieres mantener en privado. Ambas decisiones merecen respeto.',
        'No necesitas impresionar a nadie. Sé claro y auténtico al describirte.',
        'Revisa tu información de vez en cuando para mantenerla actualizada y cuidar tu privacidad.'
    ]
};

const GENERAL = [
    'Respira. Estás haciendo algo bueno por ti al entrar aquí.',
    'Pequeños pasos también cuentan. Hoy basta con uno.',
    'No tienes que cambiarlo todo hoy. Empieza por algo que puedas mantener.',
    'Date permiso de aprender sin exigirte hacerlo todo perfectamente.',
    'Tu bienestar importa incluso en los días en los que parece difícil encontrar tiempo.',
    'Haz una pausa, revisa cómo te sientes y decide qué necesitas en este momento.',
    'No estás obligado a tener todas las respuestas ahora mismo.',
    'Un momento de reflexión puede cambiar la forma en que afrontas el resto del día.',
    'Reconocer una dificultad ya puede ser un comienzo para trabajar en ella.',
    'Mereces tratarte con la misma paciencia que intentarías ofrecerle a alguien más.',
    'No todo avance se nota de inmediato. Date tiempo para aprender.',
    'Hoy puede ser una nueva oportunidad para elegir conscientemente.',
    'Tu valor no depende de tu productividad ni de cuánto tiempo pases conectado.',
    'A veces, descansar es lo más útil que puedes hacer por ti.',
    'Escucha tus necesidades y busca apoyo cuando sientas que lo necesitas.',
    'Cada día trae circunstancias diferentes. Ajustar tus metas también es válido.',
    'Recuerda que puedes disfrutar la tecnología y reservar espacio para otras cosas.',
    'No necesitas compararte con los demás para reconocer tus propios avances.',
    'Haz espacio para las actividades, las personas y los momentos que son importantes para ti.',
    'Un pequeño descanso puede ayudarte a recuperar la atención y continuar con más claridad.',
    'Lo importante no es tener el control de todo, sino aprender a tomar decisiones conscientes.',
    'Puedes volver a intentarlo sin castigarte por lo que no salió como esperabas.',
    'Tener preguntas es parte del proceso de conocerte mejor.',
    'Dedicarte tiempo no es perder el tiempo.',
    'Recuerda: avanzar a tu ritmo también es avanzar.'
];

    const DURACION_VISIBLE = 7000;   // ms que se ve el aviso
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
    titulo.textContent = 'Un recordatorio para ti (att:Hey Ya!)';
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
