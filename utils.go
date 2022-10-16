package main

import (
	"fmt"
	"html/template"
	"math"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/microcosm-cc/bluemonday"
	"go.yarn.social/lextwt"
	"go.yarn.social/types"
)

var (
	customTimeMagnitudes = []humanize.RelTimeMagnitude{
		{D: time.Second, Format: "now", DivBy: time.Second},
		{D: time.Minute, Format: "%ds %s", DivBy: time.Second},
		{D: time.Hour, Format: "%dm %s", DivBy: time.Minute},
		{D: humanize.Day, Format: "%dh %s", DivBy: time.Hour},
		{D: humanize.Week, Format: "%dd %s", DivBy: humanize.Day},
		{D: humanize.Year, Format: "%dw %s", DivBy: humanize.Week},
		{D: humanize.LongTime, Format: "%dy %s", DivBy: humanize.Year},
		{D: math.MaxInt64, Format: "a long while %s", DivBy: 1},
	}

	lastseenTimeMagnitudes = []humanize.RelTimeMagnitude{
		{D: humanize.Day, Format: "today", DivBy: time.Hour},
		{D: humanize.Week, Format: "%dd %s", DivBy: humanize.Day},
		{D: humanize.Year, Format: "%dw %s", DivBy: humanize.Week},
		{D: humanize.LongTime, Format: "%dy %s", DivBy: humanize.Year},
		{D: math.MaxInt64, Format: "never", DivBy: 1},
	}
)

// CustomRelTime returns a relative textual representation of two timestamps in a
// human readable short format.
func CustomRelTime(a, b time.Time, albl, blbl string) string {
	return humanize.CustomRelTime(a, b, albl, blbl, customTimeMagnitudes)
}

// CustomTime is a template function that returns a relative time representation of
// a timestamp compared to now in a short human readable form.
func CustomTime(then time.Time) string {
	return CustomRelTime(then, time.Now(), "ago", "from now")
}

// FormatTwt formats a twt into a valid HTML snippet
func FormatTwt(twt types.Twt) template.HTML {
	extensions := parser.NoIntraEmphasis | parser.FencedCode |
		parser.Autolink | parser.Strikethrough | parser.SpaceHeadings |
		parser.NoEmptyLineBeforeBlock | parser.HardLineBreak

	mdParser := parser.NewWithExtensions(extensions)

	htmlFlags := html.Smartypants | html.SmartypantsDashes | html.SmartypantsLatexDashes

	opts := html.RendererOptions{
		Flags:     htmlFlags,
		Generator: "",
	}

	renderer := html.NewRenderer(opts)

	// copy alt to title if present.
	if cp, ok := twt.(*lextwt.Twt); ok {
		twt = cp.Clone()
		for _, m := range twt.Links() {
			if link, ok := m.(*lextwt.Link); ok {
				link.TextToTitle()
			}
		}
	}

	// XXX: Note that even through we're calling twt.FormatText(types.HTMLFmt, conf) here
	// the output is in fact probably (mostly) Markdown anyway, we're just asking the lexttwt Parser
	// to render nodes as HTML snippets (like @-mentions).
	markdownInput := twt.FormatText(types.HTMLFmt, nil)

	md := []byte(markdownInput)
	maybeUnsafeHTML := markdown.ToHTML(md, mdParser, renderer)

	p := bluemonday.StrictPolicy()
	p.AllowStandardURLs()
	// Override the allowed schemas as p.AllowStandardURLs only permits http, https and mailto
	p.AllowURLSchemes("mailto", "http", "https", "gemini", "gopher")
	p.AllowElements("a", "img", "strong", "em", "del", "p", "br", "blockquote", "ul", "ol", "li", "pre", "code", "figure", "figcaption")
	p.AllowAttrs("href").OnElements("a")
	p.AllowAttrs("src").OnElements("img")
	p.AllowAttrs("id").OnElements("dialog")
	p.AllowAttrs("id", "controls").OnElements("audio")
	p.AllowAttrs("id", "controls", "playsinline", "preload", "poster").OnElements("video")
	p.AllowAttrs("src", "type").OnElements("source")
	p.AllowAttrs("aria-label", "class", "data-target", "target").OnElements("a")
	p.AllowAttrs("class", "data-target").OnElements("i", "lightbox")
	p.AllowAttrs("alt", "title", "loading", "data-target", "data-tooltip").OnElements("a", "img")
	html := p.SanitizeBytes(maybeUnsafeHTML)

	return template.HTML(fmt.Sprintf(`<p>%s</p>`, html))
}
