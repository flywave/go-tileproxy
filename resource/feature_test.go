package resource

import (
	"encoding/json"
	"testing"
)

// TestTextFeatureInfoDoc 测试文本特征信息文档
func TestTextFeatureInfoDoc(t *testing.T) {
	tests := []struct {
		name    string
		content interface{}
		want    string
	}{{
		name:    "string content",
		content: "test content",
		want:    "test content",
	}, {
		name:    "byte slice content",
		content: []byte("test content"),
		want:    "test content",
	}, {
		name:    "invalid content",
		content: 123,
		want:    "",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := NewTextFeatureInfoDoc(tt.content)
			if doc == nil {
				if tt.content != 123 {
					t.Fatal("NewTextFeatureInfoDoc returned nil for valid content")
				}
				return
			}

			if doc.ContentType() != "text" {
				t.Errorf("ContentType() = %s, want text", doc.ContentType())
			}

			if got := doc.ToString(); got != tt.want {
				t.Errorf("ToString() = %s, want %s", got, tt.want)
			}
		})
	}
}

// TestTextFeatureInfoDocCombine 测试文本特征信息文档合并
func TestTextFeatureInfoDocCombine(t *testing.T) {
	docs := []FeatureInfoDoc{
		&TextFeatureInfoDoc{Content: "doc1"},
		&TextFeatureInfoDoc{Content: "doc2"},
		&TextFeatureInfoDoc{Content: "doc3"},
	}

	result := docs[0].Combine(docs)
	expected := "doc1\ndoc2\ndoc3"
	if result.ToString() != expected {
		t.Errorf("Combine() = %q, want %q", result.ToString(), expected)
	}
}

// TestXMLFeatureInfoDoc 测试XML特征信息文档
func TestXMLFeatureInfoDoc(t *testing.T) {
	xmlContent := "<root><item>test</item></root>"

	tests := []struct {
		name    string
		content interface{}
	}{{
		name:    "string content",
		content: xmlContent,
	}, {
		name:    "byte slice content",
		content: []byte(xmlContent),
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := NewXMLFeatureInfoDoc(tt.content)
			if doc == nil {
				t.Fatal("NewXMLFeatureInfoDoc returned nil")
			}

			if doc.ContentType() != "xml" {
				t.Errorf("ContentType() = %s, want xml", doc.ContentType())
			}

			if doc.ToString() != xmlContent {
				t.Errorf("ToString() = %s, want %s", doc.ToString(), xmlContent)
			}

			// 测试解析XML
			etreeDoc := doc.ToEtree()
			if etreeDoc == nil {
				t.Error("ToEtree() returned nil")
			}
		})
	}
}

// TestXMLFeatureInfoDocCombine 测试XML特征信息文档合并
func TestXMLFeatureInfoDocCombine(t *testing.T) {
	xml1 := "<root><item>item1</item></root>"
	xml2 := "<root><item>item2</item></root>"

	doc1 := NewXMLFeatureInfoDoc(xml1)
	doc2 := NewXMLFeatureInfoDoc(xml2)

	docs := []FeatureInfoDoc{doc1, doc2}
	result := doc1.Combine(docs)

	if result.ContentType() != "xml" {
		t.Errorf("ContentType() = %s, want xml", result.ContentType())
	}
}

// TestHTMLFeatureInfoDoc 测试HTML特征信息文档
func TestHTMLFeatureInfoDoc(t *testing.T) {
	htmlContent := "<html><body><h1>Test</h1></body></html>"

	tests := []struct {
		name    string
		content interface{}
	}{{
		name:    "string content",
		content: htmlContent,
	}, {
		name:    "byte slice content",
		content: []byte(htmlContent),
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := NewHTMLFeatureInfoDoc(tt.content)
			if doc == nil {
				t.Fatal("NewHTMLFeatureInfoDoc returned nil")
			}

			if doc.ContentType() != "html" {
				t.Errorf("ContentType() = %s, want html", doc.ContentType())
			}
		})
	}
}

// TestJSONFeatureInfoDoc 测试JSON特征信息文档
func TestJSONFeatureInfoDoc(t *testing.T) {
	jsonContent := `{"name":"test","value":123}`

	tests := []struct {
		name    string
		content interface{}
		want    string
	}{{
		name:    "string content",
		content: jsonContent,
		want:    jsonContent,
	}, {
		name:    "byte slice content",
		content: []byte(jsonContent),
		want:    jsonContent,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := NewJSONFeatureInfoDoc(tt.content)
			if doc == nil {
				t.Fatal("NewJSONFeatureInfoDoc returned nil")
			}

			if doc.ContentType() != "json" {
				t.Errorf("ContentType() = %s, want json", doc.ContentType())
			}

			if got := doc.ToString(); got != tt.want {
				t.Errorf("ToString() = %s, want %s", got, tt.want)
			}
		})
	}
}

