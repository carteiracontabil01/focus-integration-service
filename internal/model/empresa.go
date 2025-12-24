package model

// FocusEmpresaCreateRequest representa o payload esperado para criação de empresa na Focus NFe (v2).
//
// Observações:
// - Campos marcados como required aqui são os obrigatórios no fluxo da Carteira Contábil.
// - `regime_tributario` deve ser numérico conforme a tabela da Focus:
//   1 = Simples Nacional
//   2 = Simples Nacional - Excesso de sublimite de receita bruta
//   3 = Regime Normal
//   4 = MEI
type FocusEmpresaCreateRequest struct {
	Nome                string `json:"nome" binding:"required" example:"Nome da empresa Ltda"`
	NomeFantasia        string `json:"nome_fantasia" binding:"required" example:"Nome Fantasia"`
	Bairro              string `json:"bairro" binding:"required" example:"Vila Isabel"`
	CEP                 int    `json:"cep" binding:"required" example:"80210000"`
	CNPJ                string `json:"cnpj" binding:"required" example:"10964044000164"`
	Complemento         string `json:"complemento" binding:"required" example:"Loja 1"`
	Email               string `json:"email" binding:"required" example:"test@example.com"`
	// Must be string to preserve leading zeros (JSON number cannot represent "00", "00123", etc.)
	InscricaoMunicipal  string `json:"inscricao_municipal" binding:"required" example:"0046532"`
	Logradouro          string `json:"logradouro" binding:"required" example:"Rua João da Silva"`
	Numero              int    `json:"numero" binding:"required" example:"153"`
	RegimeTributario    int    `json:"regime_tributario" binding:"required" example:"1"`
	Telefone            string `json:"telefone,omitempty" example:"4130333333"`
	Municipio           string `json:"municipio" binding:"required" example:"Curitiba"`
	UF                  string `json:"uf" binding:"required" example:"PR"`
	ArquivoCertBase64   string `json:"arquivo_certificado_base64" binding:"required" example:"MIIj4gIBAzCCI54GCSqGSIb3DQEHAaCC...ASD=="`
	SenhaCertificado    string `json:"senha_certificado" binding:"required" example:"123456"`
	NomeResponsavel string `json:"nome_responsavel,omitempty" example:"Fulano de Tal"`
	CPFResponsavel  string `json:"cpf_responsavel,omitempty" example:"12345678901"`
	
	// Campo interno: ID da relação company_certificates_access para logging de erros.
	// NÃO é enviado para a Focus.
	DatabaseLocalCertificateID *string `json:"database_local_certificate_id,omitempty" swaggerignore:"true"`
}

// FocusEmpresaUpdateRequest representa o payload de atualização de empresa na Focus.
// Todos os campos são opcionais (omitir = não alterar).
type FocusEmpresaUpdateRequest struct {
	Nome               *string `json:"nome,omitempty" example:"Nome da empresa Ltda"`
	NomeFantasia       *string `json:"nome_fantasia,omitempty" example:"Nome Fantasia"`
	Bairro             *string `json:"bairro,omitempty" example:"Vila Isabel"`
	CEP                *int    `json:"cep,omitempty" example:"80210000"`
	CNPJ               *string `json:"cnpj,omitempty" example:"10964044000164"`
	Complemento        *string `json:"complemento,omitempty" example:"Loja 1"`
	Email              *string `json:"email,omitempty" example:"test@example.com"`
	// Must be string to preserve leading zeros
	InscricaoMunicipal *string `json:"inscricao_municipal,omitempty" example:"0046532"`
	Logradouro         *string `json:"logradouro,omitempty" example:"Rua João da Silva"`
	Numero             *int    `json:"numero,omitempty" example:"153"`
	RegimeTributario   *int    `json:"regime_tributario,omitempty" example:"1"`
	Telefone           *string `json:"telefone,omitempty" example:"4130333333"`
	Municipio          *string `json:"municipio,omitempty" example:"Curitiba"`
	UF                 *string `json:"uf,omitempty" example:"PR"`
	ArquivoCertBase64 *string `json:"arquivo_certificado_base64,omitempty" example:"MIIj4gIBAzCCI54GCSqGSIb3DQEHAaCC...ASD=="`
	SenhaCertificado  *string `json:"senha_certificado,omitempty" example:"123456"`
	NomeResponsavel *string `json:"nome_responsavel,omitempty" example:"Fulano de Tal"`
	CPFResponsavel  *string `json:"cpf_responsavel,omitempty" example:"12345678901"`
	HabilitaNFE                  *bool   `json:"habilita_nfe,omitempty" example:"true"`
	ProximoNumeroNFEProducao     *int    `json:"proximo_numero_nfe_producao,omitempty" example:"1"`
	HabilitaNFCE                 *bool   `json:"habilita_nfce,omitempty" example:"true"`
	ProximoNumeroNFCEProducao    *int    `json:"proximo_numero_nfce_producao,omitempty" example:"1"`
	HabilitaNFSE                 *bool   `json:"habilita_nfse,omitempty" example:"true"`
	ProximoNumeroNFSEProducao    *int    `json:"proximo_numero_nfse_producao,omitempty" example:"1"`
	HabilitaNFSENProducao        *bool   `json:"habilita_nfsen_producao,omitempty" example:"true"`
	ProximoNumeroNFSENProducao   *int    `json:"proximo_numero_nfsen_producao,omitempty" example:"1"`
	HabilitaCTE                  *bool   `json:"habilita_cte,omitempty" example:"true"`
	ProximoNumeroCTEProducao     *int    `json:"proximo_numero_cte_producao,omitempty" example:"1"`
	HabilitaMDFE                 *bool   `json:"habilita_mdfe,omitempty" example:"true"`
	ProximoNumeroMDFEProducao    *int    `json:"proximo_numero_mdfe_producao,omitempty" example:"1"`
	LoginResponsavel            *string `json:"login_responsavel,omitempty" example:"usuario@example.com"`
	SenhaResponsavel            *string `json:"senha_responsavel,omitempty" example:"senha123"`
	SenhaResponsavelPreenchida  *bool   `json:"senha_responsavel_preenchida,omitempty" example:"false"`
}