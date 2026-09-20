# DETOX_WEB en Ruby

La versión Ruby usa Sinatra y reutiliza los recursos existentes del proyecto (`css/`, `img/`, `fuentes/` y `js/`).

## Requisitos

- Ruby 3.1 o superior
- Bundler

## Ejecutar

Desde esta carpeta:

```bash
bundle install
bundle exec ruby app.rb
```

Después abre `http://localhost:4567`.

## Rutas

- `/` Inicio
- `/inicio_sesion` Inicio de sesión
- `/registro` Crear cuenta
- `/recuperar_contraseña` Recuperación de contraseña
- `/nosotros` Nosotros
- `/comunidad` Comunidad
- `/mapa` Mapa interactivo con Leaflet y OpenStreetMap
- `/test` Medidor de bienestar digital

La carpeta tiene un espacio al final en su nombre (`ruby `). Para evitar problemas en terminal, entra usando comillas: `cd "ruby "`.
