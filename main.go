package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"unicode"

	"github.com/ghodss/yaml"
	"github.com/mattn/go-isatty"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

func main() {
	// Disable timestamp
	log.SetFlags(0)
	fmt.Print(highlight(parse(os.Stdin)))
}

func errFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func parse(r io.Reader) string {
	bytes, err := ioutil.ReadAll(r)
	errFatal(err)

	obj, _, err := scheme.Codecs.UniversalDeserializer().Decode(bytes, nil, nil)
	errFatal(err)

	switch o := obj.(type) {
	case *v1.Secret:
		errFatal(yaml.Unmarshal(bytes, &o))
		decode(o)
		bytes, err = yaml.Marshal(o)
	case *v1.List:
		var oo v1.SecretList
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

func decode(s *v1.Secret) {
	s.StringData = make(map[string]string, len(s.Data))
	for k, v := range s.Data {
		s.StringData[k] = toStringData(v)
		delete(s.Data, k)
	}
}

func decodeList(sl *v1.SecretList) {
	for i := range sl.Items {
		decode(&sl.Items[i])
	}
}

func highlight(input string) string {
	if !isatty.IsTerminal(os.Stdout.Fd()) {
		return input
	}
	bat, err := exec.LookPath("bat")
	if err != nil {
		return input
	}
	cmd := exec.Command(
		bat,
		"--color=always",
		"--language=yaml",
		"--paging=never",
		"--plain",
	)
	cmd.Stdin = strings.NewReader(input)
	var out bytes.Buffer
	cmd.Stdout = &out
	errFatal(cmd.Run())
	return out.String()
}
