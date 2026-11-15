=== RUN   TestQueryBrowserSession
=== RUN   TestQueryBrowserSession/PreAuthAnonymous
    iam_operations_e2e_test.go:1007: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query { browserSession { authenticated principalId csrfToken } }
        Input Variables:
        {}
        Result JSON:
        {
          "data": {
            "browserSession": {
              "authenticated": false,
              "csrfToken": "hKqQF0gTC_2TYGOL1V9ZDeUTbawMDRAGGH7XXcn4e7Y",
              "principalId": null
            }
          }
        }
=== RUN   TestQueryBrowserSession/AuthenticatedAllFieldsNonNull
    iam_operations_e2e_test.go:1015: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query { browserSession { authenticated principalId csrfToken } }
        Input Variables:
        {}
        Result JSON:
        {
          "data": {
            "browserSession": {
              "authenticated": true,
              "csrfToken": "lyOTd4pSTezIVZqstvz7rodqCARfA6FjnKOBuvjYuY0",
              "principalId": "853aac77-32e7-4a88-b2d6-7d68db8d13d7"
            }
          }
        }
=== RUN   TestQueryBrowserSession/AfterLogoutReturnsUnauthenticated
    iam_operations_e2e_test.go:1022: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateEmployeeUserInput!) { createEmployeeUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676826229229703_1.png",
            "display_name": "Fixture u_employee_1794676826229229703_1",
            "email": "u_employee_1794676826229229703_1@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createEmployeeUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676826229229703_1.png",
              "created_at": "2026-11-15T00:20:26.230149245+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_employee_1794676826229229703_1",
              "id": "24fb1b50-7f49-4ecd-9041-cd3344a05c6d",
              "kind": "EMPLOYEE",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:26.230149245+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1023: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_employee_1794676826229229703_1@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "A8nSjVouWvtg-b6TpwLluGO2W1JupAwVAb0EXzp2yf8",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/account",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:1024: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation { userLogout }
        Input Variables:
        {}
        Result JSON:
        {
          "data": {
            "userLogout": true
          }
        }
    iam_operations_e2e_test.go:1026: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query { browserSession { authenticated principalId csrfToken } }
        Input Variables:
        {}
        Result JSON:
        {
          "data": {
            "browserSession": {
              "authenticated": false,
              "csrfToken": "F-Dp-ZvUaXWCIv_wrBpvLJTnrIBnjNavkm9EwTT5Ib0",
              "principalId": null
            }
          }
        }
--- PASS: TestQueryBrowserSession (0.34s)
    --- PASS: TestQueryBrowserSession/PreAuthAnonymous (0.01s)
    --- PASS: TestQueryBrowserSession/AuthenticatedAllFieldsNonNull (0.07s)
    --- PASS: TestQueryBrowserSession/AfterLogoutReturnsUnauthenticated (0.26s)
=== RUN   TestQueryMyWorkspaceAccess
=== RUN   TestQueryMyWorkspaceAccess/SuperAdminIdentityWorkspace
    iam_operations_e2e_test.go:1039: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query { myWorkspaceAccess { canAccessIdentity defaultLocation } }
        Input Variables:
        {}
        Result JSON:
        {
          "data": {
            "myWorkspaceAccess": {
              "canAccessIdentity": true,
              "defaultLocation": "/identity"
            }
          }
        }
=== RUN   TestQueryMyWorkspaceAccess/CustomerStorefrontAccess
    iam_operations_e2e_test.go:1052: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateUserInput!) { createUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676826522452175_2.png",
            "display_name": "Fixture u_customer_1794676826522452175_2",
            "email": "u_customer_1794676826522452175_2@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676826522452175_2.png",
              "created_at": "2026-11-15T00:20:26.523219168+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_customer_1794676826522452175_2",
              "id": "15d7884b-f0fd-4be9-8905-529e43777786",
              "kind": "CUSTOMER",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:26.523219168+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1052: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_customer_1794676826522452175_2@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "pricrkyJqyshB3xqLyrhxFYShOTkDnq1OnWFZPY11UM",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:1053: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query { myWorkspaceAccess { canAccessIdentity defaultLocation } }
        Input Variables:
        {}
        Result JSON:
        {
          "data": {
            "myWorkspaceAccess": {
              "canAccessIdentity": false,
              "defaultLocation": "/"
            }
          }
        }
=== RUN   TestQueryMyWorkspaceAccess/AnonymousUnauthenticated
    iam_operations_e2e_test.go:1066: 
        [UNAUTHENTICATED_GUARD] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query { myWorkspaceAccess { canAccessIdentity defaultLocation } }
        Input Variables:
        {}
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "session is missing, expired, or revoked",
              "extensions": {
                "code": "UNAUTHENTICATED"
              }
            }
          ]
        }
--- PASS: TestQueryMyWorkspaceAccess (0.29s)
    --- PASS: TestQueryMyWorkspaceAccess/SuperAdminIdentityWorkspace (0.05s)
    --- PASS: TestQueryMyWorkspaceAccess/CustomerStorefrontAccess (0.23s)
    --- PASS: TestQueryMyWorkspaceAccess/AnonymousUnauthenticated (0.00s)
=== RUN   TestQueryMyProfile
=== RUN   TestQueryMyProfile/EmployeeProfileCompleteNonNull
    iam_operations_e2e_test.go:1077: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateEmployeeUserInput!) { createEmployeeUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676826763430185_3.png",
            "display_name": "Fixture u_employee_1794676826763430185_3",
            "email": "u_employee_1794676826763430185_3@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createEmployeeUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676826763430185_3.png",
              "created_at": "2026-11-15T00:20:26.764257271+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_employee_1794676826763430185_3",
              "id": "7ca11bc1-6b96-436b-a648-60b5f759b4c3",
              "kind": "EMPLOYEE",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:26.764257271+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1078: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_employee_1794676826763430185_3@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "bPgaA9bKo1q-ypx_QJVNOL9yNPlTa9O2ZIGsGNQpb8k",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/account",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:1080: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query { myProfile { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {}
        Result JSON:
        {
          "data": {
            "myProfile": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676826763430185_3.png",
              "created_at": "2026-11-15T00:20:26.764257+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_employee_1794676826763430185_3",
              "id": "7ca11bc1-6b96-436b-a648-60b5f759b4c3",
              "kind": "EMPLOYEE",
              "last_login_at": "2026-11-15T00:20:26.811208+07:00",
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:26.855193+07:00"
            }
          }
        }
=== RUN   TestQueryMyProfile/AnonymousUnauthenticated
    iam_operations_e2e_test.go:1091: 
        [UNAUTHENTICATED_GUARD] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query { myProfile { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {}
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "session is missing, expired, or revoked",
              "extensions": {
                "code": "UNAUTHENTICATED"
              }
            }
          ]
        }
--- PASS: TestQueryMyProfile (0.24s)
    --- PASS: TestQueryMyProfile/EmployeeProfileCompleteNonNull (0.24s)
    --- PASS: TestQueryMyProfile/AnonymousUnauthenticated (0.00s)
=== RUN   TestQueryMyAccessSnapshot
=== RUN   TestQueryMyAccessSnapshot/SuperAdminFullSnapshot
    iam_operations_e2e_test.go:1102: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query { myAccessSnapshot { principal_id user_kind status mfa_enabled auth_epoch perm_revision recovery_epoch can_access_identity default_redirect_path allowed_redirect_paths roles { id code name priority default_redirect_path } permissions { permission_code is_sensitive scope { type id } } synced_at_unix fresh_until_unix } }
        Input Variables:
        {}
        Result JSON:
        {
          "data": {
            "myAccessSnapshot": {
              "allowed_redirect_paths": [
                "/",
                "/identity",
                "/human-resource",
                "/procurement",
                "/warehouse",
                "/finance",
                "/accounting",
                "/product",
                "/supplier",
                "/vendor",
                "/commerce",
                "/account",
                "/profile"
              ],
              "auth_epoch": 1,
              "can_access_identity": true,
              "default_redirect_path": "/identity",
              "fresh_until_unix": 1794676856,
              "mfa_enabled": false,
              "perm_revision": 2,
              "permissions": [
                {
                  "is_sensitive": false,
                  "permission_code": "iam.dashboard.view",
                  "scope": {
                    "id": null,
                    "type": "GLOBAL"
                  }
                },
                {
                  "is_sensitive": false,
                  "permission_code": "iam.groups.read",
                  "scope": {
                    "id": null,
                    "type": "GLOBAL"
                  }
                },
                {
                  "is_sensitive": true,
                  "permission_code": "iam.groups.write",
                  "scope": {
                    "id": null,
                    "type": "GLOBAL"
                  }
                },
                {
                  "is_sensitive": false,
                  "permission_code": "iam.permissions.read",
                  "scope": {
                    "id": null,
                    "type": "GLOBAL"
                  }
                },
                {
                  "is_sensitive": false,
                  "permission_code": "iam.roles.read",
                  "scope": {
                    "id": null,
                    "type": "GLOBAL"
                  }
                },
                {
                  "is_sensitive": true,
                  "permission_code": "iam.roles.write",
                  "scope": {
                    "id": null,
                    "type": "GLOBAL"
                  }
                },
                {
                  "is_sensitive": false,
                  "permission_code": "iam.sessions.read",
                  "scope": {
                    "id": null,
                    "type": "GLOBAL"
                  }
                },
                {
                  "is_sensitive": true,
                  "permission_code": "iam.sessions.revoke",
                  "scope": {
                    "id": null,
                    "type": "GLOBAL"
                  }
                },
                {
                  "is_sensitive": false,
                  "permission_code": "iam.users.read",
                  "scope": {
                    "id": null,
                    "type": "GLOBAL"
                  }
                },
                {
                  "is_sensitive": true,
                  "permission_code": "iam.users.write",
                  "scope": {
                    "id": null,
                    "type": "GLOBAL"
                  }
                }
              ],
              "principal_id": "853aac77-32e7-4a88-b2d6-7d68db8d13d7",
              "recovery_epoch": 24,
              "roles": [
                {
                  "code": "superadmin",
                  "default_redirect_path": "/identity",
                  "id": "d8ffac94-b437-4aa3-a937-27556b86d2dd",
                  "name": "Super Administrator",
                  "priority": 1
                }
              ],
              "status": "active",
              "synced_at_unix": 1794676826,
              "user_kind": "system"
            }
          }
        }
=== RUN   TestQueryMyAccessSnapshot/CustomerSnapshotDeniedIdentity
    iam_operations_e2e_test.go:1115: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateUserInput!) { createUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676827052821408_4.png",
            "display_name": "Fixture u_customer_1794676827052821408_4",
            "email": "u_customer_1794676827052821408_4@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676827052821408_4.png",
              "created_at": "2026-11-15T00:20:27.053626424+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_customer_1794676827052821408_4",
              "id": "7b510f29-fbb2-4cbf-bd23-e78c29fada3b",
              "kind": "CUSTOMER",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:27.053626424+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1115: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_customer_1794676827052821408_4@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "PHOTra6hGS20IJnjSyYAaCWJKmEfwhvimbOvAnm4z_Q",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:1116: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query { myAccessSnapshot { principal_id user_kind status mfa_enabled auth_epoch perm_revision recovery_epoch can_access_identity default_redirect_path allowed_redirect_paths roles { id code name priority default_redirect_path } permissions { permission_code is_sensitive scope { type id } } synced_at_unix fresh_until_unix } }
        Input Variables:
        {}
        Result JSON:
        {
          "data": {
            "myAccessSnapshot": {
              "allowed_redirect_paths": [
                "/",
                "/identity",
                "/human-resource",
                "/procurement",
                "/warehouse",
                "/finance",
                "/accounting",
                "/product",
                "/supplier",
                "/vendor",
                "/commerce",
                "/account",
                "/profile"
              ],
              "auth_epoch": 1,
              "can_access_identity": false,
              "default_redirect_path": "/",
              "fresh_until_unix": 1794676857,
              "mfa_enabled": false,
              "perm_revision": 1,
              "permissions": [],
              "principal_id": "7b510f29-fbb2-4cbf-bd23-e78c29fada3b",
              "recovery_epoch": 24,
              "roles": [],
              "status": "active",
              "synced_at_unix": 1794676827,
              "user_kind": "customer"
            }
          }
        }
=== RUN   TestQueryMyAccessSnapshot/AnonymousUnauthenticated
    iam_operations_e2e_test.go:1126: 
        [UNAUTHENTICATED_GUARD] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query { myAccessSnapshot { principal_id user_kind status mfa_enabled auth_epoch perm_revision recovery_epoch can_access_identity default_redirect_path allowed_redirect_paths roles { id code name priority default_redirect_path } permissions { permission_code is_sensitive scope { type id } } synced_at_unix fresh_until_unix } }
        Input Variables:
        {}
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "session is missing, expired, or revoked",
              "extensions": {
                "code": "UNAUTHENTICATED"
              }
            }
          ]
        }
--- PASS: TestQueryMyAccessSnapshot (0.29s)
    --- PASS: TestQueryMyAccessSnapshot/SuperAdminFullSnapshot (0.05s)
    --- PASS: TestQueryMyAccessSnapshot/CustomerSnapshotDeniedIdentity (0.24s)
    --- PASS: TestQueryMyAccessSnapshot/AnonymousUnauthenticated (0.00s)
=== RUN   TestQueryMySessions
=== RUN   TestQueryMySessions/ActiveAndHistoricalSessionsComplete
    iam_operations_e2e_test.go:1142: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateEmployeeUserInput!) { createEmployeeUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676827291284424_5.png",
            "display_name": "Fixture u_employee_1794676827291284424_5",
            "email": "u_employee_1794676827291284424_5@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createEmployeeUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676827291284424_5.png",
              "created_at": "2026-11-15T00:20:27.292462395+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_employee_1794676827291284424_5",
              "id": "cf8e3c11-f9cf-487d-89ff-bde3b943a85f",
              "kind": "EMPLOYEE",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:27.292462395+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1144: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_employee_1794676827291284424_5@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "leskZYg7bclpdUg07GzOBn6PpePXxGVnv3ynFrMexsM",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/account",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:1145: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_employee_1794676827291284424_5@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "9GoAFaGRFQn9YgDmvU7oyUNU2n-bGRKwOWemnhQkI3M",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/account",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:1146: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation { userLogout }
        Input Variables:
        {}
        Result JSON:
        {
          "data": {
            "userLogout": true
          }
        }
    iam_operations_e2e_test.go:1148: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation { enrollMyTotp { factor_id otp_auth_uri } }
        Input Variables:
        {}
        Result JSON:
        {
          "data": {
            "enrollMyTotp": {
              "factor_id": "af4a93f6-0f08-4ec8-ac7b-ac9c8bef95fe",
              "otp_auth_uri": "otpauth://totp/IntegriOne%20ERP:Fixture%20u_employee_1794676827291284424_5?algorithm=SHA1\u0026digits=6\u0026issuer=IntegriOne%20ERP\u0026period=30\u0026secret=YRPBVBVZYDUF35JOS2NZUWHOOGR7CUN7"
            }
          }
        }
    iam_operations_e2e_test.go:1148: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: ConfirmTotpInput!) { confirmMyTotp(input: $input) { enabled recovery_codes } }
        Input Variables:
        {
          "input": {
            "code": "664570"
          }
        }
        Result JSON:
        {
          "data": {
            "confirmMyTotp": {
              "enabled": true,
              "recovery_codes": [
                "MUXX-ZMBF-KN39",
                "HHZZ-WKM7-VRDB",
                "JXG7-2AVN-JABB",
                "4E87-KWKV-QHX7",
                "6FAQ-UMCY-K7XW",
                "5YAN-KD4M-Q5EV",
                "45CW-N9VV-8UV3",
                "YT2T-TE9M-J983"
              ]
            }
          }
        }
    iam_operations_e2e_test.go:1149: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($code: String!) { verifyMySessionMfa(totp_code: $code) }
        Input Variables:
        {
          "code": "859554"
        }
        Result JSON:
        {
          "data": {
            "verifyMySessionMfa": true
          }
        }
    iam_operations_e2e_test.go:1153: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($active_only: Boolean, $page: PageInput) { mySessions(active_only: $active_only, page: $page) { items { id user_id mfa_verified mfa_verified_at ip_address user_agent authenticated_at last_active_at idle_expires_at absolute_expires_at ended_at end_reason created_at generation } page { total limit offset has_more } } }
        Input Variables:
        {
          "active_only": true,
          "page": {
            "limit": 10,
            "offset": 0
          }
        }
        Result JSON:
        {
          "data": {
            "mySessions": {
              "items": [
                {
                  "absolute_expires_at": "2026-11-15T08:20:27.382507+07:00",
                  "authenticated_at": "2026-11-15T00:20:27.382507+07:00",
                  "created_at": "2026-11-15T00:20:27.382507+07:00",
                  "end_reason": null,
                  "ended_at": null,
                  "generation": 2,
                  "id": "2daa4299-02d3-4200-82f2-67585b3b818b",
                  "idle_expires_at": "2026-11-15T00:50:27.382507+07:00",
                  "ip_address": "172.22.0.1",
                  "last_active_at": "2026-11-15T00:20:27.382507+07:00",
                  "mfa_verified": true,
                  "mfa_verified_at": "2026-11-15T00:20:27.441983+07:00",
                  "user_agent": "Go-http-client/1.1",
                  "user_id": "cf8e3c11-f9cf-487d-89ff-bde3b943a85f"
                }
              ],
              "page": {
                "has_more": false,
                "limit": 10,
                "offset": 0,
                "total": 1
              }
            }
          }
        }
    iam_operations_e2e_test.go:1167: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($active_only: Boolean, $page: PageInput) { mySessions(active_only: $active_only, page: $page) { items { id user_id mfa_verified mfa_verified_at ip_address user_agent authenticated_at last_active_at idle_expires_at absolute_expires_at ended_at end_reason created_at generation } page { total limit offset has_more } } }
        Input Variables:
        {
          "active_only": false,
          "page": {
            "limit": 10,
            "offset": 0
          }
        }
        Result JSON:
        {
          "data": {
            "mySessions": {
              "items": [
                {
                  "absolute_expires_at": "2026-11-15T08:20:27.431508+07:00",
                  "authenticated_at": "2026-11-15T00:20:27.431508+07:00",
                  "created_at": "2026-11-15T00:20:27.431508+07:00",
                  "end_reason": "LOGOUT",
                  "ended_at": "2026-11-15T00:20:27.435582+07:00",
                  "generation": 1,
                  "id": "9c65bd25-6397-4d94-b45a-14e66a1a6702",
                  "idle_expires_at": "2026-11-15T00:50:27.431508+07:00",
                  "ip_address": "172.22.0.1",
                  "last_active_at": "2026-11-15T00:20:27.431508+07:00",
                  "mfa_verified": false,
                  "mfa_verified_at": null,
                  "user_agent": "Go-http-client/1.1",
                  "user_id": "cf8e3c11-f9cf-487d-89ff-bde3b943a85f"
                },
                {
                  "absolute_expires_at": "2026-11-15T08:20:27.382507+07:00",
                  "authenticated_at": "2026-11-15T00:20:27.382507+07:00",
                  "created_at": "2026-11-15T00:20:27.382507+07:00",
                  "end_reason": null,
                  "ended_at": null,
                  "generation": 2,
                  "id": "2daa4299-02d3-4200-82f2-67585b3b818b",
                  "idle_expires_at": "2026-11-15T00:50:27.382507+07:00",
                  "ip_address": "172.22.0.1",
                  "last_active_at": "2026-11-15T00:20:27.382507+07:00",
                  "mfa_verified": true,
                  "mfa_verified_at": "2026-11-15T00:20:27.441983+07:00",
                  "user_agent": "Go-http-client/1.1",
                  "user_id": "cf8e3c11-f9cf-487d-89ff-bde3b943a85f"
                }
              ],
              "page": {
                "has_more": false,
                "limit": 10,
                "offset": 0,
                "total": 2
              }
            }
          }
        }
=== RUN   TestQueryMySessions/AnonymousUnauthenticated
    iam_operations_e2e_test.go:1192: 
        [UNAUTHENTICATED_GUARD] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($active_only: Boolean, $page: PageInput) { mySessions(active_only: $active_only, page: $page) { items { id user_id mfa_verified mfa_verified_at ip_address user_agent authenticated_at last_active_at idle_expires_at absolute_expires_at ended_at end_reason created_at generation } page { total limit offset has_more } } }
        Input Variables:
        {
          "active_only": true
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "session is missing, expired, or revoked",
              "extensions": {
                "code": "UNAUTHENTICATED"
              }
            }
          ]
        }
--- PASS: TestQueryMySessions (0.29s)
    --- PASS: TestQueryMySessions/ActiveAndHistoricalSessionsComplete (0.29s)
    --- PASS: TestQueryMySessions/AnonymousUnauthenticated (0.00s)
=== RUN   TestQueryUser
=== RUN   TestQueryUser/ValidUserWithLoginNonNull
    iam_operations_e2e_test.go:1203: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateEmployeeUserInput!) { createEmployeeUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676827582920781_6.png",
            "display_name": "Fixture u_employee_1794676827582920781_6",
            "email": "u_employee_1794676827582920781_6@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createEmployeeUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676827582920781_6.png",
              "created_at": "2026-11-15T00:20:27.583728711+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_employee_1794676827582920781_6",
              "id": "7d9e763f-fff3-48a5-af2b-7b9b2015a1ff",
              "kind": "EMPLOYEE",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:27.583728711+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1204: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_employee_1794676827582920781_6@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "qFK6NMjr4faD-Bx5CC-iGpHJbo58xl6D5HdJOUgtMR8",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/account",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:1206: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($id: ID!) { user(id: $id) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "id": "7d9e763f-fff3-48a5-af2b-7b9b2015a1ff"
        }
        Result JSON:
        {
          "data": {
            "user": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676827582920781_6.png",
              "created_at": "2026-11-15T00:20:27.583728+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_employee_1794676827582920781_6",
              "id": "7d9e763f-fff3-48a5-af2b-7b9b2015a1ff",
              "kind": "EMPLOYEE",
              "last_login_at": "2026-11-15T00:20:27.631732+07:00",
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:27.675393+07:00"
            }
          }
        }
=== RUN   TestQueryUser/RBACCustomerForbidden
    iam_operations_e2e_test.go:1217: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateEmployeeUserInput!) { createEmployeeUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676827819148161_7.png",
            "display_name": "Fixture u_employee_1794676827819148161_7",
            "email": "u_employee_1794676827819148161_7@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createEmployeeUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676827819148161_7.png",
              "created_at": "2026-11-15T00:20:27.820000936+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_employee_1794676827819148161_7",
              "id": "c29bbc64-27b4-4686-ab0d-8219bca0e393",
              "kind": "EMPLOYEE",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:27.820000936+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1218: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateUserInput!) { createUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676827864461914_8.png",
            "display_name": "Fixture u_customer_1794676827864461914_8",
            "email": "u_customer_1794676827864461914_8@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676827864461914_8.png",
              "created_at": "2026-11-15T00:20:27.86527829+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_customer_1794676827864461914_8",
              "id": "3331d5fd-7d2e-4956-a340-8ac916dcecf9",
              "kind": "CUSTOMER",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:27.86527829+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1218: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_customer_1794676827864461914_8@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "jd7AqKvd0j41oMVfZkfeyRZJJUWFYUe6hIUvmz3W0m4",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:1219: 
        [RBAC_TEST_AB_FORBIDDEN] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($id: ID!) { user(id: $id) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "id": "c29bbc64-27b4-4686-ab0d-8219bca0e393"
        }
        Result JSON:
        {
          "data": {
            "user": null
          },
          "errors": [
            {
              "message": "missing required permission: iam.users.read",
              "extensions": {
                "code": "FORBIDDEN"
              }
            }
          ]
        }
=== RUN   TestQueryUser/NotFoundError
    iam_operations_e2e_test.go:1224: 
        [GQL_EXPECTED_ERROR] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($id: ID!) { user(id: $id) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "id": "00000000-0000-4000-8000-000000000999"
        }
        Result JSON:
        {
          "data": {
            "user": null
          },
          "errors": [
            {
              "message": "user not found",
              "extensions": {
                "code": "NOT_FOUND"
              }
            }
          ]
        }
--- PASS: TestQueryUser (0.65s)
    --- PASS: TestQueryUser/ValidUserWithLoginNonNull (0.23s)
    --- PASS: TestQueryUser/RBACCustomerForbidden (0.36s)
    --- PASS: TestQueryUser/NotFoundError (0.05s)
=== RUN   TestQueryUsers
=== RUN   TestQueryUsers/FilterSearchKindStatusPaginationAndDeleted
    iam_operations_e2e_test.go:1240: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateVendorUserInput!) { createVendorUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_vendor_1794676828237508404_9.png",
            "display_name": "Fixture u_vendor_1794676828237508404_9",
            "email": "u_vendor_1794676828237508404_9@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createVendorUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_vendor_1794676828237508404_9.png",
              "created_at": "2026-11-15T00:20:28.238301828+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_vendor_1794676828237508404_9",
              "id": "d743c073-d8b4-497c-9f06-b34174b39bec",
              "kind": "VENDOR",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:28.238301828+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1241: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateEmployeeUserInput!) { createEmployeeUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676828286446252_10.png",
            "display_name": "Fixture u_employee_1794676828286446252_10",
            "email": "u_employee_1794676828286446252_10@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createEmployeeUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676828286446252_10.png",
              "created_at": "2026-11-15T00:20:28.287285222+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_employee_1794676828286446252_10",
              "id": "6267df00-8ee7-48e1-a0ce-80e64aa6999a",
              "kind": "EMPLOYEE",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:28.287285222+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1243: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($query: String, $kind: UserKind, $status: UserStatus, $include_deleted: Boolean, $page: PageInput) { users(query: $query, kind: $kind, status: $status, include_deleted: $include_deleted, page: $page) { items { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } page { total limit offset has_more } } }
        Input Variables:
        {
          "include_deleted": false,
          "kind": "VENDOR",
          "page": {
            "limit": 10,
            "offset": 0
          },
          "query": "u_vendor_1794676828237508404_9@integrione.test",
          "status": "ACTIVE"
        }
        Result JSON:
        {
          "data": {
            "users": {
              "items": [
                {
                  "auth_epoch": 1,
                  "avatar_url": "https://cdn.integrione.test/avatar/u_vendor_1794676828237508404_9.png",
                  "created_at": "2026-11-15T00:20:28.238301+07:00",
                  "deleted_at": null,
                  "display_name": "Fixture u_vendor_1794676828237508404_9",
                  "id": "d743c073-d8b4-497c-9f06-b34174b39bec",
                  "kind": "VENDOR",
                  "last_login_at": null,
                  "locale": "id-ID",
                  "perm_revision": 1,
                  "status": "ACTIVE",
                  "timezone": "Asia/Jakarta",
                  "updated_at": "2026-11-15T00:20:28.238301+07:00"
                }
              ],
              "page": {
                "has_more": false,
                "limit": 10,
                "offset": 0,
                "total": 1
              }
            }
          }
        }
    iam_operations_e2e_test.go:1262: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($id: ID!) { softDeleteUser(user_id: $id) }
        Input Variables:
        {
          "id": "6267df00-8ee7-48e1-a0ce-80e64aa6999a"
        }
        Result JSON:
        {
          "data": {
            "softDeleteUser": true
          }
        }
    iam_operations_e2e_test.go:1263: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($query: String, $kind: UserKind, $status: UserStatus, $include_deleted: Boolean, $page: PageInput) { users(query: $query, kind: $kind, status: $status, include_deleted: $include_deleted, page: $page) { items { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } page { total limit offset has_more } } }
        Input Variables:
        {
          "include_deleted": true,
          "page": {
            "limit": 10,
            "offset": 0
          },
          "query": "u_employee_1794676828286446252_10@integrione.test"
        }
        Result JSON:
        {
          "data": {
            "users": {
              "items": [
                {
                  "auth_epoch": 2,
                  "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676828286446252_10.png",
                  "created_at": "2026-11-15T00:20:28.287285+07:00",
                  "deleted_at": "2026-11-15T00:20:28.333531+07:00",
                  "display_name": "Fixture u_employee_1794676828286446252_10",
                  "id": "6267df00-8ee7-48e1-a0ce-80e64aa6999a",
                  "kind": "EMPLOYEE",
                  "last_login_at": null,
                  "locale": "id-ID",
                  "perm_revision": 1,
                  "status": "DISABLED",
                  "timezone": "Asia/Jakarta",
                  "updated_at": "2026-11-15T00:20:28.333431+07:00"
                }
              ],
              "page": {
                "has_more": false,
                "limit": 10,
                "offset": 0,
                "total": 1
              }
            }
          }
        }
=== RUN   TestQueryUsers/RBACCustomerForbidden
    iam_operations_e2e_test.go:1281: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateUserInput!) { createUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676828554343939_11.png",
            "display_name": "Fixture u_customer_1794676828554343939_11",
            "email": "u_customer_1794676828554343939_11@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676828554343939_11.png",
              "created_at": "2026-11-15T00:20:28.555281315+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_customer_1794676828554343939_11",
              "id": "f1c16375-3f89-4631-bc9a-432cdbced99d",
              "kind": "CUSTOMER",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:28.555281315+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1281: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_customer_1794676828554343939_11@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "4k3O1gP9ecPznQiC_7dPA2hceeoDoHyqNmNV0ygCcAo",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:1282: 
        [RBAC_TEST_AB_FORBIDDEN] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($query: String, $kind: UserKind, $status: UserStatus, $include_deleted: Boolean, $page: PageInput) { users(query: $query, kind: $kind, status: $status, include_deleted: $include_deleted, page: $page) { items { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } page { total limit offset has_more } } }
        Input Variables:
        {
          "page": {
            "limit": 5,
            "offset": 0
          }
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "missing required permission: iam.users.read",
              "extensions": {
                "code": "FORBIDDEN"
              }
            }
          ]
        }
--- PASS: TestQueryUsers (0.55s)
    --- PASS: TestQueryUsers/FilterSearchKindStatusPaginationAndDeleted (0.32s)
    --- PASS: TestQueryUsers/RBACCustomerForbidden (0.23s)
=== RUN   TestQueryUserIdentities
=== RUN   TestQueryUserIdentities/MultipleVerifiedIdentities
    iam_operations_e2e_test.go:1293: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateEmployeeUserInput!) { createEmployeeUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676828789778487_12.png",
            "display_name": "Fixture u_employee_1794676828789778487_12",
            "email": "u_employee_1794676828789778487_12@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta",
            "username": "u_employee_1794676828789778487_12"
          }
        }
        Result JSON:
        {
          "data": {
            "createEmployeeUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676828789778487_12.png",
              "created_at": "2026-11-15T00:20:28.790562382+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_employee_1794676828789778487_12",
              "id": "551f2219-0478-4bf7-931c-3151579dd729",
              "kind": "EMPLOYEE",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:28.790562382+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1295: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($user_id: ID!) { userIdentities(user_id: $user_id) { id user_id kind value is_verified is_primary verified_at created_at updated_at } }
        Input Variables:
        {
          "user_id": "551f2219-0478-4bf7-931c-3151579dd729"
        }
        Result JSON:
        {
          "data": {
            "userIdentities": [
              {
                "created_at": "2026-11-15T00:20:28.790562+07:00",
                "id": "460fa44a-c133-462f-b2bf-4dc0a352aacf",
                "is_primary": true,
                "is_verified": true,
                "kind": "EMAIL",
                "updated_at": "2026-11-15T00:20:28.790562+07:00",
                "user_id": "551f2219-0478-4bf7-931c-3151579dd729",
                "value": "u_employee_1794676828789778487_12@integrione.test",
                "verified_at": "2026-11-15T00:20:28.790562+07:00"
              },
              {
                "created_at": "2026-11-15T00:20:28.790562+07:00",
                "id": "82123473-ce8f-418e-9c2e-bae80b326a30",
                "is_primary": false,
                "is_verified": true,
                "kind": "USERNAME",
                "updated_at": "2026-11-15T00:20:28.790562+07:00",
                "user_id": "551f2219-0478-4bf7-931c-3151579dd729",
                "value": "u_employee_1794676828789778487_12",
                "verified_at": "2026-11-15T00:20:28.790562+07:00"
              }
            ]
          }
        }
=== RUN   TestQueryUserIdentities/RBACCustomerForbidden
    iam_operations_e2e_test.go:1309: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateUserInput!) { createUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676828974017981_13.png",
            "display_name": "Fixture u_customer_1794676828974017981_13",
            "email": "u_customer_1794676828974017981_13@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676828974017981_13.png",
              "created_at": "2026-11-15T00:20:28.974795965+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_customer_1794676828974017981_13",
              "id": "75ee362f-012f-4443-a257-93d76c9c4cdf",
              "kind": "CUSTOMER",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:28.974795965+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1309: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_customer_1794676828974017981_13@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "n4IbcneAE0dR9VTeE6T-8qAh9DLcCRYpCTQXjbj2mlE",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:1310: 
        [RBAC_TEST_AB_FORBIDDEN] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($user_id: ID!) { userIdentities(user_id: $user_id) { id user_id kind value is_verified is_primary verified_at created_at updated_at } }
        Input Variables:
        {
          "user_id": "75ee362f-012f-4443-a257-93d76c9c4cdf"
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "missing required permission: iam.users.read",
              "extensions": {
                "code": "FORBIDDEN"
              }
            }
          ]
        }
--- PASS: TestQueryUserIdentities (0.42s)
    --- PASS: TestQueryUserIdentities/MultipleVerifiedIdentities (0.19s)
    --- PASS: TestQueryUserIdentities/RBACCustomerForbidden (0.24s)
