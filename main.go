package main

import (
	"fmt"
	"os"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/smarthome-go/homescript/v2/homescript"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
	"github.com/tliron/glsp/server"
	"github.com/tliron/kutil/logging"

	// Must include a backend implementation. See kutil's logging/ for other options.
	_ "github.com/tliron/kutil/logging/simple"
)

const lsName = "homescript"

var version string = "2.0.0"
var handler protocol.Handler

var lspServer Lsp

type Lsp struct {
	server    *server.Server
	documents map[string]*document
}

func main() {
	// This increases logging verbosity (optional)
	logging.Configure(5, nil)

	handler = protocol.Handler{
		Initialize:             initialize,
		Initialized:            initialized,
		Shutdown:               shutdown,
		SetTrace:               setTrace,
		TextDocumentDidChange:  change,
		TextDocumentDidOpen:    open,
		TextDocumentCompletion: complete,
		TextDocumentHover:      hover,
	}

	server := server.NewServer(&handler, lsName, true)

	lspServer = Lsp{
		server:    server,
		documents: make(map[string]*document),
	}

	if err := server.RunStdio(); err != nil {
		panic(err.Error())
	}
}

func open(context *glsp.Context, params *protocol.DidOpenTextDocumentParams) error {
	doc, err := lspServer.DidOpen(*params, context.Notify)
	if err != nil {
		return err
	}
	if doc != nil {
		refreshDiagnosticsOfDocument(doc, context.Notify)
	}
	return nil
}
func change(context *glsp.Context, params *protocol.DidChangeTextDocumentParams) error {
	doc, ok := lspServer.Get(params.TextDocument.URI)
	if !ok {
		return nil
	}
	doc.ApplyChanges(params.ContentChanges)
	time.Sleep(100 * time.Millisecond)
	refreshDiagnosticsOfDocument(doc, context.Notify)
	return nil
}

func refreshDiagnosticsOfDocument(doc *document, notify glsp.NotifyFunc) {
	results, _, _ := homescript.Analyze(
		dummyExecutor{},
		doc.Content,
		make(map[string]homescript.Value),
		make([]string, 0),
		doc.Path,
	)

	os.WriteFile("/tmp/hmsls.txt", []byte(spew.Sdump(results)), 0755)

	if len(results) == 0 {
		diagnostics := []protocol.Diagnostic{}
		go notify(protocol.ServerTextDocumentPublishDiagnostics, protocol.PublishDiagnosticsParams{
			URI:         doc.URI,
			Diagnostics: diagnostics,
		})
		return
	}

	diagnostics := []protocol.Diagnostic{}
	src := fmt.Sprintf("Homescript@%s", version)
	for _, err := range results {
		severity := 0
		switch err.Severity {
		case homescript.Error:
			severity = 1
		case homescript.Warning:
			severity = 2
		case homescript.Info:
			severity = 3
		}
		severe := protocol.DiagnosticSeverity(severity)
		diagnostics = append(diagnostics, protocol.Diagnostic{
			Range: protocol.Range{
				Start: protocol.Position{
					Line:      uint32(err.Span.Start.Line) - 1,
					Character: uint32(err.Span.Start.Column) - 1,
				},
				End: protocol.Position{
					Line:      uint32(err.Span.End.Line) - 1,
					Character: uint32(err.Span.End.Column),
				},
			},
			Severity: &severe,
			Source:   &src,
			Message:  fmt.Sprintf("%s: %s", err.Kind, err.Message),
		})
	}

	go notify(protocol.ServerTextDocumentPublishDiagnostics, protocol.PublishDiagnosticsParams{
		URI:         doc.URI,
		Diagnostics: diagnostics,
	})
}

func initialize(context *glsp.Context, params *protocol.InitializeParams) (any, error) {
	capabilities := handler.CreateServerCapabilities()

	if params.Trace != nil {
		protocol.SetTraceValue(*params.Trace)
	}

	capabilities.TextDocumentSync = protocol.TextDocumentSyncKindIncremental
	capabilities.CompletionProvider = &protocol.CompletionOptions{}

	return &protocol.InitializeResult{
		Capabilities: capabilities,
		ServerInfo: &protocol.InitializeResultServerInfo{
			Name:    "homescript-language-server",
			Version: &version,
		},
	}, nil
}

func hover(context *glsp.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
	doc, ok := lspServer.Get(params.TextDocument.URI)
	if !ok {
		return nil, nil
	}

	_, symbols, _ := homescript.Analyze(
		dummyExecutor{},
		doc.Content,
		make(map[string]homescript.Value),
		make([]string, 0),
		doc.Path,
	)

	for _, symbol := range symbols {
		// Detect if the line overlaps
		if symbol.Span.Start.Line == uint(params.Position.Line)+1 || symbol.Span.End.Line == uint(params.Position.Line)+1 {
			// Check that the column is also overlapping
			if params.Position.Character+1 <= uint32(symbol.Span.End.Column) && params.Position.Character+1 >= uint32(symbol.Span.Start.Column) {
				return &protocol.Hover{
					Contents: protocol.MarkupContent{
						Kind:  protocol.MarkupKindMarkdown,
						Value: string(symbol.Type),
					},
				}, nil
			}
		}

	}

	return nil, nil
}

func complete(context *glsp.Context, params *protocol.CompletionParams) (interface{}, error) {
	doc, ok := lspServer.Get(params.TextDocument.URI)
	if !ok {
		return nil, nil
	}
	fmt.Println(doc)

	items := make([]protocol.CompletionItem, 0)

	items = append(items, protocol.CompletionItem{
		Label:      "switch",
		InsertText: strPtr("switch"),
		Detail:     strPtr("Switch builting function"),
		Kind:       kindPtr(protocol.CompletionItemKindFunction),
	})

	format := protocol.InsertTextFormatSnippet
	items = append(items, protocol.CompletionItem{
		Label:            "for",
		InsertText:       strPtr("for i in ${1:0}..${2:upper} {$0}"),
		Detail:           strPtr("For loop"),
		Kind:             kindPtr(protocol.CompletionItemKindSnippet),
		InsertTextFormat: &format,
	})

	items = append(items, protocol.CompletionItem{
		Label:      "return",
		InsertText: strPtr("return"),
		Detail:     strPtr("Return from function"),
		Kind:       kindPtr(protocol.CompletionItemKindKeyword),
	})

	items = append(items, protocol.CompletionItem{
		Label:      "break",
		InsertText: strPtr("break"),
		Detail:     strPtr("Break in loop"),
		Kind:       kindPtr(protocol.CompletionItemKindKeyword),
	})

	items = append(items, protocol.CompletionItem{
		Label:      "continue",
		InsertText: strPtr("continue;"),
		Detail:     strPtr("Continue in loop"),
		Kind:       kindPtr(protocol.CompletionItemKindKeyword),
	})

	if params.Context != nil && params.Context.TriggerKind == protocol.CompletionTriggerKindInvoked {
		//return server.buildInvokedCompletionList(notebook, doc, params.Position)
		return items, nil
	} else {
		//return server.buildTriggerCompletionList(notebook, doc, params.Position)
		return items, nil
	}
}

func kindPtr(in protocol.CompletionItemKind) *protocol.CompletionItemKind {
	return &in
}

func strPtr(in string) *string {
	return &in
}

func initialized(context *glsp.Context, params *protocol.InitializedParams) error {
	return nil
}

func shutdown(context *glsp.Context) error {
	protocol.SetTraceValue(protocol.TraceValueOff)
	return nil
}

func setTrace(context *glsp.Context, params *protocol.SetTraceParams) error {
	protocol.SetTraceValue(params.Value)
	return nil
}
