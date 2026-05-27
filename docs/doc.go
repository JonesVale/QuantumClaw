// Package docs provides the swagger specification for QuantumClaw API.
//
// The embedded swagger.json is served at /api/swagger/doc.json.
// Swagger UI is available at /api/swagger/ when the master node is running.
package docs

import _ "embed"

// SwaggerJSON contains the OpenAPI 2.0 specification.
//go:embed swagger.json
var SwaggerJSON []byte
