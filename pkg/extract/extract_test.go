package extract

import (
	"archive/zip"
	"bytes"
	"testing"
)

func TestDocx(t *testing.T) {
	docxXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body>
    <w:p>
      <w:r><w:t>Hello World</w:t></w:r>
    </w:p>
    <w:p>
      <w:r><w:t>This is a test document.</w:t></w:r>
    </w:p>
  </w:body>
</w:document>`

	zipBuf := &bytes.Buffer{}
	z := zip.NewWriter(zipBuf)
	w, _ := z.Create("word/document.xml")
	w.Write([]byte(docxXML))
	z.Close()

	text, err := Docx(zipBuf.Bytes())
	if err != nil {
		t.Fatalf("Docx failed: %v", err)
	}

	expected := "Hello World This is a test document."
	if text != expected {
		t.Errorf("Expected %q, got %q", expected, text)
	}
}

func TestXlsx(t *testing.T) {
	sharedStringsXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<sst xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <si><t>Cell 1</t></si>
  <si><t>Cell 2</t></si>
  <si><t>Cell 3</t></si>
</sst>`

	zipBuf := &bytes.Buffer{}
	z := zip.NewWriter(zipBuf)
	w, _ := z.Create("xl/sharedStrings.xml")
	w.Write([]byte(sharedStringsXML))
	z.Close()

	text, err := Xlsx(zipBuf.Bytes())
	if err != nil {
		t.Fatalf("Xlsx failed: %v", err)
	}

	expected := "Cell 1\nCell 2\nCell 3"
	if text != expected {
		t.Errorf("Expected %q, got %q", expected, text)
	}
}

func TestPptx(t *testing.T) {
	slide1XML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:sld xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">
  <p:sp>
    <p:txBody><a:t>Slide 1 Title</a:t></p:txBody>
  </p:sp>
  <p:sp>
    <p:txBody><a:t>Slide 1 Content</a:t></p:txBody>
  </p:sp>
</p:sld>`

	slide2XML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:sld xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">
  <p:sp>
    <p:txBody><a:t>Slide 2 Title</a:t></p:txBody>
  </p:sp>
</p:sld>`

	zipBuf := &bytes.Buffer{}
	z := zip.NewWriter(zipBuf)
	w1, _ := z.Create("ppt/slides/slide1.xml")
	w1.Write([]byte(slide1XML))
	w2, _ := z.Create("ppt/slides/slide2.xml")
	w2.Write([]byte(slide2XML))
	z.Close()

	text, err := Pptx(zipBuf.Bytes())
	if err != nil {
		t.Fatalf("Pptx failed: %v", err)
	}

	expected := "Slide 1 Title Slide 1 Content\n---\nSlide 2 Title"
	if text != expected {
		t.Errorf("Expected %q, got %q", expected, text)
	}
}

func TestDocxEmpty(t *testing.T) {
	docxXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body>
  </w:body>
</w:document>`

	zipBuf := &bytes.Buffer{}
	z := zip.NewWriter(zipBuf)
	w, _ := z.Create("word/document.xml")
	w.Write([]byte(docxXML))
	z.Close()

	text, err := Docx(zipBuf.Bytes())
	if err != nil {
		t.Fatalf("Docx failed: %v", err)
	}

	if text != "" {
		t.Errorf("Expected empty string, got %q", text)
	}
}

func TestDocxInvalid(t *testing.T) {
	_, err := Docx([]byte("not a zip file"))
	if err == nil {
		t.Error("Expected error for invalid zip")
	}
}

func TestXlsxInvalid(t *testing.T) {
	_, err := Xlsx([]byte("not a zip file"))
	if err == nil {
		t.Error("Expected error for invalid zip")
	}
}

func TestPptxInvalid(t *testing.T) {
	_, err := Pptx([]byte("not a zip file"))
	if err == nil {
		t.Error("Expected error for invalid zip")
	}
}

func TestDocxNotFound(t *testing.T) {
	zipBuf := &bytes.Buffer{}
	z := zip.NewWriter(zipBuf)
	z.Create("word/other.xml")
	z.Close()

	_, err := Docx(zipBuf.Bytes())
	if err == nil {
		t.Error("Expected error for missing document.xml")
	}
}

func TestXlsxNotFound(t *testing.T) {
	zipBuf := &bytes.Buffer{}
	z := zip.NewWriter(zipBuf)
	z.Create("xl/other.xml")
	z.Close()

	_, err := Xlsx(zipBuf.Bytes())
	if err == nil {
		t.Error("Expected error for missing sharedStrings.xml")
	}
}

func TestPptxNoSlides(t *testing.T) {
	zipBuf := &bytes.Buffer{}
	z := zip.NewWriter(zipBuf)
	z.Create("ppt/other.xml")
	z.Close()

	_, err := Pptx(zipBuf.Bytes())
	if err == nil {
		t.Error("Expected error for no slide files")
	}
}
