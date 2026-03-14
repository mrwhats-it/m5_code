package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

func main() {
	var prompt string

	  err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }


	tools := []openai.ChatCompletionToolUnionParam{
		{
			OfFunction: &openai.ChatCompletionFunctionToolParam{
				Function: openai.FunctionDefinitionParam{
					Name:        "Read",
					Description: openai.String("Read and return the contents of a file"),
					Parameters: openai.FunctionParameters{
						"type": "object",
						"properties": map[string]interface{}{
							"file_path": map[string]string{
								"type": "string",
								"description":"The path to the file to read",
							},
						},
						"required": []string{"file_path"},
					},
				},
			},
		},
	}


	flag.StringVar(&prompt, "p", "", "Prompt to send to LLM")
	flag.Parse()

	if prompt == "" {
		panic("Prompt must not be empty")
	}

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	baseUrl := os.Getenv("OPENROUTER_BASE_URL")
	if baseUrl == "" {
		baseUrl = "https://openrouter.ai/api/v1"
	}

	if apiKey == "" {
		panic("Env variable OPENROUTER_API_KEY not found")
	}

	client := openai.NewClient(option.WithAPIKey(apiKey), option.WithBaseURL(baseUrl))
	resp, err := client.Chat.Completions.New(context.Background(),
		openai.ChatCompletionNewParams{
			Model: "meta-llama/llama-3.3-70b-instruct:free",
			Messages: []openai.ChatCompletionMessageParamUnion{
				{
					OfUser: &openai.ChatCompletionUserMessageParam{
						Content: openai.ChatCompletionUserMessageParamContentUnion{
							OfString: openai.String(prompt),
						},
					},
				},
			
			},
			Tools: tools,
		},
	)

	raw:=resp.RawJSON()
	var data interface{}
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		panic(err)
	}

	prettyJSON,err:=json.MarshalIndent(data,"","\t")
	fmt.Println(string(prettyJSON))

	if resp.Choices[0].FinishReason=="tool_calls"{
		func_list:=resp.Choices[0].Message.ToolCalls[0].Function
		if func_list.Name=="Read"{
			var args map[string]string
			json.Unmarshal([]byte(func_list.Arguments),&args)
			err:=Read(args["file_path"])
			if err!=nil{
				log.Fatal(err)
			}
		}
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if len(resp.Choices) == 0 {
		panic("No choices in response")
	}

	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Fprintln(os.Stderr, "Logs from your program will appear here!")



	// TODO: Uncomment the line below to pass the first stage
	fmt.Print(resp.Choices[0].Message.Content)
}

func Read(file_path string) error{
	content,err:=os.ReadFile(file_path)
	if err!=nil{
		return err
	}
	fmt.Println(string(content))
	return nil
}