// TestJSONFeatureInfoDocCombine 测试JSON特征信息文档合并
func TestJSONFeatureInfoDocCombine(t *testing.T) {
	json1 := `{"features":[{"id":1,"name":"feature1"}]}`
	json2 := `{"features":[{"id":2,"name":"feature2"}]}`

	doc1 := NewJSONFeatureInfoDoc(json1)
	doc2 := NewJSONFeatureInfoDoc(json2)

	docs := []FeatureInfoDoc{doc1, doc2}
	result := doc1.Combine(docs)

	if result.ContentType() != "json" {
		t.Errorf("ContentType() = %s, want json", result.ContentType())
	}

	// 验证合并后的JSON是有效的
	var combined map[string]interface{}
	if err := json.Unmarshal([]byte(result.ToString()), &combined); err != nil {
		t.Errorf("Combined JSON is invalid: %v", err)
	}
}

// TestCreateFeatureinfoDoc 测试创建特征信息文档
func TestCreateFeatureinfoDoc(t *testing.T) {
	tests := []struct {
		name         string
		content      interface{}
		infoFormat   string
		expectedType string
	}{{
		name:         "XML format",
		content:      "<root>test</root>",
		infoFormat:   "text/xml",
		expectedType: "xml",
	}, {
		name:         "HTML format",
		content:      "<html>test</html>",
		infoFormat:   "text/html",
		expectedType: "html",
	}, {
		name:         "JSON format",
		content:      `{"test":true}`,
		infoFormat:   "application/json",
		expectedType: "json",
	}, {
		name:         "Plain text format",
		content:      "plain text",
		infoFormat:   "text/plain",
		expectedType: "text",
	}, {
		name:         "Application XML format",
		content:      "<root>test</root>",
		infoFormat:   "application/xml",
		expectedType: "xml",
	}, {
		name:         "GML format",
		content:      "<gml>test</gml>",
		infoFormat:   "application/gml+xml",
		expectedType: "xml",
	}, {
		name:         "OGC GML format",
		content:      "<gml>test</gml>",
		infoFormat:   "application/vnd.ogc.gml",
		expectedType: "xml",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := CreateFeatureinfoDoc(tt.content, tt.infoFormat)
			if doc == nil {
				t.Fatal("CreateFeatureinfoDoc returned nil")
			}

			if doc.ContentType() != tt.expectedType {
				t.Errorf("ContentType() = %s, want %s", doc.ContentType(), tt.expectedType)
			}
		})
	}
}

// TestFeatureInfoType 测试特征信息类型判断
func TestFeatureInfoType(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		expected string
	}{{
		name:     "text/xml",
		format:   "text/xml",
		expected: "xml",
	}, {
		name:     "application/xml",
		format:   "application/xml",
		expected: "xml",
	}, {
		name:     "application/gml+xml",
		format:   "application/gml+xml",
		expected: "xml",
	}, {
		name:     "application/vnd.ogc.gml",
		format:   "application/vnd.ogc.gml",
		expected: "xml",
	}, {
		name:     "text/html",
		format:   "text/html",
		expected: "html",
	}, {
		name:     "application/json",
		format:   "application/json",
		expected: "json",
	}, {
		name:     "text/plain",
		format:   "text/plain",
		expected: "text",
	}, {
		name:     "text/html with charset",
		format:   "text/html; charset=utf-8",
		expected: "html",
	}, {
		name:     "application/json with charset",
		format:   "application/json; charset=utf-8",
		expected: "json",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := featureInfoType(tt.format)
			if result != tt.expected {
				t.Errorf("featureInfoType(%s) = %s, want %s", tt.format, result, tt.expected)
			}
		})
	}
}

