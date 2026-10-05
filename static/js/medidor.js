(function () {
  var rango = document.getElementById('horas_dia');
  var rejilla = document.getElementById('rejilla_horas');
  var salida = document.getElementById('resultado');
  var valor = document.getElementById('valor_horas');
  if (!rango || !rejilla || !salida) return;

  for (var i = 0; i < 24; i++) rejilla.appendChild(document.createElement('span'));
  var celdas = rejilla.children;

  function actualizar() {
    var h = Number(rango.value);
    var dias = Math.round(h * 365 / 24);
    for (var i = 0; i < celdas.length; i++) celdas[i].classList.toggle('on', i < h);
    valor.textContent = h + (h === 1 ? ' hora' : ' horas');
    salida.innerHTML = '<strong class="cifra">' + dias + (dias === 1 ? ' día' : ' días') +
      '</strong> completos al año frente a la pantalla, si usas el celular ' + h + ' h al día.';
    rango.setAttribute('aria-valuetext', valor.textContent + ' al día');
  }
  rango.addEventListener('input', actualizar);
  actualizar();
})();
