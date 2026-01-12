package dto

// CPF Query Request
type CPFQueryRequest struct {
	CPF string `form:"cpf" binding:"required"`
}

// CPF Query Response (da API externa)
type CPFQueryResponse struct {
	Dados  []CPFData `json:"dados"`
	Status int       `json:"status"`
}

type CPFData struct {
	CPF            string `json:"CPF"`
	NASC           string `json:"NASC"`
	NOME           string `json:"NOME"`
	NOME_MAE       string `json:"NOME_MAE"`
	NOME_PAI       string `json:"NOME_PAI"`
	ORGAO_EMISSOR  string `json:"ORGAO_EMISSOR"`
	RENDA          string `json:"RENDA"`
	RG             string `json:"RG"`
	SEXO           string `json:"SEXO"`
	SO             string `json:"SO"`
	TITULO_ELEITOR string `json:"TITULO_ELEITOR"`
	UF_EMISSAO     string `json:"UF_EMISSAO"`
}
