package handler

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
	"bytes"

	"github.com/go-chi/chi/v5"
	"github.com/seuuser/focus-integration-service/internal/focus"
	"github.com/seuuser/focus-integration-service/internal/model"
	"github.com/seuuser/focus-integration-service/internal/supabase"
)

// RawPayload é usado para documentar payloads grandes/variáveis da Focus (empresas).
// No runtime, os handlers aceitam/replicam JSON como bytes, sem acoplamento rígido ao schema.
type RawPayload map[string]any

type EmpresasHandler struct {
	focus *focus.Client
}

func NewEmpresasHandler(focusClient *focus.Client) *EmpresasHandler {
	return &EmpresasHandler{focus: focusClient}
}

// CreateEmpresa godoc
// @Summary      Cria uma nova empresa na Focus
// @Description  Proxy para Focus: POST /v2/empresas
// @Tags         Empresas
// @Accept       json
// @Produce      json
// @Param        company_id  query  string  true  "ID da empresa (companies.id)"
// @Param        payload  body      model.FocusEmpresaCreateRequest  true  "Dados da empresa"
// @Success      201      {object}  model.FocusEmpresaResponse
// @Failure      400      {object}  RawPayload
// @Failure      401      {object}  RawPayload
// @Failure      429      {object}  RawPayload
// @Failure      500      {object}  RawPayload
// @Router       /v2/empresas [post]
func (h *EmpresasHandler) CreateEmpresa(w http.ResponseWriter, r *http.Request) {
	companyID := r.URL.Query().Get("company_id")
	if companyID == "" {
		writeJSONError(w, http.StatusBadRequest, "company_id é obrigatório")
		return
	}

	body, ok := readJSONBody(w, r)
	if !ok {
		return
	}

	// Extrai o database_local_certificate_id (para salvar logs em caso de erro de certificado),
	// e remove esse campo antes de enviar para a Focus.
	var databaseLocalCertificateID string
	var focusBodyBytes = body
	{
		var m map[string]any
		if err := json.Unmarshal(body, &m); err == nil {
			if v, exists := m["database_local_certificate_id"]; exists {
				if s, ok := v.(string); ok && s != "" {
					databaseLocalCertificateID = s
				}
				delete(m, "database_local_certificate_id")
				if b, err := json.Marshal(m); err == nil {
					focusBodyBytes = b
				}
			}
		}
	}

	resp, err := h.focus.CreateEmpresa(r.Context(), focusBodyBytes)
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, "Empresa cadastrada, mas houve um problema ao integrar com a Focus. Tente novamente mais tarde.")
		return
	}
	defer resp.Body.Close()

	// Lê o body para poder logar o retorno da Focus e ainda devolver para o client.
	respBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		writeJSONError(w, http.StatusInternalServerError, "erro ao ler resposta da Focus")
		return
	}

	// Se a Focus devolveu erro, salva no log e devolve mensagem amigável (empresa já foi criada no Supabase).
	if resp.StatusCode >= 400 {
		// Parseia o erro da Focus para salvar no banco
		var focusError struct {
			Codigo   string `json:"codigo"`
			Mensagem string `json:"mensagem"`
			Erros    []struct {
				Campo string `json:"campo"`
			} `json:"erros"`
		}
		
		// Verifica se é erro de certificado para incluir o company_certificates_access_id
		var certificateAccessID string
		if err := json.Unmarshal(respBytes, &focusError); err == nil {
			for _, e := range focusError.Erros {
				if e.Campo == "arquivo_certificado_base64" && databaseLocalCertificateID != "" {
					certificateAccessID = databaseLocalCertificateID
					break
				}
			}
		}

		// Re-parseia para pegar o array de erros como RawMessage
		var focusErrorRaw struct {
			Codigo   string          `json:"codigo"`
			Mensagem string          `json:"mensagem"`
			Erros    json.RawMessage `json:"erros"`
		}
		if err := json.Unmarshal(respBytes, &focusErrorRaw); err == nil {
			// Salva o erro na tabela focus_integration_errors
			if logErr := supabase.InsertFocusIntegrationError(
				companyID,
				focusErrorRaw.Codigo,
				focusErrorRaw.Mensagem,
				focusErrorRaw.Erros,
				certificateAccessID,
			); logErr != nil {
				log.Printf("[supabase] erro ao salvar log de integração Focus: %v", logErr)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error":        "Empresa cadastrada, mas houve um problema ao integrar com a Focus. Verifique os dados e tente novamente.",
			"focus_status": resp.StatusCode,
			"focus_body":   json.RawMessage(respBytes),
		})
		return
	}

	// Extrai dados para persistência no Supabase.
	var focusResp struct {
		ID                  any    `json:"id"`
		TokenProducao       string `json:"token_producao"`
		CertificadoValidoAte string `json:"certificado_valido_ate"`
		CertificadoValidoDe  string `json:"certificado_valido_de"`
	}
	_ = json.Unmarshal(respBytes, &focusResp)

	focusCompanyID := ""
	switch v := focusResp.ID.(type) {
	case float64:
		focusCompanyID = strconv.FormatInt(int64(v), 10)
	case int:
		focusCompanyID = strconv.Itoa(v)
	case string:
		focusCompanyID = v
	}

	var expDate *time.Time
	if focusResp.CertificadoValidoAte != "" {
		if t, err := time.Parse(time.RFC3339, focusResp.CertificadoValidoAte); err == nil {
			expDate = &t
		} else {
			log.Printf("[focus] warning: failed to parse certificado_valido_ate=%q as RFC3339: %v", focusResp.CertificadoValidoAte, err)
		}
	}
	var effDate *time.Time
	if focusResp.CertificadoValidoDe != "" {
		if t, err := time.Parse(time.RFC3339, focusResp.CertificadoValidoDe); err == nil {
			effDate = &t
		} else {
			log.Printf("[focus] warning: failed to parse certificado_valido_de=%q as RFC3339: %v", focusResp.CertificadoValidoDe, err)
		}
	}

	// Persiste integração no Supabase; se falhar, não quebra o retorno da Focus,
	// mas sinaliza via header para o front tratar.
	var warn string
	if focusCompanyID != "" && focusResp.TokenProducao != "" {
		if err := supabase.InsertFocusIntegration(companyID, focusCompanyID, focusResp.TokenProducao); err != nil {
			warn = "Cadastro realizado na Focus, mas não foi possível salvar os dados de integração no Supabase."
			log.Printf("[supabase] insert focus_integration failed: %v", err)
		}
		if err := supabase.UpdateCompanyFocusIntegrated(companyID, true); err != nil {
			if warn == "" {
				warn = "Cadastro realizado na Focus, mas não foi possível atualizar o status de integração no Supabase."
			}
			log.Printf("[supabase] update companies.focus_integrated failed: %v", err)
		}
		if err := supabase.UpdateCertificateDatesForCompany(companyID, effDate, expDate); err != nil {
			if warn == "" {
				warn = "Cadastro realizado na Focus, mas não foi possível atualizar as datas do certificado no Supabase."
			}
			log.Printf("[supabase] update certificates_access dates failed: %v", err)
		} else {
			log.Printf("[supabase] certificate dates updated (company_id=%s, effective_date=%v, expiration_date=%v)", companyID, effDate, expDate)
		}
		
		// Remove erros antigos do certificado (se houver) após sucesso
		if databaseLocalCertificateID != "" {
			if err := supabase.DeleteFocusIntegrationErrorsByCertificate(companyID, databaseLocalCertificateID); err != nil {
				log.Printf("[supabase] erro ao remover erros antigos do certificado: %v", err)
				// Não adiciona warning pois não é crítico
			} else {
				log.Printf("[supabase] erros antigos do certificado removidos com sucesso")
			}
		}
	} else {
		warn = "Cadastro realizado na Focus, mas não foi possível identificar id/token/datas para persistir no Supabase."
	}

	copyHeaderIfPresent(w, resp, "Content-Type")
	copyHeaderIfPresent(w, resp, "X-Total-Count")
	copyHeaderIfPresent(w, resp, "Rate-Limit-Limit")
	copyHeaderIfPresent(w, resp, "Rate-Limit-Remaining")
	copyHeaderIfPresent(w, resp, "Rate-Limit-Reset")
	if warn != "" {
		w.Header().Set("X-Integration-Warning", warn)
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(respBytes)
}

