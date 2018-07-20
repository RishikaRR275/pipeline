/*
 * Pipeline API
 *
 * Pipeline v0.3.0 swagger
 *
 * API version: 0.3.0
 * Contact: info@banzaicloud.com
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package client

type CreateUpdateDeploymentResponse struct {
	ReleaseName string `json:"releaseName,omitempty"`
	// deployment notes in base64 encoded format
	Notes string `json:"notes,omitempty"`
}
