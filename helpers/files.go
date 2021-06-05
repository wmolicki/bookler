package helpers

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// StoreFile will save the srcFile to disk, and return hashed filename and error
func StoreFile(filename string, srcFile io.ReadCloser) (string, error) {
	defer srcFile.Close()
	ext := ""
	if strings.Contains(filename, ".") {
		ext = filepath.Ext(filename)
	}

	hasher := md5.New()
	io.WriteString(hasher, filename)
	hashed := base64.URLEncoding.EncodeToString(hasher.Sum(nil)) + ext

	dstF, err := os.Create("./static/upload/" + hashed)
	defer dstF.Close()
	Must(err)
	_, err = io.Copy(dstF, srcFile)
	if err != nil {
		return "", fmt.Errorf("error writing file to disk: %v", err)
	}

	return hashed, nil
}
