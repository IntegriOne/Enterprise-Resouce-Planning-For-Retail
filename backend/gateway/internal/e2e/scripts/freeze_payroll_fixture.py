"""Uji E2E HTTPS untuk freezePayrollInput dan createPayrollRun di stack disposable."""
import base64
import datetime
import hashlib
import hmac
import http.cookiejar
import json
import os
import ssl
import struct
import subprocess
import time
import urllib.error
import urllib.request
import uuid
from pathlib import Path
from urllib.parse import parse_qs, urlsplit

BASE = 'https://erp.localhost:18443'
CTX = ssl.create_default_context(cafile='backend/deploy/certs/ca.crt')
OPENER = urllib.request.build_opener(
    urllib.request.HTTPSHandler(context=CTX),
    urllib.request.HTTPCookieProcessor(http.cookiejar.CookieJar()),
)
CSRF = ''


def graphql_raw(query, variables=None, include_csrf=True, opener=OPENER):
    global CSRF
    headers = {'Content-Type': 'application/json', 'Origin': 'https://erp.localhost:8443'}
    if include_csrf and CSRF:
        headers['X-CSRF-Token'] = CSRF
    req = urllib.request.Request(BASE + '/api/graphql', json.dumps({'query': query, 'variables': variables}).encode(), headers)
    try:
        resp = opener.open(req, timeout=10)
    except urllib.error.HTTPError as err:
        resp = err
    with resp:
        if opener is OPENER:
            CSRF = resp.headers.get('X-CSRF-Token') or CSRF
        return json.load(resp)


def graphql(query, variables=None):
    res = graphql_raw(query, variables)
    if res.get('errors'):
        raise RuntimeError([(e.get('extensions', {}).get('code'), e.get('message')) for e in res['errors']])
    return res['data']


def mutation(name, query, variables=None):
    out = graphql(query, variables)[name]
    if not out['was_successful']:
        raise RuntimeError((name, out['errors']))
    return out


env = {}
for line in Path('backend/services/iam/.env').read_text().splitlines():
    k, sep, v = line.partition('=')
    if sep:
        env[k] = v.strip().strip('"').strip("'")

USERNAME = os.getenv('E2E_SUPERADMIN_USERNAME') or env.get('IAM_SUPERADMIN_USERNAME', 'superadmin')
PASSWORD = os.getenv('E2E_SUPERADMIN_PASSWORD') or env['IAM_SUPERADMIN_PASSWORD']


def totp(secret, epoch=None):
    if epoch is None:
        epoch = time.time()
    key = base64.b32decode(secret.upper() + '=' * (-len(secret) % 8), casefold=True)
    counter = int(epoch) // 30
    digest = hmac.new(key, struct.pack('>Q', counter), hashlib.sha1).digest()
    offset = digest[-1] & 0x0F
    code = (struct.unpack('>I', digest[offset:offset + 4])[0] & 0x7FFFFFFF) % 1000000
    return f'{code:06d}'


def login_base():
    global CSRF
    CSRF = graphql('query { browserSession { csrfToken } }')['browserSession']['csrfToken']
    res = mutation(
        'userLogin',
        'mutation($i:UserLoginInput!){userLogin(input:$i){status csrfToken was_successful errors{type message}}}',
        {'i': {'identifier': USERNAME, 'password': PASSWORD}},
    )
    assert res['status'] == 'AUTHENTICATED', res
    CSRF = res['csrfToken'] or CSRF