// ListEmpresas godoc
// @Summary      Lista empresas na Focus
// @Description  Proxy para Focus: GET /v2/empresas (suporta cnpj, cpf, offset)
// @Tags         Empresas
// @Produce      json
// @Param        cnpj    query     string  false  "CNPJ (somente números)"
// @Param        cpf     query     string  false  "CPF (somente números)"
// @Param        offset  query     int     false  "Paginação (offset)"
// @Success      200     {array}   model.FocusEmpresaResponse
// @Failure      401     {object}  RawPayload
// @Failure      429     {object}  RawPayload
// @Failure      500     {object}  RawPayload
// @Router       /v2/empresas [get]
func (h *EmpresasHandler) ListEmpresas(w http.ResponseWriter, r *http.Request) {
	resp, err := h.focus.ListEmpresas(r.Context(), r.URL.RawQuery)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer resp.Body.Close()

	proxyResponse(w, resp)
}

// GetEmpresa godoc
// @Summary      Consulta uma empresa por ID na Focus
// @Description  Proxy para Focus: GET /v2/empresas/{id}
// @Tags         Empresas
// @Produce      json
// @Param        id   path      string  true  "ID da empresa na Focus"
// @Success      200  {object}  model.FocusEmpresaResponse
// @Failure      401  {object}  RawPayload
// @Failure      404  {object}  RawPayload
// @Failure      429  {object}  RawPayload
// @Failure      500  {object}  RawPayload
// @Router       /v2/empresas/{id} [get]
func (h *EmpresasHandler) GetEmpresa(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeJSONError(w, http.StatusBadRequest, "id é obrigatório")
		return
	}

	resp, err := h.focus.GetEmpresa(r.Context(), id)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer resp.Body.Close()

	proxyResponse(w, resp)
}

