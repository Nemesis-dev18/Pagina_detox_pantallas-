package main

import (
	"fmt"
	"net/http"
	"strings"
)

// robots.txt: bloquea solo los endpoints (las páginas privadas llevan meta noindex).
func robotsTxt(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "User-agent: *\nDisallow: /admin/\nDisallow: /publicaciones\nDisallow: /login\nDisallow: /logout\nDisallow: /registro\nDisallow: /recuperar\nDisallow: /restablecer\nDisallow: /yo\n\nSitemap: %s/sitemap.xml\n",
		strings.TrimRight(baseURL(), "/"))
}

// sitemap.xml: usa BASE_URL (en Render, la URL pública del sitio).
func sitemapXML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	base := strings.TrimRight(baseURL(), "/")
	fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?>`+"\n"+`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`+"\n")
	for _, p := range []string{"index", "nosotros", "comunidad", "mapa"} {
		fmt.Fprintf(w, "  <url><loc>%s/static/%s.html</loc></url>\n", base, p)
	}
	fmt.Fprint(w, "</urlset>\n")
}
