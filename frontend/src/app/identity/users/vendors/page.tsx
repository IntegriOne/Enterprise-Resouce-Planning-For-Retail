import { Suspense } from "react";
import UsersClient from "../users-client";
import { TableSkeleton } from "@/components/identity/identity-primitives";

export default function IdentityVendorsUsersPage() {
  return (
    <Suspense fallback={<TableSkeleton rows={6} cols={5} />}>
      <UsersClient segment="vendors" />
    </Suspense>
  );
}
