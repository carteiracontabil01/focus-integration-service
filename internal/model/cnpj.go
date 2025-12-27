package model

// FocusCnpjResponse represents the response from FocusNFe CNPJ lookup endpoint.
// This struct is used mainly for Swagger documentation.
type FocusCnpjResponse struct {
	RazaoSocial          string `json:"razao_social"`
	CNPJ                 string `json:"cnpj"`
	SituacaoCadastral    string `json:"situacao_cadastral"`
	CnaePrincipal        string `json:"cnae_principal"`
	OptanteSimplesNac    bool   `json:"optante_simples_nacional"`
	OptanteMEI           bool   `json:"optante_mei"`
	Endereco             FocusCnpjEndereco `json:"endereco"`
}

type FocusCnpjEndereco struct {
	CodigoMunicipio string `json:"codigo_municipio"`
	CodigoSiafi     string `json:"codigo_siafi"`
	CodigoIbge      string `json:"codigo_ibge"`
	NomeMunicipio   string `json:"nome_municipio"`
	Logradouro      string `json:"logradouro"`
	Complemento     string `json:"complemento"`
	Numero          string `json:"numero"`
	Bairro          string `json:"bairro"`
	Cep             string `json:"cep"`
	Uf              string `json:"uf"`
}


