-- ============================================================
-- RPC (service_role): update certificate dates for a company
--
-- Context:
-- - After segmented-schema migration, certificate tables live in schema `company_private`
-- - `company_private` MUST NOT be exposed via PostgREST
-- - Backend services using service_role should update sensitive data via SECURITY DEFINER RPC
--
-- This RPC:
-- - Finds one certificate linked to the company (company_private.company_certificates_access)
-- - Updates company_private.certificates_access (effective_date / expiration_date / active)
-- - Is callable ONLY by role `service_role`
-- ============================================================

create or replace function public.rpc_service_update_certificate_dates_for_company(
  p_company_id uuid,
  p_effective_date timestamp default null,
  p_expiration_date timestamp default null
)
returns void
language plpgsql
security definer
set search_path = company_private, company, public
set row_security = off
as $$
declare
  -- PostgREST exposes JWT claims via GUCs; depending on version/config,
  -- `request.jwt.claim.role` may be absent and only `request.jwt.claims` (json) is present.
  -- Use a robust role resolution with fallbacks.
  v_role text := coalesce(
    auth.role(),
    nullif(current_setting('request.jwt.claim.role', true), ''),
    nullif((current_setting('request.jwt.claims', true))::jsonb->>'role', ''),
    ''
  );
  v_certificate_access_id uuid;
begin
  -- hard guard: only backend service_role can call this function
  -- Supabase service keys typically set role=service_role; some setups may use supabase_admin.
  if v_role not in ('service_role', 'supabase_admin') then
    raise exception 'ACCESS_DENIED';
  end if;

  select cca.certificate_access_id
    into v_certificate_access_id
  from company_private.company_certificates_access cca
  where cca.company_id = p_company_id
  order by cca.created_at desc nulls last
  limit 1;

  if v_certificate_access_id is null then
    raise exception 'CERTIFICATE_NOT_FOUND';
  end if;

  update company_private.certificates_access ca
  set
    effective_date = coalesce(p_effective_date, ca.effective_date),
    expiration_date = coalesce(p_expiration_date, ca.expiration_date),
    active = true
  where ca.id = v_certificate_access_id;
end;
$$;

revoke all on function public.rpc_service_update_certificate_dates_for_company(uuid, timestamp, timestamp) from public;
revoke all on function public.rpc_service_update_certificate_dates_for_company(uuid, timestamp, timestamp) from anon;
revoke all on function public.rpc_service_update_certificate_dates_for_company(uuid, timestamp, timestamp) from authenticated;
grant execute on function public.rpc_service_update_certificate_dates_for_company(uuid, timestamp, timestamp) to service_role;
grant execute on function public.rpc_service_update_certificate_dates_for_company(uuid, timestamp, timestamp) to supabase_admin;


