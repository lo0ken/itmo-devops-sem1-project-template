package services

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"strings"
)

// ArchiveService предоставляет методы для работы с архивами
type ArchiveService struct{}

// NewArchiveService создает новый экземпляр сервиса
func NewArchiveService() *ArchiveService {
	return &ArchiveService{}
}

// Extract извлекает CSV файл из архива (ZIP или TAR)
func (s *ArchiveService) Extract(data []byte, archiveType string) ([]byte, error) {
	switch archiveType {
	case "zip":
		return s.extractZip(data)
	case "tar":
		return s.extractTar(data)
	default:
		return nil, fmt.Errorf("unsupported archive type: %s", archiveType)
	}
}

// extractZip извлекает первый CSV файл из ZIP архива
func (s *ArchiveService) extractZip(data []byte) ([]byte, error) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("failed to read zip archive: %w", err)
	}

	// Поиск первого CSV файла
	for _, file := range reader.File {
		if strings.HasSuffix(strings.ToLower(file.Name), ".csv") {
			f, err := file.Open()
			if err != nil {
				return nil, fmt.Errorf("failed to open csv file in zip: %w", err)
			}
			defer f.Close()

			csvData, err := io.ReadAll(f)
			if err != nil {
				return nil, fmt.Errorf("failed to read csv file from zip: %w", err)
			}

			return csvData, nil
		}
	}

	return nil, fmt.Errorf("no CSV file found in zip archive")
}

// extractTar извлекает первый CSV файл из TAR архива
func (s *ArchiveService) extractTar(data []byte) ([]byte, error) {
	tarReader := tar.NewReader(bytes.NewReader(data))

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read tar archive: %w", err)
		}

		if header.Typeflag == tar.TypeDir {
			continue
		}

		if strings.Contains(header.Name, "._") {
			continue
		}

		if strings.HasSuffix(strings.ToLower(header.Name), ".csv") {
			csvData, err := io.ReadAll(tarReader)
			if err != nil {
				return nil, fmt.Errorf("failed to read csv file from tar: %w", err)
			}

			return csvData, nil
		}
	}

	return nil, fmt.Errorf("no CSV file found in tar archive")
}

func (s *ArchiveService) CreateZip(csvData []byte, filename string) ([]byte, error) {
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	fileWriter, err := zipWriter.Create(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create file in zip: %w", err)
	}

	_, err = fileWriter.Write(csvData)
	if err != nil {
		return nil, fmt.Errorf("failed to write data to zip: %w", err)
	}

	if err := zipWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close zip: %w", err)
	}

	return buf.Bytes(), nil
}
