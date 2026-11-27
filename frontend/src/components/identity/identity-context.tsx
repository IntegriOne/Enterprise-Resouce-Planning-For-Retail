"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useRef,
  useState,
  type ReactNode,
} from "react";
import { fetchBrowserSession } from "@/lib/graphql/iam-browser-auth";
import {
  fetchOperatorContext,
  formatIamDashboardError,
  isStepUpRequiredError,
  isUnauthenticatedError,
} from "@/lib/graphql/iam-dashboard-queries";
import { submitUserLogout } from "@/lib/graphql/iam-dashboard-mutations";
import {
  hasAnyEfPartnerAccess,
  hasAnyEfStructureAccess,
  hasEfPermission,
} from "@/lib/graphql/ef-helpers";
import type { AccessSnapshot, IamUser } from "@/lib/graphql/iam-types";

export type WorkspaceDomain = "identity" | "human-resource" | "procurement";

export type ToastKind = "success" | "error" | "info";

export type ToastMessage = {
  id: string;
  kind: ToastKind;
  title: string;
  detail?: string;
};

type StepUpRequest = {
  reason?: string;
  onVerified?: () => Promise<void> | void;
};

type IdentityContextValue = {
  workspaceDomain: WorkspaceDomain;
  me: IamUser | null;
  snapshot: AccessSnapshot | null;
  csrfToken: string | null;
  isLoadingContext: boolean;
  accessDenied: boolean;
  contextError: string | null;
  isLoggingOut: boolean;
  isCommandOpen: boolean;
  stepUpRequest: StepUpRequest | null;
  toasts: ToastMessage[];
  refreshOperatorContext: () => Promise<void>;
  pushToast: (kind: ToastKind, title: string, detail?: string) => void;
  dismissToast: (id: string) => void;
  openCommandPalette: () => void;
  closeCommandPalette: () => void;
  openStepUpModal: (reason?: string, onVerified?: () => Promise<void> | void) => void;
  closeStepUpModal: () => void;
  handleStepUpSuccess: () => Promise<void>;
  runDashboardAction: <T>(
    action: () => Promise<T>,
    options?: { stepUpReason?: string },
  ) => Promise<T | null>;
  logout: () => Promise<void>;
};

const IdentityContext = createContext<IdentityContextValue | null>(null);

function canAccessDomain(domain: WorkspaceDomain, snap: AccessSnapshot): boolean {
  switch (domain) {
    case "human-resource":
      return hasAnyEfStructureAccess(snap) || hasEfPermission(snap, "human_resource.workers.read");
    case "procurement":
      return hasAnyEfPartnerAccess(snap);
    default:
      return Boolean(snap.can_access_identity);
  }
}

