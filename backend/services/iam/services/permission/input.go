package permission

// CreateInput membawa data pembentukan izin atomik baru pada katalog RBAC IAM.
type CreateInput struct {
	Code        string
	Resource    string
	Action      string
	Description string
	IsSensitive bool
}
