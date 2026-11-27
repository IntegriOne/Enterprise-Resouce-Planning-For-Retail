"""Menjalankan uji provisioning HTTPS pada stack disposable."""
from argparse import ArgumentParser
from pathlib import Path
import os
import json
import shutil
import subprocess
import tempfile
import time
import uuid
from urllib.parse import urlsplit, urlunsplit
scripts = Path(__file__).resolve().parent
repo = scripts.parents[4]
parser = ArgumentParser(description=__doc__)
parser.add_argument(
    '--source-root', type=Path,
    help='Root kandidat untuk membangun binary dengan GOWORK=off. Fixture memakai checkout ini.',
)
parser.add_argument('--termination', action='store_true', help='Uji terminasi memakai sertifikat workload dari overlay development.')
parser.add_argument('--guards', action='store_true', help='Uji guard semua operasi HRM pada stack disposable.')
parser.add_argument('--freeze-payroll', action='store_true', help='Uji E2E HTTPS pembekuan input payroll dan pembuatan draf.')
args = parser.parse_args()
if args.termination and shutil.which("grpcurl") is None:
    parser.error("--termination requires grpcurl for the private IAM clearance check")
source_root = args.source_root.resolve() if args.source_root else repo
build_env = dict(os.environ, GOWORK='off') if args.source_root else os.environ
modules = (
    ('services/enterprise_foundation', 'ef-current'),
    ('services/iam', 'iam-current'),
    ('services/human_resource', 'hrm-current'),
    ('gateway', 'gateway-current'),
)
for module, _ in modules:
    if not (source_root / 'backend' / module / 'go.mod').is_file():
        parser.error(f'Modul Go {module} tidak ditemukan di {source_root}')
os.chdir(repo)
artifacts = tempfile.TemporaryDirectory(prefix='erp-hrm-e2e-')
root = Path(artifacts.name)
db = 'hrm_payload_' + uuid.uuid4().hex[:10] + '_test'
ef_db = 'ef_' + db
iam_db = 'iam_' + db

def docker(*args):
    return subprocess.check_output(['docker', *args], text=True)

def inspect(name):
    return json.loads(docker('inspect', name))[0]

workload_mounts = ('/etc/integrione/mandate/', '/etc/integrione/termination/')
workload_env = ('HRM_IAM_MANDATE_', 'IAM_REQUISITION_MANDATE_', 'HRM_TERMINATION_ACCESS_', 'IAM_HRM_TERMINATION_ACCESS_')

def mounts(source, prefix='/'):
    return [
        item for m in source['Mounts']
        if m['Destination'].startswith(prefix) and not m['Destination'].startswith(workload_mounts)
        for item in ('-v', f"{m['Source']}:{m['Destination']}:ro")
    ]
started = []
created = []

def spawn(name, source, extra, env_override):
    with tempfile.NamedTemporaryFile(mode='w', dir=root) as env:
        env.writelines(
            f'{line}\n' for line in source['Config']['Env']
            if line.partition('=')[0] not in env_override and not line.startswith(workload_env)
        )
        env.writelines((f'{key}={value}\n' for key, value in env_override.items()))
        env.flush()
        docker('run', '-d', '--name', name, '--network', 'integrione-retail_default', '--env-file', env.name, *extra, source['Config']['Image'])
    started.append(name)
