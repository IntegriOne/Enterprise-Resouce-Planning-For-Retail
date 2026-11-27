"""Menyiapkan grant GraphQL, pekerja fixture, serta sesi MFA untuk uji Go."""
import json, ssl, urllib.request, http.cookiejar, os, uuid
from pathlib import Path
import datetime
base = 'https://erp.localhost:18443'
opener = urllib.request.build_opener(urllib.request.HTTPSHandler(context=ssl.create_default_context(cafile='backend/deploy/certs/ca.crt')), urllib.request.HTTPCookieProcessor(http.cookiejar.CookieJar()))
csrf = ''

def graphql(query, variables=None):
    global csrf
    headers = {'Content-Type': 'application/json', 'Origin': 'https://erp.localhost:8443'}
    if csrf:
        headers['X-CSRF-Token'] = csrf
    request = urllib.request.Request(base + '/api/graphql', json.dumps({'query': query, 'variables': variables}).encode(), headers)
    try:
        response = opener.open(request, timeout=10)
    except urllib.error.HTTPError as err:
        response = err
    with response:
        csrf = response.headers.get('X-CSRF-Token') or csrf
        result = json.load(response)
    if result.get('errors'):
        raise RuntimeError([(e.get('extensions', {}).get('code'), e.get('message')) for e in result['errors']])
    return result['data']

def mutation(name, query, variables=None):
    result = graphql(query, variables)[name]
    if not result['was_successful']:
        raise RuntimeError((name, result['errors']))
    print(name, 'OK', flush=True)
    return result
env = {}
for line in Path('backend/services/iam/.env').read_text().splitlines():
    k, sep, v = line.partition('=')
    if sep:
        env[k] = v.strip().strip('"').strip("'")

def login():
    global csrf
    csrf = graphql('query { browserSession { csrfToken } }')['browserSession']['csrfToken']
    data = mutation('userLogin', 'mutation($i: UserLoginInput!){userLogin(input:$i){status csrfToken was_successful errors{type}}}', {'i': {'identifier': os.getenv('E2E_SUPERADMIN_USERNAME') or env.get('IAM_SUPERADMIN_USERNAME', 'superadmin'), 'password': os.getenv('E2E_SUPERADMIN_PASSWORD') or env['IAM_SUPERADMIN_PASSWORD']}})
    assert data['status'] == 'AUTHENTICATED', data['status']
    csrf = data['csrfToken'] or csrf

def logout():
    mutation('userLogout', 'mutation { userLogout { was_successful errors{type} } }')
try:
    graphql('query {employeeProvisioningRequests(first:1){totalCount}}')
except RuntimeError as err:
    assert err.args[0][0][0] == 'UNAUTHENTICATED', err
    print('IAM queue rejects requests without a session', flush=True)
else:
    raise AssertionError('unauthenticated IAM queue unexpectedly returned data')
