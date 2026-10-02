// Agrega X-CSRF-Token a toda petición que modifica datos (POST/PUT/DELETE...).
// Debe cargarse ANTES que los demás scripts, sin defer.
(function () {
    function leerCookie(nombre) {
        var par = document.cookie.split('; ').find(function (c) {
            return c.indexOf(nombre + '=') === 0;
        });
        return par ? par.slice(nombre.length + 1) : '';
    }

    var fetchOriginal = window.fetch;
    window.fetch = function (entrada, opciones) {
        opciones = opciones || {};
        var metodo = (opciones.method || (entrada && entrada.method) || 'GET').toUpperCase();
        if (metodo !== 'GET' && metodo !== 'HEAD' && metodo !== 'OPTIONS') {
            var headers = new Headers(opciones.headers || (entrada && entrada.headers) || {});
            if (!headers.has('X-CSRF-Token')) {
                headers.set('X-CSRF-Token', leerCookie('csrf_token'));
            }
            opciones.headers = headers;
        }
        return fetchOriginal.call(this, entrada, opciones);
    };
})();
