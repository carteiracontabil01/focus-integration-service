package handler

import (
	"net/http"
	"regexp"

	"github.com/go-chi/chi/v5"
	"github.com/seuuser/focus-integration-service/internal/focus"
	"github.com/seuuser/focus-integration-service/internal/model"
)

var cnpjDigits14 = regexp.MustCompile(`^\d{14}$`)

type CnpjsHandler struct {
	focus *focus.Client
}

func NewCnpjsHandler(focusClient *focus.Client) *CnpjsHandler {
	return &CnpjsHandler{focus: focusClient}
}

// GetCnpj godoc
// @Summary      Consulta cadastro de CNPJ
// @Description  Proxy para Focus: GET /v2/cnpjs/{cnpj}. Informe 14 dígitos (somente números).
// @Tags         CNPJs
// @Produce      json
// @Param        cnpj  path      string  true  "CNPJ (14 dígitos, somente números)"
// @Success      200   {object}  model.FocusCnpjResponse
// @Failure      400   {object}  RawPayload
// @Failure      401   {object}  RawPayload
// @Failure      404   {object}  RawPayload
// @Failure      429   {object}  RawPayload
// @Failure      500   {object}  RawPayload
// @Router       /v2/cnpjs/{cnpj} [get]
func (h *CnpjsHandler) GetCnpj(w http.ResponseWriter, r *http.Request) {
	cnpj := chi.URLParam(r, "cnpj")
	if cnpj == "" {
		writeJSONError(w, http.StatusBadRequest, "cnpj é obrigatório")
		return
	}
	if !cnpjDigits14.MatchString(cnpj) {
		writeJSONError(w, http.StatusBadRequest, "cnpj inválido: informe 14 dígitos numéricos (somente números)")
		return
	}

	resp, err := h.focus.GetCNPJ(r.Context(), cnpj)
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, err.Error())
		return
	}
	defer resp.Body.Close()

	proxyResponse(w, resp)
}

// keep model import used for swag (avoid unused if build tags differ)
var _ = model.FocusCnpjResponse{}