login()
role = None
assignment = None
scoped_assignment = None
try:
    principal = graphql('query { browserSession { principalId } }')['browserSession']['principalId']
    codes = ['human_resource.payslips.reissue', 'human_resource.workers.activate', 'human_resource.workers.create', 'human_resource.workers.read', 'human_resource.access_requests.create', 'human_resource.access_requests.read']
    automatic_termination = os.environ.get('HRM_E2E_TERMINATION') == '1'
    if automatic_termination:
        codes.append('human_resource.workers.terminate')
    perms = []
    for code in codes:
        rows = graphql('query($q:String!){ permissions(query:$q, first:10){nodes{id code}} }', {'q': code})['permissions']['nodes']
        perms.append(next((x['id'] for x in rows if x['code'] == code)))
    role = mutation('createRole', 'mutation($i:CreateRoleInput!){createRole(input:$i){role{id} was_successful errors{type message}}}', {'i': {'code': 'hrm_e2e_' + uuid.uuid4().hex[:10], 'name': 'HRM E2E Query', 'management_domain': 'human_resource', 'default_redirect_path': '/identity', 'redirect_priority': 999}})['role']['id']
    for perm in perms:
        mutation('assignRolePermission', 'mutation($r:ID!,$p:ID!){assignRolePermission(role_id:$r,permission_id:$p){was_successful errors{type message}}}', {'r': role, 'p': perm})
    assignment = mutation('assignUserRole', 'mutation($i:AssignUserRoleInput!){assignUserRole(input:$i){user_role_edge{assignment_id} was_successful errors{type message}}}', {'i': {'user_id': principal, 'role_id': role, 'scope_type': 'GLOBAL'}})['user_role_edge']['assignment_id']
    login()
    fixture_day = datetime.datetime.now(datetime.timezone.utc).date() - datetime.timedelta(days=1)
    ref = mutation('createLegalEntity', 'mutation($i:CreateLegalEntityInput!){createLegalEntity(input:$i){legal_entity{_id status} was_successful errors{type message}}}', {'i': {'entity_code': 'E2E-' + uuid.uuid4().hex[:8].upper(), 'legal_name': 'Employee E2E Entity', 'jurisdiction_country_code': 'ID', 'status': 'ACTIVE', 'incorporation_date': fixture_day.isoformat() + 'T00:00:00Z'}})['legal_entity']
    scope_mutation = 'mutation($i:AssignUserRoleInput!){assignUserRole(input:$i){user_role_edge{assignment_id} was_successful errors{type message}}}'
    scope = {'user_id': principal, 'role_id': role, 'scope_type': 'LEGAL_ENTITY'}
    invalid = graphql(scope_mutation, {'i': dict(scope, scope_id=str(uuid.uuid4()))})['assignUserRole']
    assert invalid['was_successful'] is False and invalid['errors'][0]['type'] in ['NOT_FOUND', 'ARGUMENT', 'UNPROCESSABLE_ENTITY'], invalid
    print('IAM to EF absent scope rejected', flush=True)
    scoped_assignment = mutation('assignUserRole', scope_mutation, {'i': dict(scope, scope_id=ref['_id'])})['user_role_edge']['assignment_id']
    login()
    party = mutation('createParty', 'mutation($i:CreatePartyInput!){createParty(input:$i){party{_id status} was_successful errors{type message}}}', {'i': {'partner_code': 'E2E-' + uuid.uuid4().hex[:8].upper(), 'kind': 'INDIVIDUAL', 'display_name': 'Employee E2E Worker', 'status': 'ACTIVE'}})['party']
    worker_start = fixture_day.isoformat()
    create_worker = 'mutation($i:CreateWorkerInput!){createWorker(input:$i){worker{id _id status worker_number} was_successful errors{type message}}}'
    worker_input = {
        'worker_number': 'E2E-' + uuid.uuid4().hex[:8].upper(),
        'worker_category': 'employee', 'person_party_reference_id': party['_id'],
        'primary_legal_entity_reference_id': ref['_id'],
        'hire_date': worker_start, 'service_start_date': worker_start,
    }
    before_incorporation = (fixture_day - datetime.timedelta(days=1)).isoformat()
    invalid_worker = graphql(create_worker, {'i': dict(worker_input, hire_date=before_incorporation)})['createWorker']
    assert not invalid_worker['was_successful'] and invalid_worker['errors'][0]['type'] == 'CONFLICT', invalid_worker
    print('HRM rejected worker hire before legal entity incorporation', flush=True)
    worker = mutation('createWorker', create_worker, {'i': worker_input})['worker']
    assert worker['status'] == 'preboarding'
    reissue_query = """mutation($worker: ID!, $result: ID!) {
      reissueEmployeePayslip(worker_id: $worker, payroll_result_id: $result) {
        was_successful errors { type } payslip_pdf { content_base64 }
      }
    }"""
    reissue_vars = {"worker": worker["id"], "result": str(uuid.uuid4())}
    def expect_reissue_failure(error_type):
        reply = graphql(reissue_query, reissue_vars)["reissueEmployeePayslip"]
        assert not reply["was_successful"] and reply["payslip_pdf"] is None, reply
        assert reply["errors"][0]["type"] == error_type, reply
    expect_reissue_failure("AUTHORIZATION")
    provision = 'mutation($i:CreateEmployeeUserInput!){createEmployeeUser(input:$i){user{id _id status kind} was_successful errors{type message}}}'
    username = 'e2eemp' + uuid.uuid4().hex[:8]
    employee_input = {'worker_id': worker['id'], 'full_name': 'Employee Fixture', 'email': username + '@integrione.test', 'username': username, 'password': uuid.uuid4().hex + 'Z9!'}
    import subprocess
    db = os.environ['HRM_E2E_DB_NAME']
    assert db.startswith('hrm_payload_') and db.endswith('_test')
    worker_id = str(uuid.UUID(worker['_id']))
    account_request = 'mutation($i:RequestEmployeeAccountInput!){requestEmployeeAccount(input:$i){request{id iam_request_id status} was_successful errors{type message}}}'
    request_input = {'worker_id': worker['id']}
    no_event = graphql(account_request, {'i': request_input})['requestEmployeeAccount']
    assert no_event['was_successful'] is False and no_event['errors'][0]['type'] == 'CONFLICT', no_event
    print('HRM rejects provisioning without verified hire event', flush=True)
    activation = (
        'mutation($id:ID!){activateWorker(id:$id)'
        '{worker{status} was_successful errors{type message}}}'
    )
    activation_vars = {'id': worker['id']}
    before_mfa = graphql(activation, activation_vars)['activateWorker']
    assert not before_mfa['was_successful'] and before_mfa['errors'][0]['type'] == 'AUTHORIZATION'
    import time, hmac, hashlib, base64, struct
    from urllib.parse import urlsplit, parse_qs

    def finish_mfa(identifier, password):
        global csrf
        secret_uri = mutation('enrollMyTotp', 'mutation {enrollMyTotp{otp_auth_uri was_successful errors{type}}}')['otp_auth_uri']
        secret = parse_qs(urlsplit(secret_uri).query)['secret'][0]
        key = base64.b32decode(secret + '=' * (-len(secret) % 8), casefold=True)

        def otp(epoch):
            digest = hmac.new(key, struct.pack('>Q', int(epoch) // 30), hashlib.sha1).digest()
            offset = digest[-1] & 15
            return str((int.from_bytes(digest[offset:offset + 4], 'big') & 2147483647) % 1000000).zfill(6)
        epoch = time.time()
        mutation('confirmMyTotp', 'mutation($i:ConfirmTotpInput!){confirmMyTotp(input:$i){was_successful errors{type}}}', {'i': {'code': otp(epoch - 30)}})
        try:
            logout()
        except RuntimeError:
            pass
        csrf = graphql('query {browserSession{csrfToken}}')['browserSession']['csrfToken']
        auth = mutation('userLogin', 'mutation($i:UserLoginInput!){userLogin(input:$i){status mfaChallenge csrfToken was_successful errors{type}}}', {'i': {'identifier': identifier, 'password': password}})
        assert auth['status'] == 'MFA_REQUIRED', auth['status']
        csrf = auth['csrfToken'] or csrf
        verified = mutation('validateUserMfa', 'mutation($i:ValidateUserMfaInput!){validateUserMfa(input:$i){csrfToken was_successful errors{type}}}', {'i': {'challenge': auth['mfaChallenge'], 'totpCode': otp(time.time())}})
        csrf = verified['csrfToken'] or csrf
    finish_mfa(os.getenv('E2E_SUPERADMIN_USERNAME') or env.get('IAM_SUPERADMIN_USERNAME', 'superadmin'), os.getenv('E2E_SUPERADMIN_PASSWORD') or env['IAM_SUPERADMIN_PASSWORD'])
    expect_reissue_failure("UNAVAILABLE")
    print("Scoped payslip reissue requires MFA and fails closed without Finance and DRM", flush=True)
    without_contract = graphql(activation, activation_vars)['activateWorker']
    assert not without_contract['was_successful'] and without_contract['errors'][0]['type'] == 'CONFLICT'
    entity_id = str(uuid.UUID(ref['_id']))
    # DRM belum tersedia. Fixture menyuplai bukti kontrak yang sudah ditandatangani.
    sql = f"""INSERT INTO hr_employment_contracts
        (id, worker_id, contract_number, contract_type, contract_status,
         employing_legal_entity_reference_id, effective_from, signed_at,
         signed_document_reference_id, signed_document_version_reference_id,
         signed_content_digest, work_country_code, governing_jurisdiction_code)
        VALUES ('{uuid.uuid4()}', '{worker_id}', 'E2E-{uuid.uuid4()}', 'permanent',
                'signed', '{entity_id}', current_date - 2, now(), '{uuid.uuid4()}',
                '{uuid.uuid4()}', decode(repeat('ab', 32), 'hex'), 'ID', 'ID')"""
    subprocess.run([
        'docker', 'exec', 'integrione-retail-postgres-1', 'sh', '-c',
        'psql -Xq -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$1" -c "$2"',
        'sh', db, sql,
    ], check=True)
    activated = mutation('activateWorker', activation, activation_vars)
    assert activated['worker']['status'] == 'active'
    first = mutation('requestEmployeeAccount', account_request, {'i': request_input})['request']
    again = mutation('requestEmployeeAccount', account_request, {'i': request_input})['request']
    assert first == again and first['iam_request_id']
    print('Provisioning accepted the hire event generated by worker activation', flush=True)
    hrm_status = 'query($id:ID!){employeeAccountRequest(id:$id){id iam_request_id status}}'
    queue = graphql('query {employeeProvisioningRequests(first:1){totalCount pageInfo{hasNextPage} nodes{id worker_id worker_number legal_entity_id status}}}')['employeeProvisioningRequests']
    assert queue['totalCount'] == 1 and (not queue['pageInfo']['hasNextPage']), queue
    expected = {'id': first['iam_request_id'], 'worker_id': worker['id'], 'worker_number': worker['worker_number'], 'legal_entity_id': ref['_id'], 'status': 'pending'}
    assert queue['nodes'] == [expected], queue
    receipt = graphql('query($id:ID!){employeeProvisioningRequest(id:$id){id worker_id worker_number legal_entity_id status}}', {'id': first['iam_request_id']})['employeeProvisioningRequest']
    assert receipt == expected, receipt
    print('IAM queue and detail match submitted worker', flush=True)
    queued_input = dict(employee_input, provisioning_request_id=receipt['id'], worker_id=receipt['worker_id'])
    typed_receipt = 'gid://integrione/employee-provisioning-request/' + receipt['id']
    receipt_from_gid = graphql(
        'query($id:ID!){employeeProvisioningRequest(id:$id){id worker_id worker_number legal_entity_id status}}',
        {'id': typed_receipt},
    )['employeeProvisioningRequest']
    assert receipt_from_gid == receipt, receipt_from_gid
    wrong_receipt_type = 'gid://integrione/user/' + receipt['id']
    invalid_receipt = graphql(provision, {'i': dict(queued_input, provisioning_request_id=wrong_receipt_type)})['createEmployeeUser']
    assert not invalid_receipt['was_successful'] and invalid_receipt['errors'][0]['type'] == 'ARGUMENT', invalid_receipt
    try:
        graphql('query($id:ID!){employeeProvisioningRequest(id:$id){id}}', {'id': wrong_receipt_type})
    except RuntimeError as exc:
        assert exc.args[0][0][0] == 'INVALID_ARGUMENT', exc
    else:
        raise AssertionError('Receipt lookup accepted an IAM user GID')
    print('Provisioning and receipt query reject GID from another resource', flush=True)
    mismatch = graphql(provision, {'i': dict(queued_input, worker_id=str(uuid.uuid4()))})['createEmployeeUser']
    assert mismatch['was_successful'] is False and mismatch['errors'][0]['type'] == 'CONFLICT', mismatch
    print('mismatched worker and queue receipt rejected', flush=True)
    session_cookie = next((cookie.value for handler in opener.handlers if isinstance(handler, urllib.request.HTTPCookieProcessor) for cookie in handler.cookiejar if cookie.name == '__Host-integrione_session'))
    subprocess.run(['go', 'test', '-count=1', '-tags=e2e', '-run', '^TestHRMEmployeeProvisioningOverHTTPS$', './internal/e2e'], cwd='backend/gateway', check=True, env=dict(os.environ, E2E_HRM_PROVISIONING_URL=base, E2E_HRM_PROVISIONING_WORKER=receipt['worker_id'], E2E_HRM_PROVISIONING_REQUEST=receipt['id'], E2E_HRM_PROVISIONING_SESSION=session_cookie, E2E_HRM_PROVISIONING_USERNAME=queued_input["username"], E2E_HRM_PROVISIONING_PASSWORD=queued_input["password"]))
    employee = mutation('createEmployeeUser', provision, {'i': queued_input})['user']
    replayed = mutation('createEmployeeUser', provision, {'i': dict(queued_input, provisioning_request_id=typed_receipt)})['user']
    assert replayed == employee, (replayed, employee)
    print('Provisioning replay returned the same IAM account', flush=True)
    completed = graphql('query($id:ID!){employeeProvisioningRequest(id:$id){status}}', {'id': receipt['id']})['employeeProvisioningRequest']
    assert completed['status'] == 'linked', completed
    assert graphql('query {employeeProvisioningRequests(first:1){totalCount}}')['employeeProvisioningRequests']['totalCount'] == 0
    print('linked receipt removed from open queue', flush=True)
    hrm_completed = graphql(hrm_status, {'id': first['id']})['employeeAccountRequest']
    assert hrm_completed == dict(first, status='applied'), hrm_completed
    request_node = graphql(
        'query($id:ID!){node(id:$id){__typename ... on EmployeeAccountRequest{id iam_request_id status}}}',
        {'id': first['id']},
    )['node']
    assert request_node == dict(hrm_completed, __typename='EmployeeAccountRequest'), request_node
    print('Relay refetch returned the HRM employee account request', flush=True)
    print('HRM shows applied status after IAM links worker', flush=True)
    linked = subprocess.check_output(['docker', 'exec', 'integrione-retail-postgres-1', 'sh', '-c', 'psql -XAt -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$1" -c "$2"', 'sh', db, f"select iam_principal_reference_id from hr_workers where id='{worker_id}'"], text=True).strip()
    assert linked == employee['_id'], (linked, employee['_id'])
    print('IAM and HRM persisted matching employee principal', flush=True)
    mutation('activateUser','mutation($id:ID!){activateUser(user_id:$id){was_successful errors{type}}}',{'id':employee['id']})
    iam_db=os.environ['HRM_E2E_IAM_DB_NAME']
    principal_id=str(uuid.UUID(employee['_id']))
    subprocess.run(['docker','exec','integrione-retail-postgres-1','sh','-c','psql -XAt -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$1" -c "$2"','sh',iam_db,f"update user_identities set is_verified=true, verified_at=now() where user_id='{principal_id}'"],check=True,stdout=subprocess.DEVNULL)
    admin_opener, admin_csrf=opener,csrf
    def new_browser():
        return urllib.request.build_opener(urllib.request.HTTPSHandler(context=ssl.create_default_context(cafile='backend/deploy/certs/ca.crt')),urllib.request.HTTPCookieProcessor(http.cookiejar.CookieJar()))
    login_query='mutation($i:UserLoginInput!){userLogin(input:$i){status was_successful errors{type message}}}'
    credentials={'i':{'identifier':queued_input['username'],'password':queued_input['password']}}
    try:
        opener=new_browser()
        csrf=graphql('query { browserSession { csrfToken } }')['browserSession']['csrfToken']
        result=graphql(login_query,credentials)['userLogin']
        assert result['was_successful'] and result['status']=='AUTHENTICATED',result
        employee_opener, employee_csrf=opener,csrf
        opener,csrf=admin_opener,admin_csrf
        if automatic_termination:
            plan_id = uuid.uuid4()
            # Persetujuan independen disuplai sebagai fixture awal di database HRM disposable.
            plan_sql = f"""INSERT INTO hr_offboarding_plans
                (id, worker_id, plan_number, reason_code, notice_date, last_working_date,
                termination_effective_date, status, approved_at, approved_by_principal_id,
                created_by_principal_id, created_at)
                VALUES ('{plan_id}', '{worker_id}', 'E2E-{plan_id}', 'resignation',
                '{worker_start}', '{worker_start}', '{worker_start}', 'in_progress',
                now(), '{uuid.uuid4()}', '{uuid.uuid4()}', now() - interval '1 minute')"""
            def read_database(database, statement):
                return subprocess.check_output([
                    'docker', 'exec', 'integrione-retail-postgres-1', 'sh', '-c',
                    'psql -XAt -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$1" -c "$2"',
                    'sh', database, statement,
                ], text=True).strip()
            read_database(db, plan_sql)
            terminated = mutation(
                'terminateWorker',
                'mutation($worker:ID!,$plan:ID!){terminateWorker(id:$worker,offboarding_plan_id:$plan){worker{status} was_successful errors{type message}}}',
                {'worker': worker['id'], 'plan': str(plan_id)},
            )['worker']
            assert terminated['status'] == 'terminated', terminated
            deadline = time.monotonic() + 25
            while time.monotonic() < deadline:
                delivered = read_database(db, f"select count(*) from hr_outbox_events where aggregate_id='{plan_id}' and event_type='hr_worker_terminations.recorded' and published_at is not null")
                if delivered == '1':
                    break
                time.sleep(.5)
            else:
                raise AssertionError('IAM did not acknowledge the HRM termination event')
            status = read_database(iam_db, f"select status from users where id='{principal_id}'")
            assert status == 'disabled', status
            iam_container = json.loads(subprocess.check_output(['docker', 'inspect', 'hrm-isolated-iam'], text=True))[0]
            hrm_container = json.loads(subprocess.check_output(['docker', 'inspect', 'hrm-isolated-server'], text=True))[0]
            iam_ip = iam_container['NetworkSettings']['Networks']['integrione-retail_default']['IPAddress']
            mounts = {mount['Destination']: mount['Source'] for mount in hrm_container['Mounts']}
            clearance = {
                'workerId': worker_id, 'offboardingPlanId': str(plan_id),
                'effectiveDate': worker_start, 'workerPrincipalId': principal_id,
            }
            proof = subprocess.check_output([
                'grpcurl', '-max-time', '8', '-authority', 'iam',
                '-cacert', mounts['/etc/integrione/mandate/ca.crt'],
                '-cert', mounts['/etc/integrione/mandate/client.crt'],
                '-key', mounts['/etc/integrione/mandate/client.key'],
                '-import-path', 'backend/services/iam/contract/proto',
                '-proto', 'iam/v1/employee_termination.proto',
                '-d', json.dumps(clearance), f'{iam_ip}:9011',
                'iam.v1.EmployeeTerminationService/VerifyEmployeeTerminationClearance',
            ], text=True)
            assert json.loads(proof) == clearance
            print('IAM verified disabled-account clearance over mTLS', flush=True)

        else:
            disabled=mutation('disableUser','mutation($id:ID!){disableUser(user_id:$id){user{status} was_successful errors{type}}}',{'id':employee['id']})['user']
            assert disabled['status']=='DISABLED',disabled
        opener,csrf=employee_opener,employee_csrf
        session_deadline = time.monotonic() + (10 if automatic_termination else 0)
        while True:
            try:
                graphql('query { me { id } }')
            except RuntimeError as err:
                assert 'UNAUTHENTICATED' in str(err), err
                break
            if time.monotonic() >= session_deadline:
                raise AssertionError('employee session remained active')
            time.sleep(.25)
        opener=new_browser()
        csrf=graphql('query { browserSession { csrfToken } }')['browserSession']['csrfToken']
        rejected=graphql(login_query,credentials)['userLogin']
        assert not rejected['was_successful'],rejected
        operation = 'automatic worker termination' if automatic_termination else 'IAM disableUser'
        print(f'An employee provisioned through HRM lost the old session and could not log in after {operation}', flush=True)
    finally:
        opener,csrf=admin_opener,admin_csrf

finally:
    try:
        if scoped_assignment:
            mutation('revokeUserRole', 'mutation($id:ID!){revokeUserRole(assignment_id:$id){was_successful errors{type message}}}', {'id': scoped_assignment})
        if assignment:
            mutation('revokeUserRole', 'mutation($id:ID!){revokeUserRole(assignment_id:$id){was_successful errors{type message}}}', {'id': assignment})
        if role:
            for perm in perms:
                mutation('revokeRolePermission', 'mutation($r:ID!,$p:ID!){revokeRolePermission(role_id:$r,permission_id:$p){was_successful errors{type message}}}', {'r': role, 'p': perm})
            mutation('deleteRole', 'mutation($id:ID!){deleteRole(id:$id){was_successful errors{type message}}}', {'id': role})
    finally:
        logout()
