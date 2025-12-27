package focus

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

func NewClient(baseURL, token string) *Client {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")

	return &Client{
		baseURL: baseURL,
		token:   strings.TrimSpace(token),
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) CreateEmpresa(ctx context.Context, body []byte) (*http.Response, error) {
	return c.do(ctx, http.MethodPost, "/v2/empresas", "", body)
}

func (c *Client) ListEmpresas(ctx context.Context, rawQuery string) (*http.Response, error) {
	return c.do(ctx, http.MethodGet, "/v2/empresas", rawQuery, nil)
}

func (c *Client) GetEmpresa(ctx context.Context, id string) (*http.Response, error) {
	return c.do(ctx, http.MethodGet, "/v2/empresas/"+id, "", nil)
}

func (c *Client) UpdateEmpresa(ctx context.Context, id string, body []byte) (*http.Response, error) {
	return c.do(ctx, http.MethodPut, "/v2/empresas/"+id, "", body)
}

func (c *Client) DeleteEmpresa(ctx context.Context, id string) (*http.Response, error) {
	return c.do(ctx, http.MethodDelete, "/v2/empresas/"+id, "", nil)
}

func (c *Client) GetCNPJ(ctx context.Context, cnpj14 string) (*http.Response, error) {
	return c.do(ctx, http.MethodGet, "/v2/cnpjs/"+cnpj14, "", nil)
}

func (c *Client) ListMunicipios(ctx context.Context, rawQuery string) (*http.Response, error) {
	return c.do(ctx, http.MethodGet, "/v2/municipios", rawQuery, nil)
}

func (c *Client) GetMunicipio(ctx context.Context, codigoMunicipio string) (*http.Response, error) {
	return c.do(ctx, http.MethodGet, "/v2/municipios/"+codigoMunicipio, "", nil)
}

func (c *Client) ListMunicipioItensListaServico(ctx context.Context, codigoMunicipio, rawQuery string) (*http.Response, error) {
	return c.do(ctx, http.MethodGet, "/v2/municipios/"+codigoMunicipio+"/itens_lista_servico", rawQuery, nil)
}

func (c *Client) GetMunicipioItemListaServico(ctx context.Context, codigoMunicipio, codigoItem string) (*http.Response, error) {
	return c.do(ctx, http.MethodGet, "/v2/municipios/"+codigoMunicipio+"/itens_lista_servico/"+codigoItem, "", nil)
}

func (c *Client) ListMunicipioCodigosTributarios(ctx context.Context, codigoMunicipio, rawQuery string) (*http.Response, error) {
	return c.do(ctx, http.MethodGet, "/v2/municipios/"+codigoMunicipio+"/codigos_tributarios_municipio", rawQuery, nil)
}

func (c *Client) GetMunicipioCodigoTributario(ctx context.Context, codigoMunicipio, codigoTributario string) (*http.Response, error) {
	return c.do(ctx, http.MethodGet, "/v2/municipios/"+codigoMunicipio+"/codigos_tributarios_municipio/"+codigoTributario, "", nil)
}

func (c *Client) do(ctx context.Context, method, path, rawQuery string, body []byte) (*http.Response, error) {
	if c.baseURL == "" {
		return nil, fmt.Errorf("FOCUS_URL não definido")
	}
	if c.token == "" {
		return nil, fmt.Errorf("FOCUS_API_TOKEN não definido")
	}

	url := c.baseURL + path
	if rawQuery != "" {
		url = url + "?" + strings.TrimPrefix(rawQuery, "?")
	}

	var r io.Reader
	if body != nil {
		r = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, r)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição: %w", err)
	}

	req.SetBasicAuth(c.token, "")
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao enviar requisição para Focus: %w", err)
	}

	return resp, nil
}


