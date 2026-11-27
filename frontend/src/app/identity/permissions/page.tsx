import { Suspense } from "react";
import PermissionsClient from "./permissions-client";
import { TableSkeleton } from "@/components/identity/identity-primitives";

export default function IdentityPermissionsPage() {
  return (
    <Suspense fallback={<TableSkeleton rows={8} cols={4} />}>
      <PermissionsClient />
    </Suspense>
  );
}