=== RUN   TestQueryUserRoles
=== RUN   TestQueryUserRoles/ScopedActiveAndExpiredFilter
    iam_operations_e2e_test.go:1323: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateEmployeeUserInput!) { createEmployeeUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676829211623291_14.png",
            "display_name": "Fixture u_employee_1794676829211623291_14",
            "email": "u_employee_1794676829211623291_14@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createEmployeeUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676829211623291_14.png",
              "created_at": "2026-11-15T00:20:29.212455608+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_employee_1794676829211623291_14",
              "id": "4d3200de-a1ef-4d9d-91dc-e2fd68290533",
              "kind": "EMPLOYEE",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:29.212455608+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1324: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateRoleInput!) { createRole(input: $input) { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "role_ur_1794676829257694369_15",
            "default_redirect_path": "/identity",
            "description": "Custom E2E role role_ur_1794676829257694369_15",
            "is_system": false,
            "management_domain": "iam",
            "name": "Role role_ur_1794676829257694369_15",
            "redirect_priority": 50
          }
        }
        Result JSON:
        {
          "data": {
            "createRole": {
              "code": "role_ur_1794676829257694369_15",
              "created_at": "2026-11-15T00:20:29.258597109+07:00",
              "default_redirect_path": "/identity",
              "description": "Custom E2E role role_ur_1794676829257694369_15",
              "id": "94df2194-5d36-48cd-a466-526b358ec6b1",
              "is_active": true,
              "is_system": false,
              "management_domain": "iam",
              "name": "Role role_ur_1794676829257694369_15",
              "redirect_priority": 50,
              "updated_at": "2026-11-15T00:20:29.258597109+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1328: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: AssignUserRoleInput!) { assignUserRole(input: $input) { id user_id role_id scope { type id } granted_by expires_at created_at } }
        Input Variables:
        {
          "input": {
            "expires_at": "2026-11-14T19:20:29Z",
            "role_id": "94df2194-5d36-48cd-a466-526b358ec6b1",
            "scope_id": "11111111-1111-4111-8111-111111111111",
            "scope_type": "FACILITY",
            "user_id": "4d3200de-a1ef-4d9d-91dc-e2fd68290533"
          }
        }
        Result JSON:
        {
          "data": {
            "assignUserRole": {
              "created_at": "2026-11-15T00:20:29.261064563+07:00",
              "expires_at": "2026-11-15T02:20:29+07:00",
              "granted_by": "853aac77-32e7-4a88-b2d6-7d68db8d13d7",
              "id": "76a5a079-ae13-4339-9a08-d68a554b0ac6",
              "role_id": "94df2194-5d36-48cd-a466-526b358ec6b1",
              "scope": {
                "id": "11111111-1111-4111-8111-111111111111",
                "type": "FACILITY"
              },
              "user_id": "4d3200de-a1ef-4d9d-91dc-e2fd68290533"
            }
          }
        }
    iam_operations_e2e_test.go:1343: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($user_id: ID!, $include_expired: Boolean) { userRoles(user_id: $user_id, include_expired: $include_expired) { id user_id role_id scope { type id } granted_by expires_at created_at } }
        Input Variables:
        {
          "include_expired": false,
          "user_id": "4d3200de-a1ef-4d9d-91dc-e2fd68290533"
        }
        Result JSON:
        {
          "data": {
            "userRoles": [
              {
                "created_at": "2026-11-15T00:20:29.261064+07:00",
                "expires_at": "2026-11-15T02:20:29+07:00",
                "granted_by": "853aac77-32e7-4a88-b2d6-7d68db8d13d7",
                "id": "76a5a079-ae13-4339-9a08-d68a554b0ac6",
                "role_id": "94df2194-5d36-48cd-a466-526b358ec6b1",
                "scope": {
                  "id": "11111111-1111-4111-8111-111111111111",
                  "type": "FACILITY"
                },
                "user_id": "4d3200de-a1ef-4d9d-91dc-e2fd68290533"
              }
            ]
          }
        }
    iam_operations_e2e_test.go:1358: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($user_id: ID!, $include_expired: Boolean) { userRoles(user_id: $user_id, include_expired: $include_expired) { id user_id role_id scope { type id } granted_by expires_at created_at } }
        Input Variables:
        {
          "include_expired": false,
          "user_id": "4d3200de-a1ef-4d9d-91dc-e2fd68290533"
        }
        Result JSON:
        {
          "data": {
            "userRoles": []
          }
        }
    iam_operations_e2e_test.go:1363: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($user_id: ID!, $include_expired: Boolean) { userRoles(user_id: $user_id, include_expired: $include_expired) { id user_id role_id scope { type id } granted_by expires_at created_at } }
        Input Variables:
        {
          "include_expired": true,
          "user_id": "4d3200de-a1ef-4d9d-91dc-e2fd68290533"
        }
        Result JSON:
        {
          "data": {
            "userRoles": [
              {
                "created_at": "2026-11-15T00:20:29.261064+07:00",
                "expires_at": "2026-11-14T23:20:29.301964+07:00",
                "granted_by": "853aac77-32e7-4a88-b2d6-7d68db8d13d7",
                "id": "76a5a079-ae13-4339-9a08-d68a554b0ac6",
                "role_id": "94df2194-5d36-48cd-a466-526b358ec6b1",
                "scope": {
                  "id": "11111111-1111-4111-8111-111111111111",
                  "type": "FACILITY"
                },
                "user_id": "4d3200de-a1ef-4d9d-91dc-e2fd68290533"
              }
            ]
          }
        }
=== RUN   TestQueryUserRoles/RBACCustomerForbidden
    iam_operations_e2e_test.go:1373: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateUserInput!) { createUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676829559596513_16.png",
            "display_name": "Fixture u_customer_1794676829559596513_16",
            "email": "u_customer_1794676829559596513_16@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676829559596513_16.png",
              "created_at": "2026-11-15T00:20:29.560309634+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_customer_1794676829559596513_16",
              "id": "51a3ec4d-8617-48df-96b5-f8a42150e7ac",
              "kind": "CUSTOMER",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:29.560309634+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1373: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_customer_1794676829559596513_16@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "eY51y7amzbOseLDa-PPR8mi0JU7tbDUtN_FbkL1Ei94",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:1374: 
        [RBAC_TEST_AB_FORBIDDEN] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($user_id: ID!, $include_expired: Boolean) { userRoles(user_id: $user_id, include_expired: $include_expired) { id user_id role_id scope { type id } granted_by expires_at created_at } }
        Input Variables:
        {
          "include_expired": false,
          "user_id": "51a3ec4d-8617-48df-96b5-f8a42150e7ac"
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "missing required permission: iam.roles.read",
              "extensions": {
                "code": "FORBIDDEN"
              }
            }
          ]
        }
--- PASS: TestQueryUserRoles (0.59s)
    --- PASS: TestQueryUserRoles/ScopedActiveAndExpiredFilter (0.35s)
    --- PASS: TestQueryUserRoles/RBACCustomerForbidden (0.24s)
=== RUN   TestQueryUserEffectiveRoles
=== RUN   TestQueryUserEffectiveRoles/DirectPlusGroupInheritedRoles
    iam_operations_e2e_test.go:1385: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateEmployeeUserInput!) { createEmployeeUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676829798273544_17.png",
            "display_name": "Fixture u_employee_1794676829798273544_17",
            "email": "u_employee_1794676829798273544_17@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createEmployeeUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676829798273544_17.png",
              "created_at": "2026-11-15T00:20:29.799020068+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_employee_1794676829798273544_17",
              "id": "c0c9495a-c092-4a61-ac93-ee132eef6074",
              "kind": "EMPLOYEE",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:29.799020068+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1386: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateRoleInput!) { createRole(input: $input) { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "role_eff_direct_1794676829844156605_18",
            "default_redirect_path": "/identity",
            "description": "Custom E2E role role_eff_direct_1794676829844156605_18",
            "is_system": false,
            "management_domain": "iam",
            "name": "Role role_eff_direct_1794676829844156605_18",
            "redirect_priority": 50
          }
        }
        Result JSON:
        {
          "data": {
            "createRole": {
              "code": "role_eff_direct_1794676829844156605_18",
              "created_at": "2026-11-15T00:20:29.845115461+07:00",
              "default_redirect_path": "/identity",
              "description": "Custom E2E role role_eff_direct_1794676829844156605_18",
              "id": "915072d4-9f27-431d-ab8f-92936e8a4b24",
              "is_active": true,
              "is_system": false,
              "management_domain": "iam",
              "name": "Role role_eff_direct_1794676829844156605_18",
              "redirect_priority": 50,
              "updated_at": "2026-11-15T00:20:29.845115461+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1387: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateRoleInput!) { createRole(input: $input) { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "role_eff_grp_1794676829845969770_19",
            "default_redirect_path": "/identity",
            "description": "Custom E2E role role_eff_grp_1794676829845969770_19",
            "is_system": false,
            "management_domain": "iam",
            "name": "Role role_eff_grp_1794676829845969770_19",
            "redirect_priority": 50
          }
        }
        Result JSON:
        {
          "data": {
            "createRole": {
              "code": "role_eff_grp_1794676829845969770_19",
              "created_at": "2026-11-15T00:20:29.846564747+07:00",
              "default_redirect_path": "/identity",
              "description": "Custom E2E role role_eff_grp_1794676829845969770_19",
              "id": "0409107b-d290-455f-a48a-4c780473595c",
              "is_active": true,
              "is_system": false,
              "management_domain": "iam",
              "name": "Role role_eff_grp_1794676829845969770_19",
              "redirect_priority": 50,
              "updated_at": "2026-11-15T00:20:29.846564747+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1388: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateGroupInput!) { createGroup(input: $input) { id code name description is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "grp_eff_1794676829847815937_20",
            "description": "Group E2E grp_eff_1794676829847815937_20",
            "name": "Group grp_eff_1794676829847815937_20"
          }
        }
        Result JSON:
        {
          "data": {
            "createGroup": {
              "code": "grp_eff_1794676829847815937_20",
              "created_at": "2026-11-15T00:20:29.848840639+07:00",
              "description": "Group E2E grp_eff_1794676829847815937_20",
              "id": "fcbd2d7b-01bc-432d-b904-ca1470752caf",
              "is_active": true,
              "name": "Group grp_eff_1794676829847815937_20",
              "updated_at": "2026-11-15T00:20:29.848840639+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1390: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: AssignUserRoleInput!) { assignUserRole(input: $input) { id user_id role_id scope { type id } granted_by expires_at created_at } }
        Input Variables:
        {
          "input": {
            "role_id": "915072d4-9f27-431d-ab8f-92936e8a4b24",
            "scope_type": "GLOBAL",
            "user_id": "c0c9495a-c092-4a61-ac93-ee132eef6074"
          }
        }
        Result JSON:
        {
          "data": {
            "assignUserRole": {
              "created_at": "2026-11-15T00:20:29.850967738+07:00",
              "expires_at": null,
              "granted_by": "853aac77-32e7-4a88-b2d6-7d68db8d13d7",
              "id": "b6112ccc-b3a5-4cdc-be1e-69c56403fdc5",
              "role_id": "915072d4-9f27-431d-ab8f-92936e8a4b24",
              "scope": {
                "id": null,
                "type": "GLOBAL"
              },
              "user_id": "c0c9495a-c092-4a61-ac93-ee132eef6074"
            }
          }
        }
    iam_operations_e2e_test.go:1395: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($g: ID!, $u: ID!) { addGroupMember(group_id: $g, user_id: $u) { group_id user_id created_at } }
        Input Variables:
        {
          "g": "fcbd2d7b-01bc-432d-b904-ca1470752caf",
          "u": "c0c9495a-c092-4a61-ac93-ee132eef6074"
        }
        Result JSON:
        {
          "data": {
            "addGroupMember": {
              "created_at": "2026-11-15T00:20:29.853347687+07:00",
              "group_id": "fcbd2d7b-01bc-432d-b904-ca1470752caf",
              "user_id": "c0c9495a-c092-4a61-ac93-ee132eef6074"
            }
          }
        }
    iam_operations_e2e_test.go:1400: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: AssignGroupRoleInput!) { assignGroupRole(input: $input) { id group_id role_id scope { type id } granted_by expires_at created_at } }
        Input Variables:
        {
          "input": {
            "group_id": "fcbd2d7b-01bc-432d-b904-ca1470752caf",
            "role_id": "0409107b-d290-455f-a48a-4c780473595c",
            "scope_type": "GLOBAL"
          }
        }
        Result JSON:
        {
          "data": {
            "assignGroupRole": {
              "created_at": "2026-11-15T00:20:29.855326405+07:00",
              "expires_at": null,
              "granted_by": "853aac77-32e7-4a88-b2d6-7d68db8d13d7",
              "group_id": "fcbd2d7b-01bc-432d-b904-ca1470752caf",
              "id": "995456ac-3693-44ef-8d65-dc2e94b5b48d",
              "role_id": "0409107b-d290-455f-a48a-4c780473595c",
              "scope": {
                "id": null,
                "type": "GLOBAL"
              }
            }
          }
        }
    iam_operations_e2e_test.go:1405: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($user_id: ID!) { userEffectiveRoles(user_id: $user_id) { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } }
        Input Variables:
        {
          "user_id": "c0c9495a-c092-4a61-ac93-ee132eef6074"
        }
        Result JSON:
        {
          "data": {
            "userEffectiveRoles": [
              {
                "code": "role_eff_direct_1794676829844156605_18",
                "created_at": "2026-11-15T00:20:29.845115+07:00",
                "default_redirect_path": "/identity",
                "description": "Custom E2E role role_eff_direct_1794676829844156605_18",
                "id": "915072d4-9f27-431d-ab8f-92936e8a4b24",
                "is_active": true,
                "is_system": false,
                "management_domain": "iam",
                "name": "Role role_eff_direct_1794676829844156605_18",
                "redirect_priority": 50,
                "updated_at": "2026-11-15T00:20:29.845115+07:00"
              },
              {
                "code": "role_eff_grp_1794676829845969770_19",
                "created_at": "2026-11-15T00:20:29.846564+07:00",
                "default_redirect_path": "/identity",
                "description": "Custom E2E role role_eff_grp_1794676829845969770_19",
                "id": "0409107b-d290-455f-a48a-4c780473595c",
                "is_active": true,
                "is_system": false,
                "management_domain": "iam",
                "name": "Role role_eff_grp_1794676829845969770_19",
                "redirect_priority": 50,
                "updated_at": "2026-11-15T00:20:29.846564+07:00"
              }
            ]
          }
        }
=== RUN   TestQueryUserEffectiveRoles/RBACCustomerForbidden
    iam_operations_e2e_test.go:1420: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateUserInput!) { createUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676830280916795_21.png",
            "display_name": "Fixture u_customer_1794676830280916795_21",
            "email": "u_customer_1794676830280916795_21@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676830280916795_21.png",
              "created_at": "2026-11-15T00:20:30.281643993+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_customer_1794676830280916795_21",
              "id": "c137035d-3177-46b9-b34d-0422c48160c0",
              "kind": "CUSTOMER",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:30.281643993+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1420: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_customer_1794676830280916795_21@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "bzMFAqHs12O_f1CUhTbCRVM8MdJ1Rw3DtDqCyqGbqO0",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:1421: 
        [RBAC_TEST_AB_FORBIDDEN] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($user_id: ID!) { userEffectiveRoles(user_id: $user_id) { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } }
        Input Variables:
        {
          "user_id": "c137035d-3177-46b9-b34d-0422c48160c0"
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "missing required permission: iam.roles.read",
              "extensions": {
                "code": "FORBIDDEN"
              }
            }
          ]
        }
--- PASS: TestQueryUserEffectiveRoles (0.72s)
    --- PASS: TestQueryUserEffectiveRoles/DirectPlusGroupInheritedRoles (0.48s)
    --- PASS: TestQueryUserEffectiveRoles/RBACCustomerForbidden (0.23s)
=== RUN   TestQueryUserGroups
=== RUN   TestQueryUserGroups/MemberAndEmptySliceNonNull
    iam_operations_e2e_test.go:1432: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateEmployeeUserInput!) { createEmployeeUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676830515734345_22.png",
            "display_name": "Fixture u_employee_1794676830515734345_22",
            "email": "u_employee_1794676830515734345_22@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createEmployeeUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676830515734345_22.png",
              "created_at": "2026-11-15T00:20:30.516471762+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_employee_1794676830515734345_22",
              "id": "b2315f75-c638-4b53-8d28-ec594bd76556",
              "kind": "EMPLOYEE",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:30.516471762+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1433: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateGroupInput!) { createGroup(input: $input) { id code name description is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "grp_ug_1794676830561735910_23",
            "description": "Group E2E grp_ug_1794676830561735910_23",
            "name": "Group grp_ug_1794676830561735910_23"
          }
        }
        Result JSON:
        {
          "data": {
            "createGroup": {
              "code": "grp_ug_1794676830561735910_23",
              "created_at": "2026-11-15T00:20:30.562649391+07:00",
              "description": "Group E2E grp_ug_1794676830561735910_23",
              "id": "51eced67-562a-48c4-b7e0-1100746b19b2",
              "is_active": true,
              "name": "Group grp_ug_1794676830561735910_23",
              "updated_at": "2026-11-15T00:20:30.562649391+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1435: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($user_id: ID!) { userGroups(user_id: $user_id) { id code name description is_active created_at updated_at } }
        Input Variables:
        {
          "user_id": "b2315f75-c638-4b53-8d28-ec594bd76556"
        }
        Result JSON:
        {
          "data": {
            "userGroups": []
          }
        }
    iam_operations_e2e_test.go:1441: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($g: ID!, $u: ID!) { addGroupMember(group_id: $g, user_id: $u) { group_id user_id created_at } }
        Input Variables:
        {
          "g": "51eced67-562a-48c4-b7e0-1100746b19b2",
          "u": "b2315f75-c638-4b53-8d28-ec594bd76556"
        }
        Result JSON:
        {
          "data": {
            "addGroupMember": {
              "created_at": "2026-11-15T00:20:30.566177444+07:00",
              "group_id": "51eced67-562a-48c4-b7e0-1100746b19b2",
              "user_id": "b2315f75-c638-4b53-8d28-ec594bd76556"
            }
          }
        }
    iam_operations_e2e_test.go:1446: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($user_id: ID!) { userGroups(user_id: $user_id) { id code name description is_active created_at updated_at } }
        Input Variables:
        {
          "user_id": "b2315f75-c638-4b53-8d28-ec594bd76556"
        }
        Result JSON:
        {
          "data": {
            "userGroups": [
              {
                "code": "grp_ug_1794676830561735910_23",
                "created_at": "2026-11-15T00:20:30.562649+07:00",
                "description": "Group E2E grp_ug_1794676830561735910_23",
                "id": "51eced67-562a-48c4-b7e0-1100746b19b2",
                "is_active": true,
                "name": "Group grp_ug_1794676830561735910_23",
                "updated_at": "2026-11-15T00:20:30.562649+07:00"
              }
            ]
          }
        }
=== RUN   TestQueryUserGroups/RBACCustomerForbidden
    iam_operations_e2e_test.go:1460: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateUserInput!) { createUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676830742747629_24.png",
            "display_name": "Fixture u_customer_1794676830742747629_24",
            "email": "u_customer_1794676830742747629_24@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676830742747629_24.png",
              "created_at": "2026-11-15T00:20:30.743538167+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_customer_1794676830742747629_24",
              "id": "62297942-0a36-4801-96bc-0970828ed330",
              "kind": "CUSTOMER",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:30.743538167+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1460: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_customer_1794676830742747629_24@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "1SXwAMCxZoRzQWFYfOuocBau6cLi739nNuavVXFHZeY",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:1461: 
        [RBAC_TEST_AB_FORBIDDEN] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($user_id: ID!) { userGroups(user_id: $user_id) { id code name description is_active created_at updated_at } }
        Input Variables:
        {
          "user_id": "62297942-0a36-4801-96bc-0970828ed330"
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "missing required permission: iam.groups.read",
              "extensions": {
                "code": "FORBIDDEN"
              }
            }
          ]
        }
--- PASS: TestQueryUserGroups (0.47s)
    --- PASS: TestQueryUserGroups/MemberAndEmptySliceNonNull (0.23s)
    --- PASS: TestQueryUserGroups/RBACCustomerForbidden (0.24s)
=== RUN   TestQueryUserSessions
=== RUN   TestQueryUserSessions/AdminInspectsUserActiveAndEndedSessions
    iam_operations_e2e_test.go:1477: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateEmployeeUserInput!) { createEmployeeUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676830982741615_25.png",
            "display_name": "Fixture u_employee_1794676830982741615_25",
            "email": "u_employee_1794676830982741615_25@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createEmployeeUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676830982741615_25.png",
              "created_at": "2026-11-15T00:20:30.983544907+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_employee_1794676830982741615_25",
              "id": "67dc93ff-7262-48bc-a827-614593498ac2",
              "kind": "EMPLOYEE",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:30.983544907+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1478: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_employee_1794676830982741615_25@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "1eLLIwx9bUrOasLMzrbB8OkGtGIlfmRfL_pkLezcdX0",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/account",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:1479: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_employee_1794676830982741615_25@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "1Mf-zGg2VDYOXwCvJ-ZIGUtrNggE2ox39hqwErHwGe8",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/account",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:1480: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation { userLogout }
        Input Variables:
        {}
        Result JSON:
        {
          "data": {
            "userLogout": true
          }
        }
    iam_operations_e2e_test.go:1483: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($user_id: ID!, $active_only: Boolean, $page: PageInput) { userSessions(user_id: $user_id, active_only: $active_only, page: $page) { items { id user_id mfa_verified mfa_verified_at ip_address user_agent authenticated_at last_active_at idle_expires_at absolute_expires_at ended_at end_reason created_at generation } page { total limit offset has_more } } }
        Input Variables:
        {
          "active_only": true,
          "page": {
            "limit": 10,
            "offset": 0
          },
          "user_id": "67dc93ff-7262-48bc-a827-614593498ac2"
        }
        Result JSON:
        {
          "data": {
            "userSessions": {
              "items": [
                {
                  "absolute_expires_at": "2026-11-15T08:20:31.078935+07:00",
                  "authenticated_at": "2026-11-15T00:20:31.078935+07:00",
                  "created_at": "2026-11-15T00:20:31.078935+07:00",
                  "end_reason": null,
                  "ended_at": null,
                  "generation": 1,
                  "id": "13966daa-82e3-4d66-b086-3c38a1ee3bc3",
                  "idle_expires_at": "2026-11-15T00:50:31.078935+07:00",
                  "ip_address": "172.22.0.1",
                  "last_active_at": "2026-11-15T00:20:31.078935+07:00",
                  "mfa_verified": false,
                  "mfa_verified_at": null,
                  "user_agent": "Go-http-client/1.1",
                  "user_id": "67dc93ff-7262-48bc-a827-614593498ac2"
                }
              ],
              "page": {
                "has_more": false,
                "limit": 10,
                "offset": 0,
                "total": 1
              }
            }
          }
        }
    iam_operations_e2e_test.go:1496: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($user_id: ID!, $active_only: Boolean, $page: PageInput) { userSessions(user_id: $user_id, active_only: $active_only, page: $page) { items { id user_id mfa_verified mfa_verified_at ip_address user_agent authenticated_at last_active_at idle_expires_at absolute_expires_at ended_at end_reason created_at generation } page { total limit offset has_more } } }
        Input Variables:
        {
          "active_only": false,
          "page": {
            "limit": 10,
            "offset": 0
          },
          "user_id": "67dc93ff-7262-48bc-a827-614593498ac2"
        }
        Result JSON:
        {
          "data": {
            "userSessions": {
              "items": [
                {
                  "absolute_expires_at": "2026-11-15T08:20:31.128653+07:00",
                  "authenticated_at": "2026-11-15T00:20:31.128653+07:00",
                  "created_at": "2026-11-15T00:20:31.128653+07:00",
                  "end_reason": "LOGOUT",
                  "ended_at": "2026-11-15T00:20:31.132463+07:00",
                  "generation": 1,
                  "id": "27e10dd7-bf70-4751-9b14-d99b63f06f5e",
                  "idle_expires_at": "2026-11-15T00:50:31.128653+07:00",
                  "ip_address": "172.22.0.1",
                  "last_active_at": "2026-11-15T00:20:31.128653+07:00",
                  "mfa_verified": false,
                  "mfa_verified_at": null,
                  "user_agent": "Go-http-client/1.1",
                  "user_id": "67dc93ff-7262-48bc-a827-614593498ac2"
                },
                {
                  "absolute_expires_at": "2026-11-15T08:20:31.078935+07:00",
                  "authenticated_at": "2026-11-15T00:20:31.078935+07:00",
                  "created_at": "2026-11-15T00:20:31.078935+07:00",
                  "end_reason": null,
                  "ended_at": null,
                  "generation": 1,
                  "id": "13966daa-82e3-4d66-b086-3c38a1ee3bc3",
                  "idle_expires_at": "2026-11-15T00:50:31.078935+07:00",
                  "ip_address": "172.22.0.1",
                  "last_active_at": "2026-11-15T00:20:31.078935+07:00",
                  "mfa_verified": false,
                  "mfa_verified_at": null,
                  "user_agent": "Go-http-client/1.1",
                  "user_id": "67dc93ff-7262-48bc-a827-614593498ac2"
                }
              ],
              "page": {
                "has_more": false,
                "limit": 10,
                "offset": 0,
                "total": 2
              }
            }
          }
        }
=== RUN   TestQueryUserSessions/RBACCustomerForbidden
    iam_operations_e2e_test.go:1514: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateUserInput!) { createUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676831272164229_26.png",
            "display_name": "Fixture u_customer_1794676831272164229_26",
            "email": "u_customer_1794676831272164229_26@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676831272164229_26.png",
              "created_at": "2026-11-15T00:20:31.272824399+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_customer_1794676831272164229_26",
              "id": "7c55ef83-83e5-4ffe-945e-b1673ea84de0",
              "kind": "CUSTOMER",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:31.272824399+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1514: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_customer_1794676831272164229_26@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "Is3uapojqgiG6mN2TvD5CAv4PX1bJYuwhI4vNTcDlys",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:1515: 
        [RBAC_TEST_AB_FORBIDDEN] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($user_id: ID!, $active_only: Boolean, $page: PageInput) { userSessions(user_id: $user_id, active_only: $active_only, page: $page) { items { id user_id mfa_verified mfa_verified_at ip_address user_agent authenticated_at last_active_at idle_expires_at absolute_expires_at ended_at end_reason created_at generation } page { total limit offset has_more } } }
        Input Variables:
        {
          "active_only": true,
          "user_id": "7c55ef83-83e5-4ffe-945e-b1673ea84de0"
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "missing required permission: iam.sessions.read",
              "extensions": {
                "code": "FORBIDDEN"
              }
            }
          ]
        }
--- PASS: TestQueryUserSessions (0.53s)
    --- PASS: TestQueryUserSessions/AdminInspectsUserActiveAndEndedSessions (0.29s)
    --- PASS: TestQueryUserSessions/RBACCustomerForbidden (0.24s)
