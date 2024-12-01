package models

type ResourceConfig struct {
	BucketName string `json:"bucket_name"`
	RoleName   string `json:"role_name"`
}

type ProvisionRequest struct {
	ClientID   string `json:"client_id"`
	ClientName string `json:"client_name"`
}

type ProvisionResponse struct {
	Status       string `json:"status"`
	BucketName   string `json:"bucket_name"`
	RoleARN      string `json:"role_arn"`
	LogGroupName string `json:"log_group_name"`
	LambdaARN    string `json:"lambda_arn"`
	TopicARN     string `json:"topic_arn"`
}
