import { Suspense } from "react";
import RolesClient from "./roles-client";
import { TableSkeleton } from "@/components/identity/identity-primitives";

export default function IdentityRolesPage() {
  return (
    <Suspense fallback={<TableSkeleton rows={6} cols={5} />}>
      <RolesClient />
    </Suspense>
  );
}
