package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
)

type Request struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type Response struct {
	Model     string `json:"model"`
	CreatedAt string `json:"created_at"`
	Response  string `json:"response"`
	Done      bool   `json:"done"`
}

func openFile(filePath string) (*os.File, error) {
	if filePath == "" {
		return nil, nil
	}
	fileWriter, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	return fileWriter, nil
}

func sendRequest(address, prompt, model string) (*Response, error) {
	request := Request{
		Model:  model,
		Prompt: prompt,
		Stream: true, // Enable streaming
	}
	requestBytes, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := http.Post(address, "application/json", bytes.NewBuffer(requestBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	return processResponseStream(resp)
}

func processResponseStream(resp *http.Response) (*Response, error) {
	scanner := bufio.NewScanner(resp.Body)
	var response Response
	firstResponseReceived := false

	for scanner.Scan() {
		line := scanner.Text()
		var chunk Response
		err := json.Unmarshal([]byte(line), &chunk)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal response chunk: %v", err)
		}
		response.Response += chunk.Response
		fmt.Print(chunk.Response) // Print each chunk's response as it is received

		if !firstResponseReceived {
			response.Model = chunk.Model
			response.CreatedAt = chunk.CreatedAt
			firstResponseReceived = true
		}

		if chunk.Done {
			response.Done = true
			fmt.Printf("\n\n---\n\n")
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	return &response, nil
}

func saveToFile(fileWriter *os.File, prompt string, response *Response) error {
	if fileWriter != nil {
		_, err := fileWriter.WriteString(fmt.Sprintf(`Enter prompt (end with two empty lines):
%s---
[*] Model: %s
[*] Created At: %s
[*] Response: %s
[*] Done: %v
---
`,
			prompt,
			response.Model,
			response.CreatedAt,
			response.Response,
			response.Done,
		))
		if err != nil {
			return fmt.Errorf("failed to write to file: %v", err)
		}
	}
	return nil
}

func trimSpace(sb *strings.Builder) string {
	result := strings.TrimRightFunc(sb.String(), func(r rune) bool {
		return r == '\n' || r == '\r' || r == ' ' || r == '\t'
	})
	return strings.TrimLeftFunc(result, func(r rune) bool {
		return r == ' ' || r == '\t'
	})
}

func main() {
	ip, port, filePath, model, err := parseFlags()
	if err != nil {
		fmt.Println(err)
		return
	}

	address := fmt.Sprintf("http://%s:%s/api/generate", ip, port)
	fileWriter, err := openFile(filePath)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer fileWriter.Close()

	reader := bufio.NewReader(os.Stdin)

	for {
		var prompt strings.Builder
		fmt.Println("Enter prompt (end with two empty lines):")

		emptyLineCount := 0
		for {
			line, _ := reader.ReadString('\n')
			l := strings.TrimSpace(line)
			if l == "/bye" {
				os.Exit(0)
			}
			if l == "" {
				emptyLineCount++
				if emptyLineCount == 2 {
					break
				}
			} else {
				emptyLineCount = 0
				prompt.WriteString(line)
			}
		}

		response, err := sendRequest(address, trimSpace(&prompt), model)
		if err != nil {
			fmt.Println(err)
			break
		}

		err = saveToFile(fileWriter, prompt.String(), response)
		if err != nil {
			fmt.Println(err)
			break
		}
	}
}

func parseFlags() (string, string, string, string, error) {
	help := flag.Bool("help", false, "Print this help message")
	h := flag.Bool("h", false, "Print this help message (short)")
	ip := flag.String("ip", "127.0.0.1", "IP address of the server")
	i := flag.String("i", "127.0.0.1", "IP address of the server (short)")
	port := flag.String("port", "11434", "Port of the server")
	p := flag.String("p", "11434", "Port of the server (short)")
	file := flag.String("file", "", "File to save queries and responses")
	f := flag.String("f", "", "File to save queries and responses (short)")
	model := flag.String("model", "llama3.2", "Model to use")
	m := flag.String("m", "llama3.2", "Model to use (short)")

	flag.Parse()

	if *help || *h {
		printHelp()
		os.Exit(0)
	}

	usedIP := *ip
	if *ip == "127.0.0.1" {
		usedIP = *i
	}

	usedPort := *port
	if *port == "11434" {
		usedPort = *p
	}

	usedFile := *file
	if *file == "" {
		usedFile = *f
	}

	usedModel := *model
	if *model == "llama3.2" {
		usedModel = *m
	}

	return usedIP, usedPort, usedFile, usedModel, nil
}

func printHelp() {
	flag.Usage = func() {
		fmt.Println(`Connect to Olamma API via console.`)
		println("  Author: Brian Kellogg")
		println("  License: MIT")
		println()
		println("To exit type -> /bye")
		println()
		println("Command line arguments:")
		println("  --ip    : IP address of the server (default: 127.0.0.1)")
		println("   -i     : IP address of the server (default: 127.0.0.1)")
		println("  --port  : Port of the server (default: 11434)")
		println("   -p     : Port of the server (default: 11434)")
		println("  --file  : File to save queries and responses (default: \"\")")
		println("   -f     : File to save queries and responses (default: \"\")")
		println("  --model : Model to use (default: llama3.2)")
		println("   -m     : Model to use (default: llama3.2)")
	}
	flag.Usage()
}
