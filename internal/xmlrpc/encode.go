package xmlrpc

import (
	"fmt"
	"strings"
)

const xmlHeader = `<?xml version="1.0" encoding="UTF-8"?>` + "\n"

func encodeRequest(method string, params []Param) []byte {
	var b strings.Builder
	b.WriteString(xmlHeader)
	b.WriteString("<methodCall><methodName>")
	xmlEscapeTo(&b, method)
	b.WriteString("</methodName><params>")

	for _, p := range params {
		b.WriteString("<param>")
		encodeValue(&b, p.Value)
		b.WriteString("</param>")
	}

	b.WriteString("</params></methodCall>")
	return []byte(b.String())
}

func encodeValue(b *strings.Builder, v Value) {
	switch {
	case v.String != nil:
		b.WriteString("<value><string>")
		xmlEscapeTo(b, *v.String)
		b.WriteString("</string></value>")
	case v.Int != nil:
		fmt.Fprintf(b, "<value><int>%d</int></value>", *v.Int)
	case v.Bool != nil:
		if *v.Bool {
			b.WriteString("<value><boolean>1</boolean></value>")
		} else {
			b.WriteString("<value><boolean>0</boolean></value>")
		}
	case v.Struct != nil:
		b.WriteString("<value><struct>")
		for _, m := range v.Struct {
			b.WriteString("<member><name>")
			xmlEscapeTo(b, m.Name)
			b.WriteString("</name>")
			encodeValue(b, m.Value)
			b.WriteString("</member>")
		}
		b.WriteString("</struct></value>")
	case v.Array != nil:
		b.WriteString("<value><array><data>")
		for _, item := range v.Array {
			encodeValue(b, item)
		}
		b.WriteString("</data></array></value>")
	default:
		b.WriteString("<value><string></string></value>")
	}
}

func xmlEscapeTo(b *strings.Builder, s string) {
	for _, r := range s {
		switch r {
		case '&':
			b.WriteString("&amp;")
		case '<':
			b.WriteString("&lt;")
		case '>':
			b.WriteString("&gt;")
		case '"':
			b.WriteString("&quot;")
		case '\'':
			b.WriteString("&apos;")
		default:
			b.WriteRune(r)
		}
	}
}
