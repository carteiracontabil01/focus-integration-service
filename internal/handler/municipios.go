package handler

import (
	"net/http"
	"regexp"

	"github.com/go-chi/chi/v5"
	"github.com/seuuser/focus-integration-service/internal/focus"
	"github.com/seuuser/focus-integration-service/internal/model"
)

var municipioCodigoDigits = regexp.MustCompile(`^\d+$`)

type MunicipiosHandler struct {
	focus *focus.Client
}

func NewMunicipiosHandler(focusClient *focus.Client) *MunicipiosHandler {
	return &MunicipiosHandler{focus: focusClient}
}

// ListMunicipios godoc
// @Summary      Lista municípios (IBGE)
// @Description  Proxy para Focus: GET /v2/municipios. Suporta filtros via querystring (sigla_uf, nome_municipio, nome, status_nfse, offset).
// @Tags         Municípios
// @Produce      json
// @Param        sigla_uf        query     string  false  "Sigla UF (ex: PR)"
// @Param        nome_municipio  query     string  false  "Nome exato do município (ex: Curitiba)"
// @Param        nome            query     string  false  "Trecho do nome do município"
// @Param        status_nfse     query     string  false  "Status NFSe (ativo, fora_do_ar, pausado, em_implementacao, em_reimplementacao, inativo, nao_implementado)"
// @Param        offset          query     int     false  "Paginação (offset)"
// @Success      200             {array}   model.FocusMunicipioResponse
// @Failure      401             {object}  RawPayload
// @Failure      429             {object}  RawPayload
// @Failure      500             {object}  RawPayload
// @Router       /v2/municipios [get]
func (h *MunicipiosHandler) ListMunicipios(w http.ResponseWriter, r *http.Request) {
	resp, err := h.focus.ListMunicipios(r.Context(), r.URL.RawQuery)
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, err.Error())
		return
	}
	defer resp.Body.Close()

	proxyResponse(w, resp)
}

// GetMunicipio godoc
// @Summary      Busca município por código IBGE
// @Description  Proxy para Focus: GET /v2/municipios/{codigo_municipio}.
// @Tags         Municípios
// @Produce      json
// @Param        codigo_municipio  path      string  true  "Código do município (IBGE)"
// @Success      200               {object}  model.FocusMunicipioResponse
// @Failure      400               {object}  RawPayload
// @Failure      401               {object}  RawPayload
// @Failure      404               {object}  RawPayload
// @Failure      429               {object}  RawPayload
// @Failure      500               {object}  RawPayload
// @Router       /v2/municipios/{codigo_municipio} [get]
func (h *MunicipiosHandler) GetMunicipio(w http.ResponseWriter, r *http.Request) {
	codigo := chi.URLParam(r, "codigo_municipio")
	if codigo == "" {
		writeJSONError(w, http.StatusBadRequest, "codigo_municipio é obrigatório")
		return
	}
	if !municipioCodigoDigits.MatchString(codigo) {
		writeJSONError(w, http.StatusBadRequest, "codigo_municipio inválido: informe somente números")
		return
	}

	resp, err := h.focus.GetMunicipio(r.Context(), codigo)
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, err.Error())
		return
	}
	defer resp.Body.Close()

	proxyResponse(w, resp)
}

// ListItensListaServico godoc
// @Summary      Lista itens da lista de serviço por município
// @Description  Proxy para Focus: GET /v2/municipios/{codigo_municipio}/itens_lista_servico. Suporta filtros via querystring (codigo, descricao, offset).
// @Tags         Municípios - Itens Lista de Serviço
// @Produce      json
// @Param        codigo_municipio  path      string  true   "Código do município (IBGE)"
// @Param        codigo            query     string  false  "Trecho do código (ex: 14.)"
// @Param        descricao         query     string  false  "Trecho da descrição"
// @Param        offset            query     int     false  "Paginação (offset)"
// @Success      200               {array}   RawPayload
// @Failure      400               {object}  RawPayload
// @Failure      401               {object}  RawPayload
// @Failure      404               {object}  RawPayload
// @Failure      429               {object}  RawPayload
// @Failure      500               {object}  RawPayload
// @Router       /v2/municipios/{codigo_municipio}/itens_lista_servico [get]
func (h *MunicipiosHandler) ListItensListaServico(w http.ResponseWriter, r *http.Request) {
	codigoMunicipio := chi.URLParam(r, "codigo_municipio")
	if codigoMunicipio == "" {
		writeJSONError(w, http.StatusBadRequest, "codigo_municipio é obrigatório")
		return
	}
	if !municipioCodigoDigits.MatchString(codigoMunicipio) {
		writeJSONError(w, http.StatusBadRequest, "codigo_municipio inválido: informe somente números")
		return
	}

	resp, err := h.focus.ListMunicipioItensListaServico(r.Context(), codigoMunicipio, r.URL.RawQuery)
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, err.Error())
		return
	}
	defer resp.Body.Close()

	proxyResponse(w, resp)
}

