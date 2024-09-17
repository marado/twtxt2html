package main

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

func TestBrackets(t *testing.T) {
	var markdownInput = "an ascii heart <3 is here"
	var expectedOutput = "<p>an ascii heart &lt;3 is here</p>"

	var md = []byte(markdownInput)

	var htmlr bytes.Buffer

	gm := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Linkify,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
			html.WithUnsafe(),
		),
	)

	if err := gm.Convert(md, &htmlr); err != nil {
		t.Errorf("conversion failed!")
		panic(err)
	}

	var htmlResult = fmt.Sprintf(`%s`, htmlr.String())

	if len(expectedOutput) > len(htmlResult) {
		t.Errorf("test failed, the output (%s) is shorter than what was expected (%s).", htmlResult, expectedOutput)
	} else if htmlResult[0:len(expectedOutput)]+"" != expectedOutput {
		print("DEBUG: %s\n", htmlResult)
		t.Errorf("test failed, the output is not what was expected ('%s'), but instead '%s'", expectedOutput, htmlResult[0:len(expectedOutput)]+"")
	}
}
