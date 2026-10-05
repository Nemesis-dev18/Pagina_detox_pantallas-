(function () {
  var autorCita = document.getElementById('autor_cita');
  if (!autorCita) return;

  var autores = ['-Michel Jackson', '-Batman'];
  var indiceAutor = 0;
  window.setInterval(function () {
    indiceAutor = (indiceAutor + 1) % autores.length;
    autorCita.textContent = autores[indiceAutor];
  }, 5000);
})();