// UpdateEmpresa godoc
// @Summary      Altera uma empresa específica na Focus
// @Description  Proxy para Focus: PUT /v2/empresas/{id}. Apenas os campos enviados serão atualizados. Se atualizar certificado com sucesso, erros antigos serão removidos.
// @Tags         Empresas
// @Accept       json
// @Produce      json
// @Param        id            path      string      true   "ID da empresa na Focus"
// @Param        company_id    query     string      false  "ID da empresa no Supabase (para limpeza de erros)"
// @Param        certificate_id query    string      false  "ID do certificado no Supabase (para limpeza de erros)"
// @Param        payload       body      model.FocusEmpresaUpdateRequest  true  "Dados para atualização (campos opcionais)"
// @Success      200           {object}  model.FocusEmpresaResponse
// @Failure      400           {object}  RawPayload
// @Failure      401           {object}  RawPayload
// @Failure      404           {object}  RawPayload
// @Failure      429           {object}  RawPayload
// @Failure      500           {object}  RawPayload
// @Router       /v2/empresas/{id} [put]
func (h *EmpresasHandler) UpdateEmpresa(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeJSONError(w, http.StatusBadRequest, "id é obrigatório")
		return
	}

	// Query parameters opcionais para limpeza de erros
	companyID := r.URL.Query().Get("company_id")
	certificateID := r.URL.Query().Get("certificate_id")

	body, ok := readJSONBody(w, r)
	if !ok {
		return
	}

	// Valida e processa o payload de update, garantindo que apenas campos enviados sejam incluídos
	var updatePayload model.FocusEmpresaUpdateRequest
	if err := json.Unmarshal(body, &updatePayload); err != nil {
		writeJSONError(w, http.StatusBadRequest, "payload inválido: "+err.Error())
		return
	}

	// Re-serializa para garantir que apenas campos não-nil sejam enviados (omitempty)
	cleanBody, err := json.Marshal(updatePayload)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "erro ao processar payload")
		return
	}

	// Valida que pelo menos um campo foi enviado para atualização
	var checkEmpty map[string]interface{}
	_ = json.Unmarshal(cleanBody, &checkEmpty)
	if len(checkEmpty) == 0 {
		writeJSONError(w, http.StatusBadRequest, "nenhum campo informado para atualização")
		return
	}

	// Verifica se está atualizando certificado
	hasCertificateUpdate := updatePayload.ArquivoCertBase64 != nil || updatePayload.SenhaCertificado != nil

	log.Printf("[focus] PUT /v2/empresas/%s -> campos a atualizar: %d (certificado: %v)", id, len(checkEmpty), hasCertificateUpdate)

	resp, err := h.focus.UpdateEmpresa(r.Context(), id, cleanBody)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer resp.Body.Close()

	// Read body so we can (a) parse certificate dates and (b) still proxy the response.
	respBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		writeJSONError(w, http.StatusInternalServerError, "erro ao ler resposta da Focus")
		return
	}

	// Restore body for proxying
	resp.Body = io.NopCloser(bytes.NewReader(respBytes))

	// If success and certificate was updated, try to update certificate dates in Supabase
	if resp.StatusCode >= 200 && resp.StatusCode < 300 && hasCertificateUpdate && companyID != "" {
		// Extract cert dates (same fields used by CreateEmpresa)
		var focusResp struct {
			CertificadoValidoAte string `json:"certificado_valido_ate"`
			CertificadoValidoDe  string `json:"certificado_valido_de"`
		}
		_ = json.Unmarshal(respBytes, &focusResp)

		var expDate *time.Time
		if focusResp.CertificadoValidoAte != "" {
			if t, err := time.Parse(time.RFC3339, focusResp.CertificadoValidoAte); err == nil {
				expDate = &t
			} else {
				log.Printf("[focus] warning: failed to parse certificado_valido_ate=%q as RFC3339: %v", focusResp.CertificadoValidoAte, err)
			}
		}
		var effDate *time.Time
		if focusResp.CertificadoValidoDe != "" {
			if t, err := time.Parse(time.RFC3339, focusResp.CertificadoValidoDe); err == nil {
				effDate = &t
			} else {
				log.Printf("[focus] warning: failed to parse certificado_valido_de=%q as RFC3339: %v", focusResp.CertificadoValidoDe, err)
			}
		}

		if err := supabase.UpdateCertificateDatesForCompany(companyID, effDate, expDate); err != nil {
			log.Printf("[supabase] update certificates_access dates failed (update flow): %v", err)
		} else {
			log.Printf("[supabase] certificate dates updated (update flow) (company_id=%s, effective_date=%v, expiration_date=%v)", companyID, effDate, expDate)
		}
	}

	// If success and we have IDs, clean old certificate errors
	if resp.StatusCode >= 200 && resp.StatusCode < 300 && hasCertificateUpdate && companyID != "" && certificateID != "" {
		if err := supabase.DeleteFocusIntegrationErrorsByCertificate(companyID, certificateID); err != nil {
			log.Printf("[supabase] erro ao remover erros antigos do certificado no update: %v", err)
		} else {
			log.Printf("[supabase] erros antigos do certificado removidos com sucesso no update")
		}
	}

	proxyResponse(w, resp)
}