export function IdentityProvider({
  children,
  workspaceDomain = "identity",
}: {
  children: ReactNode;
  workspaceDomain?: WorkspaceDomain;
}) {
  const [me, setMe] = useState<IamUser | null>(null);
  const [snapshot, setSnapshot] = useState<AccessSnapshot | null>(null);
  const [csrfToken, setCsrfToken] = useState<string | null>(null);
  const [isLoadingContext, setIsLoadingContext] = useState(true);
  const [accessDenied, setAccessDenied] = useState(false);
  const [contextError, setContextError] = useState<string | null>(null);
  const [isLoggingOut, setIsLoggingOut] = useState(false);
  const [isCommandOpen, setIsCommandOpen] = useState(false);
  const [stepUpRequest, setStepUpRequest] = useState<StepUpRequest | null>(null);
  const [toasts, setToasts] = useState<ToastMessage[]>([]);
  const toastSeqRef = useRef(0);

  const dismissToast = useCallback((id: string) => {
    setToasts((prev) => prev.filter((item) => item.id !== id));
  }, []);

  const pushToast = useCallback(
    (kind: ToastKind, title: string, detail?: string) => {
      toastSeqRef.current += 1;
      const id = `iam-toast-${toastSeqRef.current}`;
      setToasts((prev) => [...prev.slice(-3), { id, kind, title, detail }]);
      window.setTimeout(() => {
        setToasts((prev) => prev.filter((item) => item.id !== id));
      }, 4800);
    },
    [],
  );

  const applyLoadedContext = useCallback(
    async (withLoadingSpinner: boolean) => {
      if (withLoadingSpinner) {
        setIsLoadingContext(true);
        setContextError(null);
      }
      try {
        const [session, operator] = await Promise.all([
          fetchBrowserSession(),
          fetchOperatorContext(),
        ]);

        if (!session.authenticated) {
          window.location.replace("/login");
          return;
        }
        if (!canAccessDomain(workspaceDomain, operator.snapshot)) {
          setCsrfToken(session.csrfToken);
          setMe(operator.me);
          setSnapshot(operator.snapshot);
          setAccessDenied(true);
          return;
        }

        setAccessDenied(false);
        setCsrfToken(session.csrfToken);
        setMe(operator.me);
        setSnapshot(operator.snapshot);
      } catch (err) {
        if (isUnauthenticatedError(err)) {
          window.location.replace("/login");
          return;
        }
        setContextError(formatIamDashboardError(err));
      } finally {
        setIsLoadingContext(false);
      }
    },
    [workspaceDomain],
  );

  const refreshOperatorContext = useCallback(async () => {
    await applyLoadedContext(true);
  }, [applyLoadedContext]);

  useEffect(() => {
    let active = true;
    async function bootstrap() {
      try {
        const [session, operator] = await Promise.all([
          fetchBrowserSession(),
          fetchOperatorContext(),
        ]);
        if (!active) return;

        if (!session.authenticated) {
          window.location.replace("/login");
          return;
        }
        if (!canAccessDomain(workspaceDomain, operator.snapshot)) {
          setCsrfToken(session.csrfToken);
          setMe(operator.me);
          setSnapshot(operator.snapshot);
          setAccessDenied(true);
          return;
        }

        setAccessDenied(false);
        setCsrfToken(session.csrfToken);
        setMe(operator.me);
        setSnapshot(operator.snapshot);
        setContextError(null);
      } catch (err) {
        if (!active) return;
        if (isUnauthenticatedError(err)) {
          window.location.replace("/login");
          return;
        }
        setContextError(formatIamDashboardError(err));
      } finally {
        if (active) setIsLoadingContext(false);
      }
    }
    void bootstrap();
    return () => {
      active = false;
    };
  }, [workspaceDomain]);

  useEffect(() => {
    function handleKeyDown(e: KeyboardEvent) {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === "k") {
        e.preventDefault();
        setIsCommandOpen((prev) => !prev);
      }
    }
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, []);

  const openCommandPalette = useCallback(() => setIsCommandOpen(true), []);
  const closeCommandPalette = useCallback(() => setIsCommandOpen(false), []);

  const openStepUpModal = useCallback(
    (reason?: string, onVerified?: () => Promise<void> | void) => {
      setStepUpRequest({ reason, onVerified });
    },
    [],
  );

  const closeStepUpModal = useCallback(() => {
    setStepUpRequest(null);
  }, []);

  const handleStepUpSuccess = useCallback(async () => {
    const callback = stepUpRequest?.onVerified;
    setStepUpRequest(null);
    await refreshOperatorContext();
    pushToast("success", "Verifikasi Step-Up Berhasil", "Sesi sensitif Anda kini terverifikasi.");
    if (callback) {
      await callback();
    }
  }, [pushToast, refreshOperatorContext, stepUpRequest]);

  const runDashboardAction = useCallback(
    async <T,>(
      action: () => Promise<T>,
      options?: { stepUpReason?: string },
    ): Promise<T | null> => {
      try {
        return await action();
      } catch (err) {
        if (isUnauthenticatedError(err)) {
          window.location.replace("/login");
          return null;
        }
        if (isStepUpRequiredError(err)) {
          openStepUpModal(
            options?.stepUpReason ||
              "Tindakan sensitif ini mewajibkan verifikasi kode MFA sesi aktif.",
            async () => {
              await action();
            },
          );
          return null;
        }
        pushToast("error", "Operasi Gagal", formatIamDashboardError(err));
        return null;
      }
    },
    [openStepUpModal, pushToast],
  );

  const logout = useCallback(async () => {
    if (isLoggingOut) return;
    setIsLoggingOut(true);
    try {
      await submitUserLogout(csrfToken);
    } catch {
      // Tetap keluar ke /login jika sesi sudah dicabut atau kedaluwarsa.
    } finally {
      setMe(null);
      setSnapshot(null);
      setCsrfToken(null);
      window.location.replace("/login");
    }
  }, [csrfToken, isLoggingOut]);

  return (
    <IdentityContext.Provider
      value={{
        workspaceDomain,
        me,
        snapshot,
        csrfToken,
        isLoadingContext,
        accessDenied,
        contextError,
        isLoggingOut,
        isCommandOpen,
        stepUpRequest,
        toasts,
        refreshOperatorContext,
        pushToast,
        dismissToast,
        openCommandPalette,
        closeCommandPalette,
        openStepUpModal,
        closeStepUpModal,
        handleStepUpSuccess,
        runDashboardAction,
        logout,
      }}
    >
      {children}
    </IdentityContext.Provider>
  );
}

export function useIdentityWorkspace(): IdentityContextValue {
  const ctx = useContext(IdentityContext);
  if (!ctx) {
    throw new Error("useIdentityWorkspace harus dipakai di dalam IdentityProvider");
  }
  return ctx;
}
