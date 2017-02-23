package handler

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/playlyfe/go-graphql"

	"golang.org/x/net/context"
)

//Shortcuts for the Content-Type header
const (
	ContentTypeJSON           = "application/json"
	ContentTypeGraphQL        = "application/graphql"
	ContentTypeFormURLEncoded = "application/x-www-form-urlencoded"
)

//Handler structure
type Handler struct {
	Executor *graphql.Executor
	Context  interface{}
	Pretty   bool
}

//RequestParameters from query like " /graphql?query=getUser($id:ID){lastName}&variables={"id":"4"} "
type RequestParameters struct {
	Query         string                 `json:"query" url:"query" schema:"query"`
	Variables     map[string]interface{} `json:"variables" url:"variables" schema:"variables"`
	OperationName string                 `json:"operationName" url:"operationName" schema:"operationName"`
}

//RequestParametersCompatibility represents an workaround for getting`variables` as a JSON string
type RequestParametersCompatibility struct {
	Query         string `json:"query" url:"query" schema:"query"`
	Variables     string `json:"variables" url:"variables" schema:"variables"`
	OperationName string `json:"operationName" url:"operationName" schema:"operationName"`
}

func getFromURL(values url.Values) *RequestParameters {
	if values.Get("query") != "" {
		// get variables map
		var variables map[string]interface{}
		variablesStr := values.Get("variables")
		json.Unmarshal([]byte(variablesStr), variables)

		return &RequestParameters{
			Query:         values.Get("query"),
			Variables:     variables,
			OperationName: values.Get("operationName"),
		}
	}

	return nil
}

// NewRequestParameters Parses a http.Request into GraphQL request options struct
func NewRequestParameters(r *http.Request) *RequestParameters {
	if reqParams := getFromURL(r.URL.Query()); reqParams != nil {
		return reqParams
	}

	if r.Method != "POST" {
		return &RequestParameters{}
	}

	if r.Body == nil {
		return &RequestParameters{}
	}

	// TODO: improve Content-Type handling
	contentTypeStr := r.Header.Get("Content-Type")
	contentTypeTokens := strings.Split(contentTypeStr, ";")
	contentType := contentTypeTokens[0]

	switch contentType {
	case ContentTypeGraphQL:
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return &RequestParameters{}
		}
		return &RequestParameters{
			Query: string(body),
		}
	case ContentTypeFormURLEncoded:
		if err := r.ParseForm(); err != nil {
			return &RequestParameters{}
		}

		if reqParams := getFromURL(r.PostForm); reqParams != nil {
			return reqParams
		}

		return &RequestParameters{}

	case ContentTypeJSON:
		fallthrough
	default:
		var params RequestParameters
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return &params
		}
		err = json.Unmarshal(body, &params)
		if err != nil {
			// Probably `variables` was sent as a string instead of an object.
			// So, we try to be polite and try to parse that as a JSON string
			var CompatibleParams RequestParametersCompatibility
			json.Unmarshal(body, &CompatibleParams)
			json.Unmarshal([]byte(CompatibleParams.Variables), &params.Variables)
		}
		return &params
	}
}

// ContextHandler provides an entrypoint into executing graphQL queries with a
// user-provided context.
func (h *Handler) ContextHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	// get query
	params := NewRequestParameters(r)
	// execute graphql query
	result, _ := (h.Executor).Execute(h.Context, params.Query, params.Variables, params.OperationName)

	if h.Pretty {
		w.WriteHeader(200) //http.StatusOK = 200
		buff, _ := json.MarshalIndent(result, "", "   ")
		w.Write(buff)
	} else {
		w.WriteHeader(200) //http.StatusOK = 200
		buff, _ := json.Marshal(result)
		w.Write(buff)
	}
}

// ServeHTTP provides an entrypoint into executing graphQL queries.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.ContextHandler(context.Background(), w, r)
}

//Config for handler of the schema
type Config Handler

//New config
func New(c *Config) *Handler {
	if c == nil {
		c = &Config{
			Executor: nil,
			Context:  "",
			Pretty:   true,
		}
	}
	if c.Executor == nil {
		panic("Undefined GraphQL Executor")
	}

	return &Handler{
		Executor: c.Executor,
		Context:  c.Context,
		Pretty:   c.Pretty,
	}
}
