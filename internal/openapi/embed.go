package openapi

import _ "embed"

// Spec is the embedded OpenAPI 3 document served at GET /openapi.yaml.
//
//go:embed spec.yaml
var Spec []byte
