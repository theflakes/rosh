# rosh
#### **R**emote **O**llama **SH**ell

Mimick the local Ollama shell remotely.  

##### [Expose Ollama service API to network](https://aident.ai/blog/how-to-expose-ollama-service-api-to-network) _(opens in a new tab)_
##### <a href="https://github.com/ollama/ollama/blob/main/docs/api.md" target="_blank">Ollama API</a>

# Usage
If `--file` is specified all console output is recorded to the specified file as well. The file is forever appended to.  
```
rosh.exe --ip 127.0.0.1 --port 11434 --file output.txt --model "llama3.2"

Connect to Olamma API via console.
  Author: Brian Kellogg
  License: MIT

Command line arguments:
  --ip    : IP address of the server (default: 127.0.0.1)
   -i     : IP address of the server (default: 127.0.0.1)
  --port  : Port of the server (default: 11434)
   -p     : Port of the server (default: 11434)
  --file  : File to save queries and responses (default: "")
   -f     : File to save queries and responses (default: "")
  --model : Model to use (default: llama3.2)
   -m     : Model to use (default: llama3.2)
```

# Compile
```
# flags reduce binary size
go build -ldflags="-s -w"
```