package provider

import (
	"fmt"
	"os"
	"sync"
)

var (
	initOnce sync.Once
	initErr  error
)

// InitProviders initializes all providers from environment variables
func InitProviders() error {
	initOnce.Do(func() {
		// OpenAI
		if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
			if err := RegisterProvider("openai", NewOpenAI(apiKey)); err != nil {
				initErr = fmt.Errorf("failed to register openai: %w", err)
				return
			}
		}

		// Anthropic
		if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
			if err := RegisterProvider("anthropic", NewAnthropic(apiKey)); err != nil {
				initErr = fmt.Errorf("failed to register anthropic: %w", err)
				return
			}
		}

		// Groq
		if apiKey := os.Getenv("GROQ_API_KEY"); apiKey != "" {
			if err := RegisterProvider("groq", NewGroq(apiKey)); err != nil {
				initErr = fmt.Errorf("failed to register groq: %w", err)
				return
			}
		}

		// Mistral
		if apiKey := os.Getenv("MISTRAL_API_KEY"); apiKey != "" {
			if err := RegisterProvider("mistral", NewMistral(apiKey)); err != nil {
				initErr = fmt.Errorf("failed to register mistral: %w", err)
				return
			}
		}

		// xAI
		if apiKey := os.Getenv("XAI_API_KEY"); apiKey != "" {
			if err := RegisterProvider("xai", NewXAI(apiKey)); err != nil {
				initErr = fmt.Errorf("failed to register xai: %w", err)
				return
			}
		}

		// Cohere
		if apiKey := os.Getenv("COHERE_API_KEY"); apiKey != "" {
			if err := RegisterProvider("cohere", NewCohere(apiKey)); err != nil {
				initErr = fmt.Errorf("failed to register cohere: %w", err)
				return
			}
		}

		// Gemini
		if apiKey := os.Getenv("GEMINI_API_KEY"); apiKey != "" {
			if err := RegisterProvider("gemini", NewGemini(apiKey)); err != nil {
				initErr = fmt.Errorf("failed to register gemini: %w", err)
				return
			}
		}

		// OpenRouter
		if apiKey := os.Getenv("OPENROUTER_API_KEY"); apiKey != "" {
			if err := RegisterProvider("openrouter", NewOpenRouter(apiKey)); err != nil {
				initErr = fmt.Errorf("failed to register openrouter: %w", err)
				return
			}
		}

		// Together
		if apiKey := os.Getenv("TOGETHER_API_KEY"); apiKey != "" {
			if err := RegisterProvider("together", NewTogether(apiKey)); err != nil {
				initErr = fmt.Errorf("failed to register together: %w", err)
				return
			}
		}

		// DeepSeek
		if apiKey := os.Getenv("DEEPSEEK_API_KEY"); apiKey != "" {
			if err := RegisterProvider("deepseek", NewDeepSeek(apiKey)); err != nil {
				initErr = fmt.Errorf("failed to register deepseek: %w", err)
				return
			}
		}

		// Moonshot (中国)
		if apiKey := os.Getenv("MOONSHOT_API_KEY"); apiKey != "" {
			if err := RegisterProvider("moonshot", NewMoonshot(apiKey)); err != nil {
				initErr = fmt.Errorf("failed to register moonshot: %w", err)
				return
			}
		}

		// Aliyun Qwen (中国)
		if apiKey := os.Getenv("ALIYUN_API_KEY"); apiKey != "" {
			if err := RegisterProvider("qwen", NewAliyunQwen(apiKey)); err != nil {
				initErr = fmt.Errorf("failed to register qwen: %w", err)
				return
			}
		}

		// Ollama (local, no API key needed)
		baseURL := os.Getenv("OLLAMA_BASE_URL")
		if baseURL == "" {
			baseURL = "http://localhost:11434"
		}
		// Always register Ollama with default or custom URL
		if err := RegisterProvider("ollama", NewOllama(baseURL)); err != nil {
			initErr = fmt.Errorf("failed to register ollama: %w", err)
			return
		}
	})

	return initErr
}

// MustInit initializes all providers and panics on error
func MustInit() {
	if err := InitProviders(); err != nil {
		panic(err)
	}
}
