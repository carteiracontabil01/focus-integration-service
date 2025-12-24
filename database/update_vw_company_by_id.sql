-- Atualização da view vw_company_by_id para incluir focus_integration

DROP VIEW IF EXISTS public.vw_company_by_id;

CREATE OR REPLACE VIEW public.vw_company_by_id AS
SELECT
  c.id,
  c.cnpj,
  c.legal_name,
  c.business_name,
  c.business_segment,
  c.activity_start_date,
  c.business_phone_number,
  c.business_email,
  c.business_contact_name,
  c.legal_nature_id,
  c.tax_regime,
  c.pro_labore_opting,
  c.created_at,
  c.tenant_id,
  c.operation_mode,
  c.updated_by,
  c.updated_at,
  c.photo_url,
  c.company_size,
  c.status,
  c.special_tax_regime,
  c.municipal_registration,
  c.nfse_municipality_specific_data,
  c.municipality_id,
  c.use_municipality_cnaes,
  c.simples_nacional_tax_regime,
  c.focus_integrated,

  -- Focus Integration (dados de integração com Focus NFe)
  CASE
    WHEN fi.id IS NOT NULL THEN
      jsonb_build_object(
        'id', fi.id,
        'focus_company_id', fi.focus_company_id,
        'token_focus_company', fi.token_focus_company,
        'created_at', fi.created_at
      )
    ELSE NULL
  END AS focus_integration,

  -- Município da empresa
  (
    SELECT to_jsonb(m)
    FROM municipios m
    WHERE m.id = c.municipality_id
  ) AS municipality,

  -- Campos específicos de NFS-e do município
  (
    SELECT to_jsonb(mnf)
    FROM municipality_nfse_specific_fields mnf
    WHERE mnf.municipality_id = c.municipality_id
    LIMIT 1
  ) AS municipality_nfse_specific_fields,

  -- Natureza legal
  jsonb_build_object(
    'id', ln.id,
    'name', ln.name,
    'registering_authoritie_id', ln.registering_authoritie_id
  ) AS legal_nature,

  -- Tenant
  jsonb_build_object('id', t.id, 'name', t.name) AS tenant,

  -- Endereços (sem município para evitar duplicação)
  (
    SELECT jsonb_agg(
      jsonb_build_object(
        'id',           a.id,
        'zip_code',     a.zip_code,
        'address',      a.address,
        'number',       a.number,
        'neighborhood', a.neighborhood,
        'complement',   a.complement,
        'city',         a.city,
        'state',        a.state,
        'address_type', ca.address_type
      )
    )
    FROM company_address ca
    JOIN addresses a ON ca.address_id = a.id
    WHERE ca.company_id = c.id
  ) AS addresses,

  -- Contatos
  (
    SELECT jsonb_agg(
      jsonb_build_object(
        'id',           ct.id,
        'name',         ct.name,
        'email',        ct.email,
        'phone',        ct.phone,
        'contact_type', ct.contact_type
      )
    )
    FROM company_contact cc
    JOIN contacts ct ON cc.contact_id = ct.id
    WHERE cc.company_id = c.id
  ) AS contacts,

  -- Sócios e administradores
  (
    SELECT jsonb_agg(
      jsonb_build_object(
        'id',              pac.id,
        'name',            pac.name,
        'cpf',             pac.cpf,
        'birthday',        pac.birthday,
        'marital_status',  pac.marital_status,
        'email',           pac.email,
        'phone',           pac.phone,
        'gov_br_password', pac.gov_br_password,
        'type',            pac.type,
        'address',
          (
            SELECT to_jsonb(
              jsonb_build_object(
                'id',           a2.id,
                'zip_code',     a2.zip_code,
                'address',      a2.address,
                'number',       a2.number,
                'neighborhood', a2.neighborhood,
                'complement',   a2.complement,
                'city',         a2.city,
                'state',        a2.state,
                'municipality',
                  (
                    SELECT to_jsonb(m2)
                    FROM municipios m2
                    WHERE m2.id = a2.municipality_id
                  )
              )
            )
            FROM partner_administrator_address paa
            JOIN addresses a2 ON a2.id = paa.address_id
            WHERE paa.partner_id = pac.id
            LIMIT 1
          )
      )
    )
    FROM partner_and_administrator_company pac
    WHERE pac.company_id = c.id
  ) AS partner_administrators,

  -- CNAEs + serviços LC 116 (incluindo CNAEs municipais) + tax_code selecionado
  (
    SELECT jsonb_agg(cnae_data)
    FROM (
      -- CNAEs padrão
      SELECT jsonb_build_object(
        'id',          cn.id,
        'code',        cn.code,
        'description', cn.description,
        'principal',   grouped.principal,
        'is_municipal', false,
        'municipality_cnae_id', NULL,
        'serviceItemsLc116', grouped.service_items
      ) AS cnae_data
      FROM (
        SELECT
          ccs.cnae_id,
          bool_or(ccs.cnae_default) AS principal,
          jsonb_agg(
            jsonb_build_object(
              'id',              s.id,
              'code_lc116',      s.code_lc116,
              'description',     s.description,
              'required_fields', s.required_fields,
              'issqn_national_service_id', s.issqn_national_service_id,

              -- >>> NOVO: tax code escolhido por serviço
              'tax_code_municipality_iss_id', ccs.tax_code_municipality_iss_id,
              'tax_code_municipality_iss',
                CASE
                  WHEN tcm.id IS NOT NULL THEN
                    jsonb_build_object(
                      'id', tcm.id,
                      'municipality_code', tcm.municipality_code,
                      'code', tcm.code,
                      'description', tcm.description,
                      'created_at', tcm.created_at
                    )
                  ELSE NULL
                END,

              'issqn_national_service',
                CASE
                  WHEN s.issqn_national_service_id IS NOT NULL THEN
                    (
                      SELECT to_jsonb(
                        jsonb_build_object(
                          'id', insc.id,
                          'code', insc.code,
                          'description', insc.description,
                          'created_at', insc.created_at
                        )
                      )
                      FROM issqn_national_service_codes insc
                      WHERE insc.id = s.issqn_national_service_id
                    )
                  ELSE NULL
                END
            )
          ) AS service_items
        FROM company_cnae_services ccs
        JOIN service_items_lc116 s ON s.id = ccs.service_id
        LEFT JOIN tax_code_municipality_iss tcm ON tcm.id = ccs.tax_code_municipality_iss_id
        WHERE ccs.company_id = c.id
          AND ccs.cnae_id IS NOT NULL
        GROUP BY ccs.cnae_id
      ) AS grouped
      JOIN cnaes cn ON cn.id = grouped.cnae_id

      UNION ALL

      -- CNAEs municipais
      SELECT jsonb_build_object(
        'id',          mc.id,
        'code',        mc.code,
        'description', mc.description,
        'principal',   grouped_municipal.principal,
        'is_municipal', true,
        'municipality_cnae_id', mc.id,
        'serviceItemsLc116', grouped_municipal.service_items
      ) AS cnae_data
      FROM (
        SELECT
          ccs.municipality_cnae_id,
          bool_or(ccs.cnae_default) AS principal,
          jsonb_agg(
            jsonb_build_object(
              'id',              s.id,
              'code_lc116',      s.code_lc116,
              'description',     s.description,
              'required_fields', s.required_fields,
              'issqn_national_service_id', s.issqn_national_service_id,

              -- >>> NOVO: tax code escolhido por serviço
              'tax_code_municipality_iss_id', ccs.tax_code_municipality_iss_id,
              'tax_code_municipality_iss',
                CASE
                  WHEN tcm.id IS NOT NULL THEN
                    jsonb_build_object(
                      'id', tcm.id,
                      'municipality_code', tcm.municipality_code,
                      'code', tcm.code,
                      'description', tcm.description,
                      'created_at', tcm.created_at
                    )
                  ELSE NULL
                END,

              'issqn_national_service',
                CASE
                  WHEN s.issqn_national_service_id IS NOT NULL THEN
                    (
                      SELECT to_jsonb(
                        jsonb_build_object(
                          'id', insc.id,
                          'code', insc.code,
                          'description', insc.description,
                          'created_at', insc.created_at
                        )
                      )
                      FROM issqn_national_service_codes insc
                      WHERE insc.id = s.issqn_national_service_id
                    )
                  ELSE NULL
                END
            )
          ) AS service_items
        FROM company_cnae_services ccs
        JOIN service_items_lc116 s ON s.id = ccs.service_id
        LEFT JOIN tax_code_municipality_iss tcm ON tcm.id = ccs.tax_code_municipality_iss_id
        WHERE ccs.company_id = c.id
          AND ccs.municipality_cnae_id IS NOT NULL
        GROUP BY ccs.municipality_cnae_id
      ) AS grouped_municipal
      JOIN municipality_cnaes mc ON mc.id = grouped_municipal.municipality_cnae_id
    ) AS all_cnaes
  ) AS cnaes,

  -- Certificados de acesso
  (
    SELECT jsonb_agg(
      jsonb_build_object(
        'id',              ca.id,
        'name',            ca.name,
        'type',            ca.type,
        'user_login',      ca.user_login,
        'password',        ca.password,
        'expiration_date', ca.expiration_date,
        'active',          ca.active,
        'certificate_url', ca.certificate_url,
        'created_at',      ca.created_at
      )
    )
    FROM company_certificates_access cca
    JOIN certificates_access ca ON cca.certificate_access_id = ca.id
    WHERE cca.company_id = c.id
  ) AS certificates_access,

  -- Documentos fiscais permitidos
  (
    SELECT jsonb_agg(
      jsonb_build_object(
        'id',                      cafd.id,
        'fiscal_document_type_id', fdt.id,
        'code',                    fdt.code,
        'name',                    fdt.name,
        'description',             fdt.description,
        'is_active',               cafd.is_active,
        'max_monthly_limit',       cafd.max_monthly_limit
      )
    )
    FROM company_allowed_fiscal_documents cafd
    JOIN fiscal_document_types fdt ON fdt.id = cafd.fiscal_document_type_id
    WHERE cafd.company_id = c.id
      AND fdt.active = true
  ) AS allowed_fiscal_documents,

  -- Clientes paginados
  public.get_company_customers_paginated(c.id, 1, 50, 'created_at', 'desc') AS company_customers

FROM companies c
LEFT JOIN legal_natures ln ON ln.id = c.legal_nature_id
LEFT JOIN tenants t ON t.id = c.tenant_id
LEFT JOIN focus_integration fi ON fi.company_id = c.id;

