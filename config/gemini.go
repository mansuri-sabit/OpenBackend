package config

import (
    "context"
    "fmt"
    "log"
    "os"

    "regexp"
    "github.com/google/generative-ai-go/genai"
    "google.golang.org/api/option"

  "net/http"
  "bytes"
  "encoding/json"
)

var GeminiClient *genai.Client

// Initialize Gemini client (call this once in main or init)
func InitGemini() {
    apiKey := os.Getenv("GEMINI_API_KEY")
    if apiKey == "" {
        log.Fatal("GEMINI_API_KEY not set in environment")
    }

    ctx := context.Background()
    client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
    if err != nil {
        log.Fatal("Failed to initialize Gemini client:", err)
    }

    GeminiClient = client
    log.Println("✅ Gemini client initialized successfully")
}

type OpenAIResponse struct {
    Choices []struct {
        Message struct {
            Content string `json:"content"`
        } `json:"message"`
    } `json:"choices"`
}


// Generates a polished, human-like response
func GenerateResponse(userPrompt string, pdfContext string) (string, error) {
    apiKey := os.Getenv("OPENAI_API_KEY")
    if apiKey == "" {
        return "", fmt.Errorf("OpenAI API key not set")
    }

    url := "https://api.openai.com/v1/chat/completions"
    payload := map[string]interface{}{
        "model": "gpt-3.5-turbo",
        "messages": []map[string]string{
            {"role": "system", "content": "You are a helpful assistant. Respond briefly and clearly."},
            {"role": "user", "content": fmt.Sprintf("Context: %s\n\nQuestion: %s", pdfContext, userPrompt)},
        },
    }

    body, _ := json.Marshal(payload)
    req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
    req.Header.Set("Authorization", "Bearer "+apiKey)
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    var res OpenAIResponse
    if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
        return "", err
    }

    if len(res.Choices) > 0 {
        return res.Choices[0].Message.Content, nil
    }

    return "No response", nil
}

func cleanResponse(raw string) string {
    cleaned := raw

    // Remove robotic or formal phrases
    cleaned = removeFirstMatch(cleaned, `(?i)^based on the .*?(document|pdf)[,:]?\s*`)
    cleaned = removeFirstMatch(cleaned, `(?i)^according to .*?[,:]?\s*`)
    cleaned = removeFirstMatch(cleaned, `(?i)^as per .*?[,:]?\s*`)
    cleaned = removeFirstMatch(cleaned, `(?i)is there anything else.*?\?$`)
    cleaned = removeFirstMatch(cleaned, `(?i)let me know if you need anything else.*?`)
    cleaned = removeFirstMatch(cleaned, `(?i)hope this helps[.!]?`)
    cleaned = removeFirstMatch(cleaned, `(?i)I'm here to assist you.*?`)

    // Optional: trim spaces
    cleaned = regexp.MustCompile(`^\s+|\s+$`).ReplaceAllString(cleaned, "")

    return cleaned
}


// Helper: simple regex match remover
func removeFirstMatch(input string, pattern string) string {
    re := regexp.MustCompile(pattern)
    return re.ReplaceAllString(input, "")
}
