import type { Metadata } from "next";
import VerifyEmailExperience from "./verify-email-experience";

export const metadata: Metadata = {
  title: "Verifikasi Email — IntegriOne",
  description: "Verifikasi alamat email akun Anda di IntegriOne.",
  robots: { index: false, follow: false },
};

export default function VerifyEmailPage() {
  return <VerifyEmailExperience />;
}
