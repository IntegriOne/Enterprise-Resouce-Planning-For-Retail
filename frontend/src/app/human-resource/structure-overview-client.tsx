"use client";

import { useCallback, useEffect, useState } from "react";
import Link from "next/link";
import styles from "@/components/identity/identity.module.css";
import { IdentityIcon } from "@/components/identity/identity-icons";
import { useIdentityWorkspace } from "@/components/identity/identity-context";
import { ErrorPanel } from "@/components/identity/identity-primitives";
import {
  EfCodePill,
  EfStatStrip,
  EfStatusBadge,
} from "@/components/workspace/ef-ui-primitives";
import { hasEfPermission } from "@/lib/graphql/ef-helpers";
import {
  fetchAllAddresses,
  fetchAllLegalEntities,
  fetchAllLocations,
  fetchAllOrganizationUnits,
  fetchPrimaryOperatingLegalEntity,
} from "@/lib/graphql/ef-structure-queries";
import type { EfLegalEntity } from "@/lib/graphql/ef-types";

export default function StructureOverviewClient() {
  const { snapshot } = useIdentityWorkspace();
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [counts, setCounts] = useState({
    legalEntities: 0,
    orgUnits: 0,
    locations: 0,
    addresses: 0,
  });
  const [primaryEntity, setPrimaryEntity] = useState<EfLegalEntity | null>(null);

  const fetchOverviewData = useCallback(async () => {
    const [leRes, ouRes, locRes, addrRes, primaryRes] = await Promise.all([
      hasEfPermission(snapshot, "enterprise_foundation.legal_entities.read")
        ? fetchAllLegalEntities(5)
        : Promise.resolve(null),
      hasEfPermission(snapshot, "enterprise_foundation.organization_units.read")
        ? fetchAllOrganizationUnits(5)
        : Promise.resolve(null),
      hasEfPermission(snapshot, "enterprise_foundation.locations.read")
        ? fetchAllLocations(5)
        : Promise.resolve(null),
      hasEfPermission(snapshot, "enterprise_foundation.addresses.read")
        ? fetchAllAddresses(5)
        : Promise.resolve(null),
      hasEfPermission(snapshot, "enterprise_foundation.legal_entities.read")
        ? fetchPrimaryOperatingLegalEntity()
        : Promise.resolve(null),
    ]);
    return { leRes, ouRes, locRes, addrRes, primaryRes };
  }, [snapshot]);

  const loadOverview = useCallback(async () => {
    setLoading(true);
    setLoadError(null);
    try {
      const { leRes, ouRes, locRes, addrRes, primaryRes } =
        await fetchOverviewData();
      setCounts({
        legalEntities: leRes?.total_count ?? 0,
        orgUnits: ouRes?.total_count ?? 0,
        locations: locRes?.total_count ?? 0,
        addresses: addrRes?.total_count ?? 0,
      });
      setPrimaryEntity(primaryRes);
    } catch {
      setLoadError("Ringkasan struktur perusahaan gagal dimuat.");
    } finally {
      setLoading(false);
    }
  }, [fetchOverviewData]);

  useEffect(() => {
    let active = true;
    async function initOverview() {
      try {
        const { leRes, ouRes, locRes, addrRes, primaryRes } =
          await fetchOverviewData();
        if (!active) return;
        setCounts({
          legalEntities: leRes?.total_count ?? 0,
          orgUnits: ouRes?.total_count ?? 0,
          locations: locRes?.total_count ?? 0,
          addresses: addrRes?.total_count ?? 0,
        });
        setPrimaryEntity(primaryRes);
        setLoadError(null);
      } catch {
        if (!active) return;
        setLoadError("Ringkasan struktur perusahaan gagal dimuat.");
      } finally {
        if (active) setLoading(false);
      }
    }
    void initOverview();
    return () => {
      active = false;
    };
  }, [fetchOverviewData]);

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h2 className="text-xl font-bold tracking-tight text-[#0b2439]">
            Struktur Perusahaan — Enterprise Foundation
          </h2>
          <p className="mt-0.5 text-sm text-[#5a6672]">
            Kelola badan hukum, hierarki unit organisasi, fasilitas operasional,
            dan alamat perusahaan.
          </p>
        </div>
        <button
          type="button"
          onClick={() => void loadOverview()}
          disabled={loading}
          className={styles.btnSecondary}
        >
          <IdentityIcon name="refresh" size={15} />
          <span>{loading ? "Memuat..." : "Muat Ulang"}</span>
        </button>
      </div>

      {loadError ? (
        <ErrorPanel message={loadError} onRetry={() => void loadOverview()} />
      ) : null}

      <EfStatStrip
        items={[
          {
            label: "Badan Hukum",
            value: counts.legalEntities,
            helper: "Legal Entities",
          },
          {
            label: "Unit Organisasi",
            value: counts.orgUnits,
            helper: "Divisi / Dept / Cabang",
          },
          {
            label: "Lokasi & Fasilitas",
            value: counts.locations,
            helper: "HQ / Gudang / Toko",
          },
          {
            label: "Master Alamat",
            value: counts.addresses,
            helper: "Geografi terpusat",
          },
        ]}
      />

      <section className={`${styles.hairlinePanel} space-y-4 p-5`}>
        <div className="flex items-start justify-between gap-3">
          <h3 className="text-base font-semibold text-[#0b2439]">
            Badan Hukum Operasional Utama
          </h3>
          {primaryEntity ? (
            <EfStatusBadge label={primaryEntity.status} />
          ) : null}
        </div>

        {primaryEntity ? (
          <div className="space-y-2 rounded-lg border border-[#dde3e8] bg-[#f9fafb] p-4">
            <div className="flex flex-wrap items-center gap-2">
              <EfCodePill code={primaryEntity.entity_code} />
              <span className="text-sm font-bold text-[#0b2439]">
                {primaryEntity.legal_name}
              </span>
            </div>
            <p className={`text-xs text-[#5a6672] ${styles.monoText}`}>
              NPWP: {primaryEntity.tax_id_number || "-"} · NIB:{" "}
              {primaryEntity.nib || "-"} · Yurisdiksi:{" "}
              {primaryEntity.jurisdiction_country_code}
            </p>
          </div>
        ) : (
          <p className="rounded-lg border border-dashed border-[#dde3e8] p-4 text-xs text-[#5a6672]">
            Belum ada badan hukum operasional utama. Tetapkan satu di halaman
            Badan Hukum.
          </p>
        )}

        <div className="grid grid-cols-2 gap-2 pt-1">
          <Link href="/human-resource/legal-entities" className={styles.btnPrimary}>
            Kelola Badan Hukum
          </Link>
          <Link href="/human-resource/organization-units" className={styles.btnSecondary}>
            Unit Organisasi
          </Link>
          <Link href="/human-resource/locations" className={styles.btnSecondary}>
            Lokasi &amp; Gudang
          </Link>
          <Link href="/human-resource/addresses" className={styles.btnSecondary}>
            Master Alamat
          </Link>
        </div>
      </section>
    </div>
  );
}
