package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

func main() {
	// Replace with your API key and secret
	apiKey := "AIzaSyCwW--D5rpE8_4Egnt6QP3FmG3JmUBAwzc"

	ctx := context.Background()
	// Access your API key as an environment variable (see "Set up your API key" above)
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// For text-only input, use the gemini-pro model
	model := client.GenerativeModel("gemini-pro")

	chat := model.StartChat()

	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Println("-------X-------X--------")
		fmt.Println("You :")
		message, _ := reader.ReadString('\n')
		fmt.Println(message)

		response := chat.SendMessageStream(context.Background(), genai.Text(message))

		fmt.Println("Gemini :")
		for {
			res, err := response.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				log.Fatal(err)
			}
			printResponse(res)
		}

		fmt.Println("------------")

	}

}

func printResponse(resp *genai.GenerateContentResponse) {
	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				fmt.Println(part)
			}
		}
	}
	fmt.Println("---")
}
