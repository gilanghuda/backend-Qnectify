package utils

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"time"
)

func SaveFile(idFile, filename string, data []byte) error {
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	if err := writer.WriteField("id_file", idFile); err != nil {
		return err
	}

	// Create form file part with explicit Content-Type header set to application/pdf
	head := textproto.MIMEHeader{}
	head.Set("Content-Disposition", fmt.Sprintf("form-data; name=\"file\"; filename=\"%s\"", filename))
	head.Set("Content-Type", "application/pdf")

	part, err := writer.CreatePart(head)
	if err != nil {
		return err
	}

	if _, err := part.Write(data); err != nil {
		return err
	}

	if err := writer.Close(); err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://api.gilanghuda.my.id/filesystem/files", &b)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		log.Printf("upload endpoint returned status %d: %s", resp.StatusCode, string(respBody))
		return fmt.Errorf("upload failed: status %d", resp.StatusCode)
	}

	return nil
}

func GetFile(idFile string) ([]byte, string, error) {
	url := fmt.Sprintf("https://api.gilanghuda.my.id/filesystem/files/%s", idFile)
	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		log.Printf("fetch endpoint returned status %d: %s", resp.StatusCode, string(respBody))
		return nil, "", fmt.Errorf("failed to fetch file: status %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	return data, contentType, nil
}