=== RUN   TestQueryRoles
=== RUN   TestQueryRoles/FilterActiveAndInactiveCustomRoles
    iam_operations_e2e_test.go:1531: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateRoleInput!) { createRole(input: $input) { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "role_rq_act_1794676831510825900_27",
            "default_redirect_path": "/identity",
            "description": "Custom E2E role role_rq_act_1794676831510825900_27",
            "is_system": false,
            "management_domain": "iam",
            "name": "Role role_rq_act_1794676831510825900_27",
            "redirect_priority": 50
          }
        }
        Result JSON:
        {
          "data": {
            "createRole": {
              "code": "role_rq_act_1794676831510825900_27",
              "created_at": "2026-11-15T00:20:31.511566032+07:00",
              "default_redirect_path": "/identity",
              "description": "Custom E2E role role_rq_act_1794676831510825900_27",
              "id": "0501e1ba-352b-43c0-b0f0-d703c23bb018",
              "is_active": true,
              "is_system": false,
              "management_domain": "iam",
              "name": "Role role_rq_act_1794676831510825900_27",
              "redirect_priority": 50,
              "updated_at": "2026-11-15T00:20:31.511566032+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1532: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateRoleInput!) { createRole(input: $input) { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "role_rq_inact_1794676831512361028_28",
            "default_redirect_path": "/identity",
            "description": "Custom E2E role role_rq_inact_1794676831512361028_28",
            "is_system": false,
            "management_domain": "iam",
            "name": "Role role_rq_inact_1794676831512361028_28",
            "redirect_priority": 50
          }
        }
        Result JSON:
        {
          "data": {
            "createRole": {
              "code": "role_rq_inact_1794676831512361028_28",
              "created_at": "2026-11-15T00:20:31.512781655+07:00",
              "default_redirect_path": "/identity",
              "description": "Custom E2E role role_rq_inact_1794676831512361028_28",
              "id": "ceafe6c8-7a93-4681-a357-98dd7a47cd41",
              "is_active": true,
              "is_system": false,
              "management_domain": "iam",
              "name": "Role role_rq_inact_1794676831512361028_28",
              "redirect_priority": 50,
              "updated_at": "2026-11-15T00:20:31.512781655+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1533: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($id: ID!) { deactivateRole(id: $id) { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } }
        Input Variables:
        {
          "id": "ceafe6c8-7a93-4681-a357-98dd7a47cd41"
        }
        Result JSON:
        {
          "data": {
            "deactivateRole": {
              "code": "role_rq_inact_1794676831512361028_28",
              "created_at": "2026-11-15T00:20:31.512781+07:00",
              "default_redirect_path": "/identity",
              "description": "Custom E2E role role_rq_inact_1794676831512361028_28",
              "id": "ceafe6c8-7a93-4681-a357-98dd7a47cd41",
              "is_active": false,
              "is_system": false,
              "management_domain": "iam",
              "name": "Role role_rq_inact_1794676831512361028_28",
              "redirect_priority": 50,
              "updated_at": "2026-11-15T00:20:31.514019971+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1536: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($query: String, $management_domain: String, $active_only: Boolean, $page: PageInput) { roles(query: $query, management_domain: $management_domain, active_only: $active_only, page: $page) { items { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } page { total limit offset has_more } } }
        Input Variables:
        {
          "active_only": true,
          "management_domain": "iam",
          "page": {
            "limit": 10,
            "offset": 0
          },
          "query": "role_rq_act_1794676831510825900_27"
        }
        Result JSON:
        {
          "data": {
            "roles": {
              "items": [
                {
                  "code": "role_rq_act_1794676831510825900_27",
                  "created_at": "2026-11-15T00:20:31.511566+07:00",
                  "default_redirect_path": "/identity",
                  "description": "Custom E2E role role_rq_act_1794676831510825900_27",
                  "id": "0501e1ba-352b-43c0-b0f0-d703c23bb018",
                  "is_active": true,
                  "is_system": false,
                  "management_domain": "iam",
                  "name": "Role role_rq_act_1794676831510825900_27",
                  "redirect_priority": 50,
                  "updated_at": "2026-11-15T00:20:31.511566+07:00"
                }
              ],
              "page": {
                "has_more": false,
                "limit": 10,
                "offset": 0,
                "total": 1
              }
            }
          }
        }
    iam_operations_e2e_test.go:1550: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($query: String, $management_domain: String, $active_only: Boolean, $page: PageInput) { roles(query: $query, management_domain: $management_domain, active_only: $active_only, page: $page) { items { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } page { total limit offset has_more } } }
        Input Variables:
        {
          "active_only": true,
          "page": {
            "limit": 10,
            "offset": 0
          },
          "query": "role_rq_inact_1794676831512361028_28"
        }
        Result JSON:
        {
          "data": {
            "roles": {
              "items": [],
              "page": {
                "has_more": false,
                "limit": 10,
                "offset": 0,
                "total": 0
              }
            }
          }
        }
=== RUN   TestQueryRoles/RBACCustomerForbidden
    iam_operations_e2e_test.go:1564: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateUserInput!) { createUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676831826313741_29.png",
            "display_name": "Fixture u_customer_1794676831826313741_29",
            "email": "u_customer_1794676831826313741_29@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676831826313741_29.png",
              "created_at": "2026-11-15T00:20:31.827089922+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_customer_1794676831826313741_29",
              "id": "06db8e19-a8e8-46f5-ad6b-896850ec6756",
              "kind": "CUSTOMER",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:31.827089922+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1564: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_customer_1794676831826313741_29@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "4xb5VC6O_2oHK4l9KTGWzZhPUXErH0fY7R6rZMBuwV8",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:1565: 
        [RBAC_TEST_AB_FORBIDDEN] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($query: String, $management_domain: String, $active_only: Boolean, $page: PageInput) { roles(query: $query, management_domain: $management_domain, active_only: $active_only, page: $page) { items { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } page { total limit offset has_more } } }
        Input Variables:
        {
          "page": {
            "limit": 5,
            "offset": 0
          }
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "missing required permission: iam.roles.read",
              "extensions": {
                "code": "FORBIDDEN"
              }
            }
          ]
        }
--- PASS: TestQueryRoles (0.55s)
    --- PASS: TestQueryRoles/FilterActiveAndInactiveCustomRoles (0.31s)
    --- PASS: TestQueryRoles/RBACCustomerForbidden (0.24s)
=== RUN   TestQueryRole
=== RUN   TestQueryRole/ValidRoleNonNull
    iam_operations_e2e_test.go:1576: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateRoleInput!) { createRole(input: $input) { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "role_single_1794676832062514952_30",
            "default_redirect_path": "/identity",
            "description": "Custom E2E role role_single_1794676832062514952_30",
            "is_system": false,
            "management_domain": "iam",
            "name": "Role role_single_1794676832062514952_30",
            "redirect_priority": 50
          }
        }
        Result JSON:
        {
          "data": {
            "createRole": {
              "code": "role_single_1794676832062514952_30",
              "created_at": "2026-11-15T00:20:32.063272798+07:00",
              "default_redirect_path": "/identity",
              "description": "Custom E2E role role_single_1794676832062514952_30",
              "id": "9ed4e168-e931-4c6f-95c6-97220b65b314",
              "is_active": true,
              "is_system": false,
              "management_domain": "iam",
              "name": "Role role_single_1794676832062514952_30",
              "redirect_priority": 50,
              "updated_at": "2026-11-15T00:20:32.063272798+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1577: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($id: ID!) { role(id: $id) { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } }
        Input Variables:
        {
          "id": "9ed4e168-e931-4c6f-95c6-97220b65b314"
        }
        Result JSON:
        {
          "data": {
            "role": {
              "code": "role_single_1794676832062514952_30",
              "created_at": "2026-11-15T00:20:32.063272+07:00",
              "default_redirect_path": "/identity",
              "description": "Custom E2E role role_single_1794676832062514952_30",
              "id": "9ed4e168-e931-4c6f-95c6-97220b65b314",
              "is_active": true,
              "is_system": false,
              "management_domain": "iam",
              "name": "Role role_single_1794676832062514952_30",
              "redirect_priority": 50,
              "updated_at": "2026-11-15T00:20:32.063272+07:00"
            }
          }
        }
=== RUN   TestQueryRole/RBACCustomerForbidden
    iam_operations_e2e_test.go:1587: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateRoleInput!) { createRole(input: $input) { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "role_rbac_1794676832236147282_31",
            "default_redirect_path": "/identity",
            "description": "Custom E2E role role_rbac_1794676832236147282_31",
            "is_system": false,
            "management_domain": "iam",
            "name": "Role role_rbac_1794676832236147282_31",
            "redirect_priority": 50
          }
        }
        Result JSON:
        {
          "data": {
            "createRole": {
              "code": "role_rbac_1794676832236147282_31",
              "created_at": "2026-11-15T00:20:32.236876253+07:00",
              "default_redirect_path": "/identity",
              "description": "Custom E2E role role_rbac_1794676832236147282_31",
              "id": "5eae7620-f6ee-4c6a-b787-61aaf405f77d",
              "is_active": true,
              "is_system": false,
              "management_domain": "iam",
              "name": "Role role_rbac_1794676832236147282_31",
              "redirect_priority": 50,
              "updated_at": "2026-11-15T00:20:32.236876253+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1588: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateUserInput!) { createUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676832237669966_32.png",
            "display_name": "Fixture u_customer_1794676832237669966_32",
            "email": "u_customer_1794676832237669966_32@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676832237669966_32.png",
              "created_at": "2026-11-15T00:20:32.238191054+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_customer_1794676832237669966_32",
              "id": "4bfa518a-fcd4-414c-a4b5-edb5fd32ae7d",
              "kind": "CUSTOMER",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:32.238191054+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1588: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_customer_1794676832237669966_32@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "Y5Ga5pi14Z7JTp9VGpnQxNHL8veySeP5F4TYji_TSuA",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:1589: 
        [RBAC_TEST_AB_FORBIDDEN] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($id: ID!) { role(id: $id) { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } }
        Input Variables:
        {
          "id": "5eae7620-f6ee-4c6a-b787-61aaf405f77d"
        }
        Result JSON:
        {
          "data": {
            "role": null
          },
          "errors": [
            {
              "message": "missing required permission: iam.roles.read",
              "extensions": {
                "code": "FORBIDDEN"
              }
            }
          ]
        }
--- PASS: TestQueryRole (0.55s)
    --- PASS: TestQueryRole/ValidRoleNonNull (0.18s)
    --- PASS: TestQueryRole/RBACCustomerForbidden (0.37s)
=== RUN   TestQueryRoleByCode
=== RUN   TestQueryRoleByCode/SystemAndCustomRoleByCode
    iam_operations_e2e_test.go:1600: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($code: String!) { roleByCode(code: $code) { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } }
        Input Variables:
        {
          "code": "superadmin"
        }
        Result JSON:
        {
          "data": {
            "roleByCode": {
              "code": "superadmin",
              "created_at": "2026-11-14T23:13:35.92827+07:00",
              "default_redirect_path": "/identity",
              "description": "IntegriOne full platform and IAM governance super administrator",
              "id": "d8ffac94-b437-4aa3-a937-27556b86d2dd",
              "is_active": true,
              "is_system": true,
              "management_domain": "iam",
              "name": "Super Administrator",
              "redirect_priority": 1,
              "updated_at": "2026-11-14T23:13:35.92827+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1607: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateRoleInput!) { createRole(input: $input) { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "role_bycode_1794676832614439540_33",
            "default_redirect_path": "/identity",
            "description": "Custom E2E role role_bycode_1794676832614439540_33",
            "is_system": false,
            "management_domain": "iam",
            "name": "Role role_bycode_1794676832614439540_33",
            "redirect_priority": 50
          }
        }
        Result JSON:
        {
          "data": {
            "createRole": {
              "code": "role_bycode_1794676832614439540_33",
              "created_at": "2026-11-15T00:20:32.615443723+07:00",
              "default_redirect_path": "/identity",
              "description": "Custom E2E role role_bycode_1794676832614439540_33",
              "id": "8f09a6e9-d865-4062-a92d-3d1cccb992e3",
              "is_active": true,
              "is_system": false,
              "management_domain": "iam",
              "name": "Role role_bycode_1794676832614439540_33",
              "redirect_priority": 50,
              "updated_at": "2026-11-15T00:20:32.615443723+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1608: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($code: String!) { roleByCode(code: $code) { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } }
        Input Variables:
        {
          "code": "role_bycode_1794676832614439540_33"
        }
        Result JSON:
        {
          "data": {
            "roleByCode": {
              "code": "role_bycode_1794676832614439540_33",
              "created_at": "2026-11-15T00:20:32.615443+07:00",
              "default_redirect_path": "/identity",
              "description": "Custom E2E role role_bycode_1794676832614439540_33",
              "id": "8f09a6e9-d865-4062-a92d-3d1cccb992e3",
              "is_active": true,
              "is_system": false,
              "management_domain": "iam",
              "name": "Role role_bycode_1794676832614439540_33",
              "redirect_priority": 50,
              "updated_at": "2026-11-15T00:20:32.615443+07:00"
            }
          }
        }
=== RUN   TestQueryRoleByCode/RBACCustomerForbidden
    iam_operations_e2e_test.go:1618: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateUserInput!) { createUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676832804563011_34.png",
            "display_name": "Fixture u_customer_1794676832804563011_34",
            "email": "u_customer_1794676832804563011_34@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676832804563011_34.png",
              "created_at": "2026-11-15T00:20:32.805327059+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_customer_1794676832804563011_34",
              "id": "3db693c2-e87e-48ff-94ad-d892976bf495",
              "kind": "CUSTOMER",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:32.805327059+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1618: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_customer_1794676832804563011_34@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "Bd-K3TOpXnQ9CWU4MKNwmXiY-rP1-SifzZ6l9RiHdZM",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:1619: 
        [RBAC_TEST_AB_FORBIDDEN] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($code: String!) { roleByCode(code: $code) { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } }
        Input Variables:
        {
          "code": "superadmin"
        }
        Result JSON:
        {
          "data": {
            "roleByCode": null
          },
          "errors": [
            {
              "message": "missing required permission: iam.roles.read",
              "extensions": {
                "code": "FORBIDDEN"
              }
            }
          ]
        }
--- PASS: TestQueryRoleByCode (0.44s)
    --- PASS: TestQueryRoleByCode/SystemAndCustomRoleByCode (0.19s)
    --- PASS: TestQueryRoleByCode/RBACCustomerForbidden (0.25s)
=== RUN   TestQueryRolePermissions
=== RUN   TestQueryRolePermissions/AssignedAndUnassignedPermissionsNonNull
    iam_operations_e2e_test.go:1630: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateRoleInput!) { createRole(input: $input) { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "role_rperms_1794676833052635490_35",
            "default_redirect_path": "/identity",
            "description": "Custom E2E role role_rperms_1794676833052635490_35",
            "is_system": false,
            "management_domain": "iam",
            "name": "Role role_rperms_1794676833052635490_35",
            "redirect_priority": 50
          }
        }
        Result JSON:
        {
          "data": {
            "createRole": {
              "code": "role_rperms_1794676833052635490_35",
              "created_at": "2026-11-15T00:20:33.053542839+07:00",
              "default_redirect_path": "/identity",
              "description": "Custom E2E role role_rperms_1794676833052635490_35",
              "id": "cc53d5a8-89a8-4b29-8e32-178cfc0ce638",
              "is_active": true,
              "is_system": false,
              "management_domain": "iam",
              "name": "Role role_rperms_1794676833052635490_35",
              "redirect_priority": 50,
              "updated_at": "2026-11-15T00:20:33.053542839+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1631: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($query: String) { permissions(query: $query, page: { limit: 25, offset: 0 }) { items { id code resource action description is_sensitive created_at } page { total limit offset has_more } } }
        Input Variables:
        {
          "query": "iam.users.read"
        }
        Result JSON:
        {
          "data": {
            "permissions": {
              "items": [
                {
                  "action": "read",
                  "code": "iam.users.read",
                  "created_at": "2026-11-14T23:13:35.92544+07:00",
                  "description": "Read IAM user principals and identities",
                  "id": "92f3f3db-af8f-49d8-85f9-7287262abbdf",
                  "is_sensitive": false,
                  "resource": "users"
                }
              ],
              "page": {
                "has_more": false,
                "limit": 25,
                "offset": 0,
                "total": 1
              }
            }
          }
        }
    iam_operations_e2e_test.go:1633: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($role_id: ID!) { rolePermissions(role_id: $role_id) { id code resource action description is_sensitive created_at } }
        Input Variables:
        {
          "role_id": "cc53d5a8-89a8-4b29-8e32-178cfc0ce638"
        }
        Result JSON:
        {
          "data": {
            "rolePermissions": []
          }
        }
    iam_operations_e2e_test.go:1638: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($r: ID!, $p: ID!) { assignRolePermission(role_id: $r, permission_id: $p) { role_id permission_id created_at } }
        Input Variables:
        {
          "p": "92f3f3db-af8f-49d8-85f9-7287262abbdf",
          "r": "cc53d5a8-89a8-4b29-8e32-178cfc0ce638"
        }
        Result JSON:
        {
          "data": {
            "assignRolePermission": {
              "created_at": "2026-11-15T00:20:33.058499379+07:00",
              "permission_id": "92f3f3db-af8f-49d8-85f9-7287262abbdf",
              "role_id": "cc53d5a8-89a8-4b29-8e32-178cfc0ce638"
            }
          }
        }
    iam_operations_e2e_test.go:1643: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($role_id: ID!) { rolePermissions(role_id: $role_id) { id code resource action description is_sensitive created_at } }
        Input Variables:
        {
          "role_id": "cc53d5a8-89a8-4b29-8e32-178cfc0ce638"
        }
        Result JSON:
        {
          "data": {
            "rolePermissions": [
              {
                "action": "read",
                "code": "iam.users.read",
                "created_at": "2026-11-14T23:13:35.92544+07:00",
                "description": "Read IAM user principals and identities",
                "id": "92f3f3db-af8f-49d8-85f9-7287262abbdf",
                "is_sensitive": false,
                "resource": "users"
              }
            ]
          }
        }
=== RUN   TestQueryRolePermissions/RBACCustomerForbidden
    iam_operations_e2e_test.go:1657: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateRoleInput!) { createRole(input: $input) { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "role_rperms_rbac_1794676833249567609_36",
            "default_redirect_path": "/identity",
            "description": "Custom E2E role role_rperms_rbac_1794676833249567609_36",
            "is_system": false,
            "management_domain": "iam",
            "name": "Role role_rperms_rbac_1794676833249567609_36",
            "redirect_priority": 50
          }
        }
        Result JSON:
        {
          "data": {
            "createRole": {
              "code": "role_rperms_rbac_1794676833249567609_36",
              "created_at": "2026-11-15T00:20:33.250550351+07:00",
              "default_redirect_path": "/identity",
              "description": "Custom E2E role role_rperms_rbac_1794676833249567609_36",
              "id": "f794930c-6321-4f76-bd80-0b90f8de4194",
              "is_active": true,
              "is_system": false,
              "management_domain": "iam",
              "name": "Role role_rperms_rbac_1794676833249567609_36",
              "redirect_priority": 50,
              "updated_at": "2026-11-15T00:20:33.250550351+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1658: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateUserInput!) { createUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676833251470625_37.png",
            "display_name": "Fixture u_customer_1794676833251470625_37",
            "email": "u_customer_1794676833251470625_37@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676833251470625_37.png",
              "created_at": "2026-11-15T00:20:33.252047467+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_customer_1794676833251470625_37",
              "id": "c9bdeec7-0bef-40e2-9f13-3aac311916d3",
              "kind": "CUSTOMER",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:33.252047467+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1658: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_customer_1794676833251470625_37@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "nhuUnKs_B-qIAL6KC0ytvTeVohzcsMJNstEQe9BUzSI",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:1659: 
        [RBAC_TEST_AB_FORBIDDEN] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($role_id: ID!) { rolePermissions(role_id: $role_id) { id code resource action description is_sensitive created_at } }
        Input Variables:
        {
          "role_id": "f794930c-6321-4f76-bd80-0b90f8de4194"
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "missing required permission: iam.roles.read",
              "extensions": {
                "code": "FORBIDDEN"
              }
            }
          ]
        }
--- PASS: TestQueryRolePermissions (0.57s)
    --- PASS: TestQueryRolePermissions/AssignedAndUnassignedPermissionsNonNull (0.20s)
    --- PASS: TestQueryRolePermissions/RBACCustomerForbidden (0.37s)
=== RUN   TestQueryPermissions
=== RUN   TestQueryPermissions/FilterByResourceSensitiveAndPagination
    iam_operations_e2e_test.go:1675: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($query: String, $resource: String, $is_sensitive: Boolean, $page: PageInput) { permissions(query: $query, resource: $resource, is_sensitive: $is_sensitive, page: $page) { items { id code resource action description is_sensitive created_at } page { total limit offset has_more } } }
        Input Variables:
        {
          "is_sensitive": true,
          "page": {
            "limit": 10,
            "offset": 0
          },
          "query": "iam.",
          "resource": "users"
        }
        Result JSON:
        {
          "data": {
            "permissions": {
              "items": [
                {
                  "action": "write",
                  "code": "iam.users.write",
                  "created_at": "2026-11-14T23:13:35.925869+07:00",
                  "description": "Create, update, suspend, or disable IAM users",
                  "id": "902a937b-958d-4a5d-86cb-cfd64c128e64",
                  "is_sensitive": true,
                  "resource": "users"
                }
              ],
              "page": {
                "has_more": false,
                "limit": 10,
                "offset": 0,
                "total": 1
              }
            }
          }
        }
=== RUN   TestQueryPermissions/RBACCustomerForbidden
    iam_operations_e2e_test.go:1698: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateUserInput!) { createUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676833680519256_38.png",
            "display_name": "Fixture u_customer_1794676833680519256_38",
            "email": "u_customer_1794676833680519256_38@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676833680519256_38.png",
              "created_at": "2026-11-15T00:20:33.681274136+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_customer_1794676833680519256_38",
              "id": "121a0792-4a52-446b-ac39-241ec3d957f0",
              "kind": "CUSTOMER",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:33.681274136+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1698: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_customer_1794676833680519256_38@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "JqAi3lmJwqDDyhRRWnOe0n_8YGwY1JiiL3OGkz3LEhk",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:1699: 
        [RBAC_TEST_AB_FORBIDDEN] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($query: String, $resource: String, $is_sensitive: Boolean, $page: PageInput) { permissions(query: $query, resource: $resource, is_sensitive: $is_sensitive, page: $page) { items { id code resource action description is_sensitive created_at } page { total limit offset has_more } } }
        Input Variables:
        {
          "page": {
            "limit": 5,
            "offset": 0
          }
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "missing required permission: iam.permissions.read",
              "extensions": {
                "code": "FORBIDDEN"
              }
            }
          ]
        }
--- PASS: TestQueryPermissions (0.30s)
    --- PASS: TestQueryPermissions/FilterByResourceSensitiveAndPagination (0.06s)
    --- PASS: TestQueryPermissions/RBACCustomerForbidden (0.24s)
=== RUN   TestQueryPermission
=== RUN   TestQueryPermission/ValidPermissionByID
    iam_operations_e2e_test.go:1710: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($query: String) { permissions(query: $query, page: { limit: 25, offset: 0 }) { items { id code resource action description is_sensitive created_at } page { total limit offset has_more } } }
        Input Variables:
        {
          "query": "iam.users.read"
        }
        Result JSON:
        {
          "data": {
            "permissions": {
              "items": [
                {
                  "action": "read",
                  "code": "iam.users.read",
                  "created_at": "2026-11-14T23:13:35.92544+07:00",
                  "description": "Read IAM user principals and identities",
                  "id": "92f3f3db-af8f-49d8-85f9-7287262abbdf",
                  "is_sensitive": false,
                  "resource": "users"
                }
              ],
              "page": {
                "has_more": false,
                "limit": 25,
                "offset": 0,
                "total": 1
              }
            }
          }
        }
    iam_operations_e2e_test.go:1711: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($id: ID!) { permission(id: $id) { id code resource action description is_sensitive created_at } }
        Input Variables:
        {
          "id": "92f3f3db-af8f-49d8-85f9-7287262abbdf"
        }
        Result JSON:
        {
          "data": {
            "permission": {
              "action": "read",
              "code": "iam.users.read",
              "created_at": "2026-11-14T23:13:35.92544+07:00",
              "description": "Read IAM user principals and identities",
              "id": "92f3f3db-af8f-49d8-85f9-7287262abbdf",
              "is_sensitive": false,
              "resource": "users"
            }
          }
        }
=== RUN   TestQueryPermission/RBACCustomerForbidden
    iam_operations_e2e_test.go:1721: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($query: String) { permissions(query: $query, page: { limit: 25, offset: 0 }) { items { id code resource action description is_sensitive created_at } page { total limit offset has_more } } }
        Input Variables:
        {
          "query": "iam.users.read"
        }
        Result JSON:
        {
          "data": {
            "permissions": {
              "items": [
                {
                  "action": "read",
                  "code": "iam.users.read",
                  "created_at": "2026-11-14T23:13:35.92544+07:00",
                  "description": "Read IAM user principals and identities",
                  "id": "92f3f3db-af8f-49d8-85f9-7287262abbdf",
                  "is_sensitive": false,
                  "resource": "users"
                }
              ],
              "page": {
                "has_more": false,
                "limit": 25,
                "offset": 0,
                "total": 1
              }
            }
          }
        }
    iam_operations_e2e_test.go:1722: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateUserInput!) { createUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676833980577287_39.png",
            "display_name": "Fixture u_customer_1794676833980577287_39",
            "email": "u_customer_1794676833980577287_39@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676833980577287_39.png",
              "created_at": "2026-11-15T00:20:33.981111679+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_customer_1794676833980577287_39",
              "id": "233564e6-405c-4b43-96fc-47b067f9bff3",
              "kind": "CUSTOMER",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:33.981111679+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1722: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_customer_1794676833980577287_39@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "mmSXp-IwdNKOiZjIFqX4gc8vU0dbei12RbOC1pa0hhc",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:1723: 
        [RBAC_TEST_AB_FORBIDDEN] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($id: ID!) { permission(id: $id) { id code resource action description is_sensitive created_at } }
        Input Variables:
        {
          "id": "92f3f3db-af8f-49d8-85f9-7287262abbdf"
        }
        Result JSON:
        {
          "data": {
            "permission": null
          },
          "errors": [
            {
              "message": "missing required permission: iam.permissions.read",
              "extensions": {
                "code": "FORBIDDEN"
              }
            }
          ]
        }
--- PASS: TestQueryPermission (0.29s)
    --- PASS: TestQueryPermission/ValidPermissionByID (0.06s)
    --- PASS: TestQueryPermission/RBACCustomerForbidden (0.23s)
=== RUN   TestQueryGroups
=== RUN   TestQueryGroups/FilterActiveAndInactiveGroups
    iam_operations_e2e_test.go:1739: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateGroupInput!) { createGroup(input: $input) { id code name description is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "grp_act_1794676834211868847_40",
            "description": "Group E2E grp_act_1794676834211868847_40",
            "name": "Group grp_act_1794676834211868847_40"
          }
        }
        Result JSON:
        {
          "data": {
            "createGroup": {
              "code": "grp_act_1794676834211868847_40",
              "created_at": "2026-11-15T00:20:34.212679272+07:00",
              "description": "Group E2E grp_act_1794676834211868847_40",
              "id": "ce0583e7-47d1-43d6-b8f6-a01cb6a87554",
              "is_active": true,
              "name": "Group grp_act_1794676834211868847_40",
              "updated_at": "2026-11-15T00:20:34.212679272+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1740: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateGroupInput!) { createGroup(input: $input) { id code name description is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "grp_inact_1794676834213912839_41",
            "description": "Group E2E grp_inact_1794676834213912839_41",
            "name": "Group grp_inact_1794676834213912839_41"
          }
        }
        Result JSON:
        {
          "data": {
            "createGroup": {
              "code": "grp_inact_1794676834213912839_41",
              "created_at": "2026-11-15T00:20:34.214637832+07:00",
              "description": "Group E2E grp_inact_1794676834213912839_41",
              "id": "799eb7d7-3910-43bd-8655-619c984558ed",
              "is_active": true,
              "name": "Group grp_inact_1794676834213912839_41",
              "updated_at": "2026-11-15T00:20:34.214637832+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1741: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($id: ID!) { deactivateGroup(id: $id) { id code name description is_active created_at updated_at } }
        Input Variables:
        {
          "id": "799eb7d7-3910-43bd-8655-619c984558ed"
        }
        Result JSON:
        {
          "data": {
            "deactivateGroup": {
              "code": "grp_inact_1794676834213912839_41",
              "created_at": "2026-11-15T00:20:34.214637+07:00",
              "description": "Group E2E grp_inact_1794676834213912839_41",
              "id": "799eb7d7-3910-43bd-8655-619c984558ed",
              "is_active": false,
              "name": "Group grp_inact_1794676834213912839_41",
              "updated_at": "2026-11-15T00:20:34.216022796+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1744: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($query: String, $active_only: Boolean, $page: PageInput) { groups(query: $query, active_only: $active_only, page: $page) { items { id code name description is_active created_at updated_at } page { total limit offset has_more } } }
        Input Variables:
        {
          "active_only": true,
          "page": {
            "limit": 10,
            "offset": 0
          },
          "query": "grp_act_1794676834211868847_40"
        }
        Result JSON:
        {
          "data": {
            "groups": {
              "items": [
                {
                  "code": "grp_act_1794676834211868847_40",
                  "created_at": "2026-11-15T00:20:34.212679+07:00",
                  "description": "Group E2E grp_act_1794676834211868847_40",
                  "id": "ce0583e7-47d1-43d6-b8f6-a01cb6a87554",
                  "is_active": true,
                  "name": "Group grp_act_1794676834211868847_40",
                  "updated_at": "2026-11-15T00:20:34.212679+07:00"
                }
              ],
              "page": {
                "has_more": false,
                "limit": 10,
                "offset": 0,
                "total": 1
              }
            }
          }
        }
    iam_operations_e2e_test.go:1757: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($query: String, $active_only: Boolean, $page: PageInput) { groups(query: $query, active_only: $active_only, page: $page) { items { id code name description is_active created_at updated_at } page { total limit offset has_more } } }
        Input Variables:
        {
          "active_only": false,
          "page": {
            "limit": 10,
            "offset": 0
          },
          "query": "grp_inact_1794676834213912839_41"
        }
        Result JSON:
        {
          "data": {
            "groups": {
              "items": [
                {
                  "code": "grp_inact_1794676834213912839_41",
                  "created_at": "2026-11-15T00:20:34.214637+07:00",
                  "description": "Group E2E grp_inact_1794676834213912839_41",
                  "id": "799eb7d7-3910-43bd-8655-619c984558ed",
                  "is_active": false,
                  "name": "Group grp_inact_1794676834213912839_41",
                  "updated_at": "2026-11-15T00:20:34.215879+07:00"
                }
              ],
              "page": {
                "has_more": false,
                "limit": 10,
                "offset": 0,
                "total": 1
              }
            }
          }
        }
=== RUN   TestQueryGroups/RBACCustomerForbidden
    iam_operations_e2e_test.go:1773: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateUserInput!) { createUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676834354302253_42.png",
            "display_name": "Fixture u_customer_1794676834354302253_42",
            "email": "u_customer_1794676834354302253_42@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676834354302253_42.png",
              "created_at": "2026-11-15T00:20:34.355011337+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_customer_1794676834354302253_42",
              "id": "8b598a90-72f3-46bc-8137-ba5d61ec45c6",
              "kind": "CUSTOMER",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:34.355011337+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1773: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_customer_1794676834354302253_42@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "5g7g0yN7iGP0JUDXOOj3SpFtqkMI2Y992SP0vUWx4Lw",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:1774: 
        [RBAC_TEST_AB_FORBIDDEN] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($query: String, $active_only: Boolean, $page: PageInput) { groups(query: $query, active_only: $active_only, page: $page) { items { id code name description is_active created_at updated_at } page { total limit offset has_more } } }
        Input Variables:
        {
          "page": {
            "limit": 5,
            "offset": 0
          }
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "missing required permission: iam.groups.read",
              "extensions": {
                "code": "FORBIDDEN"
              }
            }
          ]
        }
--- PASS: TestQueryGroups (0.38s)
    --- PASS: TestQueryGroups/FilterActiveAndInactiveGroups (0.14s)
    --- PASS: TestQueryGroups/RBACCustomerForbidden (0.24s)
=== RUN   TestQueryGroup
=== RUN   TestQueryGroup/ValidGroupNonNull
    iam_operations_e2e_test.go:1785: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateGroupInput!) { createGroup(input: $input) { id code name description is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "grp_single_1794676834592322410_43",
            "description": "Group E2E grp_single_1794676834592322410_43",
            "name": "Group grp_single_1794676834592322410_43"
          }
        }
        Result JSON:
        {
          "data": {
            "createGroup": {
              "code": "grp_single_1794676834592322410_43",
              "created_at": "2026-11-15T00:20:34.593026784+07:00",
              "description": "Group E2E grp_single_1794676834592322410_43",
              "id": "3a91ea50-d54b-4d36-bf52-8b5581cb6c07",
              "is_active": true,
              "name": "Group grp_single_1794676834592322410_43",
              "updated_at": "2026-11-15T00:20:34.593026784+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1786: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($id: ID!) { group(id: $id) { id code name description is_active created_at updated_at } }
        Input Variables:
        {
          "id": "3a91ea50-d54b-4d36-bf52-8b5581cb6c07"
        }
        Result JSON:
        {
          "data": {
            "group": {
              "code": "grp_single_1794676834592322410_43",
              "created_at": "2026-11-15T00:20:34.593026+07:00",
              "description": "Group E2E grp_single_1794676834592322410_43",
              "id": "3a91ea50-d54b-4d36-bf52-8b5581cb6c07",
              "is_active": true,
              "name": "Group grp_single_1794676834592322410_43",
              "updated_at": "2026-11-15T00:20:34.593026+07:00"
            }
          }
        }
=== RUN   TestQueryGroup/RBACCustomerForbidden
    iam_operations_e2e_test.go:1796: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateGroupInput!) { createGroup(input: $input) { id code name description is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "grp_rbac_1794676834687230374_44",
            "description": "Group E2E grp_rbac_1794676834687230374_44",
            "name": "Group grp_rbac_1794676834687230374_44"
          }
        }
        Result JSON:
        {
          "data": {
            "createGroup": {
              "code": "grp_rbac_1794676834687230374_44",
              "created_at": "2026-11-15T00:20:34.687918768+07:00",
              "description": "Group E2E grp_rbac_1794676834687230374_44",
              "id": "771f2ea8-7425-4d47-b3e4-61cbf938de0f",
              "is_active": true,
              "name": "Group grp_rbac_1794676834687230374_44",
              "updated_at": "2026-11-15T00:20:34.687918768+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1797: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateUserInput!) { createUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676834688727580_45.png",
            "display_name": "Fixture u_customer_1794676834688727580_45",
            "email": "u_customer_1794676834688727580_45@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676834688727580_45.png",
              "created_at": "2026-11-15T00:20:34.689222959+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_customer_1794676834688727580_45",
              "id": "b9b53d70-62d8-4056-b6d4-456d2bd4eac9",
              "kind": "CUSTOMER",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:34.689222959+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1797: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_customer_1794676834688727580_45@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "jtNVQnOkbkFB21p3BYmgN_aKdO6cRmrFfKiR3Zvw6iY",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:1798: 
        [RBAC_TEST_AB_FORBIDDEN] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($id: ID!) { group(id: $id) { id code name description is_active created_at updated_at } }
        Input Variables:
        {
          "id": "771f2ea8-7425-4d47-b3e4-61cbf938de0f"
        }
        Result JSON:
        {
          "data": {
            "group": null
          },
          "errors": [
            {
              "message": "missing required permission: iam.groups.read",
              "extensions": {
                "code": "FORBIDDEN"
              }
            }
          ]
        }
--- PASS: TestQueryGroup (0.38s)
    --- PASS: TestQueryGroup/ValidGroupNonNull (0.10s)
    --- PASS: TestQueryGroup/RBACCustomerForbidden (0.28s)
=== RUN   TestQueryGroupMembers
=== RUN   TestQueryGroupMembers/PaginatedMembersNonNull
    iam_operations_e2e_test.go:1814: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateGroupInput!) { createGroup(input: $input) { id code name description is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "grp_members_1794676834967427871_46",
            "description": "Group E2E grp_members_1794676834967427871_46",
            "name": "Group grp_members_1794676834967427871_46"
          }
        }
        Result JSON:
        {
          "data": {
            "createGroup": {
              "code": "grp_members_1794676834967427871_46",
              "created_at": "2026-11-15T00:20:34.968129781+07:00",
              "description": "Group E2E grp_members_1794676834967427871_46",
              "id": "6bb975d0-226b-4729-8612-d19baf17a815",
              "is_active": true,
              "name": "Group grp_members_1794676834967427871_46",
              "updated_at": "2026-11-15T00:20:34.968129781+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1815: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateEmployeeUserInput!) { createEmployeeUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676834969211310_47.png",
            "display_name": "Fixture u_employee_1794676834969211310_47",
            "email": "u_employee_1794676834969211310_47@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createEmployeeUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676834969211310_47.png",
              "created_at": "2026-11-15T00:20:34.969693603+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_employee_1794676834969211310_47",
              "id": "2af3b7e1-c433-4279-be27-6b847b029842",
              "kind": "EMPLOYEE",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:34.969693603+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1816: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateEmployeeUserInput!) { createEmployeeUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676835015667566_48.png",
            "display_name": "Fixture u_employee_1794676835015667566_48",
            "email": "u_employee_1794676835015667566_48@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createEmployeeUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676835015667566_48.png",
              "created_at": "2026-11-15T00:20:35.016682008+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_employee_1794676835015667566_48",
              "id": "68385beb-8783-4460-9e82-12ad6e5ea521",
              "kind": "EMPLOYEE",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:35.016682008+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1819: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($g: ID!, $u: ID!) { addGroupMember(group_id: $g, user_id: $u) { group_id user_id created_at } }
        Input Variables:
        {
          "g": "6bb975d0-226b-4729-8612-d19baf17a815",
          "u": "2af3b7e1-c433-4279-be27-6b847b029842"
        }
        Result JSON:
        {
          "data": {
            "addGroupMember": {
              "created_at": "2026-11-15T00:20:35.061074375+07:00",
              "group_id": "6bb975d0-226b-4729-8612-d19baf17a815",
              "user_id": "2af3b7e1-c433-4279-be27-6b847b029842"
            }
          }
        }
    iam_operations_e2e_test.go:1819: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($g: ID!, $u: ID!) { addGroupMember(group_id: $g, user_id: $u) { group_id user_id created_at } }
        Input Variables:
        {
          "g": "6bb975d0-226b-4729-8612-d19baf17a815",
          "u": "68385beb-8783-4460-9e82-12ad6e5ea521"
        }
        Result JSON:
        {
          "data": {
            "addGroupMember": {
              "created_at": "2026-11-15T00:20:35.063119289+07:00",
              "group_id": "6bb975d0-226b-4729-8612-d19baf17a815",
              "user_id": "68385beb-8783-4460-9e82-12ad6e5ea521"
            }
          }
        }
    iam_operations_e2e_test.go:1825: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($group_id: ID!, $page: PageInput) { groupMembers(group_id: $group_id, page: $page) { items { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } page { total limit offset has_more } } }
        Input Variables:
        {
          "group_id": "6bb975d0-226b-4729-8612-d19baf17a815",
          "page": {
            "limit": 1,
            "offset": 0
          }
        }
        Result JSON:
        {
          "data": {
            "groupMembers": {
              "items": [
                {
                  "auth_epoch": 0,
                  "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676834969211310_47.png",
                  "created_at": "2026-11-15T00:20:34.969693+07:00",
                  "deleted_at": null,
                  "display_name": "Fixture u_employee_1794676834969211310_47",
                  "id": "2af3b7e1-c433-4279-be27-6b847b029842",
                  "kind": "EMPLOYEE",
                  "last_login_at": null,
                  "locale": "id-ID",
                  "perm_revision": 0,
                  "status": "ACTIVE",
                  "timezone": "Asia/Jakarta",
                  "updated_at": "2026-11-15T00:20:35.060564+07:00"
                }
              ],
              "page": {
                "has_more": true,
                "limit": 1,
                "offset": 0,
                "total": 2
              }
            }
          }
        }
=== RUN   TestQueryGroupMembers/RBACCustomerForbidden
    iam_operations_e2e_test.go:1840: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateGroupInput!) { createGroup(input: $input) { id code name description is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "grp_members_rbac_1794676835318668062_49",
            "description": "Group E2E grp_members_rbac_1794676835318668062_49",
            "name": "Group grp_members_rbac_1794676835318668062_49"
          }
        }
        Result JSON:
        {
          "data": {
            "createGroup": {
              "code": "grp_members_rbac_1794676835318668062_49",
              "created_at": "2026-11-15T00:20:35.319363409+07:00",
              "description": "Group E2E grp_members_rbac_1794676835318668062_49",
              "id": "ba48cdef-03d7-4339-b86e-39d4bd39d93a",
              "is_active": true,
              "name": "Group grp_members_rbac_1794676835318668062_49",
              "updated_at": "2026-11-15T00:20:35.319363409+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1841: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateUserInput!) { createUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676835320062353_50.png",
            "display_name": "Fixture u_customer_1794676835320062353_50",
            "email": "u_customer_1794676835320062353_50@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676835320062353_50.png",
              "created_at": "2026-11-15T00:20:35.320594311+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_customer_1794676835320062353_50",
              "id": "dfc0866d-567b-436c-acf7-3ee05ed329b5",
              "kind": "CUSTOMER",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:35.320594311+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1841: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_customer_1794676835320062353_50@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "PpHAJvIXh4suhwjj4390fG3fm5iyNYpkd-xFNOzto3M",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:1842: 
        [RBAC_TEST_AB_FORBIDDEN] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($group_id: ID!, $page: PageInput) { groupMembers(group_id: $group_id, page: $page) { items { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } page { total limit offset has_more } } }
        Input Variables:
        {
          "group_id": "ba48cdef-03d7-4339-b86e-39d4bd39d93a"
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "missing required permission: iam.groups.read",
              "extensions": {
                "code": "FORBIDDEN"
              }
            }
          ]
        }
--- PASS: TestQueryGroupMembers (0.62s)
    --- PASS: TestQueryGroupMembers/PaginatedMembersNonNull (0.35s)
    --- PASS: TestQueryGroupMembers/RBACCustomerForbidden (0.27s)
=== RUN   TestQueryGroupRoles
=== RUN   TestQueryGroupRoles/ScopedActiveAndExpiredGroupRoles
    iam_operations_e2e_test.go:1855: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateGroupInput!) { createGroup(input: $input) { id code name description is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "grp_gr_1794676835596198266_51",
            "description": "Group E2E grp_gr_1794676835596198266_51",
            "name": "Group grp_gr_1794676835596198266_51"
          }
        }
        Result JSON:
        {
          "data": {
            "createGroup": {
              "code": "grp_gr_1794676835596198266_51",
              "created_at": "2026-11-15T00:20:35.597013611+07:00",
              "description": "Group E2E grp_gr_1794676835596198266_51",
              "id": "79b66541-6c1a-4bce-a803-3adb8a879a99",
              "is_active": true,
              "name": "Group grp_gr_1794676835596198266_51",
              "updated_at": "2026-11-15T00:20:35.597013611+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1856: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateRoleInput!) { createRole(input: $input) { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "role_gr_1794676835597738644_52",
            "default_redirect_path": "/identity",
            "description": "Custom E2E role role_gr_1794676835597738644_52",
            "is_system": false,
            "management_domain": "iam",
            "name": "Role role_gr_1794676835597738644_52",
            "redirect_priority": 50
          }
        }
        Result JSON:
        {
          "data": {
            "createRole": {
              "code": "role_gr_1794676835597738644_52",
              "created_at": "2026-11-15T00:20:35.598309636+07:00",
              "default_redirect_path": "/identity",
              "description": "Custom E2E role role_gr_1794676835597738644_52",
              "id": "778386f0-f6eb-4ecb-98aa-0b7d0132ec67",
              "is_active": true,
              "is_system": false,
              "management_domain": "iam",
              "name": "Role role_gr_1794676835597738644_52",
              "redirect_priority": 50,
              "updated_at": "2026-11-15T00:20:35.598309636+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1860: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: AssignGroupRoleInput!) { assignGroupRole(input: $input) { id group_id role_id scope { type id } granted_by expires_at created_at } }
        Input Variables:
        {
          "input": {
            "expires_at": "2026-11-14T19:20:35Z",
            "group_id": "79b66541-6c1a-4bce-a803-3adb8a879a99",
            "role_id": "778386f0-f6eb-4ecb-98aa-0b7d0132ec67",
            "scope_id": "22222222-2222-4222-8222-222222222222",
            "scope_type": "FACILITY"
          }
        }
        Result JSON:
        {
          "data": {
            "assignGroupRole": {
              "created_at": "2026-11-15T00:20:35.600355402+07:00",
              "expires_at": "2026-11-15T02:20:35+07:00",
              "granted_by": "853aac77-32e7-4a88-b2d6-7d68db8d13d7",
              "group_id": "79b66541-6c1a-4bce-a803-3adb8a879a99",
              "id": "619e3eef-2fef-44fc-a225-34cff7a1d346",
              "role_id": "778386f0-f6eb-4ecb-98aa-0b7d0132ec67",
              "scope": {
                "id": "22222222-2222-4222-8222-222222222222",
                "type": "FACILITY"
              }
            }
          }
        }
    iam_operations_e2e_test.go:1875: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($group_id: ID!, $include_expired: Boolean) { groupRoles(group_id: $group_id, include_expired: $include_expired) { id group_id role_id scope { type id } granted_by expires_at created_at } }
        Input Variables:
        {
          "group_id": "79b66541-6c1a-4bce-a803-3adb8a879a99",
          "include_expired": false
        }
        Result JSON:
        {
          "data": {
            "groupRoles": [
              {
                "created_at": "2026-11-15T00:20:35.600355+07:00",
                "expires_at": "2026-11-15T02:20:35+07:00",
                "granted_by": "853aac77-32e7-4a88-b2d6-7d68db8d13d7",
                "group_id": "79b66541-6c1a-4bce-a803-3adb8a879a99",
                "id": "619e3eef-2fef-44fc-a225-34cff7a1d346",
                "role_id": "778386f0-f6eb-4ecb-98aa-0b7d0132ec67",
                "scope": {
                  "id": "22222222-2222-4222-8222-222222222222",
                  "type": "FACILITY"
                }
              }
            ]
          }
        }
    iam_operations_e2e_test.go:1888: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($group_id: ID!, $include_expired: Boolean) { groupRoles(group_id: $group_id, include_expired: $include_expired) { id group_id role_id scope { type id } granted_by expires_at created_at } }
        Input Variables:
        {
          "group_id": "79b66541-6c1a-4bce-a803-3adb8a879a99",
          "include_expired": false
        }
        Result JSON:
        {
          "data": {
            "groupRoles": []
          }
        }
    iam_operations_e2e_test.go:1891: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($group_id: ID!, $include_expired: Boolean) { groupRoles(group_id: $group_id, include_expired: $include_expired) { id group_id role_id scope { type id } granted_by expires_at created_at } }
        Input Variables:
        {
          "group_id": "79b66541-6c1a-4bce-a803-3adb8a879a99",
          "include_expired": true
        }
        Result JSON:
        {
          "data": {
            "groupRoles": [
              {
                "created_at": "2026-11-15T00:20:35.600355+07:00",
                "expires_at": "2026-11-14T23:20:35.642704+07:00",
                "granted_by": "853aac77-32e7-4a88-b2d6-7d68db8d13d7",
                "group_id": "79b66541-6c1a-4bce-a803-3adb8a879a99",
                "id": "619e3eef-2fef-44fc-a225-34cff7a1d346",
                "role_id": "778386f0-f6eb-4ecb-98aa-0b7d0132ec67",
                "scope": {
                  "id": "22222222-2222-4222-8222-222222222222",
                  "type": "FACILITY"
                }
              }
            ]
          }
        }
=== RUN   TestQueryGroupRoles/RBACCustomerForbidden
    iam_operations_e2e_test.go:1900: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateGroupInput!) { createGroup(input: $input) { id code name description is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "grp_gr_rbac_1794676835856743845_53",
            "description": "Group E2E grp_gr_rbac_1794676835856743845_53",
            "name": "Group grp_gr_rbac_1794676835856743845_53"
          }
        }
        Result JSON:
        {
          "data": {
            "createGroup": {
              "code": "grp_gr_rbac_1794676835856743845_53",
              "created_at": "2026-11-15T00:20:35.857581602+07:00",
              "description": "Group E2E grp_gr_rbac_1794676835856743845_53",
              "id": "00d2b063-a139-4a31-907b-add4f03a5287",
              "is_active": true,
              "name": "Group grp_gr_rbac_1794676835856743845_53",
              "updated_at": "2026-11-15T00:20:35.857581602+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1901: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateUserInput!) { createUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676835858412877_54.png",
            "display_name": "Fixture u_customer_1794676835858412877_54",
            "email": "u_customer_1794676835858412877_54@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676835858412877_54.png",
              "created_at": "2026-11-15T00:20:35.858994699+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_customer_1794676835858412877_54",
              "id": "978dd83b-01f5-44b6-938d-0e51fa25d705",
              "kind": "CUSTOMER",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:35.858994699+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1901: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_customer_1794676835858412877_54@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "3XuT1Od0-r7UeHMMIUJjxNux22pFFO__YkjFlDP1-oQ",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:1902: 
        [RBAC_TEST_AB_FORBIDDEN] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($group_id: ID!, $include_expired: Boolean) { groupRoles(group_id: $group_id, include_expired: $include_expired) { id group_id role_id scope { type id } granted_by expires_at created_at } }
        Input Variables:
        {
          "group_id": "00d2b063-a139-4a31-907b-add4f03a5287",
          "include_expired": false
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "missing required permission: iam.groups.read",
              "extensions": {
                "code": "FORBIDDEN"
              }
            }
          ]
        }
--- PASS: TestQueryGroupRoles (0.53s)
    --- PASS: TestQueryGroupRoles/ScopedActiveAndExpiredGroupRoles (0.27s)
    --- PASS: TestQueryGroupRoles/RBACCustomerForbidden (0.27s)
=== RUN   TestMutationUserLogin
=== RUN   TestMutationUserLogin/EmailAndUsernameLoginSuccess
    iam_operations_e2e_test.go:1919: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateEmployeeUserInput!) { createEmployeeUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676836126418107_55.png",
            "display_name": "Fixture u_employee_1794676836126418107_55",
            "email": "u_employee_1794676836126418107_55@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta",
            "username": "u_employee_1794676836126418107_55"
          }
        }
        Result JSON:
        {
          "data": {
            "createEmployeeUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676836126418107_55.png",
              "created_at": "2026-11-15T00:20:36.127181633+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_employee_1794676836126418107_55",
              "id": "11466854-f985-46eb-8e86-59f1d9a61a38",
              "kind": "EMPLOYEE",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:36.127181633+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1922: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_employee_1794676836126418107_55@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "HPMQdV-wiSVeltp-UQomo32CrNjVH-d01juYxusPX4Y",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/account",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:1928: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_employee_1794676836126418107_55",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "NaqSy1FnGvzHX4WPlKZ7Dpmz0e5DfdH82Yc-RdLpvVc",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/account",
              "status": "AUTHENTICATED"
            }
          }
        }
=== RUN   TestMutationUserLogin/MfaRequiredAndInvalidCredentials
    iam_operations_e2e_test.go:1936: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateEmployeeUserInput!) { createEmployeeUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676836406342787_56.png",
            "display_name": "Fixture u_employee_1794676836406342787_56",
            "email": "u_employee_1794676836406342787_56@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createEmployeeUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676836406342787_56.png",
              "created_at": "2026-11-15T00:20:36.406926974+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_employee_1794676836406342787_56",
              "id": "b0f8944c-4b91-484e-b1ac-cfba08b1a4bd",
              "kind": "EMPLOYEE",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:36.406926974+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1937: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_employee_1794676836406342787_56@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "IVeRf4CHEAOyy3fIgH3DVEcG5ZeAcaxccKmFe-ND5DA",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/account",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:1938: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation { enrollMyTotp { factor_id otp_auth_uri } }
        Input Variables:
        {}
        Result JSON:
        {
          "data": {
            "enrollMyTotp": {
              "factor_id": "df949e1a-0cbc-472b-967e-e01ee40bbfd8",
              "otp_auth_uri": "otpauth://totp/IntegriOne%20ERP:Fixture%20u_employee_1794676836406342787_56?algorithm=SHA1\u0026digits=6\u0026issuer=IntegriOne%20ERP\u0026period=30\u0026secret=KHKZLHY6RVLO3XCQQKTNGHTTWGGJSQJP"
            }
          }
        }
    iam_operations_e2e_test.go:1938: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: ConfirmTotpInput!) { confirmMyTotp(input: $input) { enabled recovery_codes } }
        Input Variables:
        {
          "input": {
            "code": "703382"
          }
        }
        Result JSON:
        {
          "data": {
            "confirmMyTotp": {
              "enabled": true,
              "recovery_codes": [
                "R5FZ-RH2Y-XZ23",
                "TTJF-B4PC-PRNN",
                "QF2H-8SMY-PQVW",
                "KDJS-ASZT-WMDW",
                "34RW-AHXK-6V6X",
                "C4MB-BZ2V-ZB6R",
                "DPR2-BBZL-MBSV",
                "HSHH-SM9D-MRD7"
              ]
            }
          }
        }
    iam_operations_e2e_test.go:1941: 
        [GQL_EXPECTED_ERROR] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_employee_1794676836406342787_56@integrione.test",
            "password": "WrongPassword#1"
          }
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "invalid credentials",
              "extensions": {
                "code": "INVALID_CREDENTIALS"
              }
            }
          ]
        }
    iam_operations_e2e_test.go:1944: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_employee_1794676836406342787_56@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": null,
              "mfaChallenge": "sSkAKrCCb30MG-6C5jZWPq1hVCMeXtdHpeIuQi-cyys",
              "mfaExpiresAt": "2026-11-15T00:25:36+07:00",
              "redirectLocation": null,
              "status": "MFA_REQUIRED"
            }
          }
        }
--- PASS: TestMutationUserLogin (0.62s)
    --- PASS: TestMutationUserLogin/EmailAndUsernameLoginSuccess (0.28s)
    --- PASS: TestMutationUserLogin/MfaRequiredAndInvalidCredentials (0.34s)
=== RUN   TestMutationValidateUserMfa
=== RUN   TestMutationValidateUserMfa/TotpAndRecoveryCodeWithReplayProtection
    iam_operations_e2e_test.go:1962: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateEmployeeUserInput!) { createEmployeeUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676836744341523_57.png",
            "display_name": "Fixture u_employee_1794676836744341523_57",
            "email": "u_employee_1794676836744341523_57@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createEmployeeUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676836744341523_57.png",
              "created_at": "2026-11-15T00:20:36.745135888+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_employee_1794676836744341523_57",
              "id": "6d495ae4-b99a-41f5-8ef3-9e3e5b801c07",
              "kind": "EMPLOYEE",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:36.745135888+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:1963: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_employee_1794676836744341523_57@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "37NbfU_Dc1lmb49fCBpO1TgL563fA-vl7Q5Q0O6tpg0",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/account",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:1964: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation { enrollMyTotp { factor_id otp_auth_uri } }
        Input Variables:
        {}
        Result JSON:
        {
          "data": {
            "enrollMyTotp": {
              "factor_id": "a377a089-15a6-48ba-9f9d-76d6b27cb848",
              "otp_auth_uri": "otpauth://totp/IntegriOne%20ERP:Fixture%20u_employee_1794676836744341523_57?algorithm=SHA1\u0026digits=6\u0026issuer=IntegriOne%20ERP\u0026period=30\u0026secret=D3IOY327B3GGRZV4EPRLY2DTXJ6Z5BLE"
            }
          }
        }
    iam_operations_e2e_test.go:1964: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: ConfirmTotpInput!) { confirmMyTotp(input: $input) { enabled recovery_codes } }
        Input Variables:
        {
          "input": {
            "code": "341446"
          }
        }
        Result JSON:
        {
          "data": {
            "confirmMyTotp": {
              "enabled": true,
              "recovery_codes": [
                "VML9-N2MD-FT2K",
                "J6TK-9B4K-SNTM",
                "FBGG-HAP6-L4UN",
                "3ZPB-UVW9-9VQR",
                "2A6V-G7MC-L8CD",
                "BT4V-23GD-8TJT",
                "B4VQ-BHXL-3Q89",
                "UYY9-NNRZ-Z6LH"
              ]
            }
          }
        }
    iam_operations_e2e_test.go:1968: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_employee_1794676836744341523_57@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": null,
              "mfaChallenge": "VEDlFHyDxCv_x-av8VZ4blIiH6q-iOSfe7L_HvEpwF4",
              "mfaExpiresAt": "2026-11-15T00:25:36+07:00",
              "redirectLocation": null,
              "status": "MFA_REQUIRED"
            }
          }
        }
    iam_operations_e2e_test.go:1971: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: ValidateUserMfaInput!) { validateUserMfa(input: $input) { csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "challenge": "VEDlFHyDxCv_x-av8VZ4blIiH6q-iOSfe7L_HvEpwF4",
            "totpCode": "687783"
          }
        }
        Result JSON:
        {
          "data": {
            "validateUserMfa": {
              "csrfToken": "xLEnMdncIuxwBq7VQ89ns9iu5KvAutjTaGbfKyGJn6I",
              "redirectLocation": "/account"
            }
          }
        }
    iam_operations_e2e_test.go:1981: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_employee_1794676836744341523_57@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": null,
              "mfaChallenge": "LXzFUji6F-P-F0l0D87D-RgCuZzie9WZSQFj1G5E2Cg",
              "mfaExpiresAt": "2026-11-15T00:25:36+07:00",
              "redirectLocation": null,
              "status": "MFA_REQUIRED"
            }
          }
        }
    iam_operations_e2e_test.go:1984: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: ValidateUserMfaInput!) { validateUserMfa(input: $input) { csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "challenge": "LXzFUji6F-P-F0l0D87D-RgCuZzie9WZSQFj1G5E2Cg",
            "recoveryCode": "VML9-N2MD-FT2K"
          }
        }
        Result JSON:
        {
          "data": {
            "validateUserMfa": {
              "csrfToken": "wyZiW_wMJ6qCTee9KbLBNKC_tJZ8Yi4T4i4tmFbWK-0",
              "redirectLocation": "/account"
            }
          }
        }
    iam_operations_e2e_test.go:1994: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_employee_1794676836744341523_57@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": null,
              "mfaChallenge": "R0wKCWXXfdVVlWM_AoCrUjg7VZB0oBBsWip42XsB12o",
              "mfaExpiresAt": "2026-11-15T00:25:36+07:00",
              "redirectLocation": null,
              "status": "MFA_REQUIRED"
            }
          }
        }
    iam_operations_e2e_test.go:1997: 
        [GQL_EXPECTED_ERROR] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: ValidateUserMfaInput!) { validateUserMfa(input: $input) { csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "challenge": "R0wKCWXXfdVVlWM_AoCrUjg7VZB0oBBsWip42XsB12o",
            "recoveryCode": "VML9-N2MD-FT2K"
          }
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "invalid recovery code",
              "extensions": {
                "code": "INVALID_CREDENTIALS"
              }
            }
          ]
        }
