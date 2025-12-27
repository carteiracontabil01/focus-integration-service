package model

// FocusMunicipioResponse represents a municipality record returned by FocusNFe.
// It is mainly used for Swagger documentation.
type FocusMunicipioResponse struct {
	CodigoMunicipio string `json:"codigo_municipio"`
	NomeMunicipio   string `json:"nome_municipio"`
	SiglaUF         string `json:"sigla_uf"`
	NomeUF          string `json:"nome_uf"`

	NfseHabilitada bool `json:"nfse_habilitada"`
	StatusNfse     string `json:"status_nfse"`

	RequerCertificadoNfse           *bool   `json:"requer_certificado_nfse"`
	PossuiAmbienteHomologacaoNfse  *bool   `json:"possui_ambiente_homologacao_nfse"`
	PossuiCancelamentoNfse         *bool   `json:"possui_cancelamento_nfse"`
	ProvedorNfse                   *string `json:"provedor_nfse"`
	EnderecoObrigatorioNfse        *bool   `json:"endereco_obrigatorio_nfse"`
	CpfCnpjObrigatorioNfse         *bool   `json:"cpf_cnpj_obrigatorio_nfse"`
	CodigoCnaeObrigatorioNfse      *bool   `json:"codigo_cnae_obrigatorio_nfse"`
	ItemListaServicoObrigatorioNfse *bool  `json:"item_lista_servico_obrigatorio_nfse"`
	CodigoTributarioMunicipioObrigatorioNfse *bool `json:"codigo_tributario_municipio_obrigatorio_nfse"`

	DataPrevisaoReimplementacaoNfse *string `json:"data_previsao_reimplementacao_nfse"`
	UltimaEmissaoNfse               *string `json:"ultima_emissao_nfse"`
}


