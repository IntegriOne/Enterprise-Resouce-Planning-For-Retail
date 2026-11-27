# Verifikasi guard HTTPS seluruh schema HRM

Mengirim 77 operasi yang dibangun dari schema checkout melalui Caddy HTTPS. Semua 38 Query mencapai guard sesi dan 39 Mutation mencapai guard CSRF. Bukti tidak mencakup akses employee dengan grant resmi.