--- PASS: TestMutationValidateUserMfa (0.40s)
    --- PASS: TestMutationValidateUserMfa/TotpAndRecoveryCodeWithReplayProtection (0.40s)
=== RUN   TestMutationRegisterCustomer
=== RUN   TestMutationRegisterCustomer/RegisterSuccessAndDuplicateConflict
    iam_operations_e2e_test.go:2022: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: RegisterCustomerInput!) { registerCustomer(input: $input) { verificationPendingEmail requiresEmailVerification } }
        Input Variables:
        {
          "input": {
            "displayName": "Registered reg_cust_1794676837094266733_58",
            "email": "reg_cust_1794676837094266733_58@integrione.test",
            "locale": "id-ID",
            "password": "Zx9#mQp2!vLk8@Rt",
            "timezone": "Asia/Jakarta",
            "username": "reg_cust_1794676837094266733_58"
          }
        }
        Result JSON:
        {
          "data": {
            "registerCustomer": {
              "requiresEmailVerification": true,
              "verificationPendingEmail": "reg_cust_1794676837094266733_58@integrione.test"
            }
          }
        }
    iam_operations_e2e_test.go:2038: 
        [GQL_EXPECTED_ERROR] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: RegisterCustomerInput!) { registerCustomer(input: $input) { verificationPendingEmail requiresEmailVerification } }
        Input Variables:
        {
          "input": {
            "displayName": "Duplicate reg_cust_1794676837094266733_58",
            "email": "reg_cust_1794676837094266733_58@integrione.test",
            "password": "Zx9#mQp2!vLk8@Rt"
          }
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "email is already registered",
              "extensions": {
                "code": "CONFLICT"
              }
            }
          ]
        }
--- PASS: TestMutationRegisterCustomer (0.22s)
    --- PASS: TestMutationRegisterCustomer/RegisterSuccessAndDuplicateConflict (0.22s)
=== RUN   TestMutationVerifyCustomerEmail
=== RUN   TestMutationVerifyCustomerEmail/VerifiedAlreadyUsedExpiredAndInvalid
    iam_operations_e2e_test.go:2060: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: RegisterCustomerInput!) { registerCustomer(input: $input) { verificationPendingEmail requiresEmailVerification } }
        Input Variables:
        {
          "input": {
            "displayName": "Verify One",
            "email": "ver1_1794676837310832466_59@integrione.test",
            "password": "Zx9#mQp2!vLk8@Rt"
          }
        }
        Result JSON:
        {
          "data": {
            "registerCustomer": {
              "requiresEmailVerification": true,
              "verificationPendingEmail": "ver1_1794676837310832466_59@integrione.test"
            }
          }
        }
    iam_operations_e2e_test.go:2074: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: VerifyCustomerEmailInput!) { verifyCustomerEmail(input: $input) { status verified } }
        Input Variables:
        {
          "input": {
            "token": "e2e-verify-token-ver1_1794676837310832466_59"
          }
        }
        Result JSON:
        {
          "data": {
            "verifyCustomerEmail": {
              "status": "VERIFIED",
              "verified": true
            }
          }
        }
    iam_operations_e2e_test.go:2077: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "ver1_1794676837310832466_59@integrione.test",
            "password": "Zx9#mQp2!vLk8@Rt"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "In7g5vkG6jS4ob3UFslwsh7DD8RvWsrJDvsjwC7ZDWU",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:2081: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: VerifyCustomerEmailInput!) { verifyCustomerEmail(input: $input) { status verified } }
        Input Variables:
        {
          "input": {
            "token": "e2e-verify-token-ver1_1794676837310832466_59"
          }
        }
        Result JSON:
        {
          "data": {
            "verifyCustomerEmail": {
              "status": "ALREADY_USED",
              "verified": false
            }
          }
        }
    iam_operations_e2e_test.go:2087: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: RegisterCustomerInput!) { registerCustomer(input: $input) { verificationPendingEmail requiresEmailVerification } }
        Input Variables:
        {
          "input": {
            "displayName": "Verify Expired",
            "email": "ver2_1794676837490515228_60@integrione.test",
            "password": "Zx9#mQp2!vLk8@Rt"
          }
        }
        Result JSON:
        {
          "data": {
            "registerCustomer": {
              "requiresEmailVerification": true,
              "verificationPendingEmail": "ver2_1794676837490515228_60@integrione.test"
            }
          }
        }
    iam_operations_e2e_test.go:2099: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: VerifyCustomerEmailInput!) { verifyCustomerEmail(input: $input) { status verified } }
        Input Variables:
        {
          "input": {
            "token": "e2e-expired-token-ver2_1794676837490515228_60"
          }
        }
        Result JSON:
        {
          "data": {
            "verifyCustomerEmail": {
              "status": "EXPIRED",
              "verified": false
            }
          }
        }
    iam_operations_e2e_test.go:2103: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: VerifyCustomerEmailInput!) { verifyCustomerEmail(input: $input) { status verified } }
        Input Variables:
        {
          "input": {
            "token": "invalid-nonexistent-token-xyz"
          }
        }
        Result JSON:
        {
          "data": {
            "verifyCustomerEmail": {
              "status": "INVALID",
              "verified": false
            }
          }
        }
--- PASS: TestMutationVerifyCustomerEmail (0.48s)
    --- PASS: TestMutationVerifyCustomerEmail/VerifiedAlreadyUsedExpiredAndInvalid (0.48s)
=== RUN   TestMutationRequestPasswordReset
=== RUN   TestMutationRequestPasswordReset/ExistingAndNonExistentReturnTrue
    iam_operations_e2e_test.go:2115: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateEmployeeUserInput!) { createEmployeeUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676837837174290_61.png",
            "display_name": "Fixture u_employee_1794676837837174290_61",
            "email": "u_employee_1794676837837174290_61@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createEmployeeUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676837837174290_61.png",
              "created_at": "2026-11-15T00:20:37.837883354+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_employee_1794676837837174290_61",
              "id": "f2f1c587-4c36-4a3b-9dee-c58695d5cc26",
              "kind": "EMPLOYEE",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:37.837883354+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2118: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: IAMRequestPasswordResetInput!) { requestPasswordReset(input: $input) }
        Input Variables:
        {
          "input": {
            "identifier": "u_employee_1794676837837174290_61@integrione.test"
          }
        }
        Result JSON:
        {
          "data": {
            "requestPasswordReset": true
          }
        }
    iam_operations_e2e_test.go:2124: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: IAMRequestPasswordResetInput!) { requestPasswordReset(input: $input) }
        Input Variables:
        {
          "input": {
            "identifier": "ghost-none_1794676837926060637_62@integrione.test"
          }
        }
        Result JSON:
        {
          "data": {
            "requestPasswordReset": true
          }
        }
--- PASS: TestMutationRequestPasswordReset (0.23s)
    --- PASS: TestMutationRequestPasswordReset/ExistingAndNonExistentReturnTrue (0.23s)
=== RUN   TestMutationResetPassword
=== RUN   TestMutationResetPassword/ValidTokenUpdatesPasswordAndInvalidRejected
    iam_operations_e2e_test.go:2136: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateEmployeeUserInput!) { createEmployeeUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676838071279480_63.png",
            "display_name": "Fixture u_employee_1794676838071279480_63",
            "email": "u_employee_1794676838071279480_63@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createEmployeeUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676838071279480_63.png",
              "created_at": "2026-11-15T00:20:38.071962704+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_employee_1794676838071279480_63",
              "id": "95b374e0-5334-4531-9faa-46fefb301657",
              "kind": "EMPLOYEE",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:38.071962704+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2139: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: IAMRequestPasswordResetInput!) { requestPasswordReset(input: $input) }
        Input Variables:
        {
          "input": {
            "identifier": "u_employee_1794676838071279480_63@integrione.test"
          }
        }
        Result JSON:
        {
          "data": {
            "requestPasswordReset": true
          }
        }
    iam_operations_e2e_test.go:2147: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: IAMResetPasswordInput!) { resetPassword(input: $input) }
        Input Variables:
        {
          "input": {
            "newPassword": "NewResetPass#2026!",
            "token": "e2e-reset-token-rst_1794676838121301823_64"
          }
        }
        Result JSON:
        {
          "data": {
            "resetPassword": true
          }
        }
    iam_operations_e2e_test.go:2151: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_employee_1794676838071279480_63@integrione.test",
            "password": "NewResetPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "u7uEhPu3axMns1IpajeXi1MK6hIihnHLKewuJ23j4HE",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/account",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:2154: 
        [GQL_EXPECTED_ERROR] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: IAMResetPasswordInput!) { resetPassword(input: $input) }
        Input Variables:
        {
          "input": {
            "newPassword": "AnotherPass#2026!",
            "token": "invalid-reset-token-xyz"
          }
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "token not found",
              "extensions": {
                "code": "NOT_FOUND"
              }
            }
          ]
        }
--- PASS: TestMutationResetPassword (0.33s)
    --- PASS: TestMutationResetPassword/ValidTokenUpdatesPasswordAndInvalidRejected (0.33s)
=== RUN   TestMutationUserLogout
=== RUN   TestMutationUserLogout/AuthenticatedLogoutRevokesSession
    iam_operations_e2e_test.go:2167: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateEmployeeUserInput!) { createEmployeeUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676838399070809_65.png",
            "display_name": "Fixture u_employee_1794676838399070809_65",
            "email": "u_employee_1794676838399070809_65@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createEmployeeUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676838399070809_65.png",
              "created_at": "2026-11-15T00:20:38.399834787+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_employee_1794676838399070809_65",
              "id": "43d453c6-22df-4e82-aa00-05c80030cea2",
              "kind": "EMPLOYEE",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:38.399834787+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2168: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_employee_1794676838399070809_65@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "r6kLDn3qp7Fi-b3ov0oqD1l14wqnev2MYFAkDcDOpSk",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/account",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:2170: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation { userLogout }
        Input Variables:
        {}
        Result JSON:
        {
          "data": {
            "userLogout": true
          }
        }
    iam_operations_e2e_test.go:2171: 
        [GQL_EXPECTED_ERROR] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query { myProfile { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {}
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "session cookie is required",
              "extensions": {
                "code": "UNAUTHENTICATED"
              }
            }
          ]
        }
=== RUN   TestMutationUserLogout/AnonymousRejected
    iam_operations_e2e_test.go:2176: 
        [UNAUTHENTICATED_GUARD] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation { userLogout }
        Input Variables:
        {}
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "session is missing, expired, or revoked",
              "extensions": {
                "code": "UNAUTHENTICATED"
              }
            }
          ]
        }
--- PASS: TestMutationUserLogout (0.24s)
    --- PASS: TestMutationUserLogout/AuthenticatedLogoutRevokesSession (0.23s)
    --- PASS: TestMutationUserLogout/AnonymousRejected (0.00s)
=== RUN   TestMutationUpdateMyProfile
=== RUN   TestMutationUpdateMyProfile/UpdateAllProfileFields
    iam_operations_e2e_test.go:2187: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateEmployeeUserInput!) { createEmployeeUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676838634778014_66.png",
            "display_name": "Fixture u_employee_1794676838634778014_66",
            "email": "u_employee_1794676838634778014_66@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createEmployeeUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676838634778014_66.png",
              "created_at": "2026-11-15T00:20:38.635858872+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_employee_1794676838634778014_66",
              "id": "abeb313e-06ed-4d44-aa0a-487c122300c0",
              "kind": "EMPLOYEE",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:38.635858872+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2188: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_employee_1794676838634778014_66@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "GJnZuXddgvLDlhR5Flejsx3Qb4tjuUxAC4Xpu-gp8vU",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/account",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:2191: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UpdateMyProfileInput!) { updateMyProfile(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/new-avatar.png",
            "display_name": "Updated Name prof_1794676838729255272_67",
            "locale": "en-US",
            "timezone": "UTC"
          }
        }
        Result JSON:
        {
          "data": {
            "updateMyProfile": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/new-avatar.png",
              "created_at": "2026-11-15T00:20:38.635858+07:00",
              "deleted_at": null,
              "display_name": "Updated Name prof_1794676838729255272_67",
              "id": "abeb313e-06ed-4d44-aa0a-487c122300c0",
              "kind": "EMPLOYEE",
              "last_login_at": "2026-11-15T00:20:38.683059+07:00",
              "locale": "en-US",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "UTC",
              "updated_at": "2026-11-15T00:20:38.730561747+07:00"
            }
          }
        }
=== RUN   TestMutationUpdateMyProfile/AnonymousRejected
    iam_operations_e2e_test.go:2209: 
        [UNAUTHENTICATED_GUARD] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UpdateMyProfileInput!) { updateMyProfile(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "display_name": "Anon"
          }
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "session is missing, expired, or revoked",
              "extensions": {
                "code": "UNAUTHENTICATED"
              }
            }
          ]
        }
--- PASS: TestMutationUpdateMyProfile (0.23s)
    --- PASS: TestMutationUpdateMyProfile/UpdateAllProfileFields (0.23s)
    --- PASS: TestMutationUpdateMyProfile/AnonymousRejected (0.00s)
=== RUN   TestMutationChangeMyPassword
=== RUN   TestMutationChangeMyPassword/FreshSuccessWrongCurrentAndStaleDenied
    iam_operations_e2e_test.go:2222: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateEmployeeUserInput!) { createEmployeeUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676838869251892_68.png",
            "display_name": "Fixture u_employee_1794676838869251892_68",
            "email": "u_employee_1794676838869251892_68@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createEmployeeUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676838869251892_68.png",
              "created_at": "2026-11-15T00:20:38.87027462+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_employee_1794676838869251892_68",
              "id": "833ab222-57c2-4ab1-90d1-33004bba928b",
              "kind": "EMPLOYEE",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:38.87027462+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2223: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_employee_1794676838869251892_68@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "9cRgQftujZSpwymTOUIhPOUL-bBiSm86EEBW2c_etkQ",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/account",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:2225: 
        [GQL_EXPECTED_ERROR] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: ChangeMyPasswordInput!) { changeMyPassword(input: $input) }
        Input Variables:
        {
          "input": {
            "current_password": "WrongCurrent#1",
            "new_password": "NextPass#2026!"
          }
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "invalid credentials",
              "extensions": {
                "code": "INVALID_CREDENTIALS"
              }
            }
          ]
        }
    iam_operations_e2e_test.go:2230: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: ChangeMyPasswordInput!) { changeMyPassword(input: $input) }
        Input Variables:
        {
          "input": {
            "current_password": "StrongPass#2026!",
            "new_password": "ChangedPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "changeMyPassword": true
          }
        }
    iam_operations_e2e_test.go:2235: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_employee_1794676838869251892_68@integrione.test",
            "password": "ChangedPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "IbcuSHnhvNRZOeNNpP5ujAQG06Hm484NC9w-eaRrTV0",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/account",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:2237: 
        [GQL_EXPECTED_ERROR] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: ChangeMyPasswordInput!) { changeMyPassword(input: $input) }
        Input Variables:
        {
          "input": {
            "current_password": "ChangedPass#2026!",
            "new_password": "ThirdPass#2026!"
          }
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "recent authentication within 10 minutes is required for sensitive action",
              "extensions": {
                "code": "FORBIDDEN"
              }
            }
          ]
        }
--- PASS: TestMutationChangeMyPassword (0.48s)
    --- PASS: TestMutationChangeMyPassword/FreshSuccessWrongCurrentAndStaleDenied (0.48s)
=== RUN   TestMutationVerifyMySessionMfa
=== RUN   TestMutationVerifyMySessionMfa/ValidTotpRotatesCsrfAndRejectsInvalidCode
    iam_operations_e2e_test.go:2250: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateEmployeeUserInput!) { createEmployeeUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676839355954283_69.png",
            "display_name": "Fixture u_employee_1794676839355954283_69",
            "email": "u_employee_1794676839355954283_69@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createEmployeeUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676839355954283_69.png",
              "created_at": "2026-11-15T00:20:39.356641816+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_employee_1794676839355954283_69",
              "id": "a3b88ff0-c0c8-49b3-9ea4-f4efe83bd1e9",
              "kind": "EMPLOYEE",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:39.356641816+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2251: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_employee_1794676839355954283_69@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "TEGXWPNkDqh_HtZ-ulYAI6kgU-YFm6RuI1AlkazhWQM",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/account",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:2252: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation { enrollMyTotp { factor_id otp_auth_uri } }
        Input Variables:
        {}
        Result JSON:
        {
          "data": {
            "enrollMyTotp": {
              "factor_id": "1e782278-1f02-4db6-ae37-c8963f554625",
              "otp_auth_uri": "otpauth://totp/IntegriOne%20ERP:Fixture%20u_employee_1794676839355954283_69?algorithm=SHA1\u0026digits=6\u0026issuer=IntegriOne%20ERP\u0026period=30\u0026secret=7YV5UVIYOS2PDZ3CZ4FXNSLL3P5WJVEJ"
            }
          }
        }
    iam_operations_e2e_test.go:2252: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: ConfirmTotpInput!) { confirmMyTotp(input: $input) { enabled recovery_codes } }
        Input Variables:
        {
          "input": {
            "code": "979907"
          }
        }
        Result JSON:
        {
          "data": {
            "confirmMyTotp": {
              "enabled": true,
              "recovery_codes": [
                "AXV6-QZXJ-HTX7",
                "EFQA-NPDL-PGKX",
                "TXGE-PKGR-GWPJ",
                "HJ9C-LFV8-MSE5",
                "8JVD-3WQE-SGRB",
                "WG33-S3QX-MJGP",
                "4ZK3-CQNZ-J339",
                "Q2U8-96P5-N9A4"
              ]
            }
          }
        }
    iam_operations_e2e_test.go:2255: 
        [GQL_EXPECTED_ERROR] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($code: String!) { verifyMySessionMfa(totp_code: $code) }
        Input Variables:
        {
          "code": "000000"
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "invalid totp verification code",
              "extensions": {
                "code": "INVALID_CREDENTIALS"
              }
            }
          ]
        }
    iam_operations_e2e_test.go:2257: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($code: String!) { verifyMySessionMfa(totp_code: $code) }
        Input Variables:
        {
          "code": "216748"
        }
        Result JSON:
        {
          "data": {
            "verifyMySessionMfa": true
          }
        }
    iam_operations_e2e_test.go:2263: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query { mySessions(active_only: true) { items { id user_id mfa_verified mfa_verified_at ip_address user_agent authenticated_at last_active_at idle_expires_at absolute_expires_at ended_at end_reason created_at generation } page { total limit offset has_more } } }
        Input Variables:
        {}
        Result JSON:
        {
          "data": {
            "mySessions": {
              "items": [
                {
                  "absolute_expires_at": "2026-11-15T08:20:39.448286+07:00",
                  "authenticated_at": "2026-11-15T00:20:39.448286+07:00",
                  "created_at": "2026-11-15T00:20:39.448286+07:00",
                  "end_reason": null,
                  "ended_at": null,
                  "generation": 2,
                  "id": "697dd0b4-874f-44ea-a335-28eb6f93e053",
                  "idle_expires_at": "2026-11-15T00:50:39.448286+07:00",
                  "ip_address": "172.22.0.1",
                  "last_active_at": "2026-11-15T00:20:39.448286+07:00",
                  "mfa_verified": true,
                  "mfa_verified_at": "2026-11-15T00:20:39.457439+07:00",
                  "user_agent": "Go-http-client/1.1",
                  "user_id": "a3b88ff0-c0c8-49b3-9ea4-f4efe83bd1e9"
                }
              ],
              "page": {
                "has_more": false,
                "limit": 20,
                "offset": 0,
                "total": 1
              }
            }
          }
        }
--- PASS: TestMutationVerifyMySessionMfa (0.24s)
    --- PASS: TestMutationVerifyMySessionMfa/ValidTotpRotatesCsrfAndRejectsInvalidCode (0.24s)
=== RUN   TestMutationEnrollMyTotp
=== RUN   TestMutationEnrollMyTotp/FreshSessionSuccessAndStaleForbidden
    iam_operations_e2e_test.go:2283: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateEmployeeUserInput!) { createEmployeeUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676839595036692_70.png",
            "display_name": "Fixture u_employee_1794676839595036692_70",
            "email": "u_employee_1794676839595036692_70@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createEmployeeUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676839595036692_70.png",
              "created_at": "2026-11-15T00:20:39.595710969+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_employee_1794676839595036692_70",
              "id": "7da6c8b1-698c-465b-82fb-19acfab8ed54",
              "kind": "EMPLOYEE",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:39.595710969+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2284: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_employee_1794676839595036692_70@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "8uwE6Q1EoslXf2XbwBweJNlotT-HRBCW_nzsCklpooM",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/account",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:2286: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation { enrollMyTotp { factor_id otp_auth_uri } }
        Input Variables:
        {}
        Result JSON:
        {
          "data": {
            "enrollMyTotp": {
              "factor_id": "af46bcff-1bde-4387-962c-e8f5de466838",
              "otp_auth_uri": "otpauth://totp/IntegriOne%20ERP:Fixture%20u_employee_1794676839595036692_70?algorithm=SHA1\u0026digits=6\u0026issuer=IntegriOne%20ERP\u0026period=30\u0026secret=55PBKVBNU2HYT3XH2BLGN5ZMSF3E2ZXV"
            }
          }
        }
    iam_operations_e2e_test.go:2293: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateEmployeeUserInput!) { createEmployeeUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676839701030096_71.png",
            "display_name": "Fixture u_employee_1794676839701030096_71",
            "email": "u_employee_1794676839701030096_71@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createEmployeeUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676839701030096_71.png",
              "created_at": "2026-11-15T00:20:39.701656232+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_employee_1794676839701030096_71",
              "id": "fca5c046-1ee1-4ec2-9cac-75ff472ab1b6",
              "kind": "EMPLOYEE",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:39.701656232+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2294: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_employee_1794676839701030096_71@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "aWkvue_zZyxirOYzncryYMJLV8p4tc_RSmJJ7OCIVdE",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/account",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:2296: 
        [GQL_EXPECTED_ERROR] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation { enrollMyTotp { factor_id otp_auth_uri } }
        Input Variables:
        {}
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "recent authentication within 10 minutes is required for sensitive action",
              "extensions": {
                "code": "FORBIDDEN"
              }
            }
          ]
        }
--- PASS: TestMutationEnrollMyTotp (0.47s)
    --- PASS: TestMutationEnrollMyTotp/FreshSessionSuccessAndStaleForbidden (0.47s)
=== RUN   TestMutationConfirmMyTotp
=== RUN   TestMutationConfirmMyTotp/ValidCodeEnablesAndInvalidRejected
    iam_operations_e2e_test.go:2307: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateEmployeeUserInput!) { createEmployeeUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676840069365643_72.png",
            "display_name": "Fixture u_employee_1794676840069365643_72",
            "email": "u_employee_1794676840069365643_72@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createEmployeeUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676840069365643_72.png",
              "created_at": "2026-11-15T00:20:40.070168884+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_employee_1794676840069365643_72",
              "id": "9d31bf6f-934b-4ff2-8d15-5fa4d09d28eb",
              "kind": "EMPLOYEE",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:40.070168884+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2308: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_employee_1794676840069365643_72@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "QkZIuJXhPFy6DAIdidD3rlVnjZDEmdQpuH5VXnNfY6g",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/account",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:2310: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation { enrollMyTotp { factor_id otp_auth_uri } }
        Input Variables:
        {}
        Result JSON:
        {
          "data": {
            "enrollMyTotp": {
              "factor_id": "e81e4783-60f7-40ca-bc1d-f9c7e8824c9f",
              "otp_auth_uri": "otpauth://totp/IntegriOne%20ERP:Fixture%20u_employee_1794676840069365643_72?algorithm=SHA1\u0026digits=6\u0026issuer=IntegriOne%20ERP\u0026period=30\u0026secret=M4BUUTGSNNWVK73OKRQ6D2O5FPYIAN54"
            }
          }
        }
    iam_operations_e2e_test.go:2314: 
        [GQL_EXPECTED_ERROR] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: ConfirmTotpInput!) { confirmMyTotp(input: $input) { enabled recovery_codes } }
        Input Variables:
        {
          "input": {
            "code": "999999"
          }
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "invalid totp verification code",
              "extensions": {
                "code": "INVALID_CREDENTIALS"
              }
            }
          ]
        }
    iam_operations_e2e_test.go:2316: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: ConfirmTotpInput!) { confirmMyTotp(input: $input) { enabled recovery_codes } }
        Input Variables:
        {
          "input": {
            "code": "569896"
          }
        }
        Result JSON:
        {
          "data": {
            "confirmMyTotp": {
              "enabled": true,
              "recovery_codes": [
                "LXNA-3Q5U-3BQ6",
                "PEDP-AASA-QT3E",
                "P8Z8-EVCR-ERRL",
                "AS9F-TWSD-AXN5",
                "NFJS-MHVH-UGLR",
                "U8RY-WFP3-88WY",
                "BUUV-D5NZ-94H5",
                "P998-RAEN-JNKQ"
              ]
            }
          }
        }
--- PASS: TestMutationConfirmMyTotp (0.25s)
    --- PASS: TestMutationConfirmMyTotp/ValidCodeEnablesAndInvalidRejected (0.25s)
