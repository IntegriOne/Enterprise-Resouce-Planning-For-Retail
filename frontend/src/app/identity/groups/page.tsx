import { Suspense } from "react";
import GroupsClient from "./groups-client";
import { TableSkeleton } from "@/components/identity/identity-primitives";

export default function IdentityGroupsPage() {
  return (
    <Suspense fallback={<TableSkeleton rows={6} cols={4} />}>
      <GroupsClient />
    </Suspense>
  );
}
