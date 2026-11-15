import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;

  // Lindungi route /identity dari direct access anonymous
  if (pathname.startsWith("/identity")) {
    const sessionCookie = request.cookies.get("__Host-integrione_session");

    // Pastikan hanya cookie sesi __Host- yang bernilai non-kosong yang diizinkan melewati guard tepi.
    if (!sessionCookie || !sessionCookie.value) {
      const loginUrl = new URL("/login", request.url);
      return NextResponse.redirect(loginUrl);
    }
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/identity/:path*"],
};
