import type { Metadata } from "next";
import RegisterExperience from "./register-experience";

export const metadata: Metadata = {
  title: "Daftar Akun — IntegriOne",
  description:
    "Daftar akun baru IntegriOne untuk akses terpadu melalui portal identitas dan Gateway.",
  robots: { index: false, follow: false },
};

export default function RegisterPage() {
  return <RegisterExperience />;
}
