package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"unicode"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/quick"
	"github.com/alecthomas/chroma/v2/styles"
	"golang.org/x/term"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/yaml"
)

func main() {
	// Disable timestamp
	log.SetFlags(0)
	registerStyle()
	highlight(parse(os.Stdin))
}

func errFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

const (
	colorBackground = "#272822"
	colorForeground = "#f8f8f2"

	colorBlue    = "#66d9ef"
	colorGreen   = "#a6e22e"
	colorGrey    = "#75715e"
	colorPink    = "#f92672"
	colorPurple  = "#ae81ff"
	colorRed     = "#960050"
	colorRedDark = "#1e0010"
	colorYellow  = "#e6db74"
)

func registerStyle() {
	styles.Fallback = styles.Register(
		chroma.MustNewStyle(
			"ksd",
			chroma.StyleEntries{ //nolint:exhaustive // style only overrides a subset of token types
				chroma.Text:                colorForeground,
				chroma.Error:               colorRed + " bg:" + colorRedDark,
				chroma.Comment:             colorGrey,
				chroma.Keyword:             colorBlue,
				chroma.KeywordNamespace:    colorPink,
				chroma.Operator:            colorPink,
				chroma.Punctuation:         colorForeground,
				chroma.Name:                colorForeground,
				chroma.NameAttribute:       colorGreen,
				chroma.NameClass:           colorGreen,
				chroma.NameConstant:        colorBlue,
				chroma.NameDecorator:       colorGreen,
				chroma.NameException:       colorGreen,
				chroma.NameFunction:        colorGreen,
				chroma.NameOther:           colorGreen,
				chroma.NameTag:             colorPink,
				chroma.LiteralNumber:       colorPurple,
				chroma.Literal:             colorYellow,
				chroma.LiteralDate:         colorYellow,
				chroma.LiteralString:       colorYellow,
				chroma.LiteralStringEscape: colorPurple,
				chroma.GenericDeleted:      colorPink,
				chroma.GenericEmph:         "italic",
				chroma.GenericInserted:     colorGreen,
				chroma.GenericStrong:       "bold",
				chroma.GenericSubheading:   colorGrey,
				chroma.Background:          "bg:" + colorBackground,
			},
		),
	)
}

func parse(r io.Reader) string {
	bytes, err := io.ReadAll(r)
	errFatal(err)

	obj, _, err := scheme.Codecs.UniversalDeserializer().Decode(bytes, nil, nil)
	errFatal(err)

	switch o := obj.(type) {
	case *corev1.Secret:
		errFatal(yaml.Unmarshal(bytes, &o))
		decode(o)
		bytes, err = yaml.Marshal(o)
	case *corev1.List:
		var oo corev1.SecretList
		errFatal(yaml.Unmarshal(bytes, &oo))
		decodeList(&oo)
		bytes, err = yaml.Marshal(&oo)
	default:
		panic("unsupported object")
	}

	errFatal(err)

	return string(bytes)
}

func isASCII(s string) bool {
	for i := range len(s) {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}

	return true
}

func toStringData(b []byte) string {
	s := string(b)
	if !isASCII(s) {
		s = strconv.QuoteToASCII(s)

		return s[1 : len(s)-1]
	}

	return s
}

func decode(s *corev1.Secret) {
	s.StringData = make(map[string]string, len(s.Data))
	for k, v := range s.Data {
		s.StringData[k] = toStringData(v)
		delete(s.Data, k)
	}
}

func decodeList(sl *corev1.SecretList) {
	for i := range sl.Items {
		decode(&sl.Items[i])
	}
}

func highlight(data string) {
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		fmt.Println(data)

		return
	}

	errFatal(
		quick.Highlight(
			os.Stdout,
			data,
			"yaml",
			"terminal16m",
			"",
		),
	)
}
