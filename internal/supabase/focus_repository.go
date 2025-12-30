package supabase

import (
	"encoding/json"
	"fmt"
	"time"
)

func InsertFocusIntegration(companyID string, focusCompanyID string, tokenFocusCompany string) error {
	c := GetClient()
	if c == nil {
		return fmt.Errorf("supabase client não inicializado")
	}

	_, _, err := c.
		From("focus_integration").
		Insert(map[string]any{
			"company_id":          companyID,
			"focus_company_id":    focusCompanyID,
			"token_focus_company": tokenFocusCompany,
		}, false, "", "", "").
		Execute()

	return err
}

func UpdateCompanyFocusIntegrated(companyID string, integrated bool) error {
	c := GetClient()
	if c == nil {
		return fmt.Errorf("supabase client não inicializado")
	}

	update := map[string]any{
		"focus_integrated": integrated,
		"updated_at":       time.Now().UTC(),
	}

	_, _, err := c.
		From("companies").
		Update(update, "", "").
		Eq("id", companyID).
		Execute()

	return err
}

func UpdateCertificateDatesForCompany(companyID string, effectiveDate *time.Time, expirationDate *time.Time) error {
	c := GetClient()
	if c == nil {
		return fmt.Errorf("supabase client não inicializado")
	}

	// IMPORTANT:
	// After segmented-schema migration, certificate tables live under `company_private` (not exposed via PostgREST).
	// So we update via a SECURITY DEFINER RPC callable by service_role:
	//   public.rpc_service_update_certificate_dates_for_company(company_id, effective_date, expiration_date)
	payload := map[string]any{
		"p_company_id": companyID,
		// allow nulls
		"p_effective_date":   effectiveDate,
		"p_expiration_date":  expirationDate,
	}

	_, err := RpcPublic("rpc_service_update_certificate_dates_for_company", payload)
	if err != nil {
		return fmt.Errorf("erro ao atualizar datas do certificado via RPC: %w", err)
	}
	return nil
}

// InsertFocusIntegrationError salva um log de erro da Focus API na tabela focus_integration_errors.
// certificateID é opcional e só deve ser preenchido quando o erro for relacionado ao certificado.
func InsertFocusIntegrationError(companyID string, code string, message string, errors json.RawMessage, certificateID string) error {
	c := GetClient()
	if c == nil {
		return fmt.Errorf("supabase client não inicializado")
	}
	if companyID == "" {
		return fmt.Errorf("company_id é obrigatório")
	}

	payload := map[string]any{
		"company_id": companyID,
		"code":       code,
		"message":    message,
		"errors":     errors,
	}

	// Só adiciona certificates_id se foi fornecido
	if certificateID != "" {
		payload["certificates_id"] = certificateID
	}

	_, _, err := c.
		From("focus_integration_errors").
		Insert(payload, false, "", "", "").
		Execute()

	return err
}

// DeleteFocusIntegrationErrorsByCertificate remove todos os erros de integração Focus
// relacionados a um certificado específico de uma empresa.
// Deve ser chamado após upload bem-sucedido do certificado.
func DeleteFocusIntegrationErrorsByCertificate(companyID string, certificateID string) error {
	c := GetClient()
	if c == nil {
		return fmt.Errorf("supabase client não inicializado")
	}
	if companyID == "" || certificateID == "" {
		return fmt.Errorf("company_id e certificate_id são obrigatórios")
	}

	_, _, err := c.
		From("focus_integration_errors").
		Delete("", "").
		Eq("company_id", companyID).
		Eq("certificates_id", certificateID).
		Execute()

	return err
}