=== RUN   TestMutationDisableMyTotp
=== RUN   TestMutationDisableMyTotp/RequiresVerifiedSessionAndDisablesCleanly
    iam_operations_e2e_test.go:2330: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateEmployeeUserInput!) { createEmployeeUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676840318340221_73.png",
            "display_name": "Fixture u_employee_1794676840318340221_73",
            "email": "u_employee_1794676840318340221_73@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createEmployeeUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676840318340221_73.png",
              "created_at": "2026-11-15T00:20:40.319221741+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_employee_1794676840318340221_73",
              "id": "533b68e7-8b2a-44f4-9be3-69b9f9e95a69",
              "kind": "EMPLOYEE",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:40.319221741+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2331: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_employee_1794676840318340221_73@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "kFMoinvIlRvO5kD04PDaYrrKazT9FgMG50j85W5I0Zs",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/account",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:2332: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation { enrollMyTotp { factor_id otp_auth_uri } }
        Input Variables:
        {}
        Result JSON:
        {
          "data": {
            "enrollMyTotp": {
              "factor_id": "564f222d-d662-4817-b501-9d9279023f40",
              "otp_auth_uri": "otpauth://totp/IntegriOne%20ERP:Fixture%20u_employee_1794676840318340221_73?algorithm=SHA1\u0026digits=6\u0026issuer=IntegriOne%20ERP\u0026period=30\u0026secret=QDQVYSYIPW5ZZPKM5U2SUS4K75N5WTYV"
            }
          }
        }
    iam_operations_e2e_test.go:2332: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: ConfirmTotpInput!) { confirmMyTotp(input: $input) { enabled recovery_codes } }
        Input Variables:
        {
          "input": {
            "code": "863787"
          }
        }
        Result JSON:
        {
          "data": {
            "confirmMyTotp": {
              "enabled": true,
              "recovery_codes": [
                "PH6A-ZNBG-EJPV",
                "VMA2-2JT3-VR4G",
                "MWUH-MXRJ-DH7F",
                "6E8B-9JR2-7TCN",
                "UF65-5E3U-QNK2",
                "5FCZ-8KKJ-2ERB",
                "K7ZZ-4BN7-89KX",
                "3HYN-VK7X-SXRY"
              ]
            }
          }
        }
    iam_operations_e2e_test.go:2335: 
        [GQL_EXPECTED_ERROR] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation { disableMyTotp }
        Input Variables:
        {}
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "mfa step-up authentication required for sensitive action",
              "extensions": {
                "code": "FORBIDDEN"
              }
            }
          ]
        }
    iam_operations_e2e_test.go:2337: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($code: String!) { verifyMySessionMfa(totp_code: $code) }
        Input Variables:
        {
          "code": "386566"
        }
        Result JSON:
        {
          "data": {
            "verifyMySessionMfa": true
          }
        }
    iam_operations_e2e_test.go:2340: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation { disableMyTotp }
        Input Variables:
        {}
        Result JSON:
        {
          "data": {
            "disableMyTotp": true
          }
        }
    iam_operations_e2e_test.go:2342: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query { myAccessSnapshot { principal_id user_kind status mfa_enabled auth_epoch perm_revision recovery_epoch can_access_identity default_redirect_path allowed_redirect_paths roles { id code name priority default_redirect_path } permissions { permission_code is_sensitive scope { type id } } synced_at_unix fresh_until_unix } }
        Input Variables:
        {}
        Result JSON:
        {
          "data": {
            "myAccessSnapshot": {
              "allowed_redirect_paths": [
                "/",
                "/identity",
                "/human-resource",
                "/procurement",
                "/warehouse",
                "/finance",
                "/accounting",
                "/product",
                "/supplier",
                "/vendor",
                "/commerce",
                "/account",
                "/profile"
              ],
              "auth_epoch": 1,
              "can_access_identity": false,
              "default_redirect_path": "/account",
              "fresh_until_unix": 1794676870,
              "mfa_enabled": false,
              "perm_revision": 3,
              "permissions": [],
              "principal_id": "533b68e7-8b2a-44f4-9be3-69b9f9e95a69",
              "recovery_epoch": 24,
              "roles": [],
              "status": "active",
              "synced_at_unix": 1794676840,
              "user_kind": "employee"
            }
          }
        }
--- PASS: TestMutationDisableMyTotp (0.26s)
    --- PASS: TestMutationDisableMyTotp/RequiresVerifiedSessionAndDisablesCleanly (0.26s)
=== RUN   TestMutationRegenerateMyRecoveryCodes
=== RUN   TestMutationRegenerateMyRecoveryCodes/GeneratesFreshCodesAndInvalidatesPrevious
    iam_operations_e2e_test.go:2357: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateEmployeeUserInput!) { createEmployeeUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676840581336332_74.png",
            "display_name": "Fixture u_employee_1794676840581336332_74",
            "email": "u_employee_1794676840581336332_74@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createEmployeeUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676840581336332_74.png",
              "created_at": "2026-11-15T00:20:40.582137079+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_employee_1794676840581336332_74",
              "id": "d9f7b7e1-08fe-43db-9590-e4043898b451",
              "kind": "EMPLOYEE",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:40.582137079+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2358: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_employee_1794676840581336332_74@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "s23SusICt8iTPanfhlQ-GHYoUbVW4W4K9eylehbvxrg",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/account",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:2359: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation { enrollMyTotp { factor_id otp_auth_uri } }
        Input Variables:
        {}
        Result JSON:
        {
          "data": {
            "enrollMyTotp": {
              "factor_id": "d0060083-94e4-44e4-af71-c8db99362165",
              "otp_auth_uri": "otpauth://totp/IntegriOne%20ERP:Fixture%20u_employee_1794676840581336332_74?algorithm=SHA1\u0026digits=6\u0026issuer=IntegriOne%20ERP\u0026period=30\u0026secret=ICWHI5ONE6WMKLHUID32RW5QDZUY5M7M"
            }
          }
        }
    iam_operations_e2e_test.go:2359: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: ConfirmTotpInput!) { confirmMyTotp(input: $input) { enabled recovery_codes } }
        Input Variables:
        {
          "input": {
            "code": "098511"
          }
        }
        Result JSON:
        {
          "data": {
            "confirmMyTotp": {
              "enabled": true,
              "recovery_codes": [
                "T6L7-9BBY-UQQN",
                "RWBB-KRBV-JSE3",
                "XEPZ-MC48-P5FK",
                "5MMS-6H8B-TLTQ",
                "TAKZ-XRPZ-NZ26",
                "FPEB-4BDY-G5F2",
                "CBTR-KV7J-6JNF",
                "6MAE-TZQV-X6MZ"
              ]
            }
          }
        }
    iam_operations_e2e_test.go:2361: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($code: String!) { verifyMySessionMfa(totp_code: $code) }
        Input Variables:
        {
          "code": "450502"
        }
        Result JSON:
        {
          "data": {
            "verifyMySessionMfa": true
          }
        }
    iam_operations_e2e_test.go:2365: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation { regenerateMyRecoveryCodes }
        Input Variables:
        {}
        Result JSON:
        {
          "data": {
            "regenerateMyRecoveryCodes": [
              "DR5T-3WUF-2ZKD",
              "ZLPX-ZA9S-R6QW",
              "8KED-G2RB-D67Y",
              "B84G-7EUG-GR33",
              "XR7P-JTSD-GRUG",
              "M6ML-RUGT-6EJ7",
              "ACCP-GBNZ-RY23",
              "8FNM-NLU5-Z2MY"
            ]
          }
        }
    iam_operations_e2e_test.go:2376: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($in: UserLoginInput!) { userLogin(input: $in) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "in": {
            "identifier": "u_employee_1794676840581336332_74@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": null,
              "mfaChallenge": "9SJ1pNq3Q7_YfqNYIHxtpdLdLLhLErWc9_fjDc1rvLc",
              "mfaExpiresAt": "2026-11-15T00:25:40+07:00",
              "redirectLocation": null,
              "status": "MFA_REQUIRED"
            }
          }
        }
    iam_operations_e2e_test.go:2382: 
        [GQL_EXPECTED_ERROR] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($in: ValidateUserMfaInput!) { validateUserMfa(input: $in) { csrfToken redirectLocation } }
        Input Variables:
        {
          "in": {
            "challenge": "9SJ1pNq3Q7_YfqNYIHxtpdLdLLhLErWc9_fjDc1rvLc",
            "recoveryCode": "T6L7-9BBY-UQQN"
          }
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "invalid recovery code",
              "extensions": {
                "code": "INVALID_CREDENTIALS"
              }
            }
          ]
        }
    iam_operations_e2e_test.go:2386: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($in: ValidateUserMfaInput!) { validateUserMfa(input: $in) { csrfToken redirectLocation } }
        Input Variables:
        {
          "in": {
            "challenge": "9SJ1pNq3Q7_YfqNYIHxtpdLdLLhLErWc9_fjDc1rvLc",
            "recoveryCode": "DR5T-3WUF-2ZKD"
          }
        }
        Result JSON:
        {
          "data": {
            "validateUserMfa": {
              "csrfToken": "csodOLxEKOmrFTIeEoOsp7fRgyYvSCElnIq7HINCyB8",
              "redirectLocation": "/account"
            }
          }
        }
--- PASS: TestMutationRegenerateMyRecoveryCodes (0.31s)
    --- PASS: TestMutationRegenerateMyRecoveryCodes/GeneratesFreshCodesAndInvalidatesPrevious (0.31s)
=== RUN   TestMutationCreateUser
=== RUN   TestMutationCreateUser/CreateCustomerNonNull
    iam_operations_e2e_test.go:2401: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateUserInput!) { createUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676840887779800_75.png",
            "display_name": "Fixture u_customer_1794676840887779800_75",
            "email": "u_customer_1794676840887779800_75@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta",
            "username": "u_customer_1794676840887779800_75"
          }
        }
        Result JSON:
        {
          "data": {
            "createUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676840887779800_75.png",
              "created_at": "2026-11-15T00:20:40.888514632+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_customer_1794676840887779800_75",
              "id": "5ea9b9a3-32f5-47f4-ac21-9922ccb2baab",
              "kind": "CUSTOMER",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:40.888514632+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2402: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($id: ID!) { user(id: $id) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "id": "5ea9b9a3-32f5-47f4-ac21-9922ccb2baab"
        }
        Result JSON:
        {
          "data": {
            "user": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676840887779800_75.png",
              "created_at": "2026-11-15T00:20:40.888514+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_customer_1794676840887779800_75",
              "id": "5ea9b9a3-32f5-47f4-ac21-9922ccb2baab",
              "kind": "CUSTOMER",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:40.888514+07:00"
            }
          }
        }
=== RUN   TestMutationCreateUser/RBACCustomerForbidden
    iam_operations_e2e_test.go:2411: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateUserInput!) { createUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676841070185269_76.png",
            "display_name": "Fixture u_customer_1794676841070185269_76",
            "email": "u_customer_1794676841070185269_76@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676841070185269_76.png",
              "created_at": "2026-11-15T00:20:41.071230229+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_customer_1794676841070185269_76",
              "id": "0e705df2-7092-4494-95d6-2185bd4780bc",
              "kind": "CUSTOMER",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:41.071230229+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2411: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_customer_1794676841070185269_76@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "aCpkMhoa2yqnG3xpdxW0tnjFxqqElmz83EsiMGgk_to",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:2413: 
        [RBAC_TEST_AB_FORBIDDEN] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateUserInput!) { createUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "display_name": "Denied",
            "email": "denied_c_1794676841165241745_77@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "missing required permission: iam.users.write",
              "extensions": {
                "code": "FORBIDDEN"
              }
            }
          ]
        }
--- PASS: TestMutationCreateUser (0.42s)
    --- PASS: TestMutationCreateUser/CreateCustomerNonNull (0.18s)
    --- PASS: TestMutationCreateUser/RBACCustomerForbidden (0.24s)
=== RUN   TestMutationCreateEmployeeUser
=== RUN   TestMutationCreateEmployeeUser/CreateEmployeeNonNullAndLoginWorks
    iam_operations_e2e_test.go:2426: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateEmployeeUserInput!) { createEmployeeUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676841303760636_78.png",
            "display_name": "Fixture u_employee_1794676841303760636_78",
            "email": "u_employee_1794676841303760636_78@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta",
            "username": "u_employee_1794676841303760636_78"
          }
        }
        Result JSON:
        {
          "data": {
            "createEmployeeUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676841303760636_78.png",
              "created_at": "2026-11-15T00:20:41.304512309+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_employee_1794676841303760636_78",
              "id": "0f1d5530-0c9a-4e07-9a19-53eba53f01e7",
              "kind": "EMPLOYEE",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:41.304512309+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2427: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_employee_1794676841303760636_78@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "rDT04wSXCOeug-dN-1v_anobPaRJm4QZqQkkVLN0Uho",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/account",
              "status": "AUTHENTICATED"
            }
          }
        }
=== RUN   TestMutationCreateEmployeeUser/RBACCustomerForbidden
    iam_operations_e2e_test.go:2433: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateUserInput!) { createUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676841537008000_79.png",
            "display_name": "Fixture u_customer_1794676841537008000_79",
            "email": "u_customer_1794676841537008000_79@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676841537008000_79.png",
              "created_at": "2026-11-15T00:20:41.537675875+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_customer_1794676841537008000_79",
              "id": "41e18429-a849-4742-b662-9d7106667ec0",
              "kind": "CUSTOMER",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:41.537675875+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2433: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_customer_1794676841537008000_79@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "M94jIMQfl9t2myqp7a-0M0nwUIlFtWqkfRtWvKmA3QI",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:2435: 
        [RBAC_TEST_AB_FORBIDDEN] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateEmployeeUserInput!) { createEmployeeUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "display_name": "Denied",
            "email": "denied_e_1794676841632713585_80@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "missing required permission: iam.users.write",
              "extensions": {
                "code": "FORBIDDEN"
              }
            }
          ]
        }
--- PASS: TestMutationCreateEmployeeUser (0.47s)
    --- PASS: TestMutationCreateEmployeeUser/CreateEmployeeNonNullAndLoginWorks (0.23s)
    --- PASS: TestMutationCreateEmployeeUser/RBACCustomerForbidden (0.24s)
=== RUN   TestMutationCreateVendorUser
=== RUN   TestMutationCreateVendorUser/CreateVendorNonNullAndLoginWorks
    iam_operations_e2e_test.go:2448: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateVendorUserInput!) { createVendorUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_vendor_1794676841780430360_81.png",
            "display_name": "Fixture u_vendor_1794676841780430360_81",
            "email": "u_vendor_1794676841780430360_81@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta",
            "username": "u_vendor_1794676841780430360_81"
          }
        }
        Result JSON:
        {
          "data": {
            "createVendorUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_vendor_1794676841780430360_81.png",
              "created_at": "2026-11-15T00:20:41.781249843+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_vendor_1794676841780430360_81",
              "id": "1a14b19d-063c-4d33-a770-21f5fd3723a7",
              "kind": "VENDOR",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:41.781249843+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2449: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_vendor_1794676841780430360_81@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "7RDnp0uEBOGCaeT9MXbcV_KTOgW3RCYs0kgjBOQFPSk",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/vendor",
              "status": "AUTHENTICATED"
            }
          }
        }
=== RUN   TestMutationCreateVendorUser/RBACCustomerForbidden
    iam_operations_e2e_test.go:2455: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateUserInput!) { createUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676842013149304_82.png",
            "display_name": "Fixture u_customer_1794676842013149304_82",
            "email": "u_customer_1794676842013149304_82@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676842013149304_82.png",
              "created_at": "2026-11-15T00:20:42.013888174+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_customer_1794676842013149304_82",
              "id": "b3b66bae-fcea-4273-9ae4-f124f948d816",
              "kind": "CUSTOMER",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:42.013888174+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2455: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_customer_1794676842013149304_82@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "K_x-Wy7rZeSv-Cz70Uux9P9Jp5fLKslW8DK4rUemLnY",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:2457: 
        [RBAC_TEST_AB_FORBIDDEN] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateVendorUserInput!) { createVendorUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "display_name": "Denied",
            "email": "denied_v_1794676842107464545_83@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "missing required permission: iam.users.write",
              "extensions": {
                "code": "FORBIDDEN"
              }
            }
          ]
        }
--- PASS: TestMutationCreateVendorUser (0.47s)
    --- PASS: TestMutationCreateVendorUser/CreateVendorNonNullAndLoginWorks (0.23s)
    --- PASS: TestMutationCreateVendorUser/RBACCustomerForbidden (0.23s)
=== RUN   TestMutationCreateSupplierUser
=== RUN   TestMutationCreateSupplierUser/CreateSupplierNonNullAndLoginWorks
    iam_operations_e2e_test.go:2470: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateSupplierUserInput!) { createSupplierUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_supplier_1794676842246429090_84.png",
            "display_name": "Fixture u_supplier_1794676842246429090_84",
            "email": "u_supplier_1794676842246429090_84@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta",
            "username": "u_supplier_1794676842246429090_84"
          }
        }
        Result JSON:
        {
          "data": {
            "createSupplierUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_supplier_1794676842246429090_84.png",
              "created_at": "2026-11-15T00:20:42.247158282+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_supplier_1794676842246429090_84",
              "id": "c899f782-4736-4f3c-9639-ef8461ebc2ad",
              "kind": "SUPPLIER",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:42.247158282+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2471: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_supplier_1794676842246429090_84@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "PTCvIFAy51Fso4x-MAdyVUMFVpxvRU4fBykUHMEhpvg",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/supplier",
              "status": "AUTHENTICATED"
            }
          }
        }
=== RUN   TestMutationCreateSupplierUser/RBACCustomerForbidden
    iam_operations_e2e_test.go:2477: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateUserInput!) { createUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676842475734858_85.png",
            "display_name": "Fixture u_customer_1794676842475734858_85",
            "email": "u_customer_1794676842475734858_85@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676842475734858_85.png",
              "created_at": "2026-11-15T00:20:42.476586502+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_customer_1794676842475734858_85",
              "id": "075fb9ad-683d-4672-ad6c-32e760567a26",
              "kind": "CUSTOMER",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:42.476586502+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2477: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_customer_1794676842475734858_85@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "DlYQlRu0irqUMdqQdwE_JpKu3uTwqzHRe5yieI1UFJY",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:2479: 
        [RBAC_TEST_AB_FORBIDDEN] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateSupplierUserInput!) { createSupplierUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "display_name": "Denied",
            "email": "denied_s_1794676842567323885_86@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "missing required permission: iam.users.write",
              "extensions": {
                "code": "FORBIDDEN"
              }
            }
          ]
        }
--- PASS: TestMutationCreateSupplierUser (0.46s)
    --- PASS: TestMutationCreateSupplierUser/CreateSupplierNonNullAndLoginWorks (0.23s)
    --- PASS: TestMutationCreateSupplierUser/RBACCustomerForbidden (0.23s)
=== RUN   TestMutationActivateUser
=== RUN   TestMutationActivateUser/ActivateSuspendedUser
    iam_operations_e2e_test.go:2492: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateEmployeeUserInput!) { createEmployeeUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676842704714218_87.png",
            "display_name": "Fixture u_employee_1794676842704714218_87",
            "email": "u_employee_1794676842704714218_87@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createEmployeeUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676842704714218_87.png",
              "created_at": "2026-11-15T00:20:42.705414114+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_employee_1794676842704714218_87",
              "id": "3bb19574-b61d-4651-9a91-b8330c894bb0",
              "kind": "EMPLOYEE",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:42.705414114+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2493: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($id: ID!) { suspendUser(user_id: $id) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "id": "3bb19574-b61d-4651-9a91-b8330c894bb0"
        }
        Result JSON:
        {
          "data": {
            "suspendUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676842704714218_87.png",
              "created_at": "2026-11-15T00:20:42.705414+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_employee_1794676842704714218_87",
              "id": "3bb19574-b61d-4651-9a91-b8330c894bb0",
              "kind": "EMPLOYEE",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "SUSPENDED",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:42.749689359+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2496: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($user_id: ID!) { activateUser(user_id: $user_id) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "user_id": "3bb19574-b61d-4651-9a91-b8330c894bb0"
        }
        Result JSON:
        {
          "data": {
            "activateUser": {
              "auth_epoch": 2,
              "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676842704714218_87.png",
              "created_at": "2026-11-15T00:20:42.705414+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_employee_1794676842704714218_87",
              "id": "3bb19574-b61d-4651-9a91-b8330c894bb0",
              "kind": "EMPLOYEE",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:42.751939602+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2501: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_employee_1794676842704714218_87@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "q8ol_xTAJbWPgBASp0_hOYiieLqnyrZaczdvg2rvXkc",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/account",
              "status": "AUTHENTICATED"
            }
          }
        }
=== RUN   TestMutationActivateUser/RBACCustomerForbidden
    iam_operations_e2e_test.go:2507: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateUserInput!) { createUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676842940158895_88.png",
            "display_name": "Fixture u_customer_1794676842940158895_88",
            "email": "u_customer_1794676842940158895_88@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676842940158895_88.png",
              "created_at": "2026-11-15T00:20:42.940840136+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_customer_1794676842940158895_88",
              "id": "2fca6e2c-d902-48d1-9422-27ec332b6d15",
              "kind": "CUSTOMER",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:42.940840136+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2507: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_customer_1794676842940158895_88@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "ooC7iMaRhd9sgpS7xfHNORFpnbCDPrvLcXA_3OmqNr4",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:2508: 
        [RBAC_TEST_AB_FORBIDDEN] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($user_id: ID!) { activateUser(user_id: $user_id) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "user_id": "2fca6e2c-d902-48d1-9422-27ec332b6d15"
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "missing required permission: iam.users.write",
              "extensions": {
                "code": "FORBIDDEN"
              }
            }
          ]
        }
--- PASS: TestMutationActivateUser (0.47s)
    --- PASS: TestMutationActivateUser/ActivateSuspendedUser (0.23s)
    --- PASS: TestMutationActivateUser/RBACCustomerForbidden (0.23s)
=== RUN   TestMutationSuspendUser
=== RUN   TestMutationSuspendUser/SuspendActiveUserBlocksSession
    iam_operations_e2e_test.go:2519: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateEmployeeUserInput!) { createEmployeeUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676843176757749_89.png",
            "display_name": "Fixture u_employee_1794676843176757749_89",
            "email": "u_employee_1794676843176757749_89@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createEmployeeUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676843176757749_89.png",
              "created_at": "2026-11-15T00:20:43.17751298+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_employee_1794676843176757749_89",
              "id": "8a999766-4bdb-4a2d-835c-7cb271936405",
              "kind": "EMPLOYEE",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:43.17751298+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2520: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_employee_1794676843176757749_89@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "OQp0e355VTDy_D7CuS7Jf_CktFV70At1qS3eV6GAdRA",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/account",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:2522: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($user_id: ID!) { suspendUser(user_id: $user_id) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "user_id": "8a999766-4bdb-4a2d-835c-7cb271936405"
        }
        Result JSON:
        {
          "data": {
            "suspendUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676843176757749_89.png",
              "created_at": "2026-11-15T00:20:43.177512+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_employee_1794676843176757749_89",
              "id": "8a999766-4bdb-4a2d-835c-7cb271936405",
              "kind": "EMPLOYEE",
              "last_login_at": "2026-11-15T00:20:43.226654+07:00",
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "SUSPENDED",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:43.277285369+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2527: 
        [GQL_EXPECTED_ERROR] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query { myProfile { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {}
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "session is missing, expired, or revoked",
              "extensions": {
                "code": "UNAUTHENTICATED"
              }
            }
          ]
        }
=== RUN   TestMutationSuspendUser/RBACCustomerForbidden
    iam_operations_e2e_test.go:2532: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateUserInput!) { createUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676843412911369_90.png",
            "display_name": "Fixture u_customer_1794676843412911369_90",
            "email": "u_customer_1794676843412911369_90@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676843412911369_90.png",
              "created_at": "2026-11-15T00:20:43.413631704+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_customer_1794676843412911369_90",
              "id": "ff5d9184-c2a2-4a15-ba17-2a2011a090bf",
              "kind": "CUSTOMER",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:43.413631704+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2532: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_customer_1794676843412911369_90@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "NJfwDhWDH0oq4KTvZzYv__jrPxGrV9Ik7keBYAS9TDg",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:2533: 
        [RBAC_TEST_AB_FORBIDDEN] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($user_id: ID!) { suspendUser(user_id: $user_id) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "user_id": "ff5d9184-c2a2-4a15-ba17-2a2011a090bf"
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "missing required permission: iam.users.write",
              "extensions": {
                "code": "FORBIDDEN"
              }
            }
          ]
        }
--- PASS: TestMutationSuspendUser (0.47s)
    --- PASS: TestMutationSuspendUser/SuspendActiveUserBlocksSession (0.24s)
    --- PASS: TestMutationSuspendUser/RBACCustomerForbidden (0.23s)
=== RUN   TestMutationDisableUser
=== RUN   TestMutationDisableUser/DisableActiveUserBlocksSession
    iam_operations_e2e_test.go:2544: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateEmployeeUserInput!) { createEmployeeUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676843647237187_91.png",
            "display_name": "Fixture u_employee_1794676843647237187_91",
            "email": "u_employee_1794676843647237187_91@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createEmployeeUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676843647237187_91.png",
              "created_at": "2026-11-15T00:20:43.648302445+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_employee_1794676843647237187_91",
              "id": "0f1893b3-054b-43b9-9bf3-1868683802e7",
              "kind": "EMPLOYEE",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:43.648302445+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2545: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_employee_1794676843647237187_91@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "UJ825n9pAvh0637qfTjMvOt3y8b_FEsksUBhOT291Xc",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/account",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:2547: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($user_id: ID!) { disableUser(user_id: $user_id) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "user_id": "0f1893b3-054b-43b9-9bf3-1868683802e7"
        }
        Result JSON:
        {
          "data": {
            "disableUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676843647237187_91.png",
              "created_at": "2026-11-15T00:20:43.648302+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_employee_1794676843647237187_91",
              "id": "0f1893b3-054b-43b9-9bf3-1868683802e7",
              "kind": "EMPLOYEE",
              "last_login_at": "2026-11-15T00:20:43.698783+07:00",
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "DISABLED",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:43.745748026+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2552: 
        [GQL_EXPECTED_ERROR] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query { myProfile { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {}
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "session is missing, expired, or revoked",
              "extensions": {
                "code": "UNAUTHENTICATED"
              }
            }
          ]
        }
=== RUN   TestMutationDisableUser/RBACCustomerForbidden
    iam_operations_e2e_test.go:2557: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateUserInput!) { createUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676843886387614_92.png",
            "display_name": "Fixture u_customer_1794676843886387614_92",
            "email": "u_customer_1794676843886387614_92@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676843886387614_92.png",
              "created_at": "2026-11-15T00:20:43.887081419+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_customer_1794676843886387614_92",
              "id": "93f9564f-a888-491e-a28c-2d25f1ef73cc",
              "kind": "CUSTOMER",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:43.887081419+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2557: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_customer_1794676843886387614_92@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "5KjtgHoAyzgN3QTEIfpn6IqI9r_OtBduf3LiaiplUPc",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:2558: 
        [RBAC_TEST_AB_FORBIDDEN] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($user_id: ID!) { disableUser(user_id: $user_id) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "user_id": "93f9564f-a888-491e-a28c-2d25f1ef73cc"
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "missing required permission: iam.users.write",
              "extensions": {
                "code": "FORBIDDEN"
              }
            }
          ]
        }
--- PASS: TestMutationDisableUser (0.47s)
    --- PASS: TestMutationDisableUser/DisableActiveUserBlocksSession (0.24s)
    --- PASS: TestMutationDisableUser/RBACCustomerForbidden (0.23s)
=== RUN   TestMutationSoftDeleteUser
=== RUN   TestMutationSoftDeleteUser/SoftDeleteSetsDeletedAt
    iam_operations_e2e_test.go:2569: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateEmployeeUserInput!) { createEmployeeUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676844117375789_93.png",
            "display_name": "Fixture u_employee_1794676844117375789_93",
            "email": "u_employee_1794676844117375789_93@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createEmployeeUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676844117375789_93.png",
              "created_at": "2026-11-15T00:20:44.118085063+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_employee_1794676844117375789_93",
              "id": "9129678e-a0a6-4d17-869b-f7ac9ed3706e",
              "kind": "EMPLOYEE",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:44.118085063+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2571: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($user_id: ID!) { softDeleteUser(user_id: $user_id) }
        Input Variables:
        {
          "user_id": "9129678e-a0a6-4d17-869b-f7ac9ed3706e"
        }
        Result JSON:
        {
          "data": {
            "softDeleteUser": true
          }
        }
    iam_operations_e2e_test.go:2573: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($q: String) { users(query: $q, include_deleted: true, page: { limit: 10, offset: 0 }) { items { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } page { total limit offset has_more } } }
        Input Variables:
        {
          "q": "u_employee_1794676844117375789_93@integrione.test"
        }
        Result JSON:
        {
          "data": {
            "users": {
              "items": [
                {
                  "auth_epoch": 2,
                  "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676844117375789_93.png",
                  "created_at": "2026-11-15T00:20:44.118085+07:00",
                  "deleted_at": "2026-11-15T00:20:44.164357+07:00",
                  "display_name": "Fixture u_employee_1794676844117375789_93",
                  "id": "9129678e-a0a6-4d17-869b-f7ac9ed3706e",
                  "kind": "EMPLOYEE",
                  "last_login_at": null,
                  "locale": "id-ID",
                  "perm_revision": 1,
                  "status": "DISABLED",
                  "timezone": "Asia/Jakarta",
                  "updated_at": "2026-11-15T00:20:44.164142+07:00"
                }
              ],
              "page": {
                "has_more": false,
                "limit": 10,
                "offset": 0,
                "total": 1
              }
            }
          }
        }
=== RUN   TestMutationSoftDeleteUser/RBACCustomerForbidden
    iam_operations_e2e_test.go:2587: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateUserInput!) { createUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676844302439635_94.png",
            "display_name": "Fixture u_customer_1794676844302439635_94",
            "email": "u_customer_1794676844302439635_94@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676844302439635_94.png",
              "created_at": "2026-11-15T00:20:44.303224302+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_customer_1794676844302439635_94",
              "id": "778f6604-62b1-4e72-9684-1a6406eb0f02",
              "kind": "CUSTOMER",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:44.303224302+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2587: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_customer_1794676844302439635_94@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "Sygi6UXJ2q8k0LqfGXlcHGwRWngKgo_vsuktlKQjBZg",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:2588: 
        [RBAC_TEST_AB_FORBIDDEN] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($user_id: ID!) { softDeleteUser(user_id: $user_id) }
        Input Variables:
        {
          "user_id": "778f6604-62b1-4e72-9684-1a6406eb0f02"
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "missing required permission: iam.users.write",
              "extensions": {
                "code": "FORBIDDEN"
              }
            }
          ]
        }
--- PASS: TestMutationSoftDeleteUser (0.44s)
    --- PASS: TestMutationSoftDeleteUser/SoftDeleteSetsDeletedAt (0.19s)
    --- PASS: TestMutationSoftDeleteUser/RBACCustomerForbidden (0.25s)
=== RUN   TestMutationAddUserIdentity
=== RUN   TestMutationAddUserIdentity/AddSecondaryUsernameAndDuplicateConflict
    iam_operations_e2e_test.go:2599: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateEmployeeUserInput!) { createEmployeeUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676844552263332_95.png",
            "display_name": "Fixture u_employee_1794676844552263332_95",
            "email": "u_employee_1794676844552263332_95@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createEmployeeUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676844552263332_95.png",
              "created_at": "2026-11-15T00:20:44.552978066+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_employee_1794676844552263332_95",
              "id": "bf2a1d1a-7201-43af-9570-def4ddb31da6",
              "kind": "EMPLOYEE",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:44.552978066+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2602: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: AddUserIdentityInput!) { addUserIdentity(input: $input) { id user_id kind value is_verified is_primary verified_at created_at updated_at } }
        Input Variables:
        {
          "input": {
            "is_primary": false,
            "kind": "USERNAME",
            "user_id": "bf2a1d1a-7201-43af-9570-def4ddb31da6",
            "value": "sec_user_1794676844605156002_96"
          }
        }
        Result JSON:
        {
          "data": {
            "addUserIdentity": {
              "created_at": "2026-11-15T00:20:44.606044175+07:00",
              "id": "2e5311d9-b636-4f14-9f45-fd143389df15",
              "is_primary": false,
              "is_verified": false,
              "kind": "USERNAME",
              "updated_at": "2026-11-15T00:20:44.606044175+07:00",
              "user_id": "bf2a1d1a-7201-43af-9570-def4ddb31da6",
              "value": "sec_user_1794676844605156002_96",
              "verified_at": null
            }
          }
        }
    iam_operations_e2e_test.go:2616: 
        [GQL_EXPECTED_ERROR] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: AddUserIdentityInput!) { addUserIdentity(input: $input) { id user_id kind value is_verified is_primary verified_at created_at updated_at } }
        Input Variables:
        {
          "input": {
            "kind": "USERNAME",
            "user_id": "bf2a1d1a-7201-43af-9570-def4ddb31da6",
            "value": "sec_user_1794676844605156002_96"
          }
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "username is already registered",
              "extensions": {
                "code": "CONFLICT"
              }
            }
          ]
        }
=== RUN   TestMutationAddUserIdentity/RBACCustomerForbidden
    iam_operations_e2e_test.go:2623: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateUserInput!) { createUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676844742870689_97.png",
            "display_name": "Fixture u_customer_1794676844742870689_97",
            "email": "u_customer_1794676844742870689_97@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676844742870689_97.png",
              "created_at": "2026-11-15T00:20:44.743955304+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_customer_1794676844742870689_97",
              "id": "0e2a0d09-3502-449f-8992-f62f818f1cb4",
              "kind": "CUSTOMER",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:44.743955304+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2623: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_customer_1794676844742870689_97@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "D7XTX-baZ_TXTPurk3TsGU_wmk0cZcEXUNqfTxuZl6w",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:2624: 
        [RBAC_TEST_AB_FORBIDDEN] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: AddUserIdentityInput!) { addUserIdentity(input: $input) { id user_id kind value is_verified is_primary verified_at created_at updated_at } }
        Input Variables:
        {
          "input": {
            "kind": "USERNAME",
            "user_id": "0e2a0d09-3502-449f-8992-f62f818f1cb4",
            "value": "no_1794676844839510524_98"
          }
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "missing required permission: iam.users.write",
              "extensions": {
                "code": "FORBIDDEN"
              }
            }
          ]
        }
--- PASS: TestMutationAddUserIdentity (0.42s)
    --- PASS: TestMutationAddUserIdentity/AddSecondaryUsernameAndDuplicateConflict (0.19s)
    --- PASS: TestMutationAddUserIdentity/RBACCustomerForbidden (0.23s)
