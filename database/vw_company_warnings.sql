-- ========================================================================
-- VIEW: vw_company_warnings
-- Descrição: Retorna warnings/alertas para cada empresa
-- ========================================================================

CREATE OR REPLACE VIEW public.vw_company_warnings AS
WITH 
-- Erros de integração Focus mais recentes por empresa
focus_errors AS (
  SELECT DISTINCT ON (fie.company_id)
    fie.company_id,
    fie.code AS error_code,
    fie.message AS error_message,
    fie.errors AS error_details,
    fie.certificates_id,
    fie.created_at AS error_created_at
  FROM public.focus_integration_errors fie
  ORDER BY fie.company_id, fie.created_at DESC
),

-- Certificados por empresa
company_certs AS (
  SELECT 
    cca.company_id,
    COUNT(DISTINCT cca.certificate_access_id) AS total_certificates,
    MAX(ca.expiration_date) AS latest_expiration_date
  FROM public.company_certificates_access cca
  LEFT JOIN public.certificates_access ca ON ca.id = cca.certificate_access_id
  WHERE ca.active = true
  GROUP BY cca.company_id
),

-- Documentos fiscais habilitados por empresa
company_fiscal_docs AS (
  SELECT 
    cafd.company_id,
    COUNT(DISTINCT cafd.fiscal_document_type_id) AS total_fiscal_documents,
    BOOL_OR(fdt.name ILIKE '%nfse%') AS has_nfse
  FROM public.company_allowed_fiscal_documents cafd
  LEFT JOIN public.fiscal_document_types fdt ON fdt.id = cafd.fiscal_document_type_id
  WHERE cafd.is_active = true
  GROUP BY cafd.company_id
),

-- CNAEs por empresa (corrigido para usar company_cnae_services)
company_cnaes AS (
  SELECT 
    ccs.company_id,
    COUNT(DISTINCT COALESCE(ccs.cnae_id, ccs.municipality_cnae_id)) AS total_cnaes
  FROM public.company_cnae_services ccs
  WHERE ccs.cnae_id IS NOT NULL OR ccs.municipality_cnae_id IS NOT NULL
  GROUP BY ccs.company_id
),

-- Endereços por empresa (corrigido para usar company_address + addresses)
company_addresses AS (
  SELECT 
    ca.company_id,
    COUNT(DISTINCT ca.id) AS total_addresses,
    BOOL_AND(
      a.address IS NOT NULL AND a.address != '' AND
      a.number IS NOT NULL AND a.number != '' AND
      a.neighborhood IS NOT NULL AND a.neighborhood != '' AND
      a.zip_code IS NOT NULL AND a.zip_code != ''
    ) AS has_complete_address
  FROM public.company_address ca
  INNER JOIN public.addresses a ON a.id = ca.address_id
  GROUP BY ca.company_id
),