def login_with_mfa():
    global CSRF
    secret_uri = mutation('enrollMyTotp', 'mutation {enrollMyTotp{otp_auth_uri was_successful errors{type message}}}')['otp_auth_uri']
    secret = parse_qs(urlsplit(secret_uri).query)['secret'][0]
    mutation(
        'confirmMyTotp',
        'mutation($i:ConfirmTotpInput!){confirmMyTotp(input:$i){was_successful errors{type message}}}',
        {'i': {'code': totp(secret, time.time() - 30)}},
    )
    try:
        mutation('userLogout', 'mutation { userLogout { was_successful errors{type} } }')
    except RuntimeError:
        pass
    CSRF = graphql('query { browserSession { csrfToken } }')['browserSession']['csrfToken']
    auth = mutation(
        'userLogin',
        'mutation($i:UserLoginInput!){userLogin(input:$i){status mfaChallenge csrfToken was_successful errors{type message}}}',
        {'i': {'identifier': USERNAME, 'password': PASSWORD}},
    )
    assert auth['status'] == 'MFA_REQUIRED', auth
    CSRF = auth['csrfToken'] or CSRF
    verified = mutation(
        'validateUserMfa',
        'mutation($i:ValidateUserMfaInput!){validateUserMfa(input:$i){csrfToken was_successful errors{type message}}}',
        {'i': {'challenge': auth['mfaChallenge'], 'totpCode': totp(secret)}},
    )
    CSRF = verified['csrfToken'] or CSRF


def psql_exec(db_name, sql):
    subprocess.run(
        ['docker', 'exec', 'integrione-retail-postgres-1', 'sh', '-c', 'psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$1" -c "$2"', 'sh', db_name, sql],
        check=True,
        stdout=subprocess.DEVNULL,
    )


db_name = os.environ['HRM_E2E_DB_NAME']
assert db_name.startswith('hrm_payload_') and db_name.endswith('_test')

login_base()
principal = graphql('query { browserSession { principalId } }')['browserSession']['principalId']
role_id = None
assignment_id = None

