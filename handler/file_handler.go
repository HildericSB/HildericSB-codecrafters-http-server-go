package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	httpPkg "github.com/codecrafters-io/http-server-starter-go/http"
)

type FileHandler struct {
	fileDir string
}

func NewFileHandler(fileDir string) *FileHandler {
	return &FileHandler{fileDir: fileDir}
}

func (fh *FileHandler) Handle(req *httpPkg.Request, res *httpPkg.Response) {
	if req.Method == "GET" {
		fh.handleRead(req, res)
	} else if req.Method == "POST" {
		fh.handleUpload(req, res)
	} else {
		res.StatusCode = http.StatusMethodNotAllowed
	}
}

func (fh *FileHandler) handleUpload(request *httpPkg.Request, response *httpPkg.Response) {
	contentLength, err := strconv.Atoi(request.Headers["Content-Length"])
	if err != nil {
		response.StatusCode = http.StatusBadRequest
		return
	}

	if len(request.Body) < contentLength {
		response.StatusCode = http.StatusBadRequest
		return
	}

	fileName := filepath.Base(request.Path)
	filePath := filepath.Join(fh.fileDir, fileName)
	fileData := request.Body[:contentLength]

	err = os.WriteFile(filePath, []byte(fileData), 0666)
	if err != nil {
		fmt.Printf("Error writing file %s: %v\n", filePath, err)
		response.StatusCode = http.StatusInternalServerError
		return
	}

	response.StatusCode = http.StatusCreated
}

func (fh *FileHandler) handleRead(request *httpPkg.Request, response *httpPkg.Response) {
	pathsplit := strings.Split(request.Path, "/")
	if len(pathsplit) < 3 || pathsplit[2] == "" {
		response.StatusCode = http.StatusBadRequest
		return
	}

	filename := pathsplit[2]
	// Validate filename to prevent directory traversal
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") {
		response.StatusCode = http.StatusBadRequest
		return
	}

	filePath := filepath.Join(fh.fileDir, filename)

	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error reading file ", filePath, err)
		response.StatusCode = http.StatusNotFound
		return
	}

	response.StatusCode = http.StatusOK
	response.Headers["Content-Type"] = "application/octet-stream"
	response.Headers["Content-Length"] = strconv.Itoa(len(content))
	response.Body = string(content)
}
