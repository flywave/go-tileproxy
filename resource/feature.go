package resource

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/beevik/etree"
	"github.com/flywave/go-xslt"
	"golang.org/x/net/html"

	"github.com/flywave/go-tileproxy/utils"
)

type FeatureInfoDoc interface {
	ContentType() string
	ToString() string
	Combine(docs []FeatureInfoDoc) FeatureInfoDoc
}

type TextFeatureInfoDoc struct {
	FeatureInfoDoc
	Content string
}

func (d TextFeatureInfoDoc) ContentType() string {
	return "text"
}

func (d *TextFeatureInfoDoc) ToString() string {
	return d.Content
}

func (d TextFeatureInfoDoc) Combine(docs []FeatureInfoDoc) FeatureInfoDoc {
	docstrs := []string{}
	for _, d := range docs {
		if d.ContentType() == "text" {
			docstrs = append(docstrs, d.ToString())
		}
	}
	result := strings.Join(docstrs, "/n")
	return NewTextFeatureInfoDoc(result)
}

func NewTextFeatureInfoDoc(content interface{}) *TextFeatureInfoDoc {
	switch v := content.(type) {
	case string:
		return &TextFeatureInfoDoc{Content: v}
	case []byte:
		return &TextFeatureInfoDoc{Content: string(v)}
	}
	return nil
}

type XMLFeatureInfoDoc struct {
	FeatureInfoDoc
	Content string
	Doc     *etree.Document
}

func (d XMLFeatureInfoDoc) ContentType() string {
	return "xml"
}

func (d *XMLFeatureInfoDoc) ToString() string {
	return d.Content
}

func (d *XMLFeatureInfoDoc) ToEtree() *etree.Document {
	if d.Doc == nil {
		d.Doc = d.parseContext()
	}
	return d.Doc
}

func (d *XMLFeatureInfoDoc) parseContext() *etree.Document {
	doc := etree.NewDocument()
	doc.ReadFromBytes([]byte(d.Content))
	return doc
}

func NewXMLFeatureInfoDoc(content interface{}) *XMLFeatureInfoDoc {
	switch v := content.(type) {
	case string:
		return &XMLFeatureInfoDoc{Content: v}
	case []byte:
		return &XMLFeatureInfoDoc{Content: string(v)}
	case *etree.Document:
		return &XMLFeatureInfoDoc{Doc: v}
	}
	return nil
}

func (d XMLFeatureInfoDoc) Combine(docs []FeatureInfoDoc) FeatureInfoDoc {
	root := docs[0].(*XMLFeatureInfoDoc)
	result := root.ToEtree().Copy()

	for _, doc := range docs[1:] {
		xmldoc := doc.(*XMLFeatureInfoDoc)
		tree := xmldoc.ToEtree()
		for _, c := range tree.Root().Child {
			result.Root().AddChild(c)
		}
	}
	return NewXMLFeatureInfoDoc(result)
}

type HTMLFeatureInfoDoc struct {
	XMLFeatureInfoDoc
	root *html.Node
}

func (d HTMLFeatureInfoDoc) ContentType() string {
	return "html"
}

func NewHTMLFeatureInfoDoc(content interface{}) *HTMLFeatureInfoDoc {
	switch v := content.(type) {
	case string:
		{
			reader := bytes.NewBufferString(v)
			node, _ := html.Parse(reader)
			return &HTMLFeatureInfoDoc{XMLFeatureInfoDoc: XMLFeatureInfoDoc{Content: v}, root: node}
		}
	case []byte:
		{
			reader := bytes.NewBuffer(v)
			node, _ := html.Parse(reader)
			return &HTMLFeatureInfoDoc{XMLFeatureInfoDoc: XMLFeatureInfoDoc{Content: string(v)}, root: node}
		}
	case *etree.Document:
		return &HTMLFeatureInfoDoc{XMLFeatureInfoDoc: XMLFeatureInfoDoc{Doc: v}}
	}
	return nil
}

