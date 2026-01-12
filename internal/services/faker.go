package services

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type FakerService struct{}

type FakePerson struct {
	// Campos da 4devs
	Nome         string `json:"nome"`
	Sexo         string `json:"sexo"`
	Email        string `json:"email"`
	DataNasc     string `json:"data_nasc"`
	CPF          string `json:"cpf"`
	RG           string `json:"rg"`
	Telefone     string `json:"telefone"`
	Celular      string `json:"celular"`
	Altura       string `json:"altura"`
	Peso         string `json:"peso"`
	TipoSanguineo string `json:"tipo_sanguineo"`
	Mae          string `json:"mae"`
	Pai          string `json:"pai"`
	Cor          string `json:"cor"`
	Numero       int    `json:"numero"`
	CEP          string `json:"cep"`
	Endereco     string `json:"endereco"`
	Bairro       string `json:"bairro"`
	Cidade       string `json:"cidade"`
	Estado       string `json:"estado"`
}

// Resposta da API 5devs (formato diferente)
type FakePerson5devs struct {
	Nome           string `json:"nome"`
	Sexo           string `json:"sexo"`
	Email          string `json:"email"`
	DataNascimento string `json:"dataNascimento"`
	Signo          string `json:"signo"`
	CPF            string `json:"cpf"`
	RG             string `json:"rg"`
	NomePai        string `json:"nomePai"`
	NomeMae        string `json:"nomeMae"`
	Telefone       string `json:"telefone"`
}

func NewFakerService() *FakerService {
	return &FakerService{}
}

func (s *FakerService) GerarPessoa() (*FakePerson, error) {
	// Usa apenas 5devs
	pessoa, err := s.gerar5devs()
	if err != nil {
		return nil, fmt.Errorf("erro ao gerar pessoa via 5devs: %w", err)
	}

	return pessoa, nil
}

func (s *FakerService) gerar5devs() (*FakePerson, error) {
	url := "https://www.5devs.com.br/api/pessoa"

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	log.Println("📤 [5devs] Gerando dados fake de pessoa...")

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("erro ao fazer requisição: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler resposta: %w", err)
	}

	var pessoa5devs FakePerson5devs
	if err := json.Unmarshal(body, &pessoa5devs); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta: %w", err)
	}

	// Converte para o formato padrão FakePerson
	pessoa := &FakePerson{
		Nome:     pessoa5devs.Nome,
		Sexo:     pessoa5devs.Sexo,
		Email:    pessoa5devs.Email,
		CPF:      pessoa5devs.CPF,
		RG:       pessoa5devs.RG,
		Telefone: pessoa5devs.Telefone,
		Celular:  pessoa5devs.Telefone, // 5devs só retorna um telefone
		Mae:      pessoa5devs.NomeMae,
		Pai:      pessoa5devs.NomePai,
		DataNasc: pessoa5devs.DataNascimento,
	}

	log.Printf("✅ [5devs] Pessoa gerada: %s - CPF: %s", pessoa.Nome, pessoa.CPF)

	return pessoa, nil
}

// Limpa formatação de telefone (retorna apenas números)
func (s *FakerService) CleanPhone(phone string) string {
	cleaned := ""
	for _, char := range phone {
		if char >= '0' && char <= '9' {
			cleaned += string(char)
		}
	}
	return cleaned
}

// Limpa formatação de CPF (retorna apenas números)
func (s *FakerService) CleanCPF(cpf string) string {
	cleaned := ""
	for _, char := range cpf {
		if char >= '0' && char <= '9' {
			cleaned += string(char)
		}
	}
	return cleaned
}
