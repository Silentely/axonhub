package llm

type APIFormat string

const (
	APIFormatOpenAIChatCompletion  APIFormat = "openai/chat_completions"
	APIFormatOpenAIResponse        APIFormat = "openai/responses"
	APIFormatOpenAIImageGeneration APIFormat = "openai/image_generation"
	APIFormatGeminiContents        APIFormat = "gemini/contents"
	APIFormatAnthropicMessage      APIFormat = "anthropic/messages"
	APIFormatAiSDKText             APIFormat = "aisdk/text"
	APIFormatAiSDKDataStream       APIFormat = "aisdk/datastream"
)

func (f APIFormat) String() string {
	return string(f)
}

const (
	ToolType                = "function"
	ToolTypeImageGeneration = "image_generation"
	// ToolTypeGoogle indicates a Google/Gemini-specific tool.
	// When this type is set, Tool.Google field contains the specific tool configuration.
	ToolTypeGoogle = "google"
)
