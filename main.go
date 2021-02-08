package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"unicode"

	"golang.org/x/term"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/quick"
	"github.com/alecthomas/chroma/styles"
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

func registerStyle() {
	styles.Fallback = styles.Register(
		chroma.MustNewStyle(
			"ksd",
			chroma.StyleEntries{
				chroma.Text:                "#f8f8f2",
				chroma.Error:               "#960050 bg:#1e0010",
				chroma.Comment:             "#75715e",
				chroma.Keyword:             "#66d9ef",
				chroma.KeywordNamespace:    "#f92672",
				chroma.Operator:            "#f92672",
				chroma.Punctuation:         "#f8f8f2",
				chroma.Name:                "#f8f8f2",
				chroma.NameAttribute:       "#a6e22e",
				chroma.NameClass:           "#a6e22e",
				chroma.NameConstant:        "#66d9ef",
				chroma.NameDecorator:       "#a6e22e",
				chroma.NameException:       "#a6e22e",
				chroma.NameFunction:        "#a6e22e",
				chroma.NameOther:           "#a6e22e",
				chroma.NameTag:             "#f92672",
				chroma.LiteralNumber:       "#ae81ff",
				chroma.Literal:             "#e6db74",
				chroma.LiteralDate:         "#e6db74",
				chroma.LiteralString:       "#e6db74",
				chroma.LiteralStringEscape: "#ae81ff",
				chroma.GenericDeleted:      "#f92672",
				chroma.GenericEmph:         "italic",
				chroma.GenericInserted:     "#a6e22e",
				chroma.GenericStrong:       "bold",
				chroma.GenericSubheading:   "#75715e",
				chroma.Background:          "bg:#272822",
			},
		),
	)
}

func parse(r io.Reader) string {
	bytes, err := ioutil.ReadAll(r)
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
	for i := 0; i < len(s); i++ {
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
