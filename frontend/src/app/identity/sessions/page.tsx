import { Suspense } from "react";
import SessionsClient from "./sessions-client";
import { TableSkeleton } from "@/components/identity/identity-primitives";

export default function IdentitySessionsPage() {
  return (
    <Suspense fallback={<TableSkeleton rows={6} cols={4} />}>
      <SessionsClient />
    </Suspense>
  );
}
