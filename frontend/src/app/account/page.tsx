import type { Metadata } from "next";
import PayslipsClient from "./payslips-client";

export const metadata: Metadata = {
  title: "Slip Gaji Saya — IntegriOne",
  robots: { index: false, follow: false },
};

export default function EmployeeAccountPage() {
  return <PayslipsClient />;
}
