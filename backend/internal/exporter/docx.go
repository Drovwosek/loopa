package exporter

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
)

func WriteDocx(w io.Writer, text string) error {
	zipWriter := zip.NewWriter(w)

	if err := writeContentTypes(zipWriter); err != nil {
		return err
	}
	if err := writeRels(zipWriter); err != nil {
		return err
	}
	if err := writeDocument(zipWriter, text); err != nil {
		return err
	}
	return zipWriter.Close()
}

func writeContentTypes(zw *zip.Writer) error {
	content := `<?xml version="1.0" encoding="UTF-8"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
</Types>`
	return writeZipFile(zw, "[Content_Types].xml", []byte(content))
}

func writeRels(zw *zip.Writer) error {
	content := `<?xml version="1.0" encoding="UTF-8"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`
	return writeZipFile(zw, "_rels/.rels", []byte(content))
}

func writeDocument(zw *zip.Writer, text string) error {
	var escaped bytes.Buffer
	if err := xml.EscapeText(&escaped, []byte(text)); err != nil {
		return fmt.Errorf("escape text: %w", err)
	}
	content := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body>
    <w:p>
      <w:r>
        <w:t>%s</w:t>
      </w:r>
    </w:p>
  </w:body>
</w:document>`, escaped.String())
	return writeZipFile(zw, "word/document.xml", []byte(content))
}

func writeZipFile(zw *zip.Writer, name string, data []byte) error {
	file, err := zw.Create(name)
	if err != nil {
		return err
	}
	_, err = file.Write(data)
	return err
}
