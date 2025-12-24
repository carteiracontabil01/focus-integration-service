package supabase

import (
	"bytes"
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

	// Busca 1 certificado associado à empresa (associação via company_certificates_access).
	body, _, err := c.
		From("company_certificates_access").
		Select("certificate_access_id", "", false).
		Eq("company_id", companyID).
		Limit(1, "").
		Execute()
	if err != nil {
		return fmt.Errorf("erro ao buscar certificate_access_id: %w", err)
	}

	var rows []struct {
		CertificateAccessID string `json:"certificate_access_id"`
	}
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&rows); err != nil {
		return fmt.Errorf("erro ao decodificar certificate_access_id: %w", err)
	}
	if len(rows) == 0 || rows[0].CertificateAccessID == "" {
		return nil // nada para atualizar
	}

	update := map[string]any{}
	if expirationDate != nil {
		update["expiration_date"] = *expirationDate
	}
	if effectiveDate != nil {
		update["effective_date"] = *effectiveDate
	}
	// Marca o certificado como ativo após integração bem-sucedida na Focus
	update["active"] = true
	
	if len(update) == 0 {
		return nil
	}

	_, _, err = c.
		From("certificates_access").
		Update(update, "", "").
		Eq("id", rows[0].CertificateAccessID).
		Execute()

	return err
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


