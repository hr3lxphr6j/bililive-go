package charsets

import (
	"bytes"
	"golang.org/x/net/html"
	htmlcharset "golang.org/x/net/html/charset"
	"golang.org/x/text/encoding"
	"strings"
)

var boms = []struct {
	bom []byte
	enc string
}{
	{[]byte{0xfe, 0xff}, "utf-16be"},
	{[]byte{0xff, 0xfe}, "utf-16le"},
	{[]byte{0xef, 0xbb, 0xbf}, "utf-8"},
}

// FindEncoding sniff and find the encoding of the content.
func FindEncoding(content []byte) (enc encoding.Encoding, name string) {
	if len(content) == 0 {
		return
	}
	for _, b := range boms {
		if bytes.HasPrefix(content, b.bom) {
			enc, name = htmlcharset.Lookup(b.enc)
			if enc != nil {
				if strings.ToLower(name) == "utf-8" {
					enc = nil
				}
				return
			}
		}
	}
	enc, name = prescan(content)
	if strings.ToLower(name) == "utf-8" {
		enc = nil
	}
	return
}

func prescan(content []byte) (e encoding.Encoding, name string) {
	z := html.NewTokenizer(bytes.NewReader(content))
	for {
		switch z.Next() {
		case html.ErrorToken:
			return nil, ""

		case html.StartTagToken, html.SelfClosingTagToken:
			tagName, hasAttr := z.TagName()
			if !bytes.Equal(tagName, []byte("meta")) {
				continue
			}
			attrList := make(map[string]bool)
			gotPragma := false

			const (
				dontKnow = iota
				doNeedPragma
				doNotNeedPragma
			)
			needPragma := dontKnow

			name = ""
			e = nil
			for hasAttr {
				var key, val []byte
				key, val, hasAttr = z.TagAttr()
				ks := string(key)
				if attrList[ks] {
					continue
				}
				attrList[ks] = true
				for i, c := range val {
					if 'A' <= c && c <= 'Z' {
						val[i] = c + 0x20
					}
				}

				switch ks {
				case "http-equiv":
					if bytes.Equal(val, []byte("content-type")) {
						gotPragma = true
					}

				case "content":
					if e == nil {
						name = fromMetaElement(string(val))
						if name != "" {
							e, name = htmlcharset.Lookup(name)
							if e != nil {
								needPragma = doNeedPragma
							}
						}
					}

				case "charset":
					e, name = htmlcharset.Lookup(string(val))
					needPragma = doNotNeedPragma
				}
			}

			if needPragma == dontKnow || needPragma == doNeedPragma && !gotPragma {
				continue
			}

			if strings.HasPrefix(name, "utf-16") {
				name = "utf-8"
				e = encoding.Nop
			}

			if e != nil {
				return e, name
			}
		}
	}
}

func fromMetaElement(s string) string {
	for s != "" {
		csLoc := strings.Index(s, "charset")
		if csLoc == -1 {
			return ""
		}
		s = s[csLoc+len("charset"):]
		s = strings.TrimLeft(s, " \t\n\f\r")
		if !strings.HasPrefix(s, "=") {
			continue
		}
		s = s[1:]
		s = strings.TrimLeft(s, " \t\n\f\r")
		if s == "" {
			return ""
		}
		if q := s[0]; q == '"' || q == '\'' {
			s = s[1:]
			closeQuote := strings.IndexRune(s, rune(q))
			if closeQuote == -1 {
				return ""
			}
			return s[:closeQuote]
		}

		end := strings.IndexAny(s, "; \t\n\f\r")
		if end == -1 {
			end = len(s)
		}
		return s[:end]
	}
	return ""
}
