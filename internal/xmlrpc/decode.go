package xmlrpc

import (
	"encoding/xml"
	"strings"
)

type fault struct {
	Code    int
	Message string
}

func parseFault(data []byte) (fault, bool) {
	s := string(data)
	if !strings.Contains(s, "<fault>") {
		return fault{}, false
	}

	var resp struct {
		Fault struct {
			Value struct {
				Struct struct {
					Members []struct {
						Name  string `xml:"name"`
						Value struct {
							Int    *int   `xml:"int"`
							String string `xml:"string"`
						} `xml:"value"`
					} `xml:"member"`
				} `xml:"struct"`
			} `xml:"value"`
		} `xml:"fault"`
	}

	if err := xml.Unmarshal(data, &resp); err != nil {
		return fault{Message: "unknown fault"}, true
	}

	f := fault{}
	for _, m := range resp.Fault.Value.Struct.Members {
		switch m.Name {
		case "faultCode":
			if m.Value.Int != nil {
				f.Code = *m.Value.Int
			}
		case "faultString":
			f.Message = m.Value.String
		}
	}
	return f, true
}

// ExtractStringValue extracts the first string value from an XML-RPC response.
func ExtractStringValue(data []byte) string {
	var resp struct {
		Params struct {
			Param struct {
				Value struct {
					String string `xml:"string"`
				} `xml:"value"`
			} `xml:"param"`
		} `xml:"params"`
	}
	if err := xml.Unmarshal(data, &resp); err != nil {
		return ""
	}
	return resp.Params.Param.Value.String
}

// ExtractBoolValue extracts the first boolean value from an XML-RPC response.
func ExtractBoolValue(data []byte) bool {
	var resp struct {
		Params struct {
			Param struct {
				Value struct {
					Bool string `xml:"boolean"`
				} `xml:"value"`
			} `xml:"param"`
		} `xml:"params"`
	}
	if err := xml.Unmarshal(data, &resp); err != nil {
		return false
	}
	return resp.Params.Param.Value.Bool == "1" || resp.Params.Param.Value.Bool == "true"
}

// PostInfo represents a blog post from getRecentPosts.
type PostInfo struct {
	PostID string
	Title  string
	Link   string
}

// ExtractRecentPosts parses the getRecentPosts response.
func ExtractRecentPosts(data []byte) []PostInfo {
	// The response is an array of structs. We'll do simple string parsing
	// since the XML structure is deeply nested.
	type member struct {
		Name  string `xml:"name"`
		Value struct {
			String string `xml:"string"`
		} `xml:"value"`
	}
	type structVal struct {
		Members []member `xml:"member"`
	}
	type value struct {
		Struct *structVal `xml:"struct"`
	}
	type arrayData struct {
		Values []value `xml:"value"`
	}
	type param struct {
		Value struct {
			Array struct {
				Data arrayData `xml:"data"`
			} `xml:"array"`
		} `xml:"value"`
	}
	var resp struct {
		Params struct {
			Param param `xml:"param"`
		} `xml:"params"`
	}

	if err := xml.Unmarshal(data, &resp); err != nil {
		return nil
	}

	var posts []PostInfo
	for _, v := range resp.Params.Param.Value.Array.Data.Values {
		if v.Struct == nil {
			continue
		}
		p := PostInfo{}
		for _, m := range v.Struct.Members {
			switch m.Name {
			case "postid":
				p.PostID = m.Value.String
			case "title":
				p.Title = m.Value.String
			case "link":
				p.Link = m.Value.String
			}
		}
		if p.PostID != "" {
			posts = append(posts, p)
		}
	}
	return posts
}