// GetItemListaServico godoc
// @Summary      Busca item da lista de serviço por código (município)
// @Description  Proxy para Focus: GET /v2/municipios/{codigo_municipio}/itens_lista_servico/{codigo}.
// @Tags         Municípios - Itens Lista de Serviço
// @Produce      json
// @Param        codigo_municipio  path      string  true  "Código do município (IBGE)"
// @Param        codigo            path      string  true  "Código do item (ex: 14.01)"
// @Success      200               {object}  RawPayload
// @Failure      400               {object}  RawPayload
// @Failure      401               {object}  RawPayload
// @Failure      404               {object}  RawPayload
// @Failure      429               {object}  RawPayload
// @Failure      500               {object}  RawPayload
// @Router       /v2/municipios/{codigo_municipio}/itens_lista_servico/{codigo} [get]
func (h *MunicipiosHandler) GetItemListaServico(w http.ResponseWriter, r *http.Request) {
	codigoMunicipio := chi.URLParam(r, "codigo_municipio")
	codigoItem := chi.URLParam(r, "codigo")
	if codigoMunicipio == "" || codigoItem == "" {
		writeJSONError(w, http.StatusBadRequest, "codigo_municipio e codigo são obrigatórios")
		return
	}
	if !municipioCodigoDigits.MatchString(codigoMunicipio) {
		writeJSONError(w, http.StatusBadRequest, "codigo_municipio inválido: informe somente números")
		return
	}

	resp, err := h.focus.GetMunicipioItemListaServico(r.Context(), codigoMunicipio, codigoItem)
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, err.Error())
		return
	}
	defer resp.Body.Close()

	proxyResponse(w, resp)
}

// ListCodigosTributarios godoc
// @Summary      Lista códigos tributários municipais por município
// @Description  Proxy para Focus: GET /v2/municipios/{codigo_municipio}/codigos_tributarios_municipio. Suporta filtros via querystring (codigo, descricao, offset).
// @Tags         Municípios - Códigos Tributários
// @Produce      json
// @Param        codigo_municipio  path      string  true   "Código do município (IBGE)"
// @Param        codigo            query     string  false  "Trecho do código (ex: 14.)"
// @Param        descricao         query     string  false  "Trecho da descrição"
// @Param        offset            query     int     false  "Paginação (offset)"
// @Success      200               {array}   RawPayload
// @Failure      400               {object}  RawPayload
// @Failure      401               {object}  RawPayload
// @Failure      404               {object}  RawPayload
// @Failure      429               {object}  RawPayload
// @Failure      500               {object}  RawPayload
// @Router       /v2/municipios/{codigo_municipio}/codigos_tributarios_municipio [get]
func (h *MunicipiosHandler) ListCodigosTributarios(w http.ResponseWriter, r *http.Request) {
	codigoMunicipio := chi.URLParam(r, "codigo_municipio")
	if codigoMunicipio == "" {
		writeJSONError(w, http.StatusBadRequest, "codigo_municipio é obrigatório")
		return
	}
	if !municipioCodigoDigits.MatchString(codigoMunicipio) {
		writeJSONError(w, http.StatusBadRequest, "codigo_municipio inválido: informe somente números")
		return
	}

	resp, err := h.focus.ListMunicipioCodigosTributarios(r.Context(), codigoMunicipio, r.URL.RawQuery)
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, err.Error())
		return
	}
	defer resp.Body.Close()

	proxyResponse(w, resp)
}

// GetCodigoTributario godoc
// @Summary      Busca código tributário municipal por código (município)
// @Description  Proxy para Focus: GET /v2/municipios/{codigo_municipio}/codigos_tributarios_municipio/{codigo}.
// @Tags         Municípios - Códigos Tributários
// @Produce      json
// @Param        codigo_municipio  path      string  true  "Código do município (IBGE)"
// @Param        codigo            path      string  true  "Código tributário municipal"
// @Success      200               {object}  RawPayload
// @Failure      400               {object}  RawPayload
// @Failure      401               {object}  RawPayload
// @Failure      404               {object}  RawPayload
// @Failure      429               {object}  RawPayload
// @Failure      500               {object}  RawPayload
// @Router       /v2/municipios/{codigo_municipio}/codigos_tributarios_municipio/{codigo} [get]
func (h *MunicipiosHandler) GetCodigoTributario(w http.ResponseWriter, r *http.Request) {
	codigoMunicipio := chi.URLParam(r, "codigo_municipio")
	codigo := chi.URLParam(r, "codigo")
	if codigoMunicipio == "" || codigo == "" {
		writeJSONError(w, http.StatusBadRequest, "codigo_municipio e codigo são obrigatórios")
		return
	}
	if !municipioCodigoDigits.MatchString(codigoMunicipio) {
		writeJSONError(w, http.StatusBadRequest, "codigo_municipio inválido: informe somente números")
		return
	}

	resp, err := h.focus.GetMunicipioCodigoTributario(r.Context(), codigoMunicipio, codigo)
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, err.Error())
		return
	}
	defer resp.Body.Close()

	proxyResponse(w, resp)
}

// keep model import used for swag
var _ = model.FocusMunicipioResponse{}