try:
    for module, binary in modules:
        subprocess.run(['go', 'build', '-o', str(root / binary), './cmd'],
                       cwd=source_root / 'backend' / module, env=build_env, check=True)
    (root / 'Caddyfile').write_text('erp.localhost {\n    tls /etc/caddy/certs/erp.localhost.crt /etc/caddy/certs/erp.localhost.key\n    reverse_proxy /api/graphql hrm-isolated-gateway:8080\n}\n')
    docker('exec', 'integrione-retail-postgres-1', 'sh', '-c', 'createdb -U "$POSTGRES_USER" -O human_resource_svc "$1"', 'sh', db)
    created.append(db)
    docker('exec', 'integrione-retail-postgres-1', 'sh', '-c', 'createdb -U "$POSTGRES_USER" -O enterprise_foundation_svc "$1"', 'sh', ef_db)
    created.append(ef_db)
    docker('exec', 'integrione-retail-postgres-1', 'sh', '-c', 'createdb -U "$POSTGRES_USER" -O iam_management_svc "$1"', 'sh', iam_db)
    created.append(iam_db)
    docker('run', '-d', '--name', 'hrm-isolated-redis', '--network', 'integrione-retail_default', '--tmpfs', '/data', '-v', f'{repo}/backend/deploy/certs:/etc/redis/certs:ro', '-v', f'{repo}/backend/deploy/redis/redis.conf:/etc/redis/redis.conf:ro', '-v', f'{repo}/backend/deploy/redis/users.acl:/etc/redis/users.acl:ro', 'redis:latest', 'redis-server', '/etc/redis/redis.conf')
    started.append('hrm-isolated-redis')
    ef = inspect('integrione-retail-enterprise-foundation-1')
    original = urlsplit(next((value.partition('=')[2] for value in ef['Config']['Env'] if value.startswith('EF_DATABASE_DSN='))))
    spawn('hrm-isolated-ef', ef, [*mounts(ef), '-v', f'{root}/ef-current:/usr/local/bin/enterprise-foundation:ro'], {'EF_DATABASE_DSN': urlunsplit(original._replace(path='/' + ef_db))})
    ef_target = inspect('hrm-isolated-ef')['NetworkSettings']['Networks']['integrione-retail_default']['IPAddress'] + ':9002'
    iam = inspect('integrione-retail-iam-1')
    iam_dsn = urlsplit(next((value.partition('=')[2] for value in iam['Config']['Env'] if value.startswith('IAM_DATABASE_DSN='))))
    hrm = inspect('integrione-retail-human-resource-1')

    def termination_config(source, prefix, target, required_mounts):
        if not args.termination:
            return [], {}
        env = dict(line.split('=', 1) for line in source['Config']['Env'] if line.startswith(prefix))
        if not env:
            raise SystemExit('enable the development mandate and termination overlays before --termination')
        env.update(target)
        vols = [
            item for mount in source['Mounts'] if mount['Destination'].startswith(required_mounts)
            for item in ('-v', f"{mount['Source']}:{mount['Destination']}:ro")
        ]
        return vols, env

    iam_vols, iam_termination = termination_config(
        iam, ('IAM_REQUISITION_MANDATE_', 'IAM_HRM_TERMINATION_ACCESS_'),
        {'IAM_HRM_TERMINATION_ACCESS_TARGET': 'hrm-isolated-server:9012'}, workload_mounts,
    )
    hrm_vols, hrm_termination = termination_config(
        hrm, ('HRM_IAM_MANDATE_', 'HRM_TERMINATION_ACCESS_'),
        {'HRM_IAM_MANDATE_TARGET': 'hrm-isolated-iam:9011'}, workload_mounts,
    )
    spawn('hrm-isolated-iam', iam, [*mounts(iam), *iam_vols, '-v', f'{root}/iam-current:/usr/local/bin/iam-server:ro'], {'IAM_DATABASE_DSN': urlunsplit(iam_dsn._replace(path='/' + iam_db)), 'IAM_REDIS_ADDR': 'hrm-isolated-redis:6379', 'IAM_EF_GRPC_ENABLED': 'true', 'IAM_EF_GRPC_TARGET': ef_target, **iam_termination})
    spawn('hrm-isolated-server', hrm, [*mounts(hrm), *hrm_vols, '-v', f'{root}/hrm-current:/usr/local/bin/hrm-server:ro'], {'HRM_DB_NAME': db, 'HRM_REDIS_ADDR': 'hrm-isolated-redis:6379', 'HRM_EF_GRPC_TARGET': ef_target, **hrm_termination})
    iam_target = inspect('hrm-isolated-iam')['NetworkSettings']['Networks']['integrione-retail_default']['IPAddress'] + ':9001'
    hrm_target = inspect('hrm-isolated-server')['NetworkSettings']['Networks']['integrione-retail_default']['IPAddress'] + ':9003'
    gw = inspect('integrione-retail-gateway-1-1')
    spawn('hrm-isolated-gateway', gw, [*mounts(gw), '-v', f'{root}/gateway-current:/usr/local/bin/gateway-server:ro'], {'HRM_GRPC_ENABLED': 'true', 'GATEWAY_RATE_LIMIT_MUTATION_MAX': '150', 'GATEWAY_ALLOWED_ORIGINS': 'https://erp.localhost:8443,https://erp.localhost:18443', 'IAM_GRPC_TARGET': iam_target, 'GATEWAY_REDIS_ADDR': 'hrm-isolated-redis:6379', 'HRM_GRPC_TARGET': hrm_target, 'EF_GRPC_TARGET': ef_target})
    caddy = inspect('integrione-retail-caddy-1')
    spawn('hrm-isolated-caddy', caddy, ['-p', '127.0.0.1:18443:443', '-v', f'{root}/Caddyfile:/etc/caddy/Caddyfile:ro', *mounts(caddy, '/etc/caddy/certs/')], {})
    for _ in range(60):
        if all((inspect(name)['State'].get('Health', {}).get('Status') == 'healthy' for name in ('hrm-isolated-server', 'hrm-isolated-ef', 'hrm-isolated-iam', 'hrm-isolated-gateway'))):
            break
        time.sleep(0.5)
    else:
        raise SystemExit('isolated HRM server failed readiness')
    if args.guards:
        subprocess.run(['python3', str(scripts / 'check_hrm_schema_guards.py'), '--port', '18443'], check=True)
    if args.freeze_payroll:
        subprocess.run(['python3', str(scripts / 'freeze_payroll_fixture.py')], check=True, env=dict(os.environ, HRM_E2E_DB_NAME=db))
    subprocess.run(['python3', str(scripts / 'provisioning_fixture.py')], check=True, env=dict(os.environ, HRM_E2E_DB_NAME=db, HRM_E2E_IAM_DB_NAME=iam_db, HRM_E2E_TERMINATION='1' if args.termination else '0'))
    count = docker('exec', 'integrione-retail-postgres-1', 'sh', '-c', 'psql -XAt -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$1" -c "$2"', 'sh', iam_db, "select count(*) from users where display_name = 'Employee Fixture'").strip()
    assert count == '1', count
    print('IAM persisted exactly one fixture account')
    receipts = docker('exec', 'integrione-retail-postgres-1', 'sh', '-c', 'psql -XAt -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$1" -c "$2"', 'sh', iam_db, "select count(*) from employee_provisioning_requests where status = 'linked' and account_user_id is not null").strip()
    assert receipts == '1', receipts
    print('IAM persisted exactly one provisioning receipt')
finally:
    if started:
        subprocess.run(['docker', 'rm', '-f', *reversed(started)], check=False, stdout=subprocess.DEVNULL)
    for name in created:
        docker('exec', 'integrione-retail-postgres-1', 'sh', '-c', 'dropdb -U "$POSTGRES_USER" "$1"', 'sh', name)
    artifacts.cleanup()
