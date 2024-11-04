// internal/models/client.go
package models

type Client struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Resources ResourceConfig `json:"resources"`
}

type ResourceConfig struct {
	BucketName string `json:"bucket_name"`
	RoleName   string `json:"role_name"`
}

type ProvisionRequest struct {
	ClientID   string `json:"client_id"`
	ClientName string `json:"client_name"`
}

type ProvisionResponse struct {
	Status     string `json:"status"`
	BucketName string `json:"bucket_name"`
	RoleARN    string `json:"role_arn"`
}