=== RUN   TestMutationSetPrimaryUserIdentity
=== RUN   TestMutationSetPrimaryUserIdentity/SwitchPrimaryToVerifiedUsernameAndRejectUnverified
    iam_operations_e2e_test.go:2637: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateEmployeeUserInput!) { createEmployeeUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676844978759618_99.png",
            "display_name": "Fixture u_employee_1794676844978759618_99",
            "email": "u_employee_1794676844978759618_99@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta",
            "username": "u_employee_1794676844978759618_99"
          }
        }
        Result JSON:
        {
          "data": {
            "createEmployeeUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676844978759618_99.png",
              "created_at": "2026-11-15T00:20:44.979554975+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_employee_1794676844978759618_99",
              "id": "1d02187b-ce9e-4f14-99be-2e0b0fe4ea96",
              "kind": "EMPLOYEE",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:44.979554975+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2639: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($u: ID!) { userIdentities(user_id: $u) { id user_id kind value is_verified is_primary verified_at created_at updated_at } }
        Input Variables:
        {
          "u": "1d02187b-ce9e-4f14-99be-2e0b0fe4ea96"
        }
        Result JSON:
        {
          "data": {
            "userIdentities": [
              {
                "created_at": "2026-11-15T00:20:44.979554+07:00",
                "id": "88ffcc69-fa0c-4f4e-bd91-834df46a8ddf",
                "is_primary": true,
                "is_verified": true,
                "kind": "EMAIL",
                "updated_at": "2026-11-15T00:20:44.979554+07:00",
                "user_id": "1d02187b-ce9e-4f14-99be-2e0b0fe4ea96",
                "value": "u_employee_1794676844978759618_99@integrione.test",
                "verified_at": "2026-11-15T00:20:44.979554+07:00"
              },
              {
                "created_at": "2026-11-15T00:20:44.979554+07:00",
                "id": "e0b08b84-5d39-42c0-95e5-48f656fa40b7",
                "is_primary": false,
                "is_verified": true,
                "kind": "USERNAME",
                "updated_at": "2026-11-15T00:20:44.979554+07:00",
                "user_id": "1d02187b-ce9e-4f14-99be-2e0b0fe4ea96",
                "value": "u_employee_1794676844978759618_99",
                "verified_at": "2026-11-15T00:20:44.979554+07:00"
              }
            ]
          }
        }
    iam_operations_e2e_test.go:2652: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($user_id: ID!, $identity_id: ID!) { setPrimaryUserIdentity(user_id: $user_id, identity_id: $identity_id) { id user_id kind value is_verified is_primary verified_at created_at updated_at } }
        Input Variables:
        {
          "identity_id": "e0b08b84-5d39-42c0-95e5-48f656fa40b7",
          "user_id": "1d02187b-ce9e-4f14-99be-2e0b0fe4ea96"
        }
        Result JSON:
        {
          "data": {
            "setPrimaryUserIdentity": {
              "created_at": "2026-11-15T00:20:44.979554+07:00",
              "id": "e0b08b84-5d39-42c0-95e5-48f656fa40b7",
              "is_primary": true,
              "is_verified": true,
              "kind": "USERNAME",
              "updated_at": "2026-11-15T00:20:45.025829618+07:00",
              "user_id": "1d02187b-ce9e-4f14-99be-2e0b0fe4ea96",
              "value": "u_employee_1794676844978759618_99",
              "verified_at": "2026-11-15T00:20:44.979554+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2663: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($in: AddUserIdentityInput!) { addUserIdentity(input: $in) { id user_id kind value is_verified is_primary verified_at created_at updated_at } }
        Input Variables:
        {
          "in": {
            "is_primary": false,
            "kind": "EMAIL",
            "user_id": "1d02187b-ce9e-4f14-99be-2e0b0fe4ea96",
            "value": "unver_1794676845026787833_100@integrione.test"
          }
        }
        Result JSON:
        {
          "data": {
            "addUserIdentity": {
              "created_at": "2026-11-15T00:20:45.027684362+07:00",
              "id": "65ab875e-8ce6-4f23-8966-b4433276462a",
              "is_primary": false,
              "is_verified": false,
              "kind": "EMAIL",
              "updated_at": "2026-11-15T00:20:45.027684362+07:00",
              "user_id": "1d02187b-ce9e-4f14-99be-2e0b0fe4ea96",
              "value": "unver_1794676845026787833_100@integrione.test",
              "verified_at": null
            }
          }
        }
    iam_operations_e2e_test.go:2669: 
        [GQL_EXPECTED_ERROR] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($user_id: ID!, $identity_id: ID!) { setPrimaryUserIdentity(user_id: $user_id, identity_id: $identity_id) { id user_id kind value is_verified is_primary verified_at created_at updated_at } }
        Input Variables:
        {
          "identity_id": "65ab875e-8ce6-4f23-8966-b4433276462a",
          "user_id": "1d02187b-ce9e-4f14-99be-2e0b0fe4ea96"
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "user identity is not verified",
              "extensions": {
                "code": "FORBIDDEN"
              }
            }
          ]
        }
=== RUN   TestMutationSetPrimaryUserIdentity/RBACCustomerForbidden
    iam_operations_e2e_test.go:2674: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateUserInput!) { createUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676845168157785_101.png",
            "display_name": "Fixture u_customer_1794676845168157785_101",
            "email": "u_customer_1794676845168157785_101@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676845168157785_101.png",
              "created_at": "2026-11-15T00:20:45.168965836+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_customer_1794676845168157785_101",
              "id": "7a866392-1d83-49ff-a68c-6b71f2aa1bb6",
              "kind": "CUSTOMER",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:45.168965836+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2674: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_customer_1794676845168157785_101@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "1yknDrkl5hTLe4Ag9rFYvZsg9ifWi5ECvLY1Erpbqjc",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:2675: 
        [RBAC_TEST_AB_FORBIDDEN] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($user_id: ID!, $identity_id: ID!) { setPrimaryUserIdentity(user_id: $user_id, identity_id: $identity_id) { id user_id kind value is_verified is_primary verified_at created_at updated_at } }
        Input Variables:
        {
          "identity_id": "00000000-0000-4000-8000-000000000001",
          "user_id": "7a866392-1d83-49ff-a68c-6b71f2aa1bb6"
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "missing required permission: iam.users.write",
              "extensions": {
                "code": "FORBIDDEN"
              }
            }
          ]
        }
--- PASS: TestMutationSetPrimaryUserIdentity (0.43s)
    --- PASS: TestMutationSetPrimaryUserIdentity/SwitchPrimaryToVerifiedUsernameAndRejectUnverified (0.19s)
    --- PASS: TestMutationSetPrimaryUserIdentity/RBACCustomerForbidden (0.24s)
=== RUN   TestMutationRemoveUserIdentity
=== RUN   TestMutationRemoveUserIdentity/RemoveSecondaryAndProtectPrimary
    iam_operations_e2e_test.go:2686: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateEmployeeUserInput!) { createEmployeeUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676845408071859_102.png",
            "display_name": "Fixture u_employee_1794676845408071859_102",
            "email": "u_employee_1794676845408071859_102@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta",
            "username": "u_employee_1794676845408071859_102"
          }
        }
        Result JSON:
        {
          "data": {
            "createEmployeeUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676845408071859_102.png",
              "created_at": "2026-11-15T00:20:45.409282643+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_employee_1794676845408071859_102",
              "id": "cf3b1f99-7999-44ce-aa02-e2bbfe02c125",
              "kind": "EMPLOYEE",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:45.409282643+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2688: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($u: ID!) { userIdentities(user_id: $u) { id user_id kind value is_verified is_primary verified_at created_at updated_at } }
        Input Variables:
        {
          "u": "cf3b1f99-7999-44ce-aa02-e2bbfe02c125"
        }
        Result JSON:
        {
          "data": {
            "userIdentities": [
              {
                "created_at": "2026-11-15T00:20:45.409282+07:00",
                "id": "39ed605a-e50a-4495-ad19-a6cf46d498c1",
                "is_primary": true,
                "is_verified": true,
                "kind": "EMAIL",
                "updated_at": "2026-11-15T00:20:45.409282+07:00",
                "user_id": "cf3b1f99-7999-44ce-aa02-e2bbfe02c125",
                "value": "u_employee_1794676845408071859_102@integrione.test",
                "verified_at": "2026-11-15T00:20:45.409282+07:00"
              },
              {
                "created_at": "2026-11-15T00:20:45.409282+07:00",
                "id": "82c907ee-671a-414a-b1ac-a38a638c4efe",
                "is_primary": false,
                "is_verified": true,
                "kind": "USERNAME",
                "updated_at": "2026-11-15T00:20:45.409282+07:00",
                "user_id": "cf3b1f99-7999-44ce-aa02-e2bbfe02c125",
                "value": "u_employee_1794676845408071859_102",
                "verified_at": "2026-11-15T00:20:45.409282+07:00"
              }
            ]
          }
        }
    iam_operations_e2e_test.go:2700: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($user_id: ID!, $identity_id: ID!) { removeUserIdentity(user_id: $user_id, identity_id: $identity_id) }
        Input Variables:
        {
          "identity_id": "82c907ee-671a-414a-b1ac-a38a638c4efe",
          "user_id": "cf3b1f99-7999-44ce-aa02-e2bbfe02c125"
        }
        Result JSON:
        {
          "data": {
            "removeUserIdentity": true
          }
        }
    iam_operations_e2e_test.go:2702: 
        [GQL_EXPECTED_ERROR] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($user_id: ID!, $identity_id: ID!) { removeUserIdentity(user_id: $user_id, identity_id: $identity_id) }
        Input Variables:
        {
          "identity_id": "39ed605a-e50a-4495-ad19-a6cf46d498c1",
          "user_id": "cf3b1f99-7999-44ce-aa02-e2bbfe02c125"
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "cannot remove user's last remaining login identity",
              "extensions": {
                "code": "CONFLICT"
              }
            }
          ]
        }
=== RUN   TestMutationRemoveUserIdentity/RBACCustomerForbidden
    iam_operations_e2e_test.go:2707: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateUserInput!) { createUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676845599224029_103.png",
            "display_name": "Fixture u_customer_1794676845599224029_103",
            "email": "u_customer_1794676845599224029_103@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676845599224029_103.png",
              "created_at": "2026-11-15T00:20:45.600064892+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_customer_1794676845599224029_103",
              "id": "ca0c1de8-d717-4226-8be4-43f3ab41ef0b",
              "kind": "CUSTOMER",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:45.600064892+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2707: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_customer_1794676845599224029_103@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "4EHewU3dKu8vHMupIfARRKmDcwmMXjMWxkLqTVq39-k",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:2708: 
        [RBAC_TEST_AB_FORBIDDEN] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($user_id: ID!, $identity_id: ID!) { removeUserIdentity(user_id: $user_id, identity_id: $identity_id) }
        Input Variables:
        {
          "identity_id": "00000000-0000-4000-8000-000000000001",
          "user_id": "ca0c1de8-d717-4226-8be4-43f3ab41ef0b"
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "missing required permission: iam.users.write",
              "extensions": {
                "code": "FORBIDDEN"
              }
            }
          ]
        }
--- PASS: TestMutationRemoveUserIdentity (0.43s)
    --- PASS: TestMutationRemoveUserIdentity/RemoveSecondaryAndProtectPrimary (0.19s)
    --- PASS: TestMutationRemoveUserIdentity/RBACCustomerForbidden (0.24s)
=== RUN   TestMutationCreateRole
=== RUN   TestMutationCreateRole/CreateValidAndRejectInvalidRedirectOrDuplicate
    iam_operations_e2e_test.go:2719: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateRoleInput!) { createRole(input: $input) { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "role_cr_1794676845836722607_104",
            "default_redirect_path": "/identity",
            "description": "Custom E2E role role_cr_1794676845836722607_104",
            "is_system": false,
            "management_domain": "iam",
            "name": "Role role_cr_1794676845836722607_104",
            "redirect_priority": 50
          }
        }
        Result JSON:
        {
          "data": {
            "createRole": {
              "code": "role_cr_1794676845836722607_104",
              "created_at": "2026-11-15T00:20:45.837540015+07:00",
              "default_redirect_path": "/identity",
              "description": "Custom E2E role role_cr_1794676845836722607_104",
              "id": "9dd5da62-03c5-42ef-b194-cb07e3e22bea",
              "is_active": true,
              "is_system": false,
              "management_domain": "iam",
              "name": "Role role_cr_1794676845836722607_104",
              "redirect_priority": 50,
              "updated_at": "2026-11-15T00:20:45.837540015+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2722: 
        [GQL_EXPECTED_ERROR] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateRoleInput!) { createRole(input: $input) { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "role_cr_1794676845836722607_104",
            "default_redirect_path": "/identity",
            "management_domain": "iam",
            "name": "Dup"
          }
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "role code already exists",
              "extensions": {
                "code": "CONFLICT"
              }
            }
          ]
        }
    iam_operations_e2e_test.go:2726: 
        [GQL_EXPECTED_ERROR] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateRoleInput!) { createRole(input: $input) { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "bad_red_1794676845839912510_105",
            "default_redirect_path": "https://evil.example.com",
            "management_domain": "iam",
            "name": "BadRedirect"
          }
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "invalid internal redirect path",
              "extensions": {
                "code": "INVALID_ARGUMENT"
              }
            }
          ]
        }
=== RUN   TestMutationCreateRole/RBACCustomerForbidden
    iam_operations_e2e_test.go:2733: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateUserInput!) { createUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676846023663799_106.png",
            "display_name": "Fixture u_customer_1794676846023663799_106",
            "email": "u_customer_1794676846023663799_106@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676846023663799_106.png",
              "created_at": "2026-11-15T00:20:46.024548215+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_customer_1794676846023663799_106",
              "id": "c14057db-eba0-494a-a878-d7e7d4af44c1",
              "kind": "CUSTOMER",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:46.024548215+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2733: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_customer_1794676846023663799_106@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "Ykb4i76uMyepNHJEEmlC7eOA2gMvelb4AGPLNBaEaoI",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:2734: 
        [RBAC_TEST_AB_FORBIDDEN] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateRoleInput!) { createRole(input: $input) { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "no_role_1794676846120589486_107",
            "default_redirect_path": "/identity",
            "management_domain": "iam",
            "name": "No"
          }
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "missing required permission: iam.roles.write",
              "extensions": {
                "code": "FORBIDDEN"
              }
            }
          ]
        }
--- PASS: TestMutationCreateRole (0.43s)
    --- PASS: TestMutationCreateRole/CreateValidAndRejectInvalidRedirectOrDuplicate (0.19s)
    --- PASS: TestMutationCreateRole/RBACCustomerForbidden (0.24s)
=== RUN   TestMutationUpdateRoleDetails
=== RUN   TestMutationUpdateRoleDetails/UpdateAllMutableFields
    iam_operations_e2e_test.go:2747: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateRoleInput!) { createRole(input: $input) { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "role_upd_1794676846259593115_108",
            "default_redirect_path": "/identity",
            "description": "Custom E2E role role_upd_1794676846259593115_108",
            "is_system": false,
            "management_domain": "iam",
            "name": "Role role_upd_1794676846259593115_108",
            "redirect_priority": 50
          }
        }
        Result JSON:
        {
          "data": {
            "createRole": {
              "code": "role_upd_1794676846259593115_108",
              "created_at": "2026-11-15T00:20:46.260306346+07:00",
              "default_redirect_path": "/identity",
              "description": "Custom E2E role role_upd_1794676846259593115_108",
              "id": "be536b4f-bdf8-4d2d-8cc6-b8bce54fc858",
              "is_active": true,
              "is_system": false,
              "management_domain": "iam",
              "name": "Role role_upd_1794676846259593115_108",
              "redirect_priority": 50,
              "updated_at": "2026-11-15T00:20:46.260306346+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2750: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UpdateRoleDetailsInput!) { updateRoleDetails(input: $input) { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "default_redirect_path": "/finance",
            "description": "Updated E2E role description",
            "id": "be536b4f-bdf8-4d2d-8cc6-b8bce54fc858",
            "management_domain": "finance",
            "name": "Updated Role r_1794676846261282886_109",
            "redirect_priority": 85
          }
        }
        Result JSON:
        {
          "data": {
            "updateRoleDetails": {
              "code": "role_upd_1794676846259593115_108",
              "created_at": "2026-11-15T00:20:46.260306+07:00",
              "default_redirect_path": "/finance",
              "description": "Updated E2E role description",
              "id": "be536b4f-bdf8-4d2d-8cc6-b8bce54fc858",
              "is_active": true,
              "is_system": false,
              "management_domain": "finance",
              "name": "Updated Role r_1794676846261282886_109",
              "redirect_priority": 85,
              "updated_at": "2026-11-15T00:20:46.262551138+07:00"
            }
          }
        }
=== RUN   TestMutationUpdateRoleDetails/RBACCustomerForbidden
    iam_operations_e2e_test.go:2768: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateRoleInput!) { createRole(input: $input) { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "role_upd_rbac_1794676846437549676_110",
            "default_redirect_path": "/identity",
            "description": "Custom E2E role role_upd_rbac_1794676846437549676_110",
            "is_system": false,
            "management_domain": "iam",
            "name": "Role role_upd_rbac_1794676846437549676_110",
            "redirect_priority": 50
          }
        }
        Result JSON:
        {
          "data": {
            "createRole": {
              "code": "role_upd_rbac_1794676846437549676_110",
              "created_at": "2026-11-15T00:20:46.438285751+07:00",
              "default_redirect_path": "/identity",
              "description": "Custom E2E role role_upd_rbac_1794676846437549676_110",
              "id": "5d80349e-321e-41c0-9c55-38c1b9991bf3",
              "is_active": true,
              "is_system": false,
              "management_domain": "iam",
              "name": "Role role_upd_rbac_1794676846437549676_110",
              "redirect_priority": 50,
              "updated_at": "2026-11-15T00:20:46.438285751+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2769: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateUserInput!) { createUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676846439107187_111.png",
            "display_name": "Fixture u_customer_1794676846439107187_111",
            "email": "u_customer_1794676846439107187_111@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676846439107187_111.png",
              "created_at": "2026-11-15T00:20:46.439622683+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_customer_1794676846439107187_111",
              "id": "7b13cf0d-cb3c-4f83-9aa3-6cf04091e7cc",
              "kind": "CUSTOMER",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:46.439622683+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2769: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_customer_1794676846439107187_111@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "V3Vh7o-iIK_AYfvfgvG1SPpC5mkqpuQFLFI-Tl1rKyA",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:2770: 
        [RBAC_TEST_AB_FORBIDDEN] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UpdateRoleDetailsInput!) { updateRoleDetails(input: $input) { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "default_redirect_path": "/identity",
            "id": "5d80349e-321e-41c0-9c55-38c1b9991bf3",
            "management_domain": "iam",
            "name": "No"
          }
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "missing required permission: iam.roles.write",
              "extensions": {
                "code": "FORBIDDEN"
              }
            }
          ]
        }
--- PASS: TestMutationUpdateRoleDetails (0.54s)
    --- PASS: TestMutationUpdateRoleDetails/UpdateAllMutableFields (0.18s)
    --- PASS: TestMutationUpdateRoleDetails/RBACCustomerForbidden (0.36s)
