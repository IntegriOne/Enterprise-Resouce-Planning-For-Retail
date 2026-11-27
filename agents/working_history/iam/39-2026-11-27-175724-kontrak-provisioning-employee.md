# Kontrak provisioning employee

Commit `4e75dbf` mempublikasikan proto EmployeeProvisioningService dan referensi worker serta request provisioning pada CreateUserRequest. Buf lint, regenerasi, dan tes race transport IAM lulus di snapshot terisolasi. Implementasi handler dan persistensi provisioning serta penautan HRM masih perlu dipublikasikan bersama runtime.
