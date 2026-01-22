package exporter

import (
	"archive/zip"
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteDocx(t *testing.T) {
	var buf bytes.Buffer
	text := "This is a test transcript."

	err := WriteDocx(&buf, text)
	require.NoError(t, err)

	// Verify it's a valid ZIP file
	reader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	require.NoError(t, err)

	// Check required files exist
	requiredFiles := []string{
		"[Content_Types].xml",
		"_rels/.rels",
		"word/document.xml",
	}

	foundFiles := make(map[string]bool)
	for _, file := range reader.File {
		foundFiles[file.Name] = true
	}

	for _, required := range requiredFiles {
		assert.True(t, foundFiles[required], "Missing required file: %s", required)
	}
}

func TestWriteDocx_ContainsText(t *testing.T) {
	var buf bytes.Buffer
	text := "Hello world transcript text"

	err := WriteDocx(&buf, text)
	require.NoError(t, err)

	// Read the document.xml content
	reader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	require.NoError(t, err)

	var docContent string
	for _, file := range reader.File {
		if file.Name == "word/document.xml" {
			rc, err := file.Open()
			require.NoError(t, err)
			content, err := io.ReadAll(rc)
			require.NoError(t, err)
			rc.Close()
			docContent = string(content)
			break
		}
	}

	assert.Contains(t, docContent, text)
}

func TestWriteDocx_EscapesXML(t *testing.T) {
	var buf bytes.Buffer
	text := "Text with <special> & \"chars\" and 'quotes'"

	err := WriteDocx(&buf, text)
	require.NoError(t, err)

	// Read and verify the content is properly escaped
	reader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	require.NoError(t, err)

	var docContent string
	for _, file := range reader.File {
		if file.Name == "word/document.xml" {
			rc, err := file.Open()
			require.NoError(t, err)
			content, err := io.ReadAll(rc)
			require.NoError(t, err)
			rc.Close()
			docContent = string(content)
			break
		}
	}

	// XML special characters should be escaped
	assert.Contains(t, docContent, "&lt;special&gt;")
	assert.Contains(t, docContent, "&amp;")
}

func TestWriteDocx_EmptyText(t *testing.T) {
	var buf bytes.Buffer

	err := WriteDocx(&buf, "")
	require.NoError(t, err)

	// Should still be a valid ZIP
	reader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	require.NoError(t, err)
	assert.Len(t, reader.File, 3)
}

func TestWriteDocx_LongText(t *testing.T) {
	var buf bytes.Buffer
	// Create a long text (10000 characters)
	text := strings.Repeat("Lorem ipsum dolor sit amet. ", 400)

	err := WriteDocx(&buf, text)
	require.NoError(t, err)

	// Read and verify content
	reader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	require.NoError(t, err)

	var docContent string
	for _, file := range reader.File {
		if file.Name == "word/document.xml" {
			rc, err := file.Open()
			require.NoError(t, err)
			content, err := io.ReadAll(rc)
			require.NoError(t, err)
			rc.Close()
			docContent = string(content)
			break
		}
	}

	assert.Contains(t, docContent, "Lorem ipsum")
}

func TestWriteDocx_UnicodeText(t *testing.T) {
	var buf bytes.Buffer
	text := "Привет мир! 你好世界! مرحبا بالعالم"

	err := WriteDocx(&buf, text)
	require.NoError(t, err)

	reader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	require.NoError(t, err)

	var docContent string
	for _, file := range reader.File {
		if file.Name == "word/document.xml" {
			rc, err := file.Open()
			require.NoError(t, err)
			content, err := io.ReadAll(rc)
			require.NoError(t, err)
			rc.Close()
			docContent = string(content)
			break
		}
	}

	assert.Contains(t, docContent, "Привет мир!")
	assert.Contains(t, docContent, "你好世界!")
}

func TestWriteDocx_NewlinesAndTabs(t *testing.T) {
	var buf bytes.Buffer
	text := "Line 1\nLine 2\tTabbed"

	err := WriteDocx(&buf, text)
	require.NoError(t, err)

	// Should not error with special whitespace
	reader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	require.NoError(t, err)
	assert.NotEmpty(t, reader.File)
}

func TestWriteDocx_ContentTypes(t *testing.T) {
	var buf bytes.Buffer

	err := WriteDocx(&buf, "test")
	require.NoError(t, err)

	reader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	require.NoError(t, err)

	var contentTypes string
	for _, file := range reader.File {
		if file.Name == "[Content_Types].xml" {
			rc, err := file.Open()
			require.NoError(t, err)
			content, err := io.ReadAll(rc)
			require.NoError(t, err)
			rc.Close()
			contentTypes = string(content)
			break
		}
	}

	assert.Contains(t, contentTypes, "wordprocessingml.document.main+xml")
	assert.Contains(t, contentTypes, "application/xml")
}
