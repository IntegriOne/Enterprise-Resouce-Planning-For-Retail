"""Periksa inventaris Gateway terhadap proto, handler, dan katalog izin HRM."""

from collections import Counter
from pathlib import Path
import re

MODULE = Path(__file__).resolve().parents[1]
HANDOFF = MODULE.parents[2] / "agents/grpc/hrm"
COMMON = "human_resource.common.v1."
READ_ACTIONS = {"read", "list", "read_self"}
CRUD_ACTIONS = {"read", "list", "create", "update", "delete"}

# Operasi ini memakai otorisasi langsung atau membagi workflow ke beberapa metode.
EXPLICIT_OPERATIONS = {
    "AcceptOffer": ("hr_offers", "accept"),
    "ApprovePromotionRecommendation": ("hr_promotion_recommendations", "approve"),
    "DecideWorkerAssignmentTransfer": ("hr_worker_assignments", "transfer"),
    "DeclineOffer": ("hr_offers", "decline"),
    "GetRequisitionByID": ("hr_requisitions", "read"),
    "GetMyPayslip": ("hr_payslips", "read_self"),
    "ListMyPayslips": ("hr_payslips", "read_self"),
    "RecordTalentCalibrationDecision": ("hr_talent_calibration_sessions", "approve"),
    "ReissueEmployeePayslip": ("hr_payslips", "reissue"),
    "RejectPromotionRecommendation": ("hr_promotion_recommendations", "reject"),
    "ReturnPromotionRecommendation": ("hr_promotion_recommendations", "return"),
    "SubmitWorkerAssignmentTransfer": ("hr_worker_assignments", "create"),
    "TerminateWorker": ("hr_workers", "terminate"),
}


def go_functions(folder):
    functions = {}
    for path in sorted(folder.glob("*.go")):
        source = path.read_text()
        declarations = list(re.finditer(r"^func\s+(?:\([^\n]+\)\s+)?(\w+)\(", source, re.M))
        for index, declaration in enumerate(declarations):
            end = declarations[index + 1].start() if index + 1 < len(declarations) else len(source)
            functions[declaration[1]] = source[declaration.start():end]
    return functions


def resource_tables(service):
    specs = {}
    for name, body in service.items():
        match = re.search(r'\bTable: "([^"]+)"', body)
        if match:
            specs[name] = match[1]
    resources = {}
    for body in service.values():
        for field, spec in re.findall(
            r"s\.(\w+), err = operation\.NewResource\s*(?:\[[^\]]*\])?\(\s*[^,]+,\s*[^,]+,\s*(spec\w+)\(", body
        ):
            if spec in specs:
                resources[field] = specs[spec]
    return resources


def find_operation(name, handler, service, resources):
    if name in EXPLICIT_OPERATIONS:
        return EXPLICIT_OPERATIONS[name]
    matches = []
    for method in re.findall(r"[hs]\.service\.(\w+)\b", handler):
        body = service.get(method, "")
        matches.extend(re.findall(
            r'operation\.Access\{Domain: "[a-z_]+", Resource: "(hr_[a-z_]+)", Action: "([a-z_]+)"',
            body,
        ))
        for resource, verb in re.findall(r"s\.(\w+)\.(Create|Update|Delete|Get|List|Command)\(", body):
            if resource not in resources:
                continue
            if verb == "Command":
                command = re.search(r'\.Command\(ctx,\s*\w+,\s*"([^"]+)"', body)
                if not command:
                    continue
                action = command[1]
            else:
                action = "read" if verb == "Get" else verb.lower()
            matches.append((resources[resource], action))
    matches = list(dict.fromkeys(matches))
    # Command dapat menyentuh resource sekunder setelah aksi utama.
    if matches and (len(matches) == 1 or matches[0][1] not in READ_ACTIONS):
        return matches[0]
    raise ValueError(f"{name}: pemetaan operasi tidak unik: {matches}")


def main():
    catalog = (MODULE / "infrastructure/security/catalog.go").read_text()
    resources = set(re.findall(r'"([a-z_]+)"', catalog.split("readOnly :=")[0]))
    read_only = set(re.findall(r'"([a-z_]+)"\s*:\s*true', catalog.split("readOnly :=", 1)[1].split("permissions :=", 1)[0]))
    overrides = dict(re.findall(r'permissions\["([^"]+)"\] = "([^"]+)"', catalog))
    index = (HANDOFF / "1-inventaris-operasi-grpc-hrm.md").read_text()
    totals = Counter()
    errors = []
    domains = 0

    for proto in sorted((MODULE / "contract/proto/human_resource").glob("*/v1/*.proto")):
        source = proto.read_text()
        service_name = re.search(r"^service (\w+)", source, re.M)
        if not service_name:
            continue
        domains += 1
        domain = proto.stem
        service = go_functions(MODULE / "internal/service" / domain)
        handlers = go_functions(MODULE / "transport/grpc/handler" / domain)
        tables = resource_tables(service)
        expected = Counter()
        rpcs = re.findall(r"^\s*rpc\s+(\w+)\(([^)]+)\)\s*returns\s*\(([^)]+)\)", source, re.M)
        for name, request, response in rpcs:
            try:
                table, action = find_operation(name, handlers[name], service, tables)
                key = f"{table}.{action}"
                resource = table.removeprefix("hr_")
                permission = overrides.get(key)
                kind = "Query" if action in READ_ACTIONS else "Mutation"
                if permission is None:
                    if resource not in resources or action not in CRUD_ACTIONS:
                        raise ValueError(f"{name}: {key} tidak ditemukan di katalog izin")
                    if resource in read_only and kind == "Mutation":
                        kind, permission = "Nonaktif", "—"
                    else:
                        permission = f"human_resource.{resource}.{'read' if action == 'list' else action}"
                expected[(name, kind, request.removeprefix(COMMON), response.removeprefix(COMMON), permission)] += 1
            except (KeyError, ValueError) as error:
                errors.append(f"{domain}: {error}")

        documents = list(HANDOFF.glob(f"*-operasi-{domain}.md"))
        if len(documents) != 1:
            errors.append(f"{domain}: diperlukan tepat satu dokumen operasi")
            continue
        document = documents[0]
        rows = re.findall(
            r"^\| `(\w+)` \| [^|]+ \| (Query|Mutation|Nonaktif) \| `([\w.]+)` → `([\w.]+)` \| (?:`([\w.]+)`[^|]*|—) \|$",
            document.read_text(),
            re.M,
        )
        actual = Counter((*row[:4], row[4] or "—") for row in rows)
        for difference, label in ((expected - actual, "kurang"), (actual - expected, "berlebih")):
            errors.extend(f"{domain}: {label} {row}" for row in difference.elements())
        counts = Counter(row[1] for row in rows)
        totals.update(counts)
        if f"| [{service_name[1]}]({document.name}) | {len(rpcs)} | {counts['Query']} | {counts['Mutation']} | {counts['Nonaktif']} |" not in index:
            errors.append(f"{domain}: angka pada indeks tidak cocok")

    total_rpcs = sum(totals.values())
    if f"| **Total** | **{total_rpcs}** | **{totals['Query']}** | **{totals['Mutation']}** | **{totals['Nonaktif']}** |" not in index:
        errors.append("Jumlah keseluruhan pada indeks tidak cocok")
    for error in errors:
        print(error)
    print(f"{domains} service, {total_rpcs} RPC, {totals['Query']} Query, {totals['Mutation']} Mutation, {totals['Nonaktif']} nonaktif, {len(errors)} selisih")
    return bool(errors)


if __name__ == "__main__":
    raise SystemExit(main())
