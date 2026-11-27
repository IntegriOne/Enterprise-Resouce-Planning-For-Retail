import type { Metadata } from "next";
import LoginExperience from "./login-experience";

export const metadata: Metadata = {
  title: "Masuk atau Daftar — IntegriOne",
  description:
    "Masuk ke akun IntegriOne Anda atau buat akun baru secara aman.",
  robots: { index: false, follow: false },
};

export default function LoginPage() {
  return <LoginExperience />;
}
