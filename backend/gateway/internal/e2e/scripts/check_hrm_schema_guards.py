"""Memeriksa kontrak SDL serta guard sesi dan CSRF Human Resource."""

from pathlib import Path
from concurrent.futures import ThreadPoolExecutor
import json
import re
import ssl
import time
import urllib.error
import urllib.request

repo = Path(__file__).resolve().parents[5]
root = repo / 'backend/gateway/graphql/human_resource'
schema = '\n'.join(p.read_text() for p in sorted(root.glob('*.graphql')))
types = dict(re.findall(r'^input (\w+) \{\n(.*?)^\}', schema, re.M | re.S))
objects = {
    name: (interface, body)
    for name, interface, body in re.findall(
        r'^type (\w+)(?: implements (\w+))? \{\n(.*?)^\}', schema, re.M | re.S
    )
}
enums = dict(re.findall(r'^enum (\w+) \{\s*(\w+)', schema, re.M))
fixtures = {'ID':'"00000000-0000-4000-8000-000000000001"', 'String':'"fixture"', 'Int':'1', 'Float':'1.0', 'Boolean':'true', **enums}
fixtures["Time"] = json.dumps("2026-11-27T00:00:00Z")
tls = ssl.create_default_context(cafile=repo / 'backend/deploy/certs/ca.crt')
url = 'https://erp.localhost:8443/api/graphql'

def value(kind):
    if kind.startswith('['):
        return '[' + value(kind[1:-1].rstrip('!')) + ']'
    if kind in types:
        return '{' + required(types[kind]) + '}'
    return fixtures[kind]

def required(fields):
    return ', '.join(f'{name}: {value(kind)}' for name, kind in re.findall(
        r'(?:^|\n|,)\s*(\w+):\s*([\w\[\]!]+)!(?![\w!\]])(?![ \t]*=)', fields))

def probe(case):
    operation, name, args, result = case
    encoded = required(args)
    field = name + (f'({encoded})' if encoded else '')
    if result.strip("[]!") in objects:
        field += ' { __typename }'
    request = urllib.request.Request(url, json.dumps({'query': f'{operation.lower()} {{ {field} }}'}).encode(),
        {'Content-Type':'application/json','Origin':'https://erp.localhost:8443'})
    try:
        response = urllib.request.urlopen(request, context=tls, timeout=10)
    except urllib.error.HTTPError as err:
        response = err
    with response:
        payload = json.load(response)
    errors = payload.get('errors', [])
    expected = 'UNAUTHENTICATED' if operation == 'Query' else 'FORBIDDEN'
    ok = len(errors) == 1 and errors[0].get('extensions', {}).get('code') == expected
    if operation == 'Mutation':
        ok = ok and 'missing X-CSRF-Token header' in errors[0]['message']
    return (operation, name, ok, errors)

cases = []
for operation, body in re.findall(r'^extend type (Query|Mutation) \{\n(.*?)^\}', schema, re.M | re.S):
    for line in body.splitlines():
        line = line.strip()
        if not line or line.startswith('#'):
            continue
        field = re.fullmatch(r'(\w+)(?:\(([^)]*)\))?: ([\w\[\]!]+)', line)
        if field is None:
            raise SystemExit(f'Could not enumerate {operation} field: {line}')
        name, args, result = field.groups()
        cases.append((operation, name, args or '', result))
if not cases or {case[0] for case in cases} != {'Query', 'Mutation'}:
    raise SystemExit('Could not enumerate HRM schema Query and Mutation fields')
def require_fields(name, body, expected):
    fields = dict(re.findall(r'^\s*(\w+): ([\w\[\]!]+)', body, re.M))
    for field, kind in expected.items():
        if fields.get(field) != kind:
            raise SystemExit(f'{name}.{field} must be {kind}')


declared_names = re.findall(r'^(?:type|input|enum|interface|scalar|union) (\w+)', schema, re.M)
if any('hrm' in name.casefold() for name in declared_names + [name for _, name, _, _ in cases]):
    raise SystemExit('Found HRM in a public type or operation name')

pagination = {'after': 'String', 'before': 'String', 'first': 'Int', 'last': 'Int'}
connection_count = 0
node_count = 0
for name, (interface, body) in objects.items():
    if interface == 'Node':
        require_fields(name, body, {'id': 'ID!'})
        node_count += 1
    if name.endswith('Connection'):
        item = name.removesuffix('Connection')
        edge = item + 'Edge'
        if item not in objects or edge not in objects:
            raise SystemExit(f'{name} requires {item} and {edge}')
        # Slip mandiri adalah proyeksi per sesi yang dibaca lewat query khusus.
        if item != 'MyPayslip' and objects[item][0] != 'Node':
            raise SystemExit(f'{item} must implement Node')
        require_fields(name, body, {
            'edges': f'[{edge}!]!', 'nodes': f'[{item}!]!',
            'pageInfo': 'PageInfo!', 'total_count': 'Int!',
        })
        require_fields(edge, objects[edge][1], {'cursor': 'String!', 'node': f'{item}!'})
        connection_count += 1

for operation, name, args, result in cases:
    if operation == 'Query' and result.rstrip('!').endswith('Connection'):
        if result != result.rstrip('!') + '!' or result.rstrip('!') not in objects:
            raise SystemExit(f'{name} must return a non-null domain Connection')
        fields = dict(re.findall(r'(\w+): ([\w\[\]!]+)', args))
        if any(fields.get(key) != value for key, value in pagination.items()):
            raise SystemExit(f'{name} must accept after, before, first, and last')
    if operation == 'Mutation':
        payload = name[0].upper() + name[1:] + 'Payload'
        if result != payload + '!' or objects.get(payload, ('', ''))[0] != 'MutationResult':
            raise SystemExit(f'{name} must return {payload}! implementing MutationResult')
        require_fields(payload, objects[payload][1], {
            'errors': '[MutationError!]!', 'was_successful': 'Boolean!',
        })

print(f'SDL: {node_count} Node types, {connection_count} Relay connections', flush=True)
with ThreadPoolExecutor(max_workers=6) as executor:
    results = list(executor.map(probe, cases))
    limited = [
        index for index, row in enumerate(results)
        if any(error.get('extensions', {}).get('code') == 'TOO_MANY_REQUESTS' for error in row[3])
    ]
    if limited:
        print('Menunggu jendela rate limit sebelum mengulangi pemeriksaan yang tertunda.', flush=True)
        time.sleep(60)
        for index, result in zip(limited, executor.map(probe, (cases[index] for index in limited)), strict=True):
            results[index] = result
for op in ('Query','Mutation'):
    subset = [r for r in results if r[0] == op]
    print(op, sum(r[2] for r in subset), '/', len(subset))
for row in results:
    if not row[2]:
        print(row)
raise SystemExit(any(not r[2] for r in results))