=== RUN   TestMutationActivateRole
=== RUN   TestMutationActivateRole/ActivateDeactivatedRole
    iam_operations_e2e_test.go:2783: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateRoleInput!) { createRole(input: $input) { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "role_act_r_1794676846798225762_112",
            "default_redirect_path": "/identity",
            "description": "Custom E2E role role_act_r_1794676846798225762_112",
            "is_system": false,
            "management_domain": "iam",
            "name": "Role role_act_r_1794676846798225762_112",
            "redirect_priority": 50
          }
        }
        Result JSON:
        {
          "data": {
            "createRole": {
              "code": "role_act_r_1794676846798225762_112",
              "created_at": "2026-11-15T00:20:46.799085922+07:00",
              "default_redirect_path": "/identity",
              "description": "Custom E2E role role_act_r_1794676846798225762_112",
              "id": "f8dd5e01-d8a3-4124-87f1-a37edb591ddc",
              "is_active": true,
              "is_system": false,
              "management_domain": "iam",
              "name": "Role role_act_r_1794676846798225762_112",
              "redirect_priority": 50,
              "updated_at": "2026-11-15T00:20:46.799085922+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2784: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($id: ID!) { deactivateRole(id: $id) { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } }
        Input Variables:
        {
          "id": "f8dd5e01-d8a3-4124-87f1-a37edb591ddc"
        }
        Result JSON:
        {
          "data": {
            "deactivateRole": {
              "code": "role_act_r_1794676846798225762_112",
              "created_at": "2026-11-15T00:20:46.799085+07:00",
              "default_redirect_path": "/identity",
              "description": "Custom E2E role role_act_r_1794676846798225762_112",
              "id": "f8dd5e01-d8a3-4124-87f1-a37edb591ddc",
              "is_active": false,
              "is_system": false,
              "management_domain": "iam",
              "name": "Role role_act_r_1794676846798225762_112",
              "redirect_priority": 50,
              "updated_at": "2026-11-15T00:20:46.801303373+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2787: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($id: ID!) { activateRole(id: $id) { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } }
        Input Variables:
        {
          "id": "f8dd5e01-d8a3-4124-87f1-a37edb591ddc"
        }
        Result JSON:
        {
          "data": {
            "activateRole": {
              "code": "role_act_r_1794676846798225762_112",
              "created_at": "2026-11-15T00:20:46.799085+07:00",
              "default_redirect_path": "/identity",
              "description": "Custom E2E role role_act_r_1794676846798225762_112",
              "id": "f8dd5e01-d8a3-4124-87f1-a37edb591ddc",
              "is_active": true,
              "is_system": false,
              "management_domain": "iam",
              "name": "Role role_act_r_1794676846798225762_112",
              "redirect_priority": 50,
              "updated_at": "2026-11-15T00:20:46.803307159+07:00"
            }
          }
        }
=== RUN   TestMutationActivateRole/RBACCustomerForbidden
    iam_operations_e2e_test.go:2796: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateRoleInput!) { createRole(input: $input) { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "role_act_rbac_1794676846986270134_113",
            "default_redirect_path": "/identity",
            "description": "Custom E2E role role_act_rbac_1794676846986270134_113",
            "is_system": false,
            "management_domain": "iam",
            "name": "Role role_act_rbac_1794676846986270134_113",
            "redirect_priority": 50
          }
        }
        Result JSON:
        {
          "data": {
            "createRole": {
              "code": "role_act_rbac_1794676846986270134_113",
              "created_at": "2026-11-15T00:20:46.986994497+07:00",
              "default_redirect_path": "/identity",
              "description": "Custom E2E role role_act_rbac_1794676846986270134_113",
              "id": "049c2022-5bf0-4df9-9d48-cef93b3a9b04",
              "is_active": true,
              "is_system": false,
              "management_domain": "iam",
              "name": "Role role_act_rbac_1794676846986270134_113",
              "redirect_priority": 50,
              "updated_at": "2026-11-15T00:20:46.986994497+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2797: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateUserInput!) { createUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676846987756290_114.png",
            "display_name": "Fixture u_customer_1794676846987756290_114",
            "email": "u_customer_1794676846987756290_114@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676846987756290_114.png",
              "created_at": "2026-11-15T00:20:46.988269001+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_customer_1794676846987756290_114",
              "id": "87fce55e-a0e1-4ecb-84e6-3bb2e7310b79",
              "kind": "CUSTOMER",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:46.988269001+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2797: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_customer_1794676846987756290_114@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "AS82YE9OkT8kefFysnsn74wWHevPY9gbdm9Rs9A5W0w",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:2798: 
        [RBAC_TEST_AB_FORBIDDEN] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($id: ID!) { activateRole(id: $id) { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } }
        Input Variables:
        {
          "id": "049c2022-5bf0-4df9-9d48-cef93b3a9b04"
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "missing required permission: iam.roles.write",
              "extensions": {
                "code": "FORBIDDEN"
              }
            }
          ]
        }
--- PASS: TestMutationActivateRole (0.55s)
    --- PASS: TestMutationActivateRole/ActivateDeactivatedRole (0.19s)
    --- PASS: TestMutationActivateRole/RBACCustomerForbidden (0.36s)
=== RUN   TestMutationDeactivateRole
=== RUN   TestMutationDeactivateRole/DeactivateCustomAndProtectSystemRole
    iam_operations_e2e_test.go:2809: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateRoleInput!) { createRole(input: $input) { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "role_deact_r_1794676847341843865_115",
            "default_redirect_path": "/identity",
            "description": "Custom E2E role role_deact_r_1794676847341843865_115",
            "is_system": false,
            "management_domain": "iam",
            "name": "Role role_deact_r_1794676847341843865_115",
            "redirect_priority": 50
          }
        }
        Result JSON:
        {
          "data": {
            "createRole": {
              "code": "role_deact_r_1794676847341843865_115",
              "created_at": "2026-11-15T00:20:47.342716598+07:00",
              "default_redirect_path": "/identity",
              "description": "Custom E2E role role_deact_r_1794676847341843865_115",
              "id": "35b6a46e-1459-496d-8d59-904799f169ed",
              "is_active": true,
              "is_system": false,
              "management_domain": "iam",
              "name": "Role role_deact_r_1794676847341843865_115",
              "redirect_priority": 50,
              "updated_at": "2026-11-15T00:20:47.342716598+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2811: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($id: ID!) { deactivateRole(id: $id) { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } }
        Input Variables:
        {
          "id": "35b6a46e-1459-496d-8d59-904799f169ed"
        }
        Result JSON:
        {
          "data": {
            "deactivateRole": {
              "code": "role_deact_r_1794676847341843865_115",
              "created_at": "2026-11-15T00:20:47.342716+07:00",
              "default_redirect_path": "/identity",
              "description": "Custom E2E role role_deact_r_1794676847341843865_115",
              "id": "35b6a46e-1459-496d-8d59-904799f169ed",
              "is_active": false,
              "is_system": false,
              "management_domain": "iam",
              "name": "Role role_deact_r_1794676847341843865_115",
              "redirect_priority": 50,
              "updated_at": "2026-11-15T00:20:47.344123553+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2817: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query { roleByCode(code: "superadmin") { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } }
        Input Variables:
        {}
        Result JSON:
        {
          "data": {
            "roleByCode": {
              "code": "superadmin",
              "created_at": "2026-11-14T23:13:35.92827+07:00",
              "default_redirect_path": "/identity",
              "description": "IntegriOne full platform and IAM governance super administrator",
              "id": "d8ffac94-b437-4aa3-a937-27556b86d2dd",
              "is_active": true,
              "is_system": true,
              "management_domain": "iam",
              "name": "Super Administrator",
              "redirect_priority": 1,
              "updated_at": "2026-11-14T23:13:35.92827+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2819: 
        [GQL_EXPECTED_ERROR] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($id: ID!) { deactivateRole(id: $id) { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } }
        Input Variables:
        {
          "id": "d8ffac94-b437-4aa3-a937-27556b86d2dd"
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "system role is immutable",
              "extensions": {
                "code": "CONFLICT"
              }
            }
          ]
        }
=== RUN   TestMutationDeactivateRole/RBACCustomerForbidden
    iam_operations_e2e_test.go:2824: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateRoleInput!) { createRole(input: $input) { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "role_deact_rbac_1794676847526983454_116",
            "default_redirect_path": "/identity",
            "description": "Custom E2E role role_deact_rbac_1794676847526983454_116",
            "is_system": false,
            "management_domain": "iam",
            "name": "Role role_deact_rbac_1794676847526983454_116",
            "redirect_priority": 50
          }
        }
        Result JSON:
        {
          "data": {
            "createRole": {
              "code": "role_deact_rbac_1794676847526983454_116",
              "created_at": "2026-11-15T00:20:47.527717675+07:00",
              "default_redirect_path": "/identity",
              "description": "Custom E2E role role_deact_rbac_1794676847526983454_116",
              "id": "ec434f95-faaf-41ad-b205-b5f04f771d3c",
              "is_active": true,
              "is_system": false,
              "management_domain": "iam",
              "name": "Role role_deact_rbac_1794676847526983454_116",
              "redirect_priority": 50,
              "updated_at": "2026-11-15T00:20:47.527717675+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2825: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateUserInput!) { createUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676847528488425_117.png",
            "display_name": "Fixture u_customer_1794676847528488425_117",
            "email": "u_customer_1794676847528488425_117@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676847528488425_117.png",
              "created_at": "2026-11-15T00:20:47.528941964+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_customer_1794676847528488425_117",
              "id": "b6d47268-fee8-44dd-850c-9803d7740807",
              "kind": "CUSTOMER",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:47.528941964+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2825: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_customer_1794676847528488425_117@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "UWK_U-RD4u_JBI-NqSk90OrcpujBZo3F5AZSgu0JVlw",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:2826: 
        [RBAC_TEST_AB_FORBIDDEN] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($id: ID!) { deactivateRole(id: $id) { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } }
        Input Variables:
        {
          "id": "ec434f95-faaf-41ad-b205-b5f04f771d3c"
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "missing required permission: iam.roles.write",
              "extensions": {
                "code": "FORBIDDEN"
              }
            }
          ]
        }
--- PASS: TestMutationDeactivateRole (0.53s)
    --- PASS: TestMutationDeactivateRole/DeactivateCustomAndProtectSystemRole (0.18s)
    --- PASS: TestMutationDeactivateRole/RBACCustomerForbidden (0.35s)
=== RUN   TestMutationDeleteRole
=== RUN   TestMutationDeleteRole/DeleteCustomAndProtectSystemRole
    iam_operations_e2e_test.go:2837: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateRoleInput!) { createRole(input: $input) { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "role_del_r_1794676847878823491_118",
            "default_redirect_path": "/identity",
            "description": "Custom E2E role role_del_r_1794676847878823491_118",
            "is_system": false,
            "management_domain": "iam",
            "name": "Role role_del_r_1794676847878823491_118",
            "redirect_priority": 50
          }
        }
        Result JSON:
        {
          "data": {
            "createRole": {
              "code": "role_del_r_1794676847878823491_118",
              "created_at": "2026-11-15T00:20:47.879520421+07:00",
              "default_redirect_path": "/identity",
              "description": "Custom E2E role role_del_r_1794676847878823491_118",
              "id": "85374ea3-b99d-44a3-bee7-38236c51837d",
              "is_active": true,
              "is_system": false,
              "management_domain": "iam",
              "name": "Role role_del_r_1794676847878823491_118",
              "redirect_priority": 50,
              "updated_at": "2026-11-15T00:20:47.879520421+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2839: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($id: ID!) { deleteRole(id: $id) }
        Input Variables:
        {
          "id": "85374ea3-b99d-44a3-bee7-38236c51837d"
        }
        Result JSON:
        {
          "data": {
            "deleteRole": true
          }
        }
    iam_operations_e2e_test.go:2840: 
        [GQL_EXPECTED_ERROR] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($id: ID!) { role(id: $id) { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } }
        Input Variables:
        {
          "id": "85374ea3-b99d-44a3-bee7-38236c51837d"
        }
        Result JSON:
        {
          "data": {
            "role": null
          },
          "errors": [
            {
              "message": "role not found",
              "extensions": {
                "code": "NOT_FOUND"
              }
            }
          ]
        }
    iam_operations_e2e_test.go:2842: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query { roleByCode(code: "superadmin") { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } }
        Input Variables:
        {}
        Result JSON:
        {
          "data": {
            "roleByCode": {
              "code": "superadmin",
              "created_at": "2026-11-14T23:13:35.92827+07:00",
              "default_redirect_path": "/identity",
              "description": "IntegriOne full platform and IAM governance super administrator",
              "id": "d8ffac94-b437-4aa3-a937-27556b86d2dd",
              "is_active": true,
              "is_system": true,
              "management_domain": "iam",
              "name": "Super Administrator",
              "redirect_priority": 1,
              "updated_at": "2026-11-14T23:13:35.92827+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2844: 
        [GQL_EXPECTED_ERROR] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($id: ID!) { deleteRole(id: $id) }
        Input Variables:
        {
          "id": "d8ffac94-b437-4aa3-a937-27556b86d2dd"
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "system role is immutable",
              "extensions": {
                "code": "CONFLICT"
              }
            }
          ]
        }
=== RUN   TestMutationDeleteRole/RBACCustomerForbidden
    iam_operations_e2e_test.go:2849: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateRoleInput!) { createRole(input: $input) { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "role_del_rbac_1794676848067658561_119",
            "default_redirect_path": "/identity",
            "description": "Custom E2E role role_del_rbac_1794676848067658561_119",
            "is_system": false,
            "management_domain": "iam",
            "name": "Role role_del_rbac_1794676848067658561_119",
            "redirect_priority": 50
          }
        }
        Result JSON:
        {
          "data": {
            "createRole": {
              "code": "role_del_rbac_1794676848067658561_119",
              "created_at": "2026-11-15T00:20:48.068763464+07:00",
              "default_redirect_path": "/identity",
              "description": "Custom E2E role role_del_rbac_1794676848067658561_119",
              "id": "d4b08fcb-a8e7-48bc-8957-573922737e87",
              "is_active": true,
              "is_system": false,
              "management_domain": "iam",
              "name": "Role role_del_rbac_1794676848067658561_119",
              "redirect_priority": 50,
              "updated_at": "2026-11-15T00:20:48.068763464+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2850: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateUserInput!) { createUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676848069633262_120.png",
            "display_name": "Fixture u_customer_1794676848069633262_120",
            "email": "u_customer_1794676848069633262_120@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676848069633262_120.png",
              "created_at": "2026-11-15T00:20:48.070177794+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_customer_1794676848069633262_120",
              "id": "5058b7a5-bd66-4413-859a-0f388246dfd2",
              "kind": "CUSTOMER",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:48.070177794+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2850: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_customer_1794676848069633262_120@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "u7sBdiEO0ORO9LRu0vp_3jMVpvmlYBfK1tQEhm6lHyA",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:2851: 
        [RBAC_TEST_AB_FORBIDDEN] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($id: ID!) { deleteRole(id: $id) }
        Input Variables:
        {
          "id": "d4b08fcb-a8e7-48bc-8957-573922737e87"
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "missing required permission: iam.roles.write",
              "extensions": {
                "code": "FORBIDDEN"
              }
            }
          ]
        }
--- PASS: TestMutationDeleteRole (0.54s)
    --- PASS: TestMutationDeleteRole/DeleteCustomAndProtectSystemRole (0.19s)
    --- PASS: TestMutationDeleteRole/RBACCustomerForbidden (0.35s)
=== RUN   TestMutationAssignRolePermission
=== RUN   TestMutationAssignRolePermission/AssignPermissionReturnsNonNull
    iam_operations_e2e_test.go:2864: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateRoleInput!) { createRole(input: $input) { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "role_arp_1794676848420240213_121",
            "default_redirect_path": "/identity",
            "description": "Custom E2E role role_arp_1794676848420240213_121",
            "is_system": false,
            "management_domain": "iam",
            "name": "Role role_arp_1794676848420240213_121",
            "redirect_priority": 50
          }
        }
        Result JSON:
        {
          "data": {
            "createRole": {
              "code": "role_arp_1794676848420240213_121",
              "created_at": "2026-11-15T00:20:48.420948415+07:00",
              "default_redirect_path": "/identity",
              "description": "Custom E2E role role_arp_1794676848420240213_121",
              "id": "759f2387-3bbc-4fec-ada5-2574867d351f",
              "is_active": true,
              "is_system": false,
              "management_domain": "iam",
              "name": "Role role_arp_1794676848420240213_121",
              "redirect_priority": 50,
              "updated_at": "2026-11-15T00:20:48.420948415+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2865: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($query: String) { permissions(query: $query, page: { limit: 25, offset: 0 }) { items { id code resource action description is_sensitive created_at } page { total limit offset has_more } } }
        Input Variables:
        {
          "query": "iam.users.read"
        }
        Result JSON:
        {
          "data": {
            "permissions": {
              "items": [
                {
                  "action": "read",
                  "code": "iam.users.read",
                  "created_at": "2026-11-14T23:13:35.92544+07:00",
                  "description": "Read IAM user principals and identities",
                  "id": "92f3f3db-af8f-49d8-85f9-7287262abbdf",
                  "is_sensitive": false,
                  "resource": "users"
                }
              ],
              "page": {
                "has_more": false,
                "limit": 25,
                "offset": 0,
                "total": 1
              }
            }
          }
        }
    iam_operations_e2e_test.go:2867: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($role_id: ID!, $permission_id: ID!) { assignRolePermission(role_id: $role_id, permission_id: $permission_id) { role_id permission_id created_at } }
        Input Variables:
        {
          "permission_id": "92f3f3db-af8f-49d8-85f9-7287262abbdf",
          "role_id": "759f2387-3bbc-4fec-ada5-2574867d351f"
        }
        Result JSON:
        {
          "data": {
            "assignRolePermission": {
              "created_at": "2026-11-15T00:20:48.423394298+07:00",
              "permission_id": "92f3f3db-af8f-49d8-85f9-7287262abbdf",
              "role_id": "759f2387-3bbc-4fec-ada5-2574867d351f"
            }
          }
        }
=== RUN   TestMutationAssignRolePermission/RBACCustomerForbidden
    iam_operations_e2e_test.go:2876: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateRoleInput!) { createRole(input: $input) { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "role_arp_rbac_1794676848605510730_122",
            "default_redirect_path": "/identity",
            "description": "Custom E2E role role_arp_rbac_1794676848605510730_122",
            "is_system": false,
            "management_domain": "iam",
            "name": "Role role_arp_rbac_1794676848605510730_122",
            "redirect_priority": 50
          }
        }
        Result JSON:
        {
          "data": {
            "createRole": {
              "code": "role_arp_rbac_1794676848605510730_122",
              "created_at": "2026-11-15T00:20:48.606236966+07:00",
              "default_redirect_path": "/identity",
              "description": "Custom E2E role role_arp_rbac_1794676848605510730_122",
              "id": "5bc66400-56ff-4b07-b392-20994e5af64a",
              "is_active": true,
              "is_system": false,
              "management_domain": "iam",
              "name": "Role role_arp_rbac_1794676848605510730_122",
              "redirect_priority": 50,
              "updated_at": "2026-11-15T00:20:48.606236966+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2877: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($query: String) { permissions(query: $query, page: { limit: 25, offset: 0 }) { items { id code resource action description is_sensitive created_at } page { total limit offset has_more } } }
        Input Variables:
        {
          "query": "iam.users.read"
        }
        Result JSON:
        {
          "data": {
            "permissions": {
              "items": [
                {
                  "action": "read",
                  "code": "iam.users.read",
                  "created_at": "2026-11-14T23:13:35.92544+07:00",
                  "description": "Read IAM user principals and identities",
                  "id": "92f3f3db-af8f-49d8-85f9-7287262abbdf",
                  "is_sensitive": false,
                  "resource": "users"
                }
              ],
              "page": {
                "has_more": false,
                "limit": 25,
                "offset": 0,
                "total": 1
              }
            }
          }
        }
    iam_operations_e2e_test.go:2878: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateUserInput!) { createUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676848608267573_123.png",
            "display_name": "Fixture u_customer_1794676848608267573_123",
            "email": "u_customer_1794676848608267573_123@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676848608267573_123.png",
              "created_at": "2026-11-15T00:20:48.608767831+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_customer_1794676848608267573_123",
              "id": "6e858695-b3c1-488a-9fd2-3fb6a273de87",
              "kind": "CUSTOMER",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:48.608767831+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2878: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_customer_1794676848608267573_123@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "Knxow3fCyYueDFmXlTIuWfYEpv7vO7A3_py_QYYA_9I",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:2879: 
        [RBAC_TEST_AB_FORBIDDEN] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($role_id: ID!, $permission_id: ID!) { assignRolePermission(role_id: $role_id, permission_id: $permission_id) { role_id permission_id created_at } }
        Input Variables:
        {
          "permission_id": "92f3f3db-af8f-49d8-85f9-7287262abbdf",
          "role_id": "5bc66400-56ff-4b07-b392-20994e5af64a"
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "missing required permission: iam.roles.write",
              "extensions": {
                "code": "FORBIDDEN"
              }
            }
          ]
        }
--- PASS: TestMutationAssignRolePermission (0.54s)
    --- PASS: TestMutationAssignRolePermission/AssignPermissionReturnsNonNull (0.19s)
    --- PASS: TestMutationAssignRolePermission/RBACCustomerForbidden (0.35s)
=== RUN   TestMutationRevokeRolePermission
=== RUN   TestMutationRevokeRolePermission/RevokeAssignedPermissionReturnsTrue
    iam_operations_e2e_test.go:2890: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateRoleInput!) { createRole(input: $input) { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "role_rrp_1794676848958262354_124",
            "default_redirect_path": "/identity",
            "description": "Custom E2E role role_rrp_1794676848958262354_124",
            "is_system": false,
            "management_domain": "iam",
            "name": "Role role_rrp_1794676848958262354_124",
            "redirect_priority": 50
          }
        }
        Result JSON:
        {
          "data": {
            "createRole": {
              "code": "role_rrp_1794676848958262354_124",
              "created_at": "2026-11-15T00:20:48.959069023+07:00",
              "default_redirect_path": "/identity",
              "description": "Custom E2E role role_rrp_1794676848958262354_124",
              "id": "805f1885-5a4a-43d9-abb7-1e8d04c7fe9f",
              "is_active": true,
              "is_system": false,
              "management_domain": "iam",
              "name": "Role role_rrp_1794676848958262354_124",
              "redirect_priority": 50,
              "updated_at": "2026-11-15T00:20:48.959069023+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2891: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($query: String) { permissions(query: $query, page: { limit: 25, offset: 0 }) { items { id code resource action description is_sensitive created_at } page { total limit offset has_more } } }
        Input Variables:
        {
          "query": "iam.users.read"
        }
        Result JSON:
        {
          "data": {
            "permissions": {
              "items": [
                {
                  "action": "read",
                  "code": "iam.users.read",
                  "created_at": "2026-11-14T23:13:35.92544+07:00",
                  "description": "Read IAM user principals and identities",
                  "id": "92f3f3db-af8f-49d8-85f9-7287262abbdf",
                  "is_sensitive": false,
                  "resource": "users"
                }
              ],
              "page": {
                "has_more": false,
                "limit": 25,
                "offset": 0,
                "total": 1
              }
            }
          }
        }
    iam_operations_e2e_test.go:2892: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($r: ID!, $p: ID!) { assignRolePermission(role_id: $r, permission_id: $p) { role_id permission_id created_at } }
        Input Variables:
        {
          "p": "92f3f3db-af8f-49d8-85f9-7287262abbdf",
          "r": "805f1885-5a4a-43d9-abb7-1e8d04c7fe9f"
        }
        Result JSON:
        {
          "data": {
            "assignRolePermission": {
              "created_at": "2026-11-15T00:20:48.96158547+07:00",
              "permission_id": "92f3f3db-af8f-49d8-85f9-7287262abbdf",
              "role_id": "805f1885-5a4a-43d9-abb7-1e8d04c7fe9f"
            }
          }
        }
    iam_operations_e2e_test.go:2897: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($role_id: ID!, $permission_id: ID!) { revokeRolePermission(role_id: $role_id, permission_id: $permission_id) }
        Input Variables:
        {
          "permission_id": "92f3f3db-af8f-49d8-85f9-7287262abbdf",
          "role_id": "805f1885-5a4a-43d9-abb7-1e8d04c7fe9f"
        }
        Result JSON:
        {
          "data": {
            "revokeRolePermission": true
          }
        }
    iam_operations_e2e_test.go:2898: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($r: ID!) { rolePermissions(role_id: $r) { id code resource action description is_sensitive created_at } }
        Input Variables:
        {
          "r": "805f1885-5a4a-43d9-abb7-1e8d04c7fe9f"
        }
        Result JSON:
        {
          "data": {
            "rolePermissions": []
          }
        }
=== RUN   TestMutationRevokeRolePermission/RBACCustomerForbidden
    iam_operations_e2e_test.go:2906: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateRoleInput!) { createRole(input: $input) { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "role_rrp_rbac_1794676849148794990_125",
            "default_redirect_path": "/identity",
            "description": "Custom E2E role role_rrp_rbac_1794676849148794990_125",
            "is_system": false,
            "management_domain": "iam",
            "name": "Role role_rrp_rbac_1794676849148794990_125",
            "redirect_priority": 50
          }
        }
        Result JSON:
        {
          "data": {
            "createRole": {
              "code": "role_rrp_rbac_1794676849148794990_125",
              "created_at": "2026-11-15T00:20:49.149543087+07:00",
              "default_redirect_path": "/identity",
              "description": "Custom E2E role role_rrp_rbac_1794676849148794990_125",
              "id": "59e6566b-f4cd-4205-b08c-6eee5e22cacc",
              "is_active": true,
              "is_system": false,
              "management_domain": "iam",
              "name": "Role role_rrp_rbac_1794676849148794990_125",
              "redirect_priority": 50,
              "updated_at": "2026-11-15T00:20:49.149543087+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2907: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($query: String) { permissions(query: $query, page: { limit: 25, offset: 0 }) { items { id code resource action description is_sensitive created_at } page { total limit offset has_more } } }
        Input Variables:
        {
          "query": "iam.users.read"
        }
        Result JSON:
        {
          "data": {
            "permissions": {
              "items": [
                {
                  "action": "read",
                  "code": "iam.users.read",
                  "created_at": "2026-11-14T23:13:35.92544+07:00",
                  "description": "Read IAM user principals and identities",
                  "id": "92f3f3db-af8f-49d8-85f9-7287262abbdf",
                  "is_sensitive": false,
                  "resource": "users"
                }
              ],
              "page": {
                "has_more": false,
                "limit": 25,
                "offset": 0,
                "total": 1
              }
            }
          }
        }
    iam_operations_e2e_test.go:2908: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateUserInput!) { createUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676849152278980_126.png",
            "display_name": "Fixture u_customer_1794676849152278980_126",
            "email": "u_customer_1794676849152278980_126@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676849152278980_126.png",
              "created_at": "2026-11-15T00:20:49.152825756+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_customer_1794676849152278980_126",
              "id": "878bb081-c435-40f9-b110-045e96e93ede",
              "kind": "CUSTOMER",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:49.152825756+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2908: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_customer_1794676849152278980_126@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "-jOAJlcHe8DAomLSpOsXhPh6lvrAUscmYYpRx7uOaJw",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:2909: 
        [RBAC_TEST_AB_FORBIDDEN] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($role_id: ID!, $permission_id: ID!) { revokeRolePermission(role_id: $role_id, permission_id: $permission_id) }
        Input Variables:
        {
          "permission_id": "92f3f3db-af8f-49d8-85f9-7287262abbdf",
          "role_id": "59e6566b-f4cd-4205-b08c-6eee5e22cacc"
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "missing required permission: iam.roles.write",
              "extensions": {
                "code": "FORBIDDEN"
              }
            }
          ]
        }
--- PASS: TestMutationRevokeRolePermission (0.56s)
    --- PASS: TestMutationRevokeRolePermission/RevokeAssignedPermissionReturnsTrue (0.19s)
    --- PASS: TestMutationRevokeRolePermission/RBACCustomerForbidden (0.37s)
=== RUN   TestMutationAssignUserRole
=== RUN   TestMutationAssignUserRole/AssignGlobalAndScopedRole
    iam_operations_e2e_test.go:2922: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateEmployeeUserInput!) { createEmployeeUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676849514616224_127.png",
            "display_name": "Fixture u_employee_1794676849514616224_127",
            "email": "u_employee_1794676849514616224_127@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createEmployeeUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676849514616224_127.png",
              "created_at": "2026-11-15T00:20:49.515294689+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_employee_1794676849514616224_127",
              "id": "33171ef8-62a0-47f4-906d-69aefd6a402c",
              "kind": "EMPLOYEE",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:49.515294689+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2923: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateRoleInput!) { createRole(input: $input) { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "role_aur_g_1794676849562135725_128",
            "default_redirect_path": "/identity",
            "description": "Custom E2E role role_aur_g_1794676849562135725_128",
            "is_system": false,
            "management_domain": "iam",
            "name": "Role role_aur_g_1794676849562135725_128",
            "redirect_priority": 50
          }
        }
        Result JSON:
        {
          "data": {
            "createRole": {
              "code": "role_aur_g_1794676849562135725_128",
              "created_at": "2026-11-15T00:20:49.563017295+07:00",
              "default_redirect_path": "/identity",
              "description": "Custom E2E role role_aur_g_1794676849562135725_128",
              "id": "f341484b-6f91-4824-849f-c12e367967d2",
              "is_active": true,
              "is_system": false,
              "management_domain": "iam",
              "name": "Role role_aur_g_1794676849562135725_128",
              "redirect_priority": 50,
              "updated_at": "2026-11-15T00:20:49.563017295+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2924: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateRoleInput!) { createRole(input: $input) { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "role_aur_f_1794676849564274226_129",
            "default_redirect_path": "/identity",
            "description": "Custom E2E role role_aur_f_1794676849564274226_129",
            "is_system": false,
            "management_domain": "iam",
            "name": "Role role_aur_f_1794676849564274226_129",
            "redirect_priority": 50
          }
        }
        Result JSON:
        {
          "data": {
            "createRole": {
              "code": "role_aur_f_1794676849564274226_129",
              "created_at": "2026-11-15T00:20:49.565037452+07:00",
              "default_redirect_path": "/identity",
              "description": "Custom E2E role role_aur_f_1794676849564274226_129",
              "id": "ea8e2938-cbc8-4b91-8197-81f4d081139f",
              "is_active": true,
              "is_system": false,
              "management_domain": "iam",
              "name": "Role role_aur_f_1794676849564274226_129",
              "redirect_priority": 50,
              "updated_at": "2026-11-15T00:20:49.565037452+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2926: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: AssignUserRoleInput!) { assignUserRole(input: $input) { id user_id role_id scope { type id } granted_by expires_at created_at } }
        Input Variables:
        {
          "input": {
            "role_id": "f341484b-6f91-4824-849f-c12e367967d2",
            "scope_type": "GLOBAL",
            "user_id": "33171ef8-62a0-47f4-906d-69aefd6a402c"
          }
        }
        Result JSON:
        {
          "data": {
            "assignUserRole": {
              "created_at": "2026-11-15T00:20:49.56681038+07:00",
              "expires_at": null,
              "granted_by": "853aac77-32e7-4a88-b2d6-7d68db8d13d7",
              "id": "09705125-b505-49d4-821c-d2ba7437b8a1",
              "role_id": "f341484b-6f91-4824-849f-c12e367967d2",
              "scope": {
                "id": null,
                "type": "GLOBAL"
              },
              "user_id": "33171ef8-62a0-47f4-906d-69aefd6a402c"
            }
          }
        }
    iam_operations_e2e_test.go:2933: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: AssignUserRoleInput!) { assignUserRole(input: $input) { id user_id role_id scope { type id } granted_by expires_at created_at } }
        Input Variables:
        {
          "input": {
            "expires_at": "2026-11-14T21:20:49Z",
            "role_id": "ea8e2938-cbc8-4b91-8197-81f4d081139f",
            "scope_id": "33333333-3333-4333-8333-333333333333",
            "scope_type": "FACILITY",
            "user_id": "33171ef8-62a0-47f4-906d-69aefd6a402c"
          }
        }
        Result JSON:
        {
          "data": {
            "assignUserRole": {
              "created_at": "2026-11-15T00:20:49.568880562+07:00",
              "expires_at": "2026-11-15T04:20:49+07:00",
              "granted_by": "853aac77-32e7-4a88-b2d6-7d68db8d13d7",
              "id": "afa2713c-ff28-4138-a6ad-d0707bb814c7",
              "role_id": "ea8e2938-cbc8-4b91-8197-81f4d081139f",
              "scope": {
                "id": "33333333-3333-4333-8333-333333333333",
                "type": "FACILITY"
              },
              "user_id": "33171ef8-62a0-47f4-906d-69aefd6a402c"
            }
          }
        }
=== RUN   TestMutationAssignUserRole/RBACCustomerForbidden
    iam_operations_e2e_test.go:2953: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateRoleInput!) { createRole(input: $input) { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "role_aur_rbac_1794676849938290764_130",
            "default_redirect_path": "/identity",
            "description": "Custom E2E role role_aur_rbac_1794676849938290764_130",
            "is_system": false,
            "management_domain": "iam",
            "name": "Role role_aur_rbac_1794676849938290764_130",
            "redirect_priority": 50
          }
        }
        Result JSON:
        {
          "data": {
            "createRole": {
              "code": "role_aur_rbac_1794676849938290764_130",
              "created_at": "2026-11-15T00:20:49.939114735+07:00",
              "default_redirect_path": "/identity",
              "description": "Custom E2E role role_aur_rbac_1794676849938290764_130",
              "id": "5fc96122-224d-42d1-afea-35635c6bcf78",
              "is_active": true,
              "is_system": false,
              "management_domain": "iam",
              "name": "Role role_aur_rbac_1794676849938290764_130",
              "redirect_priority": 50,
              "updated_at": "2026-11-15T00:20:49.939114735+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2954: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateUserInput!) { createUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676849940258462_131.png",
            "display_name": "Fixture u_customer_1794676849940258462_131",
            "email": "u_customer_1794676849940258462_131@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676849940258462_131.png",
              "created_at": "2026-11-15T00:20:49.94099113+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_customer_1794676849940258462_131",
              "id": "7149460c-72bc-4f45-abf0-0ff6a760534c",
              "kind": "CUSTOMER",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:49.94099113+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2954: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_customer_1794676849940258462_131@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "kIOOXgTw_oDGaKfEBWZC369J0mPwOAz52S6fkgvDkGE",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:2955: 
        [RBAC_TEST_AB_FORBIDDEN] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: AssignUserRoleInput!) { assignUserRole(input: $input) { id user_id role_id scope { type id } granted_by expires_at created_at } }
        Input Variables:
        {
          "input": {
            "role_id": "5fc96122-224d-42d1-afea-35635c6bcf78",
            "scope_type": "GLOBAL",
            "user_id": "7149460c-72bc-4f45-abf0-0ff6a760534c"
          }
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "missing required permission: iam.roles.write",
              "extensions": {
                "code": "FORBIDDEN"
              }
            }
          ]
        }
--- PASS: TestMutationAssignUserRole (0.78s)
    --- PASS: TestMutationAssignUserRole/AssignGlobalAndScopedRole (0.42s)
    --- PASS: TestMutationAssignUserRole/RBACCustomerForbidden (0.36s)
=== RUN   TestMutationRevokeUserRole
=== RUN   TestMutationRevokeUserRole/RevokeReturnsTrue
    iam_operations_e2e_test.go:2968: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateEmployeeUserInput!) { createEmployeeUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676850294107545_132.png",
            "display_name": "Fixture u_employee_1794676850294107545_132",
            "email": "u_employee_1794676850294107545_132@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createEmployeeUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676850294107545_132.png",
              "created_at": "2026-11-15T00:20:50.294840223+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_employee_1794676850294107545_132",
              "id": "4621366c-9398-449d-aed1-ad16194df2a8",
              "kind": "EMPLOYEE",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:50.294840223+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2969: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateRoleInput!) { createRole(input: $input) { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "role_rur_1794676850338736290_133",
            "default_redirect_path": "/identity",
            "description": "Custom E2E role role_rur_1794676850338736290_133",
            "is_system": false,
            "management_domain": "iam",
            "name": "Role role_rur_1794676850338736290_133",
            "redirect_priority": 50
          }
        }
        Result JSON:
        {
          "data": {
            "createRole": {
              "code": "role_rur_1794676850338736290_133",
              "created_at": "2026-11-15T00:20:50.339692642+07:00",
              "default_redirect_path": "/identity",
              "description": "Custom E2E role role_rur_1794676850338736290_133",
              "id": "9dda76b0-7a68-4757-a0c2-4abd18eb51ce",
              "is_active": true,
              "is_system": false,
              "management_domain": "iam",
              "name": "Role role_rur_1794676850338736290_133",
              "redirect_priority": 50,
              "updated_at": "2026-11-15T00:20:50.339692642+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2970: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($in: AssignUserRoleInput!) { assignUserRole(input: $in) { id user_id role_id scope { type id } granted_by expires_at created_at } }
        Input Variables:
        {
          "in": {
            "role_id": "9dda76b0-7a68-4757-a0c2-4abd18eb51ce",
            "scope_type": "GLOBAL",
            "user_id": "4621366c-9398-449d-aed1-ad16194df2a8"
          }
        }
        Result JSON:
        {
          "data": {
            "assignUserRole": {
              "created_at": "2026-11-15T00:20:50.342417383+07:00",
              "expires_at": null,
              "granted_by": "853aac77-32e7-4a88-b2d6-7d68db8d13d7",
              "id": "3409b2dd-c745-4dd8-b4e4-84006d6df957",
              "role_id": "9dda76b0-7a68-4757-a0c2-4abd18eb51ce",
              "scope": {
                "id": null,
                "type": "GLOBAL"
              },
              "user_id": "4621366c-9398-449d-aed1-ad16194df2a8"
            }
          }
        }
    iam_operations_e2e_test.go:2976: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($assignment_id: ID!) { revokeUserRole(assignment_id: $assignment_id) }
        Input Variables:
        {
          "assignment_id": "3409b2dd-c745-4dd8-b4e4-84006d6df957"
        }
        Result JSON:
        {
          "data": {
            "revokeUserRole": true
          }
        }
    iam_operations_e2e_test.go:2977: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($u: ID!) { userRoles(user_id: $u) { id user_id role_id scope { type id } granted_by expires_at created_at } }
        Input Variables:
        {
          "u": "4621366c-9398-449d-aed1-ad16194df2a8"
        }
        Result JSON:
        {
          "data": {
            "userRoles": []
          }
        }
=== RUN   TestMutationRevokeUserRole/RBACCustomerForbidden
    iam_operations_e2e_test.go:2985: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateUserInput!) { createUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676850607958185_134.png",
            "display_name": "Fixture u_customer_1794676850607958185_134",
            "email": "u_customer_1794676850607958185_134@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676850607958185_134.png",
              "created_at": "2026-11-15T00:20:50.608694069+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_customer_1794676850607958185_134",
              "id": "f877b517-5c2e-4d81-982f-22bf9bcdd33b",
              "kind": "CUSTOMER",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:50.608694069+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2985: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_customer_1794676850607958185_134@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "DGFCRVsntfrY5HnBmh6GrN5vceUld66FNgPs1lAKqSU",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:2986: 
        [RBAC_TEST_AB_FORBIDDEN] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($assignment_id: ID!) { revokeUserRole(assignment_id: $assignment_id) }
        Input Variables:
        {
          "assignment_id": "00000000-0000-4000-8000-000000000001"
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "missing required permission: iam.roles.write",
              "extensions": {
                "code": "FORBIDDEN"
              }
            }
          ]
        }
--- PASS: TestMutationRevokeUserRole (0.55s)
    --- PASS: TestMutationRevokeUserRole/RevokeReturnsTrue (0.32s)
    --- PASS: TestMutationRevokeUserRole/RBACCustomerForbidden (0.23s)
=== RUN   TestMutationCreateGroup
=== RUN   TestMutationCreateGroup/CreateValidAndDuplicateConflict
    iam_operations_e2e_test.go:2997: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateGroupInput!) { createGroup(input: $input) { id code name description is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "grp_cg_1794676850842227686_135",
            "description": "Group E2E grp_cg_1794676850842227686_135",
            "name": "Group grp_cg_1794676850842227686_135"
          }
        }
        Result JSON:
        {
          "data": {
            "createGroup": {
              "code": "grp_cg_1794676850842227686_135",
              "created_at": "2026-11-15T00:20:50.842915188+07:00",
              "description": "Group E2E grp_cg_1794676850842227686_135",
              "id": "528c54ea-ce74-4492-99e6-015439c117c9",
              "is_active": true,
              "name": "Group grp_cg_1794676850842227686_135",
              "updated_at": "2026-11-15T00:20:50.842915188+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:2999: 
        [GQL_EXPECTED_ERROR] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateGroupInput!) { createGroup(input: $input) { id code name description is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "grp_cg_1794676850842227686_135",
            "description": "dup",
            "name": "Duplicate"
          }
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "group code already exists",
              "extensions": {
                "code": "CONFLICT"
              }
            }
          ]
        }
=== RUN   TestMutationCreateGroup/RBACCustomerForbidden
    iam_operations_e2e_test.go:3006: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateUserInput!) { createUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676850938096450_136.png",
            "display_name": "Fixture u_customer_1794676850938096450_136",
            "email": "u_customer_1794676850938096450_136@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676850938096450_136.png",
              "created_at": "2026-11-15T00:20:50.938833597+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_customer_1794676850938096450_136",
              "id": "ba1f844a-4b4b-407d-b01f-3c76c01a9628",
              "kind": "CUSTOMER",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:50.938833597+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:3006: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_customer_1794676850938096450_136@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "Way_i6RlAvE7QgoZigcfjpmyTlp1mwUPqQzDgpSmJBk",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:3007: 
        [RBAC_TEST_AB_FORBIDDEN] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateGroupInput!) { createGroup(input: $input) { id code name description is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "no_g_1794676851034555883_137",
            "name": "No"
          }
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "missing required permission: iam.groups.write",
              "extensions": {
                "code": "FORBIDDEN"
              }
            }
          ]
        }
--- PASS: TestMutationCreateGroup (0.34s)
    --- PASS: TestMutationCreateGroup/CreateValidAndDuplicateConflict (0.10s)
    --- PASS: TestMutationCreateGroup/RBACCustomerForbidden (0.24s)
=== RUN   TestMutationUpdateGroupDetails
=== RUN   TestMutationUpdateGroupDetails/UpdateNameAndDescription
    iam_operations_e2e_test.go:3020: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateGroupInput!) { createGroup(input: $input) { id code name description is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "grp_ugd_1794676851179534361_138",
            "description": "Group E2E grp_ugd_1794676851179534361_138",
            "name": "Group grp_ugd_1794676851179534361_138"
          }
        }
        Result JSON:
        {
          "data": {
            "createGroup": {
              "code": "grp_ugd_1794676851179534361_138",
              "created_at": "2026-11-15T00:20:51.18025152+07:00",
              "description": "Group E2E grp_ugd_1794676851179534361_138",
              "id": "79bc1f4c-85a5-4000-aca2-4dd93334f970",
              "is_active": true,
              "name": "Group grp_ugd_1794676851179534361_138",
              "updated_at": "2026-11-15T00:20:51.18025152+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:3023: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UpdateGroupDetailsInput!) { updateGroupDetails(input: $input) { id code name description is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "description": "Updated E2E desc",
            "id": "79bc1f4c-85a5-4000-aca2-4dd93334f970",
            "name": "Updated Group g_1794676851181006700_139"
          }
        }
        Result JSON:
        {
          "data": {
            "updateGroupDetails": {
              "code": "grp_ugd_1794676851179534361_138",
              "created_at": "2026-11-15T00:20:51.180251+07:00",
              "description": "Updated E2E desc",
              "id": "79bc1f4c-85a5-4000-aca2-4dd93334f970",
              "is_active": true,
              "name": "Updated Group g_1794676851181006700_139",
              "updated_at": "2026-11-15T00:20:51.18163979+07:00"
            }
          }
        }
=== RUN   TestMutationUpdateGroupDetails/RBACCustomerForbidden
    iam_operations_e2e_test.go:3034: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateGroupInput!) { createGroup(input: $input) { id code name description is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "grp_ugd_rbac_1794676851275997201_140",
            "description": "Group E2E grp_ugd_rbac_1794676851275997201_140",
            "name": "Group grp_ugd_rbac_1794676851275997201_140"
          }
        }
        Result JSON:
        {
          "data": {
            "createGroup": {
              "code": "grp_ugd_rbac_1794676851275997201_140",
              "created_at": "2026-11-15T00:20:51.276740469+07:00",
              "description": "Group E2E grp_ugd_rbac_1794676851275997201_140",
              "id": "f254a3ae-0992-4dad-a56c-acf8b15904ce",
              "is_active": true,
              "name": "Group grp_ugd_rbac_1794676851275997201_140",
              "updated_at": "2026-11-15T00:20:51.276740469+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:3035: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateUserInput!) { createUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676851277458108_141.png",
            "display_name": "Fixture u_customer_1794676851277458108_141",
            "email": "u_customer_1794676851277458108_141@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676851277458108_141.png",
              "created_at": "2026-11-15T00:20:51.277943989+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_customer_1794676851277458108_141",
              "id": "f03ba941-c1fd-4de6-9397-a7e3a063988e",
              "kind": "CUSTOMER",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:51.277943989+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:3035: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_customer_1794676851277458108_141@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "Rm9mHBLNs-lef2O3bwNenagjiPdryWA1chz_IppXauU",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:3036: 
        [RBAC_TEST_AB_FORBIDDEN] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UpdateGroupDetailsInput!) { updateGroupDetails(input: $input) { id code name description is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "id": "f254a3ae-0992-4dad-a56c-acf8b15904ce",
            "name": "No"
          }
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "missing required permission: iam.groups.write",
              "extensions": {
                "code": "FORBIDDEN"
              }
            }
          ]
        }
--- PASS: TestMutationUpdateGroupDetails (0.38s)
    --- PASS: TestMutationUpdateGroupDetails/UpdateNameAndDescription (0.10s)
    --- PASS: TestMutationUpdateGroupDetails/RBACCustomerForbidden (0.28s)
=== RUN   TestMutationActivateGroup
=== RUN   TestMutationActivateGroup/ActivateDeactivatedGroup
    iam_operations_e2e_test.go:3047: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateGroupInput!) { createGroup(input: $input) { id code name description is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "grp_act_g_1794676851554176525_142",
            "description": "Group E2E grp_act_g_1794676851554176525_142",
            "name": "Group grp_act_g_1794676851554176525_142"
          }
        }
        Result JSON:
        {
          "data": {
            "createGroup": {
              "code": "grp_act_g_1794676851554176525_142",
              "created_at": "2026-11-15T00:20:51.554813421+07:00",
              "description": "Group E2E grp_act_g_1794676851554176525_142",
              "id": "15d6b528-1597-4442-95f3-b0f68cf703f8",
              "is_active": true,
              "name": "Group grp_act_g_1794676851554176525_142",
              "updated_at": "2026-11-15T00:20:51.554813421+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:3048: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($id: ID!) { deactivateGroup(id: $id) { id code name description is_active created_at updated_at } }
        Input Variables:
        {
          "id": "15d6b528-1597-4442-95f3-b0f68cf703f8"
        }
        Result JSON:
        {
          "data": {
            "deactivateGroup": {
              "code": "grp_act_g_1794676851554176525_142",
              "created_at": "2026-11-15T00:20:51.554813+07:00",
              "description": "Group E2E grp_act_g_1794676851554176525_142",
              "id": "15d6b528-1597-4442-95f3-b0f68cf703f8",
              "is_active": false,
              "name": "Group grp_act_g_1794676851554176525_142",
              "updated_at": "2026-11-15T00:20:51.556202092+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:3051: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($id: ID!) { activateGroup(id: $id) { id code name description is_active created_at updated_at } }
        Input Variables:
        {
          "id": "15d6b528-1597-4442-95f3-b0f68cf703f8"
        }
        Result JSON:
        {
          "data": {
            "activateGroup": {
              "code": "grp_act_g_1794676851554176525_142",
              "created_at": "2026-11-15T00:20:51.554813+07:00",
              "description": "Group E2E grp_act_g_1794676851554176525_142",
              "id": "15d6b528-1597-4442-95f3-b0f68cf703f8",
              "is_active": true,
              "name": "Group grp_act_g_1794676851554176525_142",
              "updated_at": "2026-11-15T00:20:51.557888727+07:00"
            }
          }
        }
=== RUN   TestMutationActivateGroup/RBACCustomerForbidden
    iam_operations_e2e_test.go:3060: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateGroupInput!) { createGroup(input: $input) { id code name description is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "grp_act_g_rbac_1794676851652683827_143",
            "description": "Group E2E grp_act_g_rbac_1794676851652683827_143",
            "name": "Group grp_act_g_rbac_1794676851652683827_143"
          }
        }
        Result JSON:
        {
          "data": {
            "createGroup": {
              "code": "grp_act_g_rbac_1794676851652683827_143",
              "created_at": "2026-11-15T00:20:51.653509381+07:00",
              "description": "Group E2E grp_act_g_rbac_1794676851652683827_143",
              "id": "315145a6-4900-499d-b773-e0150b72f697",
              "is_active": true,
              "name": "Group grp_act_g_rbac_1794676851652683827_143",
              "updated_at": "2026-11-15T00:20:51.653509381+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:3061: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateUserInput!) { createUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676851654462397_144.png",
            "display_name": "Fixture u_customer_1794676851654462397_144",
            "email": "u_customer_1794676851654462397_144@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676851654462397_144.png",
              "created_at": "2026-11-15T00:20:51.655190526+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_customer_1794676851654462397_144",
              "id": "bb9db9fb-5fa4-4c96-a3eb-de9db4f6d396",
              "kind": "CUSTOMER",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:51.655190526+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:3061: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_customer_1794676851654462397_144@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "8WGqUAqZZzut07ZR8EIo6aMwoHSVHq98_DlKQaqmtHE",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:3062: 
        [RBAC_TEST_AB_FORBIDDEN] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($id: ID!) { activateGroup(id: $id) { id code name description is_active created_at updated_at } }
        Input Variables:
        {
          "id": "315145a6-4900-499d-b773-e0150b72f697"
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "missing required permission: iam.groups.write",
              "extensions": {
                "code": "FORBIDDEN"
              }
            }
          ]
        }
--- PASS: TestMutationActivateGroup (0.37s)
    --- PASS: TestMutationActivateGroup/ActivateDeactivatedGroup (0.10s)
    --- PASS: TestMutationActivateGroup/RBACCustomerForbidden (0.28s)
=== RUN   TestMutationDeactivateGroup
=== RUN   TestMutationDeactivateGroup/DeactivateActiveGroup
    iam_operations_e2e_test.go:3073: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateGroupInput!) { createGroup(input: $input) { id code name description is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "grp_deact_g_1794676851928611907_145",
            "description": "Group E2E grp_deact_g_1794676851928611907_145",
            "name": "Group grp_deact_g_1794676851928611907_145"
          }
        }
        Result JSON:
        {
          "data": {
            "createGroup": {
              "code": "grp_deact_g_1794676851928611907_145",
              "created_at": "2026-11-15T00:20:51.929174893+07:00",
              "description": "Group E2E grp_deact_g_1794676851928611907_145",
              "id": "279fa7db-f6ec-4516-95aa-ce79ba066bf1",
              "is_active": true,
              "name": "Group grp_deact_g_1794676851928611907_145",
              "updated_at": "2026-11-15T00:20:51.929174893+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:3075: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($id: ID!) { deactivateGroup(id: $id) { id code name description is_active created_at updated_at } }
        Input Variables:
        {
          "id": "279fa7db-f6ec-4516-95aa-ce79ba066bf1"
        }
        Result JSON:
        {
          "data": {
            "deactivateGroup": {
              "code": "grp_deact_g_1794676851928611907_145",
              "created_at": "2026-11-15T00:20:51.929174+07:00",
              "description": "Group E2E grp_deact_g_1794676851928611907_145",
              "id": "279fa7db-f6ec-4516-95aa-ce79ba066bf1",
              "is_active": false,
              "name": "Group grp_deact_g_1794676851928611907_145",
              "updated_at": "2026-11-15T00:20:51.93088877+07:00"
            }
          }
        }
=== RUN   TestMutationDeactivateGroup/RBACCustomerForbidden
    iam_operations_e2e_test.go:3084: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateGroupInput!) { createGroup(input: $input) { id code name description is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "grp_deact_g_rbac_1794676852027039338_146",
            "description": "Group E2E grp_deact_g_rbac_1794676852027039338_146",
            "name": "Group grp_deact_g_rbac_1794676852027039338_146"
          }
        }
        Result JSON:
        {
          "data": {
            "createGroup": {
              "code": "grp_deact_g_rbac_1794676852027039338_146",
              "created_at": "2026-11-15T00:20:52.027788597+07:00",
              "description": "Group E2E grp_deact_g_rbac_1794676852027039338_146",
              "id": "7b10d64e-50bb-455c-9d3b-97c4c2fccee5",
              "is_active": true,
              "name": "Group grp_deact_g_rbac_1794676852027039338_146",
              "updated_at": "2026-11-15T00:20:52.027788597+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:3085: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateUserInput!) { createUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676852028676279_147.png",
            "display_name": "Fixture u_customer_1794676852028676279_147",
            "email": "u_customer_1794676852028676279_147@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676852028676279_147.png",
              "created_at": "2026-11-15T00:20:52.029232172+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_customer_1794676852028676279_147",
              "id": "bfcb9f5d-5839-4502-990f-1f3ee4c98c97",
              "kind": "CUSTOMER",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:52.029232172+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:3085: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_customer_1794676852028676279_147@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "2rAhTcXmVNaZdrkjHLcDK7UnhXdWmVBSMQzI2prfrrg",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:3086: 
        [RBAC_TEST_AB_FORBIDDEN] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($id: ID!) { deactivateGroup(id: $id) { id code name description is_active created_at updated_at } }
        Input Variables:
        {
          "id": "7b10d64e-50bb-455c-9d3b-97c4c2fccee5"
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "missing required permission: iam.groups.write",
              "extensions": {
                "code": "FORBIDDEN"
              }
            }
          ]
        }
--- PASS: TestMutationDeactivateGroup (0.37s)
    --- PASS: TestMutationDeactivateGroup/DeactivateActiveGroup (0.10s)
    --- PASS: TestMutationDeactivateGroup/RBACCustomerForbidden (0.28s)
=== RUN   TestMutationDeleteGroup
=== RUN   TestMutationDeleteGroup/DeleteGroupReturnsTrue
    iam_operations_e2e_test.go:3097: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateGroupInput!) { createGroup(input: $input) { id code name description is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "grp_del_g_1794676852304734866_148",
            "description": "Group E2E grp_del_g_1794676852304734866_148",
            "name": "Group grp_del_g_1794676852304734866_148"
          }
        }
        Result JSON:
        {
          "data": {
            "createGroup": {
              "code": "grp_del_g_1794676852304734866_148",
              "created_at": "2026-11-15T00:20:52.30547081+07:00",
              "description": "Group E2E grp_del_g_1794676852304734866_148",
              "id": "b3d620a3-d488-46f4-98f5-9690a3821a88",
              "is_active": true,
              "name": "Group grp_del_g_1794676852304734866_148",
              "updated_at": "2026-11-15T00:20:52.30547081+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:3099: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($id: ID!) { deleteGroup(id: $id) }
        Input Variables:
        {
          "id": "b3d620a3-d488-46f4-98f5-9690a3821a88"
        }
        Result JSON:
        {
          "data": {
            "deleteGroup": true
          }
        }
    iam_operations_e2e_test.go:3100: 
        [GQL_EXPECTED_ERROR] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($id: ID!) { group(id: $id) { id code name description is_active created_at updated_at } }
        Input Variables:
        {
          "id": "b3d620a3-d488-46f4-98f5-9690a3821a88"
        }
        Result JSON:
        {
          "data": {
            "group": null
          },
          "errors": [
            {
              "message": "group not found",
              "extensions": {
                "code": "NOT_FOUND"
              }
            }
          ]
        }
=== RUN   TestMutationDeleteGroup/RBACCustomerForbidden
    iam_operations_e2e_test.go:3105: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateGroupInput!) { createGroup(input: $input) { id code name description is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "grp_del_g_rbac_1794676852403734411_149",
            "description": "Group E2E grp_del_g_rbac_1794676852403734411_149",
            "name": "Group grp_del_g_rbac_1794676852403734411_149"
          }
        }
        Result JSON:
        {
          "data": {
            "createGroup": {
              "code": "grp_del_g_rbac_1794676852403734411_149",
              "created_at": "2026-11-15T00:20:52.404456258+07:00",
              "description": "Group E2E grp_del_g_rbac_1794676852403734411_149",
              "id": "241e1e65-3b2e-4869-8045-b4f605b98bca",
              "is_active": true,
              "name": "Group grp_del_g_rbac_1794676852403734411_149",
              "updated_at": "2026-11-15T00:20:52.404456258+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:3106: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateUserInput!) { createUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676852405256845_150.png",
            "display_name": "Fixture u_customer_1794676852405256845_150",
            "email": "u_customer_1794676852405256845_150@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676852405256845_150.png",
              "created_at": "2026-11-15T00:20:52.405736183+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_customer_1794676852405256845_150",
              "id": "af959f15-b6c1-49d5-aad5-441509c8b14b",
              "kind": "CUSTOMER",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:52.405736183+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:3106: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_customer_1794676852405256845_150@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "q_Z1x-AU4Lqi7HQgYuAc4L2IXfK2vyiChJFmNao0Xmw",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:3107: 
        [RBAC_TEST_AB_FORBIDDEN] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($id: ID!) { deleteGroup(id: $id) }
        Input Variables:
        {
          "id": "241e1e65-3b2e-4869-8045-b4f605b98bca"
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "missing required permission: iam.groups.write",
              "extensions": {
                "code": "FORBIDDEN"
              }
            }
          ]
        }
--- PASS: TestMutationDeleteGroup (0.38s)
    --- PASS: TestMutationDeleteGroup/DeleteGroupReturnsTrue (0.10s)
    --- PASS: TestMutationDeleteGroup/RBACCustomerForbidden (0.28s)
=== RUN   TestMutationAddGroupMember
=== RUN   TestMutationAddGroupMember/AddMemberReturnsNonNull
    iam_operations_e2e_test.go:3120: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateGroupInput!) { createGroup(input: $input) { id code name description is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "grp_agm_1794676852684561060_151",
            "description": "Group E2E grp_agm_1794676852684561060_151",
            "name": "Group grp_agm_1794676852684561060_151"
          }
        }
        Result JSON:
        {
          "data": {
            "createGroup": {
              "code": "grp_agm_1794676852684561060_151",
              "created_at": "2026-11-15T00:20:52.685620106+07:00",
              "description": "Group E2E grp_agm_1794676852684561060_151",
              "id": "f8bba3c6-8ddc-4d4a-835e-7780534ad5ea",
              "is_active": true,
              "name": "Group grp_agm_1794676852684561060_151",
              "updated_at": "2026-11-15T00:20:52.685620106+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:3121: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateEmployeeUserInput!) { createEmployeeUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676852686715862_152.png",
            "display_name": "Fixture u_employee_1794676852686715862_152",
            "email": "u_employee_1794676852686715862_152@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createEmployeeUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676852686715862_152.png",
              "created_at": "2026-11-15T00:20:52.687848828+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_employee_1794676852686715862_152",
              "id": "2549a531-c266-4ab8-9ae6-d2983efc26dc",
              "kind": "EMPLOYEE",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:52.687848828+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:3123: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($group_id: ID!, $user_id: ID!) { addGroupMember(group_id: $group_id, user_id: $user_id) { group_id user_id created_at } }
        Input Variables:
        {
          "group_id": "f8bba3c6-8ddc-4d4a-835e-7780534ad5ea",
          "user_id": "2549a531-c266-4ab8-9ae6-d2983efc26dc"
        }
        Result JSON:
        {
          "data": {
            "addGroupMember": {
              "created_at": "2026-11-15T00:20:52.733516411+07:00",
              "group_id": "f8bba3c6-8ddc-4d4a-835e-7780534ad5ea",
              "user_id": "2549a531-c266-4ab8-9ae6-d2983efc26dc"
            }
          }
        }
=== RUN   TestMutationAddGroupMember/RBACCustomerForbidden
    iam_operations_e2e_test.go:3132: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateGroupInput!) { createGroup(input: $input) { id code name description is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "grp_agm_rbac_1794676852909186651_153",
            "description": "Group E2E grp_agm_rbac_1794676852909186651_153",
            "name": "Group grp_agm_rbac_1794676852909186651_153"
          }
        }
        Result JSON:
        {
          "data": {
            "createGroup": {
              "code": "grp_agm_rbac_1794676852909186651_153",
              "created_at": "2026-11-15T00:20:52.909939808+07:00",
              "description": "Group E2E grp_agm_rbac_1794676852909186651_153",
              "id": "eb725c5a-2b04-4b7e-9e55-dfcf6888d405",
              "is_active": true,
              "name": "Group grp_agm_rbac_1794676852909186651_153",
              "updated_at": "2026-11-15T00:20:52.909939808+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:3133: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateUserInput!) { createUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676852910862115_154.png",
            "display_name": "Fixture u_customer_1794676852910862115_154",
            "email": "u_customer_1794676852910862115_154@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676852910862115_154.png",
              "created_at": "2026-11-15T00:20:52.911570798+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_customer_1794676852910862115_154",
              "id": "bb3d2eb4-862d-45d1-b0a6-fc42a46a0b23",
              "kind": "CUSTOMER",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:52.911570798+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:3133: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_customer_1794676852910862115_154@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "E8zosWGNOrswnxzLfzc1Bc5qWuBIP7LiGXg36widSvA",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:3134: 
        [RBAC_TEST_AB_FORBIDDEN] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($group_id: ID!, $user_id: ID!) { addGroupMember(group_id: $group_id, user_id: $user_id) { group_id user_id created_at } }
        Input Variables:
        {
          "group_id": "eb725c5a-2b04-4b7e-9e55-dfcf6888d405",
          "user_id": "bb3d2eb4-862d-45d1-b0a6-fc42a46a0b23"
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "missing required permission: iam.groups.write",
              "extensions": {
                "code": "FORBIDDEN"
              }
            }
          ]
        }
--- PASS: TestMutationAddGroupMember (0.50s)
    --- PASS: TestMutationAddGroupMember/AddMemberReturnsNonNull (0.23s)
    --- PASS: TestMutationAddGroupMember/RBACCustomerForbidden (0.28s)
=== RUN   TestMutationRemoveGroupMember
=== RUN   TestMutationRemoveGroupMember/RemoveMemberReturnsTrue
    iam_operations_e2e_test.go:3145: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateGroupInput!) { createGroup(input: $input) { id code name description is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "grp_rgm_1794676853188355029_155",
            "description": "Group E2E grp_rgm_1794676853188355029_155",
            "name": "Group grp_rgm_1794676853188355029_155"
          }
        }
        Result JSON:
        {
          "data": {
            "createGroup": {
              "code": "grp_rgm_1794676853188355029_155",
              "created_at": "2026-11-15T00:20:53.189050427+07:00",
              "description": "Group E2E grp_rgm_1794676853188355029_155",
              "id": "18e9b830-4253-4123-9e8d-373d5ca5b0a7",
              "is_active": true,
              "name": "Group grp_rgm_1794676853188355029_155",
              "updated_at": "2026-11-15T00:20:53.189050427+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:3146: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateEmployeeUserInput!) { createEmployeeUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676853189825194_156.png",
            "display_name": "Fixture u_employee_1794676853189825194_156",
            "email": "u_employee_1794676853189825194_156@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createEmployeeUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676853189825194_156.png",
              "created_at": "2026-11-15T00:20:53.19034524+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_employee_1794676853189825194_156",
              "id": "08976d28-f048-4eef-a60d-11a2658f9042",
              "kind": "EMPLOYEE",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:53.19034524+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:3147: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($g: ID!, $u: ID!) { addGroupMember(group_id: $g, user_id: $u) { group_id user_id created_at } }
        Input Variables:
        {
          "g": "18e9b830-4253-4123-9e8d-373d5ca5b0a7",
          "u": "08976d28-f048-4eef-a60d-11a2658f9042"
        }
        Result JSON:
        {
          "data": {
            "addGroupMember": {
              "created_at": "2026-11-15T00:20:53.237984037+07:00",
              "group_id": "18e9b830-4253-4123-9e8d-373d5ca5b0a7",
              "user_id": "08976d28-f048-4eef-a60d-11a2658f9042"
            }
          }
        }
    iam_operations_e2e_test.go:3152: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($group_id: ID!, $user_id: ID!) { removeGroupMember(group_id: $group_id, user_id: $user_id) }
        Input Variables:
        {
          "group_id": "18e9b830-4253-4123-9e8d-373d5ca5b0a7",
          "user_id": "08976d28-f048-4eef-a60d-11a2658f9042"
        }
        Result JSON:
        {
          "data": {
            "removeGroupMember": true
          }
        }
    iam_operations_e2e_test.go:3153: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($g: ID!) { groupMembers(group_id: $g) { items { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } page { total limit offset has_more } } }
        Input Variables:
        {
          "g": "18e9b830-4253-4123-9e8d-373d5ca5b0a7"
        }
        Result JSON:
        {
          "data": {
            "groupMembers": {
              "items": [],
              "page": {
                "has_more": false,
                "limit": 20,
                "offset": 0,
                "total": 0
              }
            }
          }
        }
=== RUN   TestMutationRemoveGroupMember/RBACCustomerForbidden
    iam_operations_e2e_test.go:3162: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateGroupInput!) { createGroup(input: $input) { id code name description is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "grp_rgm_rbac_1794676853414833341_157",
            "description": "Group E2E grp_rgm_rbac_1794676853414833341_157",
            "name": "Group grp_rgm_rbac_1794676853414833341_157"
          }
        }
        Result JSON:
        {
          "data": {
            "createGroup": {
              "code": "grp_rgm_rbac_1794676853414833341_157",
              "created_at": "2026-11-15T00:20:53.41561443+07:00",
              "description": "Group E2E grp_rgm_rbac_1794676853414833341_157",
              "id": "0fb96bf6-202b-4439-8ae5-c64d3877b10b",
              "is_active": true,
              "name": "Group grp_rgm_rbac_1794676853414833341_157",
              "updated_at": "2026-11-15T00:20:53.41561443+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:3163: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateUserInput!) { createUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676853416375542_158.png",
            "display_name": "Fixture u_customer_1794676853416375542_158",
            "email": "u_customer_1794676853416375542_158@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676853416375542_158.png",
              "created_at": "2026-11-15T00:20:53.417222236+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_customer_1794676853416375542_158",
              "id": "732d27f6-b8a7-4c84-88f2-aa805179cc59",
              "kind": "CUSTOMER",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:53.417222236+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:3163: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_customer_1794676853416375542_158@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "C26pTc8lwKsqu0ztlpxnGCku_XTdpzNQQSJ3l66KqFM",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:3164: 
        [RBAC_TEST_AB_FORBIDDEN] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($group_id: ID!, $user_id: ID!) { removeGroupMember(group_id: $group_id, user_id: $user_id) }
        Input Variables:
        {
          "group_id": "0fb96bf6-202b-4439-8ae5-c64d3877b10b",
          "user_id": "732d27f6-b8a7-4c84-88f2-aa805179cc59"
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "missing required permission: iam.groups.write",
              "extensions": {
                "code": "FORBIDDEN"
              }
            }
          ]
        }
--- PASS: TestMutationRemoveGroupMember (0.51s)
    --- PASS: TestMutationRemoveGroupMember/RemoveMemberReturnsTrue (0.23s)
    --- PASS: TestMutationRemoveGroupMember/RBACCustomerForbidden (0.28s)
=== RUN   TestMutationAssignGroupRole
=== RUN   TestMutationAssignGroupRole/AssignScopedGroupRoleNonNullAndInherited
    iam_operations_e2e_test.go:3177: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateGroupInput!) { createGroup(input: $input) { id code name description is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "grp_agr_1794676853693478347_159",
            "description": "Group E2E grp_agr_1794676853693478347_159",
            "name": "Group grp_agr_1794676853693478347_159"
          }
        }
        Result JSON:
        {
          "data": {
            "createGroup": {
              "code": "grp_agr_1794676853693478347_159",
              "created_at": "2026-11-15T00:20:53.694181018+07:00",
              "description": "Group E2E grp_agr_1794676853693478347_159",
              "id": "cd0be67f-a15e-4a82-8f73-99ab21869422",
              "is_active": true,
              "name": "Group grp_agr_1794676853693478347_159",
              "updated_at": "2026-11-15T00:20:53.694181018+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:3178: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateRoleInput!) { createRole(input: $input) { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "role_agr_1794676853694929386_160",
            "default_redirect_path": "/identity",
            "description": "Custom E2E role role_agr_1794676853694929386_160",
            "is_system": false,
            "management_domain": "iam",
            "name": "Role role_agr_1794676853694929386_160",
            "redirect_priority": 50
          }
        }
        Result JSON:
        {
          "data": {
            "createRole": {
              "code": "role_agr_1794676853694929386_160",
              "created_at": "2026-11-15T00:20:53.695463938+07:00",
              "default_redirect_path": "/identity",
              "description": "Custom E2E role role_agr_1794676853694929386_160",
              "id": "f9e56d2a-d6d4-4aa8-a910-96b887d0aa41",
              "is_active": true,
              "is_system": false,
              "management_domain": "iam",
              "name": "Role role_agr_1794676853694929386_160",
              "redirect_priority": 50,
              "updated_at": "2026-11-15T00:20:53.695463938+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:3179: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateEmployeeUserInput!) { createEmployeeUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676853696126013_161.png",
            "display_name": "Fixture u_employee_1794676853696126013_161",
            "email": "u_employee_1794676853696126013_161@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createEmployeeUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676853696126013_161.png",
              "created_at": "2026-11-15T00:20:53.696564283+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_employee_1794676853696126013_161",
              "id": "30a98a26-2884-479c-9d97-e08695c8f8f2",
              "kind": "EMPLOYEE",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:53.696564283+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:3180: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($g: ID!, $u: ID!) { addGroupMember(group_id: $g, user_id: $u) { group_id user_id created_at } }
        Input Variables:
        {
          "g": "cd0be67f-a15e-4a82-8f73-99ab21869422",
          "u": "30a98a26-2884-479c-9d97-e08695c8f8f2"
        }
        Result JSON:
        {
          "data": {
            "addGroupMember": {
              "created_at": "2026-11-15T00:20:53.741370474+07:00",
              "group_id": "cd0be67f-a15e-4a82-8f73-99ab21869422",
              "user_id": "30a98a26-2884-479c-9d97-e08695c8f8f2"
            }
          }
        }
    iam_operations_e2e_test.go:3186: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: AssignGroupRoleInput!) { assignGroupRole(input: $input) { id group_id role_id scope { type id } granted_by expires_at created_at } }
        Input Variables:
        {
          "input": {
            "expires_at": "2026-11-14T21:20:53Z",
            "group_id": "cd0be67f-a15e-4a82-8f73-99ab21869422",
            "role_id": "f9e56d2a-d6d4-4aa8-a910-96b887d0aa41",
            "scope_id": "44444444-4444-4444-8444-444444444444",
            "scope_type": "FACILITY"
          }
        }
        Result JSON:
        {
          "data": {
            "assignGroupRole": {
              "created_at": "2026-11-15T00:20:53.743833219+07:00",
              "expires_at": "2026-11-15T04:20:53+07:00",
              "granted_by": "853aac77-32e7-4a88-b2d6-7d68db8d13d7",
              "group_id": "cd0be67f-a15e-4a82-8f73-99ab21869422",
              "id": "997aae68-809f-4f6b-aab1-7f1a8c1b54ef",
              "role_id": "f9e56d2a-d6d4-4aa8-a910-96b887d0aa41",
              "scope": {
                "id": "44444444-4444-4444-8444-444444444444",
                "type": "FACILITY"
              }
            }
          }
        }
    iam_operations_e2e_test.go:3203: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($u: ID!) { userEffectiveRoles(user_id: $u) { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } }
        Input Variables:
        {
          "u": "30a98a26-2884-479c-9d97-e08695c8f8f2"
        }
        Result JSON:
        {
          "data": {
            "userEffectiveRoles": [
              {
                "code": "role_agr_1794676853694929386_160",
                "created_at": "2026-11-15T00:20:53.695463+07:00",
                "default_redirect_path": "/identity",
                "description": "Custom E2E role role_agr_1794676853694929386_160",
                "id": "f9e56d2a-d6d4-4aa8-a910-96b887d0aa41",
                "is_active": true,
                "is_system": false,
                "management_domain": "iam",
                "name": "Role role_agr_1794676853694929386_160",
                "redirect_priority": 50,
                "updated_at": "2026-11-15T00:20:53.695463+07:00"
              }
            ]
          }
        }
=== RUN   TestMutationAssignGroupRole/RBACCustomerForbidden
    iam_operations_e2e_test.go:3219: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateGroupInput!) { createGroup(input: $input) { id code name description is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "grp_agr_rbac_1794676854040256715_162",
            "description": "Group E2E grp_agr_rbac_1794676854040256715_162",
            "name": "Group grp_agr_rbac_1794676854040256715_162"
          }
        }
        Result JSON:
        {
          "data": {
            "createGroup": {
              "code": "grp_agr_rbac_1794676854040256715_162",
              "created_at": "2026-11-15T00:20:54.040958595+07:00",
              "description": "Group E2E grp_agr_rbac_1794676854040256715_162",
              "id": "d32dc4c9-277f-4687-96a7-6044a81d580a",
              "is_active": true,
              "name": "Group grp_agr_rbac_1794676854040256715_162",
              "updated_at": "2026-11-15T00:20:54.040958595+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:3220: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateRoleInput!) { createRole(input: $input) { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "role_agr_rbac_1794676854041733703_163",
            "default_redirect_path": "/identity",
            "description": "Custom E2E role role_agr_rbac_1794676854041733703_163",
            "is_system": false,
            "management_domain": "iam",
            "name": "Role role_agr_rbac_1794676854041733703_163",
            "redirect_priority": 50
          }
        }
        Result JSON:
        {
          "data": {
            "createRole": {
              "code": "role_agr_rbac_1794676854041733703_163",
              "created_at": "2026-11-15T00:20:54.042272163+07:00",
              "default_redirect_path": "/identity",
              "description": "Custom E2E role role_agr_rbac_1794676854041733703_163",
              "id": "cf26af96-e9a7-48a7-8705-81ff57da9773",
              "is_active": true,
              "is_system": false,
              "management_domain": "iam",
              "name": "Role role_agr_rbac_1794676854041733703_163",
              "redirect_priority": 50,
              "updated_at": "2026-11-15T00:20:54.042272163+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:3221: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateUserInput!) { createUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676854042910593_164.png",
            "display_name": "Fixture u_customer_1794676854042910593_164",
            "email": "u_customer_1794676854042910593_164@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676854042910593_164.png",
              "created_at": "2026-11-15T00:20:54.043335808+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_customer_1794676854042910593_164",
              "id": "49a5f750-33f1-40d2-a61c-34fb53afd982",
              "kind": "CUSTOMER",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:54.043335808+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:3221: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_customer_1794676854042910593_164@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "1sCDob3WK25SQLxeBzGdGVgGbVGPk4OBMZ_im4nw718",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:3222: 
        [RBAC_TEST_AB_FORBIDDEN] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: AssignGroupRoleInput!) { assignGroupRole(input: $input) { id group_id role_id scope { type id } granted_by expires_at created_at } }
        Input Variables:
        {
          "input": {
            "group_id": "d32dc4c9-277f-4687-96a7-6044a81d580a",
            "role_id": "cf26af96-e9a7-48a7-8705-81ff57da9773",
            "scope_type": "GLOBAL"
          }
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "missing required permission: iam.groups.write",
              "extensions": {
                "code": "FORBIDDEN"
              }
            }
          ]
        }
--- PASS: TestMutationAssignGroupRole (0.74s)
    --- PASS: TestMutationAssignGroupRole/AssignScopedGroupRoleNonNullAndInherited (0.35s)
    --- PASS: TestMutationAssignGroupRole/RBACCustomerForbidden (0.39s)
=== RUN   TestMutationRevokeGroupRole
=== RUN   TestMutationRevokeGroupRole/RevokeGroupRoleReturnsTrue
    iam_operations_e2e_test.go:3235: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateGroupInput!) { createGroup(input: $input) { id code name description is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "grp_rgr_1794676854430846861_165",
            "description": "Group E2E grp_rgr_1794676854430846861_165",
            "name": "Group grp_rgr_1794676854430846861_165"
          }
        }
        Result JSON:
        {
          "data": {
            "createGroup": {
              "code": "grp_rgr_1794676854430846861_165",
              "created_at": "2026-11-15T00:20:54.431571834+07:00",
              "description": "Group E2E grp_rgr_1794676854430846861_165",
              "id": "463b09b5-fea3-4a7a-8a48-828ee7434443",
              "is_active": true,
              "name": "Group grp_rgr_1794676854430846861_165",
              "updated_at": "2026-11-15T00:20:54.431571834+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:3236: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateRoleInput!) { createRole(input: $input) { id code name description management_domain default_redirect_path redirect_priority is_system is_active created_at updated_at } }
        Input Variables:
        {
          "input": {
            "code": "role_rgr_1794676854432319701_166",
            "default_redirect_path": "/identity",
            "description": "Custom E2E role role_rgr_1794676854432319701_166",
            "is_system": false,
            "management_domain": "iam",
            "name": "Role role_rgr_1794676854432319701_166",
            "redirect_priority": 50
          }
        }
        Result JSON:
        {
          "data": {
            "createRole": {
              "code": "role_rgr_1794676854432319701_166",
              "created_at": "2026-11-15T00:20:54.432882427+07:00",
              "default_redirect_path": "/identity",
              "description": "Custom E2E role role_rgr_1794676854432319701_166",
              "id": "b7667407-6309-4e8c-a862-aa07e4643a2b",
              "is_active": true,
              "is_system": false,
              "management_domain": "iam",
              "name": "Role role_rgr_1794676854432319701_166",
              "redirect_priority": 50,
              "updated_at": "2026-11-15T00:20:54.432882427+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:3237: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($in: AssignGroupRoleInput!) { assignGroupRole(input: $in) { id group_id role_id scope { type id } granted_by expires_at created_at } }
        Input Variables:
        {
          "in": {
            "group_id": "463b09b5-fea3-4a7a-8a48-828ee7434443",
            "role_id": "b7667407-6309-4e8c-a862-aa07e4643a2b",
            "scope_type": "GLOBAL"
          }
        }
        Result JSON:
        {
          "data": {
            "assignGroupRole": {
              "created_at": "2026-11-15T00:20:54.43452065+07:00",
              "expires_at": null,
              "granted_by": "853aac77-32e7-4a88-b2d6-7d68db8d13d7",
              "group_id": "463b09b5-fea3-4a7a-8a48-828ee7434443",
              "id": "b7ff023e-be40-42be-a572-6671d6ff2c5b",
              "role_id": "b7667407-6309-4e8c-a862-aa07e4643a2b",
              "scope": {
                "id": null,
                "type": "GLOBAL"
              }
            }
          }
        }
    iam_operations_e2e_test.go:3243: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($assignment_id: ID!) { revokeGroupRole(assignment_id: $assignment_id) }
        Input Variables:
        {
          "assignment_id": "b7ff023e-be40-42be-a572-6671d6ff2c5b"
        }
        Result JSON:
        {
          "data": {
            "revokeGroupRole": true
          }
        }
    iam_operations_e2e_test.go:3244: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($g: ID!) { groupRoles(group_id: $g) { id group_id role_id scope { type id } granted_by expires_at created_at } }
        Input Variables:
        {
          "g": "463b09b5-fea3-4a7a-8a48-828ee7434443"
        }
        Result JSON:
        {
          "data": {
            "groupRoles": []
          }
        }
=== RUN   TestMutationRevokeGroupRole/RBACCustomerForbidden
    iam_operations_e2e_test.go:3252: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateUserInput!) { createUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676854657037097_167.png",
            "display_name": "Fixture u_customer_1794676854657037097_167",
            "email": "u_customer_1794676854657037097_167@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676854657037097_167.png",
              "created_at": "2026-11-15T00:20:54.657824197+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_customer_1794676854657037097_167",
              "id": "e54a0f44-792d-4c2d-94c3-c23e6d8e6c4e",
              "kind": "CUSTOMER",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:54.657824197+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:3252: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_customer_1794676854657037097_167@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "Z9Ho0fxQ67qTu290BjW0HSYyp-8SuobJgdoxdaP2LFA",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:3253: 
        [RBAC_TEST_AB_FORBIDDEN] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($assignment_id: ID!) { revokeGroupRole(assignment_id: $assignment_id) }
        Input Variables:
        {
          "assignment_id": "00000000-0000-4000-8000-000000000001"
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "missing required permission: iam.groups.write",
              "extensions": {
                "code": "FORBIDDEN"
              }
            }
          ]
        }
--- PASS: TestMutationRevokeGroupRole (0.46s)
    --- PASS: TestMutationRevokeGroupRole/RevokeGroupRoleReturnsTrue (0.23s)
    --- PASS: TestMutationRevokeGroupRole/RBACCustomerForbidden (0.23s)
=== RUN   TestMutationRevokeSession
=== RUN   TestMutationRevokeSession/RevokeTargetSessionInvalidatesClient
    iam_operations_e2e_test.go:3264: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateEmployeeUserInput!) { createEmployeeUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676854891939857_168.png",
            "display_name": "Fixture u_employee_1794676854891939857_168",
            "email": "u_employee_1794676854891939857_168@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createEmployeeUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676854891939857_168.png",
              "created_at": "2026-11-15T00:20:54.89306023+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_employee_1794676854891939857_168",
              "id": "9ee0f36a-0a89-411f-9a18-5a79407ff249",
              "kind": "EMPLOYEE",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:54.89306023+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:3265: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_employee_1794676854891939857_168@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "PHmqlOxxQSo9hayhrMjvR9hazWOpbw0PDhR26cmS-NQ",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/account",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:3267: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($u: ID!) { userSessions(user_id: $u, active_only: true) { items { id user_id mfa_verified mfa_verified_at ip_address user_agent authenticated_at last_active_at idle_expires_at absolute_expires_at ended_at end_reason created_at generation } page { total limit offset has_more } } }
        Input Variables:
        {
          "u": "9ee0f36a-0a89-411f-9a18-5a79407ff249"
        }
        Result JSON:
        {
          "data": {
            "userSessions": {
              "items": [
                {
                  "absolute_expires_at": "2026-11-15T08:20:54.985305+07:00",
                  "authenticated_at": "2026-11-15T00:20:54.985305+07:00",
                  "created_at": "2026-11-15T00:20:54.985305+07:00",
                  "end_reason": null,
                  "ended_at": null,
                  "generation": 1,
                  "id": "848c5778-905c-41b5-9269-dab4255dca9c",
                  "idle_expires_at": "2026-11-15T00:50:54.985305+07:00",
                  "ip_address": "172.22.0.1",
                  "last_active_at": "2026-11-15T00:20:54.985305+07:00",
                  "mfa_verified": false,
                  "mfa_verified_at": null,
                  "user_agent": "Go-http-client/1.1",
                  "user_id": "9ee0f36a-0a89-411f-9a18-5a79407ff249"
                }
              ],
              "page": {
                "has_more": false,
                "limit": 20,
                "offset": 0,
                "total": 1
              }
            }
          }
        }
    iam_operations_e2e_test.go:3274: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($session_id: ID!, $reason: SessionEndReason) { revokeSession(session_id: $session_id, reason: $reason) }
        Input Variables:
        {
          "reason": "ADMIN_REVOKED",
          "session_id": "848c5778-905c-41b5-9269-dab4255dca9c"
        }
        Result JSON:
        {
          "data": {
            "revokeSession": true
          }
        }
    iam_operations_e2e_test.go:3276: 
        [GQL_EXPECTED_ERROR] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query { myProfile { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {}
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "session is missing, expired, or revoked",
              "extensions": {
                "code": "UNAUTHENTICATED"
              }
            }
          ]
        }
    iam_operations_e2e_test.go:3278: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query ($u: ID!) { userSessions(user_id: $u, active_only: false) { items { id user_id mfa_verified mfa_verified_at ip_address user_agent authenticated_at last_active_at idle_expires_at absolute_expires_at ended_at end_reason created_at generation } page { total limit offset has_more } } }
        Input Variables:
        {
          "u": "9ee0f36a-0a89-411f-9a18-5a79407ff249"
        }
        Result JSON:
        {
          "data": {
            "userSessions": {
              "items": [
                {
                  "absolute_expires_at": "2026-11-15T08:20:54.985305+07:00",
                  "authenticated_at": "2026-11-15T00:20:54.985305+07:00",
                  "created_at": "2026-11-15T00:20:54.985305+07:00",
                  "end_reason": "ADMIN_REVOKED",
                  "ended_at": "2026-11-15T00:20:54.989851+07:00",
                  "generation": 1,
                  "id": "848c5778-905c-41b5-9269-dab4255dca9c",
                  "idle_expires_at": "2026-11-15T00:50:54.985305+07:00",
                  "ip_address": "172.22.0.1",
                  "last_active_at": "2026-11-15T00:20:54.985305+07:00",
                  "mfa_verified": false,
                  "mfa_verified_at": null,
                  "user_agent": "Go-http-client/1.1",
                  "user_id": "9ee0f36a-0a89-411f-9a18-5a79407ff249"
                }
              ],
              "page": {
                "has_more": false,
                "limit": 20,
                "offset": 0,
                "total": 1
              }
            }
          }
        }
=== RUN   TestMutationRevokeSession/RBACCustomerForbidden
    iam_operations_e2e_test.go:3291: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateUserInput!) { createUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676855127194514_169.png",
            "display_name": "Fixture u_customer_1794676855127194514_169",
            "email": "u_customer_1794676855127194514_169@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676855127194514_169.png",
              "created_at": "2026-11-15T00:20:55.128007344+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_customer_1794676855127194514_169",
              "id": "e541e701-8854-41ea-8470-13f79380e6d6",
              "kind": "CUSTOMER",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:55.128007344+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:3291: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_customer_1794676855127194514_169@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "O49EzN4g1r0zIX5VgjSlfsUBaat_Gv-6hfaWYM6c44c",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:3292: 
        [RBAC_TEST_AB_FORBIDDEN] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($session_id: ID!, $reason: SessionEndReason) { revokeSession(session_id: $session_id, reason: $reason) }
        Input Variables:
        {
          "reason": "ADMIN_REVOKED",
          "session_id": "00000000-0000-4000-8000-000000000001"
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "missing required permission: iam.sessions.revoke",
              "extensions": {
                "code": "FORBIDDEN"
              }
            }
          ]
        }
--- PASS: TestMutationRevokeSession (0.47s)
    --- PASS: TestMutationRevokeSession/RevokeTargetSessionInvalidatesClient (0.24s)
    --- PASS: TestMutationRevokeSession/RBACCustomerForbidden (0.23s)
=== RUN   TestMutationRevokeAllUserSessions
=== RUN   TestMutationRevokeAllUserSessions/RevokeAllActiveSessionsReturnsCountAndInvalidatesAll
    iam_operations_e2e_test.go:3305: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateEmployeeUserInput!) { createEmployeeUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676855359340603_170.png",
            "display_name": "Fixture u_employee_1794676855359340603_170",
            "email": "u_employee_1794676855359340603_170@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createEmployeeUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_employee_1794676855359340603_170.png",
              "created_at": "2026-11-15T00:20:55.360169143+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_employee_1794676855359340603_170",
              "id": "efc38c7a-1cd5-4960-9f67-fb211845a989",
              "kind": "EMPLOYEE",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:55.360169143+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:3306: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_employee_1794676855359340603_170@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "L6AmNSir9nEG40I_o-BVyjEGCkNK5cioaeOfP_6puwc",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/account",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:3307: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_employee_1794676855359340603_170@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "v4I5XKs_UtZdujc867F16hb-BFQ7Fu1ly4GByjJYmRE",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/account",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:3309: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($user_id: ID!, $reason: SessionEndReason) { revokeAllUserSessions(user_id: $user_id, reason: $reason) }
        Input Variables:
        {
          "reason": "ADMIN_REVOKED",
          "user_id": "efc38c7a-1cd5-4960-9f67-fb211845a989"
        }
        Result JSON:
        {
          "data": {
            "revokeAllUserSessions": 2
          }
        }
    iam_operations_e2e_test.go:3317: 
        [GQL_EXPECTED_ERROR] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query { myProfile { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {}
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "session is missing, expired, or revoked",
              "extensions": {
                "code": "UNAUTHENTICATED"
              }
            }
          ]
        }
    iam_operations_e2e_test.go:3317: 
        [GQL_EXPECTED_ERROR] POST /api/graphql (HTTP 200)
        Query/Mutation:
          query { myProfile { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {}
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "session is missing, expired, or revoked",
              "extensions": {
                "code": "UNAUTHENTICATED"
              }
            }
          ]
        }
=== RUN   TestMutationRevokeAllUserSessions/RBACCustomerForbidden
    iam_operations_e2e_test.go:3323: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: CreateUserInput!) { createUser(input: $input) { id kind status display_name avatar_url locale timezone last_login_at created_at updated_at deleted_at auth_epoch perm_revision } }
        Input Variables:
        {
          "input": {
            "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676855637725597_171.png",
            "display_name": "Fixture u_customer_1794676855637725597_171",
            "email": "u_customer_1794676855637725597_171@integrione.test",
            "locale": "id-ID",
            "mark_identities_verified": true,
            "password": "StrongPass#2026!",
            "status": "ACTIVE",
            "timezone": "Asia/Jakarta"
          }
        }
        Result JSON:
        {
          "data": {
            "createUser": {
              "auth_epoch": 1,
              "avatar_url": "https://cdn.integrione.test/avatar/u_customer_1794676855637725597_171.png",
              "created_at": "2026-11-15T00:20:55.63847683+07:00",
              "deleted_at": null,
              "display_name": "Fixture u_customer_1794676855637725597_171",
              "id": "8452c70e-e5c1-492d-96bc-88b6d7ee868d",
              "kind": "CUSTOMER",
              "last_login_at": null,
              "locale": "id-ID",
              "perm_revision": 1,
              "status": "ACTIVE",
              "timezone": "Asia/Jakarta",
              "updated_at": "2026-11-15T00:20:55.63847683+07:00"
            }
          }
        }
    iam_operations_e2e_test.go:3323: 
        [GQL_OK] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($input: UserLoginInput!) { userLogin(input: $input) { status mfaChallenge mfaExpiresAt csrfToken redirectLocation } }
        Input Variables:
        {
          "input": {
            "identifier": "u_customer_1794676855637725597_171@integrione.test",
            "password": "StrongPass#2026!"
          }
        }
        Result JSON:
        {
          "data": {
            "userLogin": {
              "csrfToken": "lE4tVKKW2oSQVLohO5Oqjd1Po9WuqC-363qK0yM08EI",
              "mfaChallenge": null,
              "mfaExpiresAt": null,
              "redirectLocation": "/",
              "status": "AUTHENTICATED"
            }
          }
        }
    iam_operations_e2e_test.go:3324: 
        [RBAC_TEST_AB_FORBIDDEN] POST /api/graphql (HTTP 200)
        Query/Mutation:
          mutation ($user_id: ID!, $reason: SessionEndReason) { revokeAllUserSessions(user_id: $user_id, reason: $reason) }
        Input Variables:
        {
          "reason": "ADMIN_REVOKED",
          "user_id": "8452c70e-e5c1-492d-96bc-88b6d7ee868d"
        }
        Result JSON:
        {
          "data": null,
          "errors": [
            {
              "message": "missing required permission: iam.sessions.revoke",
              "extensions": {
                "code": "FORBIDDEN"
              }
            }
          ]
        }
--- PASS: TestMutationRevokeAllUserSessions (0.51s)
    --- PASS: TestMutationRevokeAllUserSessions/RevokeAllActiveSessionsReturnsCountAndInvalidatesAll (0.28s)
    --- PASS: TestMutationRevokeAllUserSessions/RBACCustomerForbidden (0.23s)
=== RUN   TestLiveHTTPSIAMGatewayE2E
=== RUN   TestLiveHTTPSIAMGatewayE2E/SingleEndpointAndFrontendRoutes
=== RUN   TestLiveHTTPSIAMGatewayE2E/BrowserBootstrapOriginCSRFAndSuperAdminLogin
=== RUN   TestLiveHTTPSIAMGatewayE2E/UserVariantsSnakeCaseAndRBACLifecycle
=== RUN   TestLiveHTTPSIAMGatewayE2E/NegativeAuthorizationTestABAndZeroDBMutation
=== RUN   TestLiveHTTPSIAMGatewayE2E/EmployeeSelfServiceTOTPMFAAndStaleSensitiveDenial
=== RUN   TestLiveHTTPSIAMGatewayE2E/RegistrationVerificationPasswordResetAndRateLimit
=== RUN   TestLiveHTTPSIAMGatewayE2E/RedisACLEnforcementAndFailClosedRecovery
--- PASS: TestLiveHTTPSIAMGatewayE2E (23.18s)
    --- PASS: TestLiveHTTPSIAMGatewayE2E/SingleEndpointAndFrontendRoutes (0.02s)
    --- PASS: TestLiveHTTPSIAMGatewayE2E/BrowserBootstrapOriginCSRFAndSuperAdminLogin (0.05s)
    --- PASS: TestLiveHTTPSIAMGatewayE2E/UserVariantsSnakeCaseAndRBACLifecycle (0.22s)
    --- PASS: TestLiveHTTPSIAMGatewayE2E/NegativeAuthorizationTestABAndZeroDBMutation (0.29s)
    --- PASS: TestLiveHTTPSIAMGatewayE2E/EmployeeSelfServiceTOTPMFAAndStaleSensitiveDenial (4.49s)
    --- PASS: TestLiveHTTPSIAMGatewayE2E/RegistrationVerificationPasswordResetAndRateLimit (0.10s)
    --- PASS: TestLiveHTTPSIAMGatewayE2E/RedisACLEnforcementAndFailClosedRecovery (17.72s)
PASS
ok  	github.com/integrione/erp-retail/backend/gateway/internal/e2e	52.932s
