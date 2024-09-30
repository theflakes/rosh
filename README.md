# rosh
Remote shell for Ollama

Mimick the local Ollama shell remotely.  

Ollama API: https://github.com/ollama/ollama/blob/main/docs/api.md  

# Usage
If `--file` is specified all console output is recorded to the specified file as well. The file is forever appended to.  
```
rosh.exe --ip 127.0.0.1 --port 11434 --file output.txt --model "llama3.2"
```

# Compile
```
# flags reduce binary size
go build -ldflags="-s -w"
```