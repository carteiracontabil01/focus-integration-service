package model

// FocusEmpresaResponse representa uma empresa retornada pela Focus NFe (v2).
// Observação: a Focus retorna muitos campos; aqui documentamos os principais
// (os demais podem aparecer e serão repassados pelo proxy normalmente).
type FocusEmpresaResponse struct {
	// Campos principais + campos retornados pela Focus (exemplo real).
	ID               int    `json:"id" example:"170571"`
	ClientAppID      *int   `json:"client_app_id,omitempty" example:"357307"`
	Nome             string `json:"nome" example:"INFINITY CODE SOLUTIONS LTDA"`
	NomeFantasia     string `json:"nome_fantasia" example:"INFINITY CODE"`
	CNPJ             string `json:"cnpj" example:"61453926000127"`
	CPF              *string `json:"cpf,omitempty" example:"12345678901"`
	Email            string `json:"email,omitempty" example:"andre.vi.riibeiro@gmail.com"`
	Telefone         string `json:"telefone,omitempty" example:"1993675739"`
	InscricaoEstadual *string `json:"inscricao_estadual,omitempty"`
	InscricaoMunicipal string `json:"inscricao_municipal,omitempty" example:"10721193"`

	Logradouro  string  `json:"logradouro,omitempty" example:"COACIARA"`
	Numero      string  `json:"numero,omitempty" example:"1101"`
	Complemento string  `json:"complemento,omitempty" example:"BLOCO 24 APT 04 COND RES OURO VERDE"`
	Bairro      string  `json:"bairro,omitempty" example:"PARQUE DOM PEDRO II"`
	CEP         string  `json:"cep,omitempty" example:"13056430"`
	Municipio   string  `json:"municipio,omitempty" example:"CAMPINAS"`
	UF          string  `json:"uf,omitempty" example:"SP"`
	Pais        string  `json:"pais,omitempty" example:"Brasil"`
	CodigoMunicipio *string `json:"codigo_municipio,omitempty" example:"3509502"`
	CodigoPais      *string `json:"codigo_pais,omitempty" example:"1058"`
	CodigoUF        *string `json:"codigo_uf,omitempty" example:"35"`

	RegimeTributario string `json:"regime_tributario,omitempty" example:"1"`

	// Flags comuns
	DiscriminaImpostos        *bool `json:"discrimina_impostos,omitempty" example:"true"`
	EnviarEmailDestinatario   *bool `json:"enviar_email_destinatario,omitempty" example:"true"`
	EnviarEmailHomologacao    *bool `json:"enviar_email_homologacao,omitempty" example:"false"`
	HabilitaNFCE              *bool `json:"habilita_nfce,omitempty" example:"false"`
	HabilitaNFE               *bool `json:"habilita_nfe,omitempty" example:"false"`
	HabilitaNFSE              *bool `json:"habilita_nfse,omitempty" example:"false"`
	HabilitaCTE               *bool `json:"habilita_cte,omitempty" example:"false"`
	HabilitaMDFE              *bool `json:"habilita_mdfe,omitempty" example:"false"`
	HabilitaManifestacao      *bool `json:"habilita_manifestacao,omitempty" example:"false"`
	HabilitaCSRTNFe           *bool `json:"habilita_csrt_nfe,omitempty" example:"true"`
	NFeSincrono               *bool `json:"nfe_sincrono,omitempty" example:"false"`
	NFeSincronoHomologacao    *bool `json:"nfe_sincrono_homologacao,omitempty" example:"false"`

	// Certificado
	CertificadoValidoAte string `json:"certificado_valido_ate,omitempty" example:"2026-11-12T14:33:00-03:00"`
	CertificadoValidoDe  string `json:"certificado_valido_de,omitempty" example:"2025-11-12T14:33:00-03:00"`
	CertificadoCNPJ      string `json:"certificado_cnpj,omitempty" example:"61453926000127"`
	CertificadoEspecifico *bool `json:"certificado_especifico,omitempty" example:"false"`

	// Tokens
	TokenProducao    string `json:"token_producao,omitempty" example:"mQadVh1LFN7KQNESwDFLsDBq039z273s"`
	TokenHomologacao string `json:"token_homologacao,omitempty" example:"tmyEw3AL8QIEqjObIuqnbGcP7wMW961F"`

	// Outros
	NomeResponsavel *string `json:"nome_responsavel,omitempty"`
	CpfResponsavel  *string `json:"cpf_responsavel,omitempty"`
}


