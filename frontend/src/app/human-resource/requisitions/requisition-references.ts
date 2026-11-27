import { fetchLocations } from "@/lib/graphql/ef-structure-queries";
import { executeGraphQL } from "@/lib/graphql/iam-browser-auth";

export type Option = { id: string; label: string };
export type OptionsPage = { items: Option[]; next: string | null };

type RequisitionScope = { job: string; unit: string };

type Connection<T> = { nodes: T[]; pageInfo: { hasNextPage: boolean; endCursor?: string | null } };
type Position = { _id: string; position_code: string; position_title: string; position_status: string };
type Plan = { _id: string; plan_code: string; plan_version: number; status: string };
type Line = { _id: string; workforce_plan_id: string; line_number: number; action_code: string };

function optionsPage<T>(page: Connection<T>, toOption: (item: T) => Option): OptionsPage {
  return { items: page.nodes.map(toOption), next: page.pageInfo.hasNextPage ? page.pageInfo.endCursor ?? null : null };
}
function positionOption(position: Position): Option {
  return { id: position._id, label: `${position.position_code} · ${position.position_title}` };
}
function planOption(plan: Plan): Option {
  return { id: plan._id, label: `${plan.plan_code} · v${plan.plan_version}` };
}

export async function fetchPositionOptions({ job, unit }: RequisitionScope, after: string | null): Promise<OptionsPage> {
  if (!job || !unit) return { items: [], next: null };
  const { data: { positions } } = await executeGraphQL<{ positions: Connection<Position & { job_id: string; organization_unit_reference_id: string }> }>(
    `query RequisitionPositionOptions($after: String) {
      positions(first: 20, after: $after) {
        nodes { _id position_code position_title position_status job_id organization_unit_reference_id }
        pageInfo { endCursor hasNextPage }
      }
    }`, { after });
  return optionsPage({ ...positions, nodes: positions.nodes.filter((item) =>
    item.position_status === "active" && item.job_id === job && item.organization_unit_reference_id === unit) }, positionOption);
}

export async function fetchUnitLegalEntity(unit: string): Promise<string | null> {
  const { data: { organizationUnit } } = await executeGraphQL<{
    organizationUnit: { legal_entity_id: string } | null;
  }>(`query RequisitionLocationScope($unit: ID!) {
    organizationUnit(id: $unit) { legal_entity_id }
  }`, { unit });
  return organizationUnit?.legal_entity_id.replace(/^gid:\/\/integrione\/legal-entity\//, "") ?? null;
}

export async function fetchLocationOptions(legalEntity: string | null, after: string | null): Promise<OptionsPage> {
  if (!legalEntity) return { items: [], next: null };
  const locations = await fetchLocations({ legal_entity_id: legalEntity, active_only: true, after });
  return optionsPage(locations, (item) => ({ id: item._id, label: `${item.location_code} · ${item.name}` }));
}

export async function fetchPlanOptions(legalEntity: string | null, after: string | null): Promise<OptionsPage> {
  if (!legalEntity) return { items: [], next: null };
  const { data: { workforcePlans } } = await executeGraphQL<{ workforcePlans: Connection<Plan & { legal_entity_reference_id: string }> }>(
    `query RequisitionPlanOptions($after: String) {
      workforcePlans(first: 20, after: $after) {
        nodes { _id plan_code plan_version status legal_entity_reference_id }
        pageInfo { endCursor hasNextPage }
      }
    }`, { after });
  return optionsPage({ ...workforcePlans, nodes: workforcePlans.nodes.filter((item) =>
    item.status === "approved" && item.legal_entity_reference_id === legalEntity) }, planOption);
}

export async function fetchLineOptions(plan: string, { job, unit, category }: RequisitionScope & { category: string }, after: string | null): Promise<OptionsPage> {
  if (!job || !unit) return { items: [], next: null };
  const { data: { workforcePlanLines } } = await executeGraphQL<{
    workforcePlanLines: Connection<Pick<Line, "_id" | "line_number" | "action_code"> & {
      job_id: string; organization_unit_reference_id: string; worker_category: string;
    }>;
  }>(`query RequisitionPlanLineOptions($plan: ID!, $after: String) {
    workforcePlanLines(workforce_plan_id: $plan, first: 20, after: $after) {
      nodes { _id line_number action_code job_id organization_unit_reference_id worker_category }
      pageInfo { endCursor hasNextPage }
    }
  }`, { plan, after });
  return optionsPage({ ...workforcePlanLines, nodes: workforcePlanLines.nodes.filter((item) =>
    item.action_code === "recruit" && item.job_id === job &&
    item.organization_unit_reference_id === unit && item.worker_category === category) }, (item) => ({
    id: item._id, label: `Lini ${item.line_number}`,
  }));
}

type References = { position_id: string | null; work_location_reference_id: string | null; workforce_plan_line_id: string | null };
export type ReferenceLabels = { position?: Option; location?: Option; line?: Option; plan?: Option };

export async function fetchReferenceLabels(value: References): Promise<ReferenceLabels> {
  const [position, location, line] = await Promise.all([
    value.position_id ? executeGraphQL<{ position: Position | null }>(
      "query RequisitionPosition($id: ID!) { position(id: $id) { _id position_code position_title } }",
      { id: value.position_id }).then(({ data }) => data.position && positionOption(data.position)).catch(() => null) : null,
    value.work_location_reference_id ? executeGraphQL<{ location: { _id: string; location_code: string; name: string } | null }>(
      "query RequisitionLocation($id: ID!) { location(id: $id) { _id location_code name } }",
      { id: value.work_location_reference_id }).then(({ data: { location } }) => location && ({
      id: location._id, label: `${location.location_code} · ${location.name}`,
    })).catch(() => null) : null,
    value.workforce_plan_line_id ? executeGraphQL<{ workforcePlanLine: Line | null }>(
      "query RequisitionPlanLine($id: ID!) { workforcePlanLine(id: $id) { _id line_number workforce_plan_id } }",
      { id: value.workforce_plan_line_id }).then(({ data }) => data.workforcePlanLine).catch(() => null) : null,
  ]);
  const plan = line ? await executeGraphQL<{ workforcePlan: Plan | null }>(
    "query RequisitionLinePlan($id: ID!) { workforcePlan(id: $id) { _id plan_code plan_version } }",
    { id: line.workforce_plan_id }).then(({ data }) => data.workforcePlan && planOption(data.workforcePlan)).catch(() => null) : null;
  return {
    position: position ?? undefined,
    location: location ?? undefined,
    line: line ? { id: line._id, label: `Lini ${line.line_number}` } : undefined,
    plan: plan ?? (line ? { id: line.workforce_plan_id, label: "Rencana saat ini" } : undefined),
  };
}

export const employmentCategories: Record<string, string> = {
  employee: "Pegawai",
  fixed_term_employee: "Pegawai kontrak waktu tertentu",
  contractor: "Kontraktor",
  agency_worker: "Pekerja agensi",
  apprentice: "Peserta pemagangan",
  intern: "Peserta magang",
  volunteer: "Relawan",
  contingent_worker: "Pekerja sementara",
};

export const requisitionTypes: Record<string, string> = {
  new_headcount: "Penambahan tenaga kerja",
  replacement: "Pengganti",
  temporary: "Sementara",
};
