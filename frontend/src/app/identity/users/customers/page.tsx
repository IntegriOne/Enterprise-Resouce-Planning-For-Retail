import { Suspense } from "react";
import UsersClient from "../users-client";
import { TableSkeleton } from "@/components/identity/identity-primitives";

export default function IdentityCustomersUsersPage() {
  return (
    <Suspense fallback={<TableSkeleton rows={6} cols={5} />}>
      <UsersClient segment="customers" />
    </Suspense>
  );
}
