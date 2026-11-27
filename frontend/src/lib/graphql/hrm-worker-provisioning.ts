import type { PageInfo } from "./iam-types";
import { executeGraphQL, unwrapMutationPayload, type MutationResult } from "./iam-browser-auth.ts";

export type HrmWorker = {
  id: string;
  worker_number: string;
  status: string;
  hire_date: string | null;
  worker_name: string | null;
  legal_entity_name: string | null;
  primary_legal_entity_reference_id: string;
};

export type HrmAccountRequest = {
  status: string;
  requested_at: string;
};

type WorkerReference = Omit<HrmWorker, "worker_name" | "legal_entity_name"> & {
  person_party_reference_id: string | null;
};

export async function listProvisioningWorkers(workerNumber: string, after: string | null, before: string | null) {
  const { data } = await executeGraphQL<{
    workers: { total_count: number; nodes: WorkerReference[]; pageInfo: PageInfo };
  }>(
    `query ProvisioningWorkers($number: String, $first: Int, $last: Int, $after: String, $before: String) {
      workers(worker_number: $number, first: $first, last: $last, after: $after, before: $before) {
        total_count pageInfo { startCursor endCursor hasPreviousPage hasNextPage } nodes { id worker_number status hire_date person_party_reference_id primary_legal_entity_reference_id }
      }
    }`,
    { number: workerNumber || null, after, before, first: before ? null : 20, last: before ? 20 : null },
  );
  const page = data.workers;
  const partyIDs = [...new Set(page.nodes.flatMap((worker) =>
    worker.person_party_reference_id ? [worker.person_party_reference_id] : [],
  ))];
  const entityIDs = [...new Set(page.nodes.map((worker) => worker.primary_legal_entity_reference_id))];
  const [parties, entities] = await Promise.allSettled([
    partyIDs.length > 0 ? executeGraphQL<{ partiesByIds: { _id: string; display_name: string }[] }>(
      `query WorkerNames($ids: [ID!]!) { partiesByIds(ids: $ids) { _id display_name } }`,
      { ids: partyIDs },
    ) : null,
    entityIDs.length > 0 ? executeGraphQL<{ legalEntitiesByIds: { _id: string; legal_name: string }[] }>(
      `query WorkerLegalEntities($ids: [ID!]!) { legalEntitiesByIds(ids: $ids) { _id legal_name } }`,
      { ids: entityIDs },
    ) : null,
  ]);
  const partyNames = new Map(
    (parties.status === "fulfilled" ? parties.value?.data.partiesByIds ?? [] : [])
      .map((party) => [party._id, party.display_name]),
  );
  const entityNames = new Map(
    (entities.status === "fulfilled" ? entities.value?.data.legalEntitiesByIds ?? [] : [])
      .map((entity) => [entity._id, entity.legal_name]),
  );
  return {
    total_count: page.total_count,
    pageInfo: page.pageInfo,
    items: page.nodes.map((worker): HrmWorker => ({
      id: worker.id,
      worker_number: worker.worker_number,
      status: worker.status,
      hire_date: worker.hire_date,
      primary_legal_entity_reference_id: worker.primary_legal_entity_reference_id,
      worker_name: partyNames.get(worker.person_party_reference_id ?? "") ?? null,
      legal_entity_name: entityNames.get(worker.primary_legal_entity_reference_id) ?? null,
    })),
  };
}

export async function fetchWorkerDisplayLabel(workerID: string): Promise<string | null> {
  const { data } = await executeGraphQL<{
    worker: { worker_number: string; person_party_reference_id: string | null } | null;
  }>(
    `query WorkerDisplayLabel($id: ID!) {
      worker(id: $id) { worker_number person_party_reference_id }
    }`,
    { id: workerID },
  );
  const worker = data.worker;
  if (!worker) return null;

  const partyID = worker.person_party_reference_id;
  if (!partyID) return worker.worker_number;
  try {
    const { data: labels } = await executeGraphQL<{ partiesByIds: { _id: string; display_name: string }[] }>(
      `query WorkerNames($ids: [ID!]!) { partiesByIds(ids: $ids) { _id display_name } }`,
      { ids: [partyID] },
    );
    const name = labels.partiesByIds.find((party) => party._id === partyID)?.display_name;
    return name ? `${name} · ${worker.worker_number}` : worker.worker_number;
  } catch {
    return worker.worker_number;
  }
}

export async function fetchWorkerAccountRequest(workerID: string) {
  const { data } = await executeGraphQL<{
    workerEmployeeAccountRequest: HrmAccountRequest | null;
  }>(
    `query WorkerAccountRequest($id: ID!) {
      workerEmployeeAccountRequest(worker_id: $id) { status requested_at }
    }`,
    { id: workerID },
  );
  return data.workerEmployeeAccountRequest;
}

export async function submitWorkerAccountRequest(workerID: string, csrfToken: string) {
  const { data, response } = await executeGraphQL<{
    requestEmployeeAccount: MutationResult & { request?: HrmAccountRequest | null };
  }>(
    `mutation RequestWorkerAccount($id: ID!) {
      requestEmployeeAccount(input: { worker_id: $id }) {
        request { status requested_at }
        was_successful errors { type message }
      }
    }`,
    { id: workerID },
    csrfToken,
  );
  const payload = data.requestEmployeeAccount;
  try {
    return { request: unwrapMutationPayload(payload, response.status).request ?? null, deliveryFailed: false };
  } catch (error) {
    if (!payload?.request) throw error;
    return { request: payload.request, deliveryFailed: true };
  }
}
