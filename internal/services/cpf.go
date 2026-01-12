package services

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/victtorkaiser/server-apis/internal/config"
	"github.com/victtorkaiser/server-apis/internal/dto"
)

type CPFService struct {
	cfg *config.Config
}

func NewCPFService(cfg *config.Config) *CPFService {
	return &CPFService{
		cfg: cfg,
	}
}

func (s *CPFService) ConsultarCPF(cpf string) (*dto.CPFQueryResponse, error) {
	if s.cfg.CPFAPIUrl == "" || s.cfg.CPFAPIToken == "" {
		return nil, fmt.Errorf("API de consulta de CPF não configurada")
	}

	// Remove formatação do CPF (apenas números)
	cpf = cleanCPF(cpf)

	if len(cpf) != 11 {
		return nil, fmt.Errorf("CPF inválido: deve conter 11 dígitos")
	}

	// Monta URL
	url := fmt.Sprintf("%s?token_api=%s&cpf=%s", s.cfg.CPFAPIUrl, s.cfg.CPFAPIToken, cpf)

	log.Printf("📤 [CPF] Consultando CPF: %s", maskCPF(cpf))

	// Faz requisição
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		log.Printf("❌ [CPF] Erro ao consultar API: %v", err)
		return nil, fmt.Errorf("erro ao consultar API: %w", err)
	}
	defer resp.Body.Close()

	// Lê resposta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("❌ [CPF] Erro ao ler resposta: %v", err)
		return nil, fmt.Errorf("erro ao ler resposta: %w", err)
	}

	log.Printf("📡 [CPF] Response HTTP %d: %s", resp.StatusCode, string(body))

	// Decodifica JSON
	var result dto.CPFQueryResponse
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("❌ [CPF] Erro ao decodificar JSON: %v", err)
		return nil, fmt.Errorf("erro ao decodificar resposta: %w", err)
	}

	// Valida resposta
	if result.Status != 200 {
		return nil, fmt.Errorf("erro na consulta: status %d", result.Status)
	}

	if len(result.Dados) == 0 {
		return nil, fmt.Errorf("CPF não encontrado")
	}

	log.Printf("✅ [CPF] Consulta realizada com sucesso: %s", maskCPF(cpf))
	return &result, nil
}

// Remove caracteres não numéricos do CPF
func cleanCPF(cpf string) string {
	cleaned := ""
	for _, char := range cpf {
		if char >= '0' && char <= '9' {
			cleaned += string(char)
		}
	}
	return cleaned
}

// Mascara CPF para logs (XXX.XXX.XXX-XX)
func maskCPF(cpf string) string {
	if len(cpf) != 11 {
		return cpf
	}
	return fmt.Sprintf("%s.%s.%s-%s", cpf[0:3], cpf[3:6], cpf[6:9], cpf[9:11])
}
