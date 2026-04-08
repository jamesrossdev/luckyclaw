package extract

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
)

func Docx(data []byte) (string, error) {
	reader := bytes.NewReader(data)
	z, err := zip.NewReader(reader, reader.Size())
	if err != nil {
		return "", fmt.Errorf("failed to open zip: %w", err)
	}

	var docXML []byte
	for _, f := range z.File {
		if f.Name == "word/document.xml" {
			rc, err := f.Open()
			if err != nil {
				return "", fmt.Errorf("failed to open word/document.xml: %w", err)
			}
			docXML, err = io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return "", fmt.Errorf("failed to read word/document.xml: %w", err)
			}
			break
		}
	}

	if len(docXML) == 0 {
		return "", errors.New("word/document.xml not found in docx")
	}

	return extractDocxText(docXML), nil
}

func extractDocxText(xmlData []byte) string {
	re := regexp.MustCompile(`<w:t[^>]*>([^<]*)</w:t>`)
	matches := re.FindAllStringSubmatch(string(xmlData), -1)
	var result []string
	for _, match := range matches {
		if len(match) > 1 && match[1] != "" {
			result = append(result, match[1])
		}
	}
	return strings.Join(result, " ")
}

func Xlsx(data []byte) (string, error) {
	reader := bytes.NewReader(data)
	z, err := zip.NewReader(reader, reader.Size())
	if err != nil {
		return "", fmt.Errorf("failed to open zip: %w", err)
	}

	var stringsXML []byte
	for _, f := range z.File {
		if f.Name == "xl/sharedStrings.xml" {
			rc, err := f.Open()
			if err != nil {
				return "", fmt.Errorf("failed to open xl/sharedStrings.xml: %w", err)
			}
			stringsXML, err = io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return "", fmt.Errorf("failed to read xl/sharedStrings.xml: %w", err)
			}
			break
		}
	}

	if len(stringsXML) == 0 {
		return "", errors.New("xl/sharedStrings.xml not found in xlsx")
	}

	return extractXlsxText(stringsXML), nil
}

func extractXlsxText(xmlData []byte) string {
	re := regexp.MustCompile(`<t>([^<]*)</t>`)
	matches := re.FindAllStringSubmatch(string(xmlData), -1)
	var result []string
	for _, match := range matches {
		if len(match) > 1 && match[1] != "" {
			result = append(result, match[1])
		}
	}
	return strings.Join(result, "\n")
}

func Pptx(data []byte) (string, error) {
	reader := bytes.NewReader(data)
	z, err := zip.NewReader(reader, reader.Size())
	if err != nil {
		return "", fmt.Errorf("failed to open zip: %w", err)
	}

	var slideFiles []string
	for _, f := range z.File {
		if matched, _ := regexp.MatchString(`ppt/slides/slide\d+\.xml`, f.Name); matched {
			slideFiles = append(slideFiles, f.Name)
		}
	}

	if len(slideFiles) == 0 {
		return "", errors.New("no slide files found in pptx")
	}

	sort.Strings(slideFiles)

	var allText []string
	for _, slideName := range slideFiles {
		for _, f := range z.File {
			if f.Name == slideName {
				rc, err := f.Open()
				if err != nil {
					continue
				}
				slideXML, err := io.ReadAll(rc)
				rc.Close()
				if err != nil {
					continue
				}
				text := extractPptxSlideText(slideXML)
				if text != "" {
					allText = append(allText, text)
				}
				break
			}
		}
	}

	return strings.Join(allText, "\n---\n"), nil
}

func extractPptxSlideText(xmlData []byte) string {
	re := regexp.MustCompile(`<a:t>([^<]*)</a:t>`)
	matches := re.FindAllStringSubmatch(string(xmlData), -1)
	var result []string
	for _, match := range matches {
		if len(match) > 1 && match[1] != "" {
			result = append(result, match[1])
		}
	}
	return strings.Join(result, " ")
}
