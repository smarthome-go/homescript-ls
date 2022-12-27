package main

import (
	"net/url"
	"runtime"
	"strings"

	"github.com/pkg/errors"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func (s *Lsp) DidOpen(params protocol.DidOpenTextDocumentParams, notify glsp.NotifyFunc) (*document, error) {
	uri := params.TextDocument.URI
	path, err := normalizePath(uri)
	if err != nil {
		return nil, err
	}
	doc := &document{
		URI:     uri,
		Path:    path,
		Content: params.TextDocument.Text,
	}
	s.documents[path] = doc
	return doc, nil
}

func (s *Lsp) Get(pathOrURI string) (*document, bool) {
	path, err := normalizePath(pathOrURI)
	if err != nil {
		return nil, false
	}
	d, ok := s.documents[path]
	return d, ok
}

func uriToPath(uri string) (string, error) {
	s := strings.ReplaceAll(uri, "%5C", "/")
	parsed, err := url.Parse(s)
	if err != nil {
		return "", err
	}
	if parsed.Scheme != "file" {
		return "", errors.New("URI was not a file:// URI")
	}

	if runtime.GOOS == "windows" {
		return "", errors.New("Windows is not supported")
	}
	return parsed.Path, nil
}

func normalizePath(pathOrUri string) (string, error) {
	path, err := uriToPath(pathOrUri)
	if err != nil {
		return "", errors.Wrapf(err, "unable to parse URI: %s", pathOrUri)
	}
	return path, nil
}

// document represents an opened file.
type document struct {
	URI                     protocol.DocumentUri
	Path                    string
	NeedsRefreshDiagnostics bool
	Content                 string
	lines                   []string
}

// ApplyChanges updates the content of the document from LSP textDocument/didChange events.
func (d *document) ApplyChanges(changes []interface{}) {
	for _, change := range changes {
		switch c := change.(type) {
		case protocol.TextDocumentContentChangeEvent:
			startIndex, endIndex := c.Range.IndexesIn(d.Content)
			d.Content = d.Content[:startIndex] + c.Text + d.Content[endIndex:]
		case protocol.TextDocumentContentChangeEventWhole:
			d.Content = c.Text
		}
	}

	d.lines = nil
}

// WordAt returns the word found at the given location.
func (d *document) WordAt(pos protocol.Position) string {
	line, ok := d.GetLine(int(pos.Line))
	if !ok {
		return ""
	}
	return WordAt(line, int(pos.Character))
}

// ContentAtRange returns the document text at given range.
func (d *document) ContentAtRange(rng protocol.Range) string {
	startIndex, endIndex := rng.IndexesIn(d.Content)
	return d.Content[startIndex:endIndex]
}

// GetLine returns the line at the given index.
func (d *document) GetLine(index int) (string, bool) {
	lines := d.GetLines()
	if index < 0 || index > len(lines) {
		return "", false
	}
	return lines[index], true
}

// GetLines returns all the lines in the document.
func (d *document) GetLines() []string {
	if d.lines == nil {
		// We keep \r on purpose, to avoid messing up position conversions.
		d.lines = strings.Split(d.Content, "\n")
	}
	return d.lines
}

// LookBehind returns the n characters before the given position, on the same line.
func (d *document) LookBehind(pos protocol.Position, length int) string {
	line, ok := d.GetLine(int(pos.Line))
	if !ok {
		return ""
	}

	charIdx := int(pos.Character)
	if length > charIdx {
		return line[0:charIdx]
	}
	return line[(charIdx - length):charIdx]
}

// LookForward returns the n characters after the given position, on the same line.
func (d *document) LookForward(pos protocol.Position, length int) string {
	line, ok := d.GetLine(int(pos.Line))
	if !ok {
		return ""
	}

	lineLength := len(line)
	charIdx := int(pos.Character)
	if lineLength <= charIdx+length {
		return line[charIdx:]
	}
	return line[charIdx:(charIdx + length)]
}
