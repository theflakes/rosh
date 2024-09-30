package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
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

func getTerminalWidth() (int, error) {
	cmd := exec.Command("tput", "cols")
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	width, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		return 0, err
	}
	return width, nil
}

func parseFlags() (string, string, string, string, error) {
	ip := flag.String("ip", "", "IP address of the server")
	i := flag.String("i", "", "IP address of the server (short)")
	port := flag.String("port", "", "Port of the server")
	p := flag.String("p", "", "Port of the server (short)")
	file := flag.String("file", "", "File to save queries and responses")
	f := flag.String("f", "", "File to save queries and responses (short)")
	model := flag.String("model", "llama3.2", "Model to use")
	m := flag.String("m", "llama3.2", "Model to use (short)")
	flag.Parse()

	if (*ip == "" && *i == "") || (*port == "" && *p == "") {
		return "", "", "", "", fmt.Errorf("usage: rosh --ip <IP> --port <Port> [--file <File>] [--model <Model>]")
	}

	usedIP := *ip
	if usedIP == "" {
		usedIP = *i
	}

	usedPort := *port
	if usedPort == "" {
		usedPort = *p
	}

	usedFile := *file
	if usedFile == "" {
		usedFile = *f
	}

	usedModel := *model
	if usedModel == "" {
		usedModel = *m
	}

	return usedIP, usedPort, usedFile, usedModel, nil
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
		Stream: false,
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

	var response Response
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &response, nil
}

func printResponse(response *Response, line string) error {
	fmt.Println(line)
	fmt.Printf("[*] Model: %s\n", response.Model)
	fmt.Printf("[*] Created At: %s\n", response.CreatedAt)
	fmt.Printf("[*] Response: %s\n", response.Response)
	fmt.Printf("[*] Done: %v\n", response.Done)
	fmt.Println(line)
	fmt.Println("")
	return nil
}

func saveToFile(fileWriter *os.File, prompt string, response *Response, line string) error {
	if fileWriter != nil {
		_, err := fileWriter.WriteString(fmt.Sprintf("Enter prompt: %s%s\n[*] Model: %s\n[*] Created At: %s\n[*] Response: %s\n[*] Done: %v\n%s\n\n",
			prompt,
			line,
			response.Model,
			response.CreatedAt,
			response.Response,
			response.Done,
			line))
		if err != nil {
			return fmt.Errorf("failed to write to file: %v", err)
		}
	}
	return nil
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

	width, err := getTerminalWidth()
	if err != nil {
		fmt.Printf("failed to get terminal width: %v\n", err)
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("Enter prompt: ")
		prompt, _ := reader.ReadString('\n')

		response, err := sendRequest(address, strings.TrimSpace(prompt), model)
		if err != nil {
			fmt.Println(err)
			break
		}

		line := strings.Repeat("*", width)
		err = printResponse(response, line)
		if err != nil {
			fmt.Println(err)
			break
		}

		err = saveToFile(fileWriter, prompt, response, line)
		if err != nil {
			fmt.Println(err)
			break
		}
	}
}
