package controllers

import (
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

var uploadPath = "./public"

// UploadFile handles single file upload
func UploadFile(options map[string]interface{}) map[string]interface{} {
	c, ok := options["context"].(*gin.Context)
	if !ok || c == nil {
		return map[string]interface{}{
			"success": false,
			"message": "gin context required",
		}
	}

	fileField, ok := options["file_name"].(string)
	if !ok || fileField == "" {
		return map[string]interface{}{
			"success": false,
			"message": "file name required",
		}
	}

	file, err := c.FormFile(fileField)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": "File upload error: " + err.Error(),
		}
	}

	if err := os.MkdirAll(uploadPath, 0755); err != nil {
		return map[string]interface{}{
			"success": false,
			"message": "Unable to create upload directory: " + err.Error(),
		}
	}

	dst := filepath.Join(uploadPath, filepath.Base(file.Filename))
	if err := c.SaveUploadedFile(file, dst); err != nil {
		return map[string]interface{}{
			"success": false,
			"message": "Unable to save file: " + err.Error(),
		}
	}

	return map[string]interface{}{
		"success":  true,
		"message":  "File uploaded successfully",
		"filename": file.Filename,
		"path":     dst,
	}
}

// UploadMultipleFiles handles multiple file uploads
func UploadMultipleFiles(options map[string]interface{}) map[string]interface{} {
	c, ok := options["context"].(*gin.Context)
	if !ok || c == nil {
		return map[string]interface{}{
			"success": false,
			"message": "gin context required",
		}
	}

	form, err := c.MultipartForm()
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"message": "Invalid form: " + err.Error(),
		}
	}

	files := form.File["files"] // expects input name="files"
	if len(files) == 0 {
		return map[string]interface{}{
			"success": false,
			"message": "No files uploaded",
		}
	}

	if err := os.MkdirAll(uploadPath, 0755); err != nil {
		return map[string]interface{}{
			"success": false,
			"message": "Unable to create upload directory: " + err.Error(),
		}
	}

	var uploaded []string
	for _, file := range files {
		dst := filepath.Join(uploadPath, filepath.Base(file.Filename))
		if err := c.SaveUploadedFile(file, dst); err != nil {
			continue
		}
		uploaded = append(uploaded, file.Filename)
	}

	return map[string]interface{}{
		"success":  true,
		"message":  "Files uploaded successfully",
		"uploaded": uploaded,
	}
}

// DownloadFile sends a file to the client
func DownloadFile(options map[string]interface{}) map[string]interface{} {
	c, ok := options["context"].(*gin.Context)
	if !ok || c == nil {
		return map[string]interface{}{
			"success": false,
			"message": "gin context required",
		}
	}

	fileName, ok := options["file_name"].(string)
	if !ok || fileName == "" {
		return map[string]interface{}{
			"success": false,
			"message": "file name required",
		}
	}

	fullPath := filepath.Join(uploadPath, fileName)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return map[string]interface{}{
			"success": false,
			"message": "File not found",
		}
	}

	c.FileAttachment(fullPath, fileName)
	return map[string]interface{}{
		"success":  true,
		"message":  "File download started",
		"filename": fileName,
		"path":     fullPath,
	}
}

// DeleteFile removes a file
func DeleteFile(options map[string]interface{}) map[string]interface{} {
	c, ok := options["context"].(*gin.Context)
	if !ok || c == nil {
		return map[string]interface{}{
			"success": false,
			"message": "gin context required",
		}
	}

	fileName := c.Param("filename")
	if fileName == "" {
		return map[string]interface{}{
			"success": false,
			"message": "filename param required",
		}
	}

	fullPath := filepath.Join(uploadPath, fileName)
	if err := os.Remove(fullPath); err != nil {
		return map[string]interface{}{
			"success": false,
			"message": "Failed to delete file: " + err.Error(),
		}
	}

	return map[string]interface{}{
		"success": true,
		"message": "File deleted successfully",
	}
}

// CheckFile verifies if a file exists
func CheckFile(options map[string]interface{}) map[string]interface{} {
	c, ok := options["context"].(*gin.Context)
	if !ok || c == nil {
		return map[string]interface{}{
			"success": false,
			"message": "gin context required",
		}
	}

	fileName := c.Param("filename")
	if fileName == "" {
		return map[string]interface{}{
			"success": false,
			"message": "filename param required",
		}
	}

	fullPath := filepath.Join(uploadPath, fileName)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return map[string]interface{}{
			"success": false,
			"message": "File not found",
		}
	}

	return map[string]interface{}{
		"success":  true,
		"message":  "File exists",
		"filename": fileName,
		"path":     fullPath,
	}
}
