package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type DocsController struct {
	spec []byte
}

func NewDocsController(spec []byte) *DocsController {
	return &DocsController{spec: spec}
}

func (c *DocsController) OpenAPI(ectx echo.Context) error {
	return ectx.Blob(http.StatusOK, "application/json", c.spec)
}

func (c *DocsController) UI(ectx echo.Context) error {
	return ectx.HTML(http.StatusOK, swaggerUIPage)
}

const swaggerUIPage = `<!doctype html>
<html>
<head>
  <title>partner_app API</title>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.17.14/swagger-ui.css" integrity="sha384-wxLW6kwyHktdDGr6Pv1zgm/VGJh99lfUbzSn6HNHBENZlCN7W602k9VkGdxuFvPn" crossorigin="anonymous" />
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.17.14/swagger-ui-bundle.js" integrity="sha384-wmyclcVGX/WhUkdkATwhaK1X1JtiNrr2EoYJ+diV3vj4v6OC5yCeSu+yW13SYJep" crossorigin="anonymous"></script>
  <script>
    window.addEventListener("load", function () {
      window.ui = SwaggerUIBundle({
        url: "/openapi.json",
        dom_id: "#swagger-ui",
        deepLinking: true,
        persistAuthorization: true,
        tryItOutEnabled: true,
        displayRequestDuration: true
      });
    });
  </script>
</body>
</html>`
