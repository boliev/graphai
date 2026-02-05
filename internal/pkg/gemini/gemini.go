package gemini

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/boliev/graphai/internal/domain"
	"google.golang.org/genai"
)

const GEMINI_MODEL = "gemini-2.5-flash-image"
const PROMTP_END = ". Please respond by one photo, no text."

type Gemini struct {
	client *genai.Client
}

func NewGemini(ctx context.Context, token string) (*Gemini, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  token,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, err
	}

	return &Gemini{
		client: client,
	}, nil
}

func (g *Gemini) Send(ctx context.Context, description string, files []string) (*domain.AIResponse, error) {
	parts := []*genai.Part{
		genai.NewPartFromText(
			description + PROMTP_END,
		),
	}

	for _, file := range files {
		resp, err := http.Get(file)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			return nil, fmt.Errorf("cannot download photo from telegram %s: %s", file, resp.Status)
		}

		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		mime := g.guessMimeByPath(file)

		parts = append(parts, genai.NewPartFromBytes(b, mime))
	}

	contents := []*genai.Content{
		genai.NewContentFromParts(parts, genai.RoleUser),
	}

	resp, err := g.client.Models.GenerateContent(ctx, GEMINI_MODEL, contents, nil)
	if err != nil {
		return nil, err
	}

	return g.getBytes(resp)
}

func (g *Gemini) getBytes(resp *genai.GenerateContentResponse) (*domain.AIResponse, error) {
	for _, cand := range resp.Candidates {
		if cand.Content == nil {
			continue
		}
		for _, part := range cand.Content.Parts {
			if part.InlineData == nil || len(part.InlineData.Data) == 0 {
				continue
			}
			mime := part.InlineData.MIMEType
			if !strings.HasPrefix(mime, "image/") {
				continue
			}

			return &domain.AIResponse{
				Photo: part.InlineData.Data,
				Mime:  mime,
				Ext:   g.guessExt(mime),
			}, nil
		}
	}

	return nil, fmt.Errorf("no image found in gemini response")
}

func (g *Gemini) guessMimeByPath(p string) string {
	p = strings.ToLower(p)
	switch {
	case strings.HasSuffix(p, ".png"):
		return "image/png"
	case strings.HasSuffix(p, ".webp"):
		return "image/webp"
	default:
		return "image/jpeg"
	}
}

func (g *Gemini) guessExt(mime string) string {
	ext := "jpg"
	if strings.HasSuffix(mime, "png") {
		ext = "png"
	} else if strings.HasSuffix(mime, "webp") {
		ext = "webp"
	}

	return ext
}