-- Warnings gerados
warnings AS (
  SELECT 
    c.id AS company_id,
    c.cnpj,
    
    -- Warning 1: Erro de senha do certificado (Focus)
    CASE 
      WHEN fe.error_details::jsonb->0->>'mensagem' ILIKE '%senha%' THEN
        jsonb_build_object(
          'id', 'focus_certificate_password_error',
          'severity', 'error',
          'title', 'Senha do Certificado Incorreta',
          'message', 'A senha do certificado digital está incorreta. Verifique a senha e tente cadastrar novamente.',
          'icon', 'pi pi-lock',
          'action_label', 'Atualizar Certificado',
          'action_route', '/companies/' || c.id || '/info'
        )
      ELSE NULL
    END AS warning_cert_password,
    
    -- Warning 2: Certificado de outro CNPJ (Focus)
    CASE 
      WHEN fe.error_details::jsonb->0->>'mensagem' ILIKE '%não pertence ao cnpj%' THEN
        jsonb_build_object(
          'id', 'focus_certificate_wrong_cnpj',
          'severity', 'error',
          'title', 'Certificado de Outro CNPJ',
          'message', 'O certificado não pertence ao CNPJ ' || c.cnpj || '. Utilize o certificado correto desta empresa.',
          'icon', 'pi pi-times-circle',
          'action_label', 'Corrigir Certificado',
          'action_route', '/companies/' || c.id || '/info'
        )
      ELSE NULL
    END AS warning_wrong_cnpj,
    
    -- Warning 3: Certificado vencido (Focus)
    CASE 
      WHEN fe.error_details::jsonb->0->>'mensagem' ILIKE '%vencido%' OR 
           fe.error_details::jsonb->0->>'mensagem' ILIKE '%validade%' THEN
        jsonb_build_object(
          'id', 'focus_certificate_expired_error',
          'severity', 'error',
          'title', 'Certificado Vencido na Integração',
          'message', 'O certificado enviado já está vencido. Renove o certificado e cadastre novamente.',
          'icon', 'pi pi-calendar-times',
          'action_label', 'Renovar Certificado',
          'action_route', '/companies/' || c.id || '/info'
        )
      ELSE NULL
    END AS warning_cert_expired_focus,
    
    -- Warning 4: Formato do certificado inválido (Focus)
    CASE 
      WHEN fe.error_details::jsonb->0->>'mensagem' ILIKE '%formato%' OR
           fe.error_details::jsonb->0->>'mensagem' ILIKE '%pfx%' OR
           fe.error_details::jsonb->0->>'mensagem' ILIKE '%p12%' OR
           fe.error_details::jsonb->0->>'mensagem' ILIKE '%base64%' THEN
        jsonb_build_object(
          'id', 'focus_certificate_format_error',
          'severity', 'error',
          'title', 'Formato do Certificado Inválido',
          'message', 'O arquivo do certificado está em formato inválido. Use um arquivo PFX ou P12 válido, codificado em Base64.',
          'icon', 'pi pi-file-excel',
          'action_label', 'Enviar Novamente',
          'action_route', '/companies/' || c.id || '/info'
        )
      ELSE NULL
    END AS warning_cert_format,
    
    -- Warning 5: Certificado ausente
    CASE 
      WHEN COALESCE(cc.total_certificates, 0) = 0 THEN
        jsonb_build_object(
          'id', 'missing_certificate',
          'severity', 'error',
          'title', 'Certificado Digital Ausente',
          'message', 'Esta empresa não possui certificado digital cadastrado. Configure o certificado para emitir notas fiscais.',
          'icon', 'pi pi-shield',
          'action_label', 'Cadastrar Certificado',
          'action_route', '/companies/' || c.id || '/info'
        )
      ELSE NULL
    END AS warning_no_certificate,
    
    -- Warning 6: Certificado vencido (local)
    CASE 
      WHEN cc.latest_expiration_date IS NOT NULL AND cc.latest_expiration_date < CURRENT_DATE THEN
        jsonb_build_object(
          'id', 'certificate_expired',
          'severity', 'error',
          'title', 'Certificado Digital Vencido',
          'message', 'O certificado digital venceu há ' || (CURRENT_DATE - cc.latest_expiration_date) || ' dias. Atualize o certificado para continuar emitindo notas fiscais.',
          'icon', 'pi pi-shield',
          'action_label', 'Atualizar Certificado',
          'action_route', '/companies/' || c.id || '/info'
        )
      ELSE NULL
    END AS warning_cert_expired,
    
    -- Warning 7: Certificado expirando em até 7 dias (crítico)
    CASE 
      WHEN cc.latest_expiration_date IS NOT NULL AND 
           cc.latest_expiration_date >= CURRENT_DATE AND
           cc.latest_expiration_date <= (CURRENT_DATE + INTERVAL '7 days') THEN
        jsonb_build_object(
          'id', 'certificate_expiring_soon_critical',
          'severity', 'error',
          'title', 'Certificado Expirando em ' || (cc.latest_expiration_date - CURRENT_DATE) || ' Dias',
          'message', 'O certificado digital irá vencer em ' || TO_CHAR(cc.latest_expiration_date, 'DD/MM/YYYY') || '. Providencie a renovação com urgência.',
          'icon', 'pi pi-calendar-times',
          'action_label', 'Renovar Agora',
          'action_route', '/companies/' || c.id || '/info'
        )
      ELSE NULL
    END AS warning_cert_expiring_critical,
    
    -- Warning 8: Certificado expirando em 8-30 dias
    CASE 
      WHEN cc.latest_expiration_date IS NOT NULL AND 
           cc.latest_expiration_date > (CURRENT_DATE + INTERVAL '7 days') AND
           cc.latest_expiration_date <= (CURRENT_DATE + INTERVAL '30 days') THEN
        jsonb_build_object(
          'id', 'certificate_expiring_soon',
          'severity', 'warn',
          'title', 'Certificado Expirando em ' || (cc.latest_expiration_date - CURRENT_DATE) || ' Dias',
          'message', 'O certificado digital irá vencer em ' || TO_CHAR(cc.latest_expiration_date, 'DD/MM/YYYY') || '. Providencie a renovação em breve.',
          'icon', 'pi pi-calendar-times',
          'action_label', 'Renovar Certificado',
          'action_route', '/companies/' || c.id || '/info'
        )
      ELSE NULL
    END AS warning_cert_expiring,
    
    -- Warning 9: Documentos fiscais não habilitados
    CASE 
      WHEN COALESCE(cfd.total_fiscal_documents, 0) = 0 THEN
        jsonb_build_object(
          'id', 'missing_fiscal_documents',
          'severity', 'error',
          'title', 'Documentos Fiscais Não Habilitados',
          'message', 'Esta empresa não possui nenhum documento fiscal habilitado (NFSe, NFSe Nacional, etc.). Configure para começar a emitir notas.',
          'icon', 'pi pi-file',
          'action_label', 'Habilitar Documentos',
          'action_route', '/companies/' || c.id || '/info'
        )
      ELSE NULL
    END AS warning_no_fiscal_docs,
    
    -- Warning 10: CNAE não cadastrado
    CASE 
      WHEN COALESCE(cn.total_cnaes, 0) = 0 THEN
        jsonb_build_object(
          'id', 'missing_cnae',
          'severity', 'warn',
          'title', 'CNAE Não Cadastrado',
          'message', 'Esta empresa não possui CNAE cadastrado. Configure os CNAEs para emitir notas fiscais corretamente.',
          'icon', 'pi pi-list',
          'action_label', 'Cadastrar CNAE',
          'action_route', '/companies/' || c.id || '/info'
        )
      ELSE NULL
    END AS warning_no_cnae,
    
    -- Warning 11: Requisitos NFSe incompletos
    CASE 
      WHEN cfd.has_nfse = true AND 
           (c.municipal_registration IS NULL OR c.municipal_registration = '' OR c.municipality_id IS NULL) THEN
        jsonb_build_object(
          'id', 'missing_nfse_requirements',
          'severity', 'warn',
          'title', 'Requisitos NFSe Incompletos',
          'message', 'Para emitir NFSe é necessário cadastrar: ' || 
                     CASE 
                       WHEN (c.municipal_registration IS NULL OR c.municipal_registration = '') AND c.municipality_id IS NULL THEN 'Inscrição Municipal, Município'
                       WHEN c.municipal_registration IS NULL OR c.municipal_registration = '' THEN 'Inscrição Municipal'
                       WHEN c.municipality_id IS NULL THEN 'Município'
                       ELSE ''
                     END || '. Complete o cadastro.',
          'icon', 'pi pi-exclamation-circle',
          'action_label', 'Completar Cadastro',
          'action_route', '/companies/' || c.id || '/info'
        )
      ELSE NULL
    END AS warning_nfse_requirements,
    
    -- Warning 12: Inscrição municipal ausente
    CASE 
      WHEN c.municipal_registration IS NULL OR c.municipal_registration = '' THEN
        jsonb_build_object(
          'id', 'missing_municipal_registration',
          'severity', 'info',
          'title', 'Inscrição Municipal Ausente',
          'message', 'Esta empresa não possui inscrição municipal cadastrada.',
          'icon', 'pi pi-id-card',
          'action_label', 'Cadastrar',
          'action_route', '/companies/' || c.id || '/info'
        )
      ELSE NULL
    END AS warning_no_municipal_reg,
    
    -- Warning 13: Endereço não cadastrado
    CASE 
      WHEN COALESCE(ca.total_addresses, 0) = 0 THEN
        jsonb_build_object(
          'id', 'incomplete_address',
          'severity', 'info',
          'title', 'Endereço Não Cadastrado',
          'message', 'Esta empresa não possui endereço cadastrado. Complete os dados para melhor qualidade das informações.',
          'icon', 'pi pi-map-marker',
          'action_label', 'Cadastrar Endereço',
          'action_route', '/companies/' || c.id || '/info'
        )
      ELSE NULL
    END AS warning_no_address,
    
    -- Warning 14: Endereço incompleto
    CASE 
      WHEN ca.total_addresses > 0 AND ca.has_complete_address = false THEN
        jsonb_build_object(
          'id', 'incomplete_address',
          'severity', 'info',
          'title', 'Endereço Incompleto',
          'message', 'O endereço desta empresa está incompleto. Complete para melhor qualidade dos dados.',
          'icon', 'pi pi-map-marker',
          'action_label', 'Completar Endereço',
          'action_route', '/companies/' || c.id || '/info'
        )
      ELSE NULL
    END AS warning_incomplete_address
    
  FROM public.companies c
  LEFT JOIN focus_errors fe ON fe.company_id = c.id
  LEFT JOIN company_certs cc ON cc.company_id = c.id
  LEFT JOIN company_fiscal_docs cfd ON cfd.company_id = c.id
  LEFT JOIN company_cnaes cn ON cn.company_id = c.id
  LEFT JOIN company_addresses ca ON ca.company_id = c.id
  WHERE c.status = 'ACTIVE'
)

-- Resultado final: transforma colunas em array de warnings
SELECT 
  company_id,
  cnpj,
  ARRAY_REMOVE(ARRAY[
    warning_cert_password,
    warning_wrong_cnpj,
    warning_cert_expired_focus,
    warning_cert_format,
    warning_no_certificate,
    warning_cert_expired,
    warning_cert_expiring_critical,
    warning_cert_expiring,
    warning_no_fiscal_docs,
    warning_no_cnae,
    warning_nfse_requirements,
    warning_no_municipal_reg,
    warning_no_address,
    warning_incomplete_address
  ], NULL) AS warnings
FROM warnings;

-- ========================================================================
-- COMENTÁRIOS E ÍNDICES
-- ========================================================================
COMMENT ON VIEW public.vw_company_warnings IS 'View que retorna warnings/alertas para cada empresa baseado em regras de negócio';

-- Garantir que índices existam nas tabelas base (já criados anteriormente)
-- Índices em focus_integration_errors já existem
-- Índices em company_certificates_access, company_allowed_fiscal_documents, etc.
