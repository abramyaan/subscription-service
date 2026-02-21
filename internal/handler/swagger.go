package handler

import (
	"net/http"
)

// SwaggerYAML возвращает swagger.yaml файл
func SwaggerYAML() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Важно: устанавливаем правильный Content-Type для YAML
		w.Header().Set("Content-Type", "application/x-yaml")
		http.ServeFile(w, r, "docs/swagger.yaml")
	}
}

// SwaggerUI возвращает HTML страницу Swagger UI
func SwaggerUI() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		html := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Subscription Service API</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui.css">
    <style>
        html { box-sizing: border-box; overflow-y: scroll; }
        *, *:before, *:after { box-sizing: inherit; }
        body { margin:0; background: #fafafa; }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            const ui = SwaggerUIBundle({
                url: "/swagger/swagger.yaml",
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout"
            });
            window.ui = ui;
        };
    </script>
</body>
</html>`
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	}
}