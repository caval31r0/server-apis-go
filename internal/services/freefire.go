package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/victtorkaiser/server-apis/internal/dto"
)

type FreeFireService struct {
	apis []FreeFireAPI
}

type FreeFireAPI struct {
	Name string
	URL  string
	Type string // "garena" ou "region"
}

func NewFreeFireService() *FreeFireService {
	return &FreeFireService{
		apis: []FreeFireAPI{
			{Name: "Garena Fast API", URL: "https://garena-production.up.railway.app/api/player-fast/%s", Type: "garena"},
			{Name: "XZA Region API", URL: "https://xza-get-region.vercel.app/region?uid=%s", Type: "region"},
			{Name: "Get Region API", URL: "https://get-region-xza.vercel.app/region?uid=%s", Type: "region"},
		},
	}
}

func (s *FreeFireService) GetPlayer(ctx context.Context, playerID string) (*dto.FreeFireResponse, error) {
	log.Printf("🎮 [FreeFire] Buscando player: %s", playerID)

	// Tenta cada API em ordem até conseguir uma resposta
	for _, api := range s.apis {
		player, err := s.callAPI(ctx, api, playerID)
		if err != nil {
			log.Printf("⚠️ [FreeFire] %s falhou: %v", api.Name, err)
			continue // Tenta a próxima API
		}

		log.Printf("✅ [FreeFire] Dados obtidos de: %s", api.Name)
		return player, nil
	}

	return nil, fmt.Errorf("todas as APIs FreeFire estão indisponíveis")
}

func (s *FreeFireService) callAPI(ctx context.Context, api FreeFireAPI, playerID string) (*dto.FreeFireResponse, error) {
	url := fmt.Sprintf(api.URL, playerID)
	log.Printf("📤 [FreeFire] Tentando: %s", api.Name)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Timeout: 5 * time.Second, // Timeout curto para failover rápido
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	log.Printf("📡 [FreeFire] %s response: %s", api.Name, string(body))

	// Parse resposta baseado no tipo de API
	if api.Type == "garena" {
		return s.parseGarenaResponse(body, api.Name)
	}
	return s.parseRegionResponse(body, api.Name)
}

func (s *FreeFireService) parseGarenaResponse(body []byte, source string) (*dto.FreeFireResponse, error) {
	var garenaResp dto.GarenaFastAPIResponse
	if err := json.Unmarshal(body, &garenaResp); err != nil {
		return nil, err
	}

	if !garenaResp.Success {
		return nil, fmt.Errorf("API retornou success=false")
	}

	return &dto.FreeFireResponse{
		Success:   true,
		Nickname:  garenaResp.Data.Nickname,
		PlayerID:  garenaResp.Data.PlayerID,
		UID:       garenaResp.Data.PlayerID,
		Level:     garenaResp.Data.Level,
		AvatarID:  garenaResp.Data.AvatarID,
		AvatarURL: garenaResp.Data.AvatarURL,
		Source:    source,
	}, nil
}

func (s *FreeFireService) parseRegionResponse(body []byte, source string) (*dto.FreeFireResponse, error) {
	var regionResp dto.RegionAPIResponse
	if err := json.Unmarshal(body, &regionResp); err != nil {
		return nil, err
	}

	if regionResp.UID == "" || regionResp.Nickname == "" {
		return nil, fmt.Errorf("resposta inválida")
	}

	return &dto.FreeFireResponse{
		Success:  true,
		Nickname: regionResp.Nickname,
		PlayerID: regionResp.UID,
		UID:      regionResp.UID,
		Region:   regionResp.Region,
		Source:   source,
	}, nil
}