func (d HTMLFeatureInfoDoc) Combine(docs []FeatureInfoDoc) FeatureInfoDoc {
	root := docs[0].(*HTMLFeatureInfoDoc)
	result := root.ToEtree().Copy()

	for _, doc := range docs[1:] {
		xmldoc := doc.(*HTMLFeatureInfoDoc)
		tree := xmldoc.ToEtree()
		for _, c := range tree.Root().Child {
			result.Root().AddChild(c)
		}
	}
	return NewHTMLFeatureInfoDoc(result)
}

type JSONFeatureInfoDoc struct {
	FeatureInfoDoc
	Content string
}

func (d JSONFeatureInfoDoc) ContentType() string {
	return "json"
}

func (d *JSONFeatureInfoDoc) ToString() string {
	return d.Content
}

func NewJSONFeatureInfoDoc(content interface{}) *JSONFeatureInfoDoc {
	switch v := content.(type) {
	case string:
		return &JSONFeatureInfoDoc{Content: v}
	case []byte:
		return &JSONFeatureInfoDoc{Content: string(v)}
	}
	return nil
}

func merge_map(base map[string]interface{}, other map[string]interface{}) map[string]interface{} {
	for k, v := range other {
		if _, ok := base[k]; !ok {
			base[k] = v
		} else {
			if m, ok := base[k].(map[string]interface{}); ok {
				if m2, ok := v.(map[string]interface{}); ok {
					merge_map(m, m2)
				}
			} else if l, ok := base[k].([]interface{}); ok {
				if l2, ok := v.([]interface{}); ok {
					l = append(l, l2...)
				}
				base[k] = l
			} else {
				base[k] = v
			}
		}
	}
	return base
}

func (d JSONFeatureInfoDoc) Combine(docs []FeatureInfoDoc) FeatureInfoDoc {
	maps := []map[string]interface{}{}
	for _, doc := range docs {
		m := make(map[string]interface{})
		json.Unmarshal([]byte(doc.ToString()), &m)
		maps = append(maps, m)
	}
	result := make(map[string]interface{})
	for _, m := range maps {
		result = merge_map(result, m)
	}
	resultJson, _ := json.Marshal(result)
	return NewJSONFeatureInfoDoc(resultJson)
}

func CreateFeatureinfoDoc(content interface{}, info_format string) FeatureInfoDoc {
	info_type := featureInfoType(info_format)
	if info_type == "xml" {
		return NewXMLFeatureInfoDoc(content)
	}
	if info_type == "html" {
		return NewHTMLFeatureInfoDoc(content)
	}
	if info_type == "json" {
		return NewJSONFeatureInfoDoc(content)
	}
	return NewTextFeatureInfoDoc(content)
}

var (
	xml_mime = []string{
		"text/xml", "application/xml",
		"application/gml+xml", "application/vnd.ogc.gml",
	}
)

func featureInfoType(info_format string) string {
	formats := strings.Split(info_format, ";")
	info_format = strings.Trim(formats[0], "")
	if utils.ContainsString(xml_mime, info_format) {
		return "xml"
	}
	if info_format == "text/html" {
		return "html"
	}
	if info_format == "application/json" {
		return "json"
	}
	return "text"
}

type XSLTransformer struct {
	xsltscript string
	info_type  string
}

func NewXSLTransformer(xsltscript string, info_format *string) *XSLTransformer {
	if info_format != nil {
		return &XSLTransformer{xsltscript: xsltscript, info_type: *info_format}
	}
	return &XSLTransformer{xsltscript: xsltscript, info_type: "text/xml"}
}

func (t *XSLTransformer) Transform(input_doc *XMLFeatureInfoDoc) FeatureInfoDoc {
	xml := input_doc.ToString()
	xslt_tree, _ := xslt.NewStylesheet([]byte(t.xsltscript))

	output_xml, _ := xslt_tree.Transform([]byte(xml))
	if t.info_type == "html" {
		return NewHTMLFeatureInfoDoc(output_xml)
	} else {
		return NewXMLFeatureInfoDoc(output_xml)
	}
}

func CombineDocs(docs []FeatureInfoDoc, transformer *XSLTransformer) ([]byte, string) {
	infotype := docs[0].ContentType()
	doc := docs[0].Combine(docs)
	if transformer != nil {
		doc = transformer.Transform(doc.(*XMLFeatureInfoDoc))
		infotype = ""
	}
	return []byte(doc.ToString()), infotype
}
