(function () {
  var formulario = document.getElementById('form_test');
  var progreso = document.getElementById('test_progreso');
  var resultado = document.getElementById('resultado_test');
  var mensaje = document.getElementById('resultado_mensaje');
  if (!formulario || !progreso || !resultado || !mensaje) return;

  var totalPreguntas = 6;

  formulario.addEventListener('change', function () {
    var respondidas = formulario.querySelectorAll('input[type="radio"]:checked').length;
    progreso.textContent = respondidas + ' de ' + totalPreguntas + ' respondidas';
    resultado.hidden = true;
  });

  formulario.addEventListener('submit', function (evento) {
    evento.preventDefault();
    if (!formulario.reportValidity()) return;

    var respuestas = new FormData(formulario);
    var puntaje = 0;
    for (var numero = 1; numero <= totalPreguntas; numero++) {
      puntaje += Number(respuestas.get('pregunta' + numero));
    }

    if (puntaje <= 5) {
      mensaje.textContent = 'Por tus respuestas, parece que el celular suele ocupar un espacio que puedes manejar. Sigue observando qué hábitos te ayudan a mantener ese equilibrio.';
    } else if (puntaje <= 11) {
      mensaje.textContent = 'A veces el celular puede estar ocupando más espacio del que quisieras. Identifica en qué momentos ocurre y prueba un cambio pequeño, como dejarlo lejos durante una actividad.';
    } else {
      mensaje.textContent = 'Tus respuestas sugieren que el celular podría estar interfiriendo en varias áreas de tu día. Puedes probar cambios graduales y buscar apoyo si esto te preocupa.';
    }

    resultado.hidden = false;
    resultado.scrollIntoView({ behavior: 'smooth', block: 'center' });
  });

  formulario.addEventListener('reset', function () {
    progreso.textContent = '0 de ' + totalPreguntas + ' respondidas';
    resultado.hidden = true;
  });
})();