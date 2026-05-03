package trend

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"gnai/trend-detector/indicator"
)

const systemPrompt = `You are a quantitative market analyst specialising in Indian equity markets (NSE/BSE).
You will receive a snapshot of technical indicators for a single instrument and must classify the current trend.
Always return a single JSON object with keys: "trend", "confidence" (float 0-1), "reasoning" (max 2 sentences).
Do not include markdown, explanation, or any text outside the JSON object.`

// Classifier classifies market trend from an indicator snapshot.
// If apiKey is empty it runs in rule-based mode; otherwise it calls Claude.
type Classifier struct {
	client *anthropic.Client
	model  string
}

// NewClassifier creates a Classifier. Pass an empty apiKey to use rule-based mode.
func NewClassifier(apiKey string) *Classifier {
	if apiKey == "" {
		return &Classifier{}
	}
	client := anthropic.NewClient(option.WithAPIKey(apiKey))
	return &Classifier{client: &client, model: "claude-sonnet-4-6"}
}

// Classify returns a Classification for the given snapshot.
func (c *Classifier) Classify(ctx context.Context, snap indicator.Snapshot) Classification {
	if c.client == nil {
		return c.ruleBasedClassify(snap)
	}
	result, err := c.claudeClassify(ctx, snap)
	if err != nil {
		log.Printf("classifier: claude error, falling back to rule-based: %v", err)
		return c.ruleBasedClassify(snap)
	}
	return result
}

// ruleBasedClassify uses weighted indicator votes to determine trend.
func (c *Classifier) ruleBasedClassify(snap indicator.Snapshot) Classification {
	votes := snap.Votes()
	scores := map[TrendType]float64{}
	for _, v := range votes {
		scores[TrendType(v.Trend)] += v.Weight
	}

	best, bestScore := Unknown, 0.0
	total := 0.0
	for t, s := range scores {
		total += s
		if s > bestScore {
			best = t
			bestScore = s
		}
	}

	confidence := 0.0
	if total > 0 {
		confidence = bestScore / total
	}

	return Classification{
		Trend:      best,
		Confidence: confidence,
		Reasoning:  fmt.Sprintf("Rule-based weighted vote: %s leads with %.0f%% of weight.", best, confidence*100),
		Timestamp:  time.Now(),
		Mode:       "rule-based",
	}
}

// claudeClassify sends the indicator snapshot to Claude and parses the response.
func (c *Classifier) claudeClassify(ctx context.Context, snap indicator.Snapshot) (Classification, error) {
	msg, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(c.model),
		MaxTokens: 256,
		System: []anthropic.TextBlockParam{
			{Text: systemPrompt},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(snap.Text())),
		},
	})
	if err != nil {
		return Classification{}, fmt.Errorf("claude API: %w", err)
	}

	raw := ""
	for _, block := range msg.Content {
		if block.Type == "text" {
			raw = strings.TrimSpace(block.Text)
			break
		}
	}

	var parsed struct {
		Trend      string  `json:"trend"`
		Confidence float64 `json:"confidence"`
		Reasoning  string  `json:"reasoning"`
	}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return Classification{}, fmt.Errorf("parse claude response %q: %w", raw, err)
	}

	return Classification{
		Trend:      TrendType(strings.ToUpper(parsed.Trend)),
		Confidence: parsed.Confidence,
		Reasoning:  parsed.Reasoning,
		Timestamp:  time.Now(),
		Mode:       "claude-enhanced",
	}, nil
}
