// twtxt2html is a command-line tool that generates HTML pages out of Twtxt feeds
// which can be used in conjuction with static site generators or directly hosted.
package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"git.mills.io/prologic/go-gopher"
	"git.mills.io/yarnsocial/yarn"
	"github.com/Masterminds/sprig"
	"github.com/makeworld-the-better-one/go-gemini"
	sync "github.com/sasha-s/go-deadlock"
	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"go.yarn.social/lextwt"
	"go.yarn.social/types"
)

const (
	helpText = `
twtxt2html converts a twtxt feed to a static HTML page
`

	htmlTemplate = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <link rel="stylesheet" href="https://cdn.simplecss.org/simple.min.css">
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0">
    <title>{{ .Title }}</title>
  </head>
	<body class="preload">
	  <main class="container">
		{{ range $_, $twt := $.Twts }}
		  <article id="{{ $twt.Hash }}" class="h-entry">
			<div class="u-author h-card">
			  <div class="dt-publish">
				<a class="u-url" href="#{{ $twt.Hash }}">
				  <time class="dt-published" datetime="{{ $twt.Created | date "2006-01-02T15:04:05Z07:00" }}">
					{{ $twt.Created }}
				  </time>
				</a>
				<span>&nbsp;{{ $twt.Created | time }}</span>
				<a class="u-search" href="https://search.twtxt.net/twt/{{ $twt.Hash }}">(search)</a>
			  </div>
			</div>
			<div class="e-content">
			  {{ formatTwt $twt }}
			</div>
		  </article>
		{{ end }}
	  </main>
	</body>
</html>`
)

var (
	debug   bool
	version bool

	limit     int
	reverse   bool
	title     string
	tmplFIle  string
	noreldate bool
)

type context struct {
	Title     string
	Twts      types.Twts
	NoRelDate bool
}

func init() {
	baseProg := filepath.Base(os.Args[0])
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] FILE|URL\n", baseProg)
		fmt.Fprint(os.Stderr, helpText)
		flag.PrintDefaults()
	}

	flag.BoolVarP(&debug, "debug", "d", false, "enable debug logging")
	flag.BoolVarP(&version, "version", "v", false, "display version information")

	flag.IntVarP(&limit, "limit", "l", -1, "limit number ot twts (default all)")
	flag.BoolVarP(&reverse, "reverse", "r", false, "reverse the order of twts (oldest first)")
	flag.StringVarP(&title, "title", "t", "Twtxt Feed", "title of generated page")
	flag.StringVarP(&tmplFIle, "template", "T", "", "path to template file")
	flag.BoolVarP(&noreldate, "noreldate", "n", false, "disable relative timestamps")
}

func flagNameFromEnvironmentName(s string) string {
	s = strings.ToLower(s)
	s = strings.Replace(s, "_", "-", -1)
	return s
}

func parseArgs() error {
	for _, v := range os.Environ() {
		vals := strings.SplitN(v, "=", 2)
		flagName := flagNameFromEnvironmentName(vals[0])
		fn := flag.CommandLine.Lookup(flagName)
		if fn == nil || fn.Changed {
			continue
		}
		if err := fn.Value.Set(vals[1]); err != nil {
			return err
		}
	}
	flag.Parse()
	return nil
}

func main() {
	parseArgs()

	if version {
		fmt.Printf("twtxt2html version %s\n", yarn.FullVersion())
		os.Exit(0)
	}

	if debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)

		// Disable deadlock detection in production mode
		sync.Opts.Disable = true
	}

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	t := htmlTemplate

	url, err := url.Parse(flag.Arg(0))
	if err != nil {
		log.WithError(err).Error("error parsing url")
		os.Exit(2)
	}

	switch url.Scheme {
	case "", "file":
		f, err := os.Open(url.Path)
		if err != nil {
			log.WithError(err).Error("error reading file feed")
			os.Exit(2)
		}
		defer f.Close()

		twtxt2HTML(t, f)
	case "http", "https":
		f, err := http.Get(url.String())
		if err != nil {
			log.WithError(err).Error("error reading HTTP feed")
			os.Exit(2)
		}
		defer f.Body.Close()

		twtxt2HTML(t, f.Body)
	case "gopher":
		res, err := gopher.Get(url.String())
		if err != nil {
			log.WithError(err).Error("error reading Gopher feed")
			os.Exit(2)
		}
		defer res.Body.Close()

		twtxt2HTML(t, res.Body)
	case "gemini":
		res, err := gemini.Fetch(url.String())
		if err != nil {
			log.WithError(err).Error("error reading Gemini feed")
			os.Exit(2)
		}
		defer res.Body.Close()

		twtxt2HTML(t, res.Body)
	default:
		log.WithError(err).Errorf("unsupported url scheme: %s", url.Scheme)
		os.Exit(2)
	}
}

func render(tpl string, ctx context) (string, error) {
	funcMap := sprig.FuncMap()

	if ctx.NoRelDate {
		funcMap["time"] = NoCustomTime
	} else {
		funcMap["time"] = CustomTime
	}
	funcMap["formatTwt"] = FormatTwt

	t := template.Must(template.New("tpl").Funcs(funcMap).Parse(tpl))

	buf := bytes.NewBuffer([]byte{})
	err := t.Execute(buf, ctx)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func twtxt2HTML(t string, r io.Reader) {
	twter := types.NilTwt.Twter()
	tf, err := lextwt.ParseFile(r, &twter)
	if err != nil {
		log.WithError(err).Error("error parsing feed")
		os.Exit(2)
	}

	twts := tf.Twts()
	sort.Sort(twts)

	if reverse {
		sort.Sort(sort.Reverse(twts))
	}

	if limit > -1 {
		n := limit
		if n > len(twts) {
			n = len(twts)
		}
		twts = twts[:n]
	}

	ctx := context{
		Title:     title,
		Twts:      twts,
		NoRelDate: noreldate,
	}

	html, err := render(t, ctx)
	if err != nil {
		log.WithError(err).Errorf("error rendering feed")
		os.Exit(2)
	}

	fmt.Println(string(html))
}
