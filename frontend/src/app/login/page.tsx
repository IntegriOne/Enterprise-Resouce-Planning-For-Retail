import type { Metadata } from "next";
import LoginExperience from "./login-experience";

export const metadata: Metadata = {
  title: "Masuk — IntegriOne",
  description:
    "Masuk ke akun IntegriOne Anda untuk mengakses portal identitas dan layanan Gateway terpadu.",
  robots: { index: false, follow: false },
};

export default function LoginPage() {
  return <LoginExperience />;
}
