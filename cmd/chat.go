package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Start interactive chat session",
	Long: `Start an interactive chat session with an AI provider.

Examples:
  # Chat with OpenAI GPT-4
  liteclaw chat --provider openai --model gpt-4

  # One-off message
  liteclaw chat --provider anthropic --message "Hello, Claude!"

  # With system prompt
  liteclaw chat --provider openai --system "You are a helpful assistant"`,
	RunE: runChat,
}

var (
	chatProvider    string
	chatModel       string
	chatTemperature float64
	chatSystem      string
	chatMessage     string
	chatStream      bool
	chatMaxTokens   int
	chatTools       bool
)

func init() {
	rootCmd.AddCommand(chatCmd)

	chatCmd.Flags().StringVarP(&chatProvider, "provider", "p", "openai", "LLM provider")
	chatCmd.Flags().StringVarP(&chatModel, "model", "m", "", "Model name")
	chatCmd.Flags().Float64VarP(&chatTemperature, "temperature", "t", 0.7, "Temperature (0.0-2.0)")
	chatCmd.Flags().StringVarP(&chatSystem, "system", "s", "", "System prompt")
	chatCmd.Flags().StringVarP(&chatMessage, "message", "M", "", "One-off message")
	chatCmd.Flags().BoolVar(&chatStream, "stream", false, "Enable streaming")
	chatCmd.Flags().IntVar(&chatMaxTokens, "max-tokens", 2000, "Max tokens in response")
	chatCmd.Flags().BoolVar(&chatTools, "tools", false, "Enable function calling")
}

func runChat(cmd *cobra.Command, args []string) error {
	fmt.Printf("🤖 Starting chat with %s\n", chatProvider)
	if chatModel != "" {
		fmt.Printf("   Model: %s\n", chatModel)
	}
	fmt.Printf("   Temperature: %.1f\n", chatTemperature)
	fmt.Printf("   Stream: %v\n", chatStream)
	fmt.Println()

	// TODO: Implement actual chat logic
	return fmt.Errorf("chat functionality not yet implemented")
}