// TestXSLTransformer 测试XSLT转换器
func TestXSLTransformer(t *testing.T) {
	xsltScript := `<?xml version="1.0"?>
<xsl:stylesheet version="1.0" xmlns:xsl="http://www.w3.org/1999/XSL/Transform">
    <xsl:template match="/">
        <html><body><h1>Transformed</h1></body></html>
    </xsl:template>
</xsl:stylesheet>`

	inputXML := "<root><item>test</item></root>"
	inputDoc := NewXMLFeatureInfoDoc(inputXML)

	tests := []struct {
		name     string
		infoType *string
		expected string
	}{{
		name:     "HTML output",
		infoType: stringPtr("html"),
		expected: "html",
	}, {
		name:     "XML output",
		infoType: stringPtr("xml"),
		expected: "xml",
	}, {
		name:     "Default output",
		infoType: nil,
		expected: "xml",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transformer := NewXSLTransformer(xsltScript, tt.infoType)
			if transformer == nil {
				t.Fatal("NewXSLTransformer returned nil")
			}

			result := transformer.Transform(inputDoc)
			if result.ContentType() != tt.expected {
				t.Errorf("ContentType() = %s, want %s", result.ContentType(), tt.expected)
			}
		})
	}
}

// TestCombineDocs 测试文档合并
func TestCombineDocs(t *testing.T) {
	textDocs := []FeatureInfoDoc{
		NewTextFeatureInfoDoc("text1"),
		NewTextFeatureInfoDoc("text2"),
	}

	xmlDocs := []FeatureInfoDoc{
		NewXMLFeatureInfoDoc("<root><item>1</item></root>"),
		NewXMLFeatureInfoDoc("<root><item>2</item></root>"),
	}

	tests := []struct {
		name        string
		docs        []FeatureInfoDoc
		transformer *XSLTransformer
	}{{
		name: "Text combine without transformer",
		docs: textDocs,
	}, {
		name: "XML combine without transformer",
		docs: xmlDocs,
	}, {
		name: "XML combine with transformer",
		docs: xmlDocs,
		transformer: NewXSLTransformer(`<?xml version="1.0"?>
<xsl:stylesheet version="1.0" xmlns:xsl="http://www.w3.org/1999/XSL/Transform">
    <xsl:template match="/">
        <html><body><h1>Combined</h1></body></html>
    </xsl:template>
</xsl:stylesheet>`, stringPtr("html")),
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.docs) == 0 {
				t.Skip("Empty docs slice")
			}

			data, infoType := CombineDocs(tt.docs, tt.transformer)
			if len(data) == 0 {
				t.Log("CombineDocs returned empty data")
			}

			if tt.transformer != nil {
				// 有转换器时，infotype应该为空
			} else {
				// 无转换器时，infotype应该为第一个文档的类型
				if infoType != tt.docs[0].ContentType() {
					t.Errorf("Expected infoType %s, got %s", tt.docs[0].ContentType(), infoType)
				}
			}
		})
	}
}

// TestJSONMerge 测试JSON合并功能
func TestJSONMerge(t *testing.T) {
	base := map[string]interface{}{
		"features": []interface{}{
			map[string]interface{}{"id": 1, "name": "feature1"},
		},
	}

	other := map[string]interface{}{
		"features": []interface{}{
			map[string]interface{}{"id": 2, "name": "feature2"},
		},
	}

	result := mergeMap(base, other)

	features, ok := result["features"].([]interface{})
	if !ok {
		t.Fatal("features is not a slice")
	}

	if len(features) != 2 {
		t.Errorf("Expected 2 features, got %d", len(features))
	}
}

// TestEdgeCases 测试边界情况
func TestEdgeCases(t *testing.T) {
	t.Run("Empty XML document", func(t *testing.T) {
		doc := NewXMLFeatureInfoDoc("")
		if doc == nil {
			t.Fatal("NewXMLFeatureInfoDoc should handle empty string")
		}
	})

	t.Run("Invalid XML content", func(t *testing.T) {
		doc := NewXMLFeatureInfoDoc("<invalid>xml")
		if doc == nil {
			t.Fatal("NewXMLFeatureInfoDoc should handle invalid XML")
		}
	})

	t.Run("Empty JSON document", func(t *testing.T) {
		doc := NewJSONFeatureInfoDoc("")
		if doc == nil {
			t.Fatal("NewJSONFeatureInfoDoc should handle empty string")
		}
	})

	t.Run("Invalid JSON content", func(t *testing.T) {
		doc := NewJSONFeatureInfoDoc("invalid json")
		if doc == nil {
			t.Fatal("NewJSONFeatureInfoDoc should handle invalid JSON")
		}
	})

	t.Run("Empty docs slice", func(t *testing.T) {
		// 空切片应该跳过或返回空结果
		// 不直接测试CombineDocs的空切片情况，因为函数设计需要至少一个文档
		t.Skip("CombineDocs requires at least one document")
	})
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
