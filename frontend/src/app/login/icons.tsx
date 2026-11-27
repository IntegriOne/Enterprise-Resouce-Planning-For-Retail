"use client";

export type IconName =
  | "eye"
  | "eyeOff"
  | "shield"
  | "check"
  | "alert"
  | "arrowLeft"
  | "arrowRight"
  | "mail"
  | "sparkles";

export function Icon({
  name,
  size = 18,
}: {
  name: IconName;
  size?: number;
}) {
  const common = {
    width: size,
    height: size,
    viewBox: "0 0 24 24",
    fill: "none",
    stroke: "currentColor",
    strokeWidth: 1.8,
    strokeLinecap: "round" as const,
    strokeLinejoin: "round" as const,
    "aria-hidden": true as const,
    focusable: false as const,
  };

  switch (name) {
    case "eye":
      return (
        <svg {...common}>
          <path d="M2 12s3.6-6 10-6 10 6 10 6-3.6 6-10 6-10-6-10-6Z" />
          <circle cx="12" cy="12" r="2.5" />
        </svg>
      );
    case "eyeOff":
      return (
        <svg {...common}>
          <path d="m3 3 18 18M10.6 6.1A10.9 10.9 0 0 1 12 6c6.4 0 10 6 10 6a15.4 15.4 0 0 1-3.1 3.5M6.5 6.5C3.6 8.4 2 12 2 12s3.6 6 10 6c1.6 0 3-.4 4.2-1" />
          <path d="M10.6 10.6a2 2 0 0 0 2.8 2.8" />
        </svg>
      );
    case "shield":
      return (
        <svg {...common}>
          <path d="M12 2 4.5 5v6c0 5.1 3.2 8.4 7.5 11 4.3-2.6 7.5-5.9 7.5-11V5L12 2Z" />
          <path d="m9.5 12 1.8 1.8 3.7-3.8" />
        </svg>
      );
    case "check":
      return (
        <svg {...common}>
          <path d="m5 12 4 4 10-10" />
        </svg>
      );
    case "alert":
      return (
        <svg {...common}>
          <circle cx="12" cy="12" r="9" />
          <path d="M12 8v4" />
          <path d="M12 16h.01" />
        </svg>
      );
    case "arrowLeft":
      return (
        <svg {...common}>
          <path d="M19 12H5m6-6-6 6 6 6" />
        </svg>
      );
    case "arrowRight":
      return (
        <svg {...common}>
          <path d="M5 12h14m-6-6 6 6-6 6" />
        </svg>
      );
    case "mail":
      return (
        <svg {...common}>
          <rect width="20" height="16" x="2" y="4" rx="2" />
          <path d="m22 7-8.97 5.7a1.94 1.94 0 0 1-2.06 0L2 7" />
        </svg>
      );
    case "sparkles":
      return (
        <svg {...common}>
          <path d="M12 3v3m0 12v3M3 12h3m12 0h3M6.3 6.3l2.1 2.1m7.2 7.2 2.1 2.1m0-11.4-2.1 2.1m-7.2 7.2-2.1 2.1" />
        </svg>
      );
    default:
      return null;
  }
}

export function BrandMark() {
  return (
    <svg
      width="20"
      height="20"
      viewBox="0 0 24 24"
      fill="none"
      aria-hidden="true"
      focusable="false"
    >
      <rect x="3" y="3" width="8" height="8" rx="2" fill="#F9FAFB" />
      <rect x="13" y="3" width="8" height="8" rx="2" fill="#DA414A" />
      <rect
        x="3"
        y="13"
        width="8"
        height="8"
        rx="2"
        fill="rgba(249,250,251,0.65)"
      />
      <rect x="13" y="13" width="8" height="8" rx="2" fill="#F9FAFB" />
    </svg>
  );
}