// DeleteEmpresa godoc
// @Summary      Exclui uma empresa na Focus
// @Description  Proxy para Focus: DELETE /v2/empresas/{id}
// @Tags         Empresas
// @Produce      json
// @Param        id   path      string  true  "ID da empresa na Focus"
// @Success      200  {object}  model.FocusEmpresaResponse
// @Failure      401  {object}  RawPayload
// @Failure      404  {object}  RawPayload
// @Failure      429  {object}  RawPayload
// @Failure      500  {object}  RawPayload
// @Router       /v2/empresas/{id} [delete]
func (h *EmpresasHandler) DeleteEmpresa(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeJSONError(w, http.StatusBadRequest, "id é obrigatório")
		return
	}

	resp, err := h.focus.DeleteEmpresa(r.Context(), id)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer resp.Body.Close()

	proxyResponse(w, resp)
}

func readJSONBody(w http.ResponseWriter, r *http.Request) ([]byte, bool) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "não foi possível ler o corpo da requisição")
		return nil, false
	}
	if len(body) == 0 {
		writeJSONError(w, http.StatusBadRequest, "payload JSON é obrigatório")
		return nil, false
	}

	var tmp any
	if err := json.Unmarshal(body, &tmp); err != nil {
		writeJSONError(w, http.StatusBadRequest, "JSON inválido")
		return nil, false
	}

	return body, true
}

func proxyResponse(w http.ResponseWriter, resp *http.Response) {
	copyHeaderIfPresent(w, resp, "Content-Type")
	copyHeaderIfPresent(w, resp, "X-Total-Count")
	copyHeaderIfPresent(w, resp, "Rate-Limit-Limit")
	copyHeaderIfPresent(w, resp, "Rate-Limit-Remaining")
	copyHeaderIfPresent(w, resp, "Rate-Limit-Reset")

	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func copyHeaderIfPresent(w http.ResponseWriter, resp *http.Response, key string) {
	if v := resp.Header.Get(key); v != "" {
		w.Header().Set(key, v)
	}
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": message,
	})
}


