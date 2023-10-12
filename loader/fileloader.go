package loader

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/tealeg/xlsx/v3"
)

type FileLoader interface {
	GetFileContent(fileURL string) ([]string, error)
}

type fileLoader struct {
	client    *http.Client
	Content   string
	itemLimit int
}

func NewFileLoader(itemLimit int) *fileLoader {
	client := http.DefaultClient
	return &fileLoader{
		client:    client,
		itemLimit: itemLimit,
		Content:   "",
	}
}

func (f *fileLoader) GetFileContent(fileURL string) ([]string, error) {
	_, err := f.validateFile(fileURL)
	if err != nil {
		return nil, err
	}
	items := strings.Split(f.Content, ",")
	return items, nil
}

func (f *fileLoader) validateFile(fileURL string) (int, error) {
	err := f.loadFile(fileURL)
	if err != nil {
		return 0, err
	}
	if len(f.Content) == 0 {
		return 0, errors.New("empty file")
	}

	emailQuantity := f.getQuantityOfItems(f.Content)
	if emailQuantity > f.itemLimit {
		return 0, errors.New("file with too many Items")
	}

	if emailQuantity <= 1 {
		return 0, errors.New("file has few Items")
	}
	return emailQuantity, nil
}

func (f *fileLoader) loadFile(url string) error {
	res, err := f.client.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return errors.New("invalid file url")
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	contentType := res.Header.Get("Content-Type")
	fileType, ok := f.getFileType(contentType)
	if !ok {
		return errors.New("invalid file type")
	}

	var content string
	switch fileType {
	case "csv":
		content = string(body)
	case "xlsx":
		content, err = f.extractContentOfXLSX(body)
		if err != nil {
			return err
		}
	}

	f.Content = content
	return nil
}

func (f *fileLoader) getFileType(contentType string) (string, bool) {
	csvTypes := []string{"text/csv", "text/plain", "encoding/csv"}
	xlsxTypes := []string{"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", "application/vnd.ms-excel"}
	for _, csvType := range csvTypes {
		if strings.Contains(contentType, csvType) {
			return "csv", true
		}
	}
	for _, xlsxType := range xlsxTypes {
		if strings.Contains(contentType, xlsxType) {
			return "xlsx", true
		}
	}
	return "", false
}

func (s *fileLoader) getQuantityOfItems(content string) int {
	chars := len(strings.TrimSpace(content))
	commas := strings.Count(content, ",")
	if chars > 0 && commas == 0 {
		return 1
	}
	if commas > 0 {
		return commas + 1
	}
	return commas
}

func (f *fileLoader) extractContentOfXLSX(buff []byte) (string, error) {
	file, err := xlsx.OpenBinary(buff, xlsx.ValueOnly())
	if err != nil {
		return "", err
	}
	sheet := file.Sheets[0]
	if sheet == nil {
		return "", errors.New("invalid file content")
	}
	var lines []string
	err = sheet.ForEachRow(func(row *xlsx.Row) error {
		return row.ForEachCell(func(cell *xlsx.Cell) error {
			value, err := cell.FormattedValue()
			if err != nil {
				return nil
			}
			if value != "" {
				lines = append(lines, value)
			}
			return nil
		})
	})
	if err != nil {
		return "", errors.New("invalid file content")
	}
	content := strings.Join(lines, ",")
	return content, nil
}
