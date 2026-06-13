package services

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"testing"
)

func newFileHeader(t *testing.T, filename string, content []byte) *multipart.FileHeader {
	t.Helper()

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile("f", filename)
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatalf("write part: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	form, err := multipart.NewReader(&buf, w.Boundary()).ReadForm(int64(buf.Len()) + 1024)
	if err != nil {
		t.Fatalf("ReadForm: %v", err)
	}
	return form.File["f"][0]
}

func TestValidatePdfAndOpenFile(t *testing.T) {
	validPdf := append([]byte("%PDF-1.4\n"), bytes.Repeat([]byte("a"), 32)...)

	cases := []struct {
		name     string
		filename string
		content  []byte
		wantErr  bool
	}{
		{"valid pdf", "doc.pdf", validPdf, false},
		{"wrong extension", "doc.txt", validPdf, true},
		{"junk content", "doc.pdf", []byte("not a pdf at all, just text"), true},
	}

	svc := &FileService{}
	for _, c := range cases {
		fh := newFileHeader(t, c.filename, c.content)
		rc, err := svc.validatePdfAndOpenFile(fh)

		if c.wantErr {
			if !errors.Is(err, ErrPdfFileFormat) {
				t.Errorf("%s: err = %v, want ErrPdfFileFormat", c.name, err)
			}
			continue
		}

		if err != nil {
			t.Errorf("%s: unexpected err %v", c.name, err)
			continue
		}
		got, readErr := io.ReadAll(rc)
		err = rc.Close()
		if err != nil {
			t.Errorf("failed to close io.ReadCloser: %v", err)
		}
		if readErr != nil {
			t.Errorf("%s: read returned reader: %v", c.name, readErr)
		}
		if !bytes.Equal(got, c.content) {
			t.Errorf("%s: reader yielded %d bytes from offset 0, want %d (Seek rewind failed)", c.name, len(got), len(c.content))
		}
	}
}
