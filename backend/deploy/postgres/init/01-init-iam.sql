-- Inisialisasi database dan role minimum PostgreSQL 18 untuk stack backend dev lokal.
SELECT 'CREATE ROLE iam_management_svc WITH LOGIN CREATEDB PASSWORD ''d931eeb77575bfc4c1b2b99e11807050529d2756c5d1d5d2'''
WHERE NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'iam_management_svc')\gexec

SELECT 'CREATE ROLE gateway_svc WITH LOGIN PASSWORD ''d931eeb77575bfc4c1b2b99e11807050529d2756c5d1d5d2'''
WHERE NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'gateway_svc')\gexec

SELECT 'CREATE DATABASE iam OWNER iam_management_svc'
WHERE NOT EXISTS (SELECT 1 FROM pg_database WHERE datname = 'iam')\gexec

SELECT 'CREATE DATABASE iam_test OWNER iam_management_svc'
WHERE NOT EXISTS (SELECT 1 FROM pg_database WHERE datname = 'iam_test')\gexec