try:
    # Siapkan role & permission lewat GraphQL IAM.
    perm_ids = []
    for code in ('human_resource.payroll_runs.create', 'human_resource.payroll_runs.read'):
        rows = graphql('query($q:String!){permissions(query:$q,first:10){nodes{id code}}}', {'q': code})['permissions']['nodes']
        perm_ids.append(next(x['id'] for x in rows if x['code'] == code))
    role_id = mutation(
        'createRole',
        'mutation($i:CreateRoleInput!){createRole(input:$i){role{id} was_successful errors{type message}}}',
        {'i': {'code': 'hrm_freeze_e2e_' + uuid.uuid4().hex[:8], 'name': 'HRM Freeze E2E', 'management_domain': 'human_resource', 'default_redirect_path': '/identity', 'redirect_priority': 999}},
    )['role']['id']
    for pid in perm_ids:
        mutation('assignRolePermission', 'mutation($r:ID!,$p:ID!){assignRolePermission(role_id:$r,permission_id:$p){was_successful errors{type message}}}', {'r': role_id, 'p': pid})
    assignment_id = mutation(
        'assignUserRole',
        'mutation($i:AssignUserRoleInput!){assignUserRole(input:$i){user_role_edge{assignment_id} was_successful errors{type message}}}',
        {'i': {'user_id': principal, 'role_id': role_id, 'scope_type': 'GLOBAL'}},
    )['user_role_edge']['assignment_id']

    # Buat LegalEntity di EF lewat GraphQL.
    today = datetime.datetime.now(datetime.timezone.utc).date()
    incorp = (today - datetime.timedelta(days=30)).isoformat()
    legal = mutation(
        'createLegalEntity',
        'mutation($i:CreateLegalEntityInput!){createLegalEntity(input:$i){legal_entity{_id status} was_successful errors{type message}}}',
        {'i': {'entity_code': 'FRZ-' + uuid.uuid4().hex[:8].upper(), 'legal_name': 'Freeze E2E Entity', 'jurisdiction_country_code': 'ID', 'status': 'ACTIVE', 'incorporation_date': incorp + 'T00:00:00Z'}},
    )['legal_entity']
    legal_id = legal['_id']

    # Siapkan fixture grup, periode terbuka + kosong, rule set, worker, dan profil payroll di DB disposable.
    group_id = str(uuid.uuid4())
    period_id = str(uuid.uuid4())
    empty_period_id = str(uuid.uuid4())
    rule_set_id = str(uuid.uuid4())
    artifact_doc_id = str(uuid.uuid4())
    worker_id = str(uuid.uuid4())
    person_id = str(uuid.uuid4())
    profile_id = str(uuid.uuid4())
    bank_tok_id = str(uuid.uuid4())
    digest_hex = '11' * 32
    p_start = today.replace(day=1).isoformat()
    p_end = (today.replace(day=1) + datetime.timedelta(days=27)).isoformat()
    psql_exec(db_name, f"""
        INSERT INTO hr_payroll_groups (id, group_code, group_name, employing_legal_entity_reference_id, jurisdiction_code, currency_code, currency_scale, pay_frequency, timezone, is_active)
        VALUES ('{group_id}', 'GRP-{group_id[:8]}', 'Freeze E2E Group', '{legal_id}', 'ID', 'IDR', 2, 'monthly', 'Asia/Jakarta', true);
        INSERT INTO hr_payroll_periods (id, payroll_group_id, employing_legal_entity_reference_id, currency_code, jurisdiction_code, period_code, period_start, period_end, pay_date, period_status)
        VALUES
          ('{period_id}', '{group_id}', '{legal_id}', 'IDR', 'ID', '2026-11', '{p_start}', '{p_end}', '{p_end}', 'open'),
          ('{empty_period_id}', '{group_id}', '{legal_id}', 'IDR', 'ID', '2026-12', '{p_start}', '{p_end}', '{p_end}', 'closed');
        INSERT INTO hr_payroll_rule_sets (id, rule_code, rule_version, jurisdiction_code, artifact_document_version_reference_id, artifact_digest, legal_basis_document_version_reference_id, effective_from, approved_by_principal_id, approved_at)
        VALUES ('{rule_set_id}', 'RULE-{rule_set_id[:8]}', '1.0', 'ID', '{artifact_doc_id}', decode('{digest_hex}', 'hex'), '{artifact_doc_id}', '{incorp}', '{principal}', NOW() - interval '1 day');
        INSERT INTO hr_workers (id, worker_number, person_party_reference_id, worker_category, status, primary_legal_entity_reference_id, iam_principal_reference_id, hire_date, service_start_date)
        VALUES ('{worker_id}', 'FRZW-{worker_id[:8]}', '{person_id}', 'employee', 'active', '{legal_id}', '{principal}', '{incorp}', '{incorp}');
        INSERT INTO hr_payroll_profiles (id, worker_id, payroll_country_code, payroll_jurisdiction_code, payroll_group_id, employing_legal_entity_reference_id, currency_code, bank_account_token_reference_id, effective_from, is_active)
        VALUES ('{profile_id}', '{worker_id}', 'ID', 'ID', '{group_id}', '{legal_id}', 'IDR', '{bank_tok_id}', '{incorp}', true);
    """)

    FREEZE_Q = '''mutation($i:FreezePayrollInputInput!){
      freezePayrollInput(input:$i){
        was_successful
        errors{type message}
        frozen_payroll_input{
          input_snapshot_reference_id payroll_period_id payroll_group_id
          employing_legal_entity_reference_id jurisdiction_code currency_code
          input_digest_sha256 frozen_at
        }
      }
    }'''
    CREATE_Q = '''mutation($i:CreatePayrollRunInput!){
      createPayrollRun(input:$i){
        was_successful
        errors{type message}
        payroll_run{id payroll_period_id run_number status}
      }
    }'''
    idem_key = str(uuid.uuid4())

    # TC-03: Tanpa MFA ditolak AUTHORIZATION.
    no_mfa = graphql(FREEZE_Q, {'i': {'payroll_period_id': period_id, 'idempotency_key': idem_key}})['freezePayrollInput']
    assert not no_mfa['was_successful'] and no_mfa['errors'][0]['type'] == 'AUTHORIZATION', no_mfa

    # Login dengan MFA, lalu TC-03 tanpa CSRF ditolak FORBIDDEN di level GraphQL guard.
    login_with_mfa()
    no_csrf = graphql_raw(FREEZE_Q, {'i': {'payroll_period_id': period_id, 'idempotency_key': idem_key}}, include_csrf=False)
    assert no_csrf.get('errors') and no_csrf['errors'][0]['extensions']['code'] == 'FORBIDDEN', no_csrf

    # TC-02: Periode tertutup / tidak layak ditolak CONFLICT.
    closed = graphql(FREEZE_Q, {'i': {'payroll_period_id': empty_period_id, 'idempotency_key': str(uuid.uuid4())}})['freezePayrollInput']
    assert not closed['was_successful'] and closed['errors'][0]['type'] == 'CONFLICT', closed

    # TC-01 & TC-04 & TC-05: Freeze lalu replay dengan kunci sama -> metadata identik & bebas token/isi payroll.
    res1 = mutation('freezePayrollInput', FREEZE_Q, {'i': {'payroll_period_id': period_id, 'idempotency_key': idem_key}})
    snap1 = res1['frozen_payroll_input']
    raw_dump = json.dumps(res1)
    assert bank_tok_id not in raw_dump and 'salaries' not in raw_dump, raw_dump
    res2 = mutation('freezePayrollInput', FREEZE_Q, {'i': {'payroll_period_id': period_id, 'idempotency_key': idem_key}})
    snap2 = res2['frozen_payroll_input']
    assert snap1['input_snapshot_reference_id'] == snap2['input_snapshot_reference_id']
    assert snap1['input_digest_sha256'] == snap2['input_digest_sha256']

    # TC-02: Digest salah ditolak saat createPayrollRun.
    bad_run = graphql(CREATE_Q, {'i': {
        'payroll_period_id': period_id, 'payroll_group_id': group_id,
        'employing_legal_entity_reference_id': legal_id, 'currency_code': 'IDR',
        'rule_set_id': rule_set_id, 'run_number': 1, 'run_type': 'regular',
        'jurisdiction_code': 'ID', 'rule_engine_name': 'local', 'rule_engine_version': '1',
        'input_snapshot_reference_id': snap1['input_snapshot_reference_id'],
        'input_digest_sha256': 'aa' * 32, 'idempotency_key': str(uuid.uuid4()),
    }})['createPayrollRun']
    assert not bad_run['was_successful'] and bad_run['errors'][0]['type'] == 'CONFLICT', bad_run

    # TC-01: Buat draf memakai snapshot ID & digest dari freezePayrollInput.
    good_run = mutation('createPayrollRun', CREATE_Q, {'i': {
        'payroll_period_id': snap1['payroll_period_id'], 'payroll_group_id': snap1['payroll_group_id'],
        'employing_legal_entity_reference_id': snap1['employing_legal_entity_reference_id'],
        'currency_code': snap1['currency_code'], 'rule_set_id': rule_set_id, 'run_number': 1,
        'run_type': 'regular', 'jurisdiction_code': snap1['jurisdiction_code'],
        'rule_engine_name': 'local', 'rule_engine_version': '1',
        'input_snapshot_reference_id': snap1['input_snapshot_reference_id'],
        'input_digest_sha256': snap1['input_digest_sha256'], 'idempotency_key': str(uuid.uuid4()),
    }})
    assert good_run['payroll_run']['status'] == 'draft', good_run
    print('HTTPS freezePayrollInput + createPayrollRun E2E OK', flush=True)
finally:
    if assignment_id:
        try:
            graphql('mutation($id:ID!){revokeUserRole(assignment_id:$id){was_successful}}', {'id': assignment_id})
        except Exception:
            pass
    if role_id:
        try:
            graphql('mutation($id:ID!){deleteRole(id:$id){was_successful}}', {'id': role_id})
        except Exception:
            pass
    try:
        graphql('mutation { disableMyTotp { was_successful } }')
    except Exception:
        pass
    try:
        graphql('mutation { userLogout { was_successful } }')
    except Exception:
        pass
