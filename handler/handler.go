package handler

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/codecrafters-io/http-server-starter-go/http"
)

func HandleFile(request *http.Request, response *http.Response, fileDir string) {
	if request.Method == "GET" {
		HandleFileRead(request, response, fileDir)
	} else {
		HandleFileUpload(request, response, fileDir)
	}
}

func HandleFileUpload(request *http.Request, response *http.Response, fileDir string) {
	contentLength, err := strconv.Atoi(request.Headers["Content-Length"])
	if err != nil {
		response.StatusCode = 400
		return
	}

	if len(request.Body) < contentLength {
		response.StatusCode = 400
		return
	}

	fileName := filepath.Base(request.Path)
	filePath := filepath.Join(fileDir, fileName)
	fileData := request.Body[:contentLength]

	err = os.WriteFile(filePath, []byte(fileData), 0666)
	if err != nil {
		fmt.Printf("Error writing file %s: %v\n", filePath, err)
		response.StatusCode = 500
		return
	}

	response.StatusCode = 201
}
func HandleFileRead(request *http.Request, response *http.Response, fileDir string) {
	pathsplit := strings.Split(request.Path, "/")

	if len(pathsplit) < 3 {
		response.StatusCode = 400
		return
	}

	content, err := os.ReadFile(fileDir + pathsplit[2])
	if err != nil {
		fmt.Println("Error reading file ", fileDir+pathsplit[2], err)
		response.StatusCode = 404
		return
	}

	response.StatusCode = 200
	response.Headers["Content-Type"] = "application/octet-stream"
	response.Headers["Content-Length"] = strconv.Itoa(len(content))
	response.Body = string(content)
}

func HandleEcho(request *http.Request, response *http.Response) {
	body := strings.TrimPrefix(request.Path, "/echo/")
	response.StatusCode = 200
	response.Headers["Content-Type"] = "text/plain"
	response.Headers["Content-Length"] = strconv.Itoa(len(body))
	response.Body = body
}

func HandleUserAgent(request *http.Request, response *http.Response) {
	userAgent := request.Headers["User-Agent"]
	if userAgent == "" {
		response.StatusCode = 400
		return
	}

	response.StatusCode = 200
	response.Headers["Content-Type"] = "text/plain"
	response.Headers["Content-Length"] = strconv.Itoa(len(userAgent))
	response.Body = userAgent
}
