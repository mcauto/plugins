package http

import (
	"fmt"
	"regexp"
	"strings"

	"goa.design/goa/codegen"
	goadesign "goa.design/goa/design"
	"goa.design/goa/eval"
	httpcodegen "goa.design/goa/http/codegen"
	"goa.design/goa/http/codegen/openapi"
	httpdesign "goa.design/goa/http/design"
	seccodegen "goa.design/plugins/security/codegen"
	"goa.design/plugins/security/design"
)

type (
	// ServiceData contains the data necessary to render the secure endpoints
	// constructor in example.
	ServiceData struct {
		*httpcodegen.ServiceData
		// Schemes is the unique security schemes for the service.
		Schemes []*design.SchemeExpr
	}

	// SecureDecoderData contains the data necessary to render the security aware
	// HTTP decoder.
	SecureDecoderData struct {
		// ServiceName is the name of the service.
		ServiceName string
		// MethodName is the name of the method.
		MethodName string
		// SecureRequestDecoder is the name of the generated decoder function.
		SecureRequestDecoder string
		// RequestDecoder is the name of the decoder function
		// originally generated by goa.
		RequestDecoder string
		// PayloadType is the Go type name of the secured method payload.
		PayloadType string
		// Schemes contains the security scheme data needed to render the auth
		// code.
		Schemes []*seccodegen.SchemeData
	}

	// SecureEncoderData contains the data necessary to render the security aware
	// HTTP encoder.
	SecureEncoderData struct {
		// ServiceName is the name of the service.
		ServiceName string
		// MethodName is the name of the method.
		MethodName string
		// SecureRequestEncoder is the name of the generated encoder function.
		SecureRequestEncoder string
		// RequestEncoder is the name of the encoder function
		// originally generated by goa.
		RequestEncoder string
		// PayloadType is the Go type name of the secured method payload.
		PayloadType string
		// Schemes contains the security scheme data needed to render the auth
		// code.
		Schemes []*seccodegen.SchemeData
	}
)

// Register the plugin HTTP Generator function.
func init() {
	codegen.RegisterPlugin(design.PluginName, "gen", Generate)
	codegen.RegisterPlugin(design.PluginName, "example", Example)
}

// Generate produces HTTP decoders and encoders that initialize the
// security attributes.
func Generate(genpkg string, roots []eval.Root, files []*codegen.File) ([]*codegen.File, error) {
	for _, root := range roots {
		switch r := root.(type) {
		case *httpdesign.RootExpr:
			for _, f := range files {
				SecureRequestDecoders(f)
				SecureRequestEncoders(f)
				OpenAPIV2(r, f)
			}
		}
	}
	return files, nil
}

// Example modified the generated main function so that the secured endpoints
// context gets initialized with the security requirements.
func Example(genpkg string, roots []eval.Root, files []*codegen.File) ([]*codegen.File, error) {
	var (
		data    = make(map[string]*seccodegen.ServiceData)
		svcData []*ServiceData
	)
	for _, root := range roots {
		switch r := root.(type) {
		case *goadesign.RootExpr:
			for _, s := range r.Services {
				data[s.Name] = seccodegen.Data.Get(s.Name)
			}
		case *httpdesign.RootExpr:
			for _, s := range r.HTTPServices {
				sd := httpcodegen.HTTPServices.Get(s.Name())
				svcData = append(svcData, &ServiceData{ServiceData: sd})
			}
		}
	}
	for _, svc := range svcData {
		if d, ok := data[svc.ServiceData.Service.Name]; ok {
			svc.Schemes = d.Schemes
		}
	}
	for _, f := range files {
		for _, s := range f.Section("service-main") {
			s.Source = strings.Replace(
				s.Source,
				"{{- range .Services }}",
				"{{- range $svc := .Services }}",
				-1,
			)
			s.Source = strings.Replace(
				s.Source,
				"{{ .Service.PkgName }}.NewEndpoints({{ .Service.VarName }}Svc)",
				"{{ .Service.PkgName }}.NewSecureEndpoints({{ .Service.VarName }}Svc{{ range .Schemes }}, {{ $.APIPkg }}.{{ $svc.Service.StructName }}Auth{{ .Type }}Fn{{ end }})",
				1,
			)
			sData := s.Data.(map[string]interface{})
			sData["Services"] = svcData
		}
	}
	return files, nil
}

// SecureRequestDecoders initializes the security attributes for HTTP request decoders.
func SecureRequestDecoders(f *codegen.File) {
	secureDecoder(f)
	for _, s := range f.Section("response-encoder") {
		data := s.Data.(*httpcodegen.EndpointData)
		md := seccodegen.Data.Get(data.ServiceName).MethodData(data.Method.Name)
		if len(md.Requirements) > 0 {
			f.SectionTemplates = append(f.SectionTemplates, &codegen.SectionTemplate{
				Name:   "secure-request-decoder",
				Source: authDecoderT,
				Data: &SecureDecoderData{
					SecureRequestDecoder: "Secure" + data.RequestDecoder,
					RequestDecoder:       data.RequestDecoder,
					PayloadType:          data.Payload.Ref,
					Schemes:              md.Schemes,
					ServiceName:          data.ServiceName,
					MethodName:           data.Method.Name,
				},
				FuncMap: codegen.TemplateFuncs(),
			})
		}
	}
}

// SecureRequestEncoders initializes the security attributes for HTTP encoders.
func SecureRequestEncoders(f *codegen.File) {
	secureEncoder(f)
	for _, s := range f.Section("request-builder") {
		data := s.Data.(*httpcodegen.EndpointData)
		md := seccodegen.Data.Get(data.ServiceName).MethodData(data.Method.Name)
		if len(md.Requirements) > 0 {
			funcs := codegen.TemplateFuncs()
			funcs["querySchemes"] = querySchemes
			f.SectionTemplates = append(f.SectionTemplates, &codegen.SectionTemplate{
				Name:   "secure-request-encoder",
				Source: authEncoderT,
				Data: &SecureEncoderData{
					SecureRequestEncoder: "Secure" + data.RequestEncoder,
					RequestEncoder:       data.RequestEncoder,
					PayloadType:          data.Payload.Ref,
					Schemes:              md.Schemes,
					ServiceName:          data.ServiceName,
					MethodName:           data.Method.Name,
				},
				FuncMap: funcs,
			})
		}
	}
}

// OpenAPIV2 adds the security requirements for the HTTP endpoints.
func OpenAPIV2(r *httpdesign.RootExpr, f *codegen.File) {
	for _, s := range f.Section("openapi") {
		spec := s.Data.(*openapi.V2)
		for _, svc := range r.HTTPServices {
			for _, e := range svc.HTTPEndpoints {
				reqs := design.Requirements(e.MethodExpr)
				for _, route := range e.Routes {
					var (
						p  *openapi.Path
						op *openapi.Operation
					)
					for path, v := range spec.Paths {
						for _, rPath := range route.FullPaths() {
							if rPath == path {
								p = v.(*openapi.Path)
								break
							}
						}
					}
					if p == nil {
						continue
					}
					switch route.Method {
					case "GET":
						op = p.Get
					case "PUT":
						op = p.Put
					case "POST":
						op = p.Post
					case "DELETE":
						op = p.Delete
					case "OPTIONS":
						op = p.Options
					case "HEAD":
						op = p.Head
					case "PATCH":
						op = p.Patch
					}
					applySecurity(op, reqs)
				}
			}
		}
		s.Data = spec
	}
}

// applySecurity applies the security requirements to the openapi V2 operation.
func applySecurity(op *openapi.Operation, reqs []*design.EndpointSecurityExpr) {
	if len(reqs) == 0 {
		return
	}
	requirements := make([]map[string][]string, len(reqs))
	for i, req := range reqs {
		requirement := make(map[string][]string)
		for _, s := range req.Schemes {
			requirement[s.SchemeName] = []string{}
			switch s.Kind {
			case design.OAuth2Kind:
				for _, scope := range req.Scopes {
					requirement[s.SchemeName] = append(requirement[s.SchemeName], scope)
				}
			case design.JWTKind:
				lines := make([]string, 0, len(req.Scopes))
				for _, scope := range req.Scopes {
					lines = append(lines, fmt.Sprintf("  * `%s`", scope))
				}
				if op.Description != "" {
					op.Description += "\n"
				}
				op.Description += fmt.Sprintf("\nRequired security scopes:\n%s", strings.Join(lines, "\n"))
			}
		}
		requirements[i] = requirement
	}
	op.Security = requirements
}

var (
	// decoderRegexp matches occurrences of "{{ .RequestDecoder }}" in section template code.
	decoderRegexp = regexp.MustCompile(`({{.?\.RequestDecoder.?}})`)

	// encoderRegexp matches occurrences of "{{ .RequestEncoder }}" in section template code.
	encoderRegexp = regexp.MustCompile(`({{.?\.RequestEncoder.?}})`)
)

// secureDecoder prefixes all occurrences of "{{ .RequestDecoder }}" with "Secure"
// (except in server decode files) if the corresponding endpoint is secured.
func secureDecoder(f *codegen.File) {
	for _, s := range f.SectionTemplates {
		if s.Name == "request-decoder" {
			continue
		}
		if decoderRegexp.MatchString(s.Source) {
			if data, ok := s.Data.(*httpcodegen.EndpointData); ok {
				svc := seccodegen.Data.Get(data.ServiceName)
				if len(svc.MethodData(data.Method.Name).Requirements) > 0 {
					s.Source = decoderRegexp.ReplaceAllString(s.Source, "Secure${1}")
				}
			}
		}
	}
}

// secureEncoder prefixes all occurrences of "{{ .RequestDecoder }}" with "Secure"
// (except in client encode files) if the corresponding endpoint is secured.
func secureEncoder(f *codegen.File) {
	for _, s := range f.SectionTemplates {
		if s.Name == "request-encoder" {
			continue
		}
		if data, ok := s.Data.(*httpcodegen.EndpointData); ok {
			svc := seccodegen.Data.Get(data.ServiceName)
			if len(svc.MethodData(data.Method.Name).Requirements) > 0 {
				s.Source = encoderRegexp.ReplaceAllString(s.Source, "Secure${1}")
			}
		}
	}
}

func querySchemes(schemes []*seccodegen.SchemeData) bool {
	for _, s := range schemes {
		if s.Scheme.In == "query" {
			return true
		}
	}
	return false
}

// input: SecureDecoderData
const authDecoderT = `{{ printf "%s returns a decoder for requests sent to the %s %s endpoint that is security scheme aware." .SecureRequestDecoder .ServiceName .MethodName | comment }}
func {{ .SecureRequestDecoder }}(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (interface{}, error) {
	rawDecoder := {{ .RequestDecoder }}(mux, decoder)
	return func(r *http.Request) (interface{}, error) {
		p, err := rawDecoder(r)
		if err != nil {
			return nil, err
		}
		payload := p.({{ .PayloadType }})
{{- range .Schemes }}

	{{- if eq .Scheme.Kind 2 }}{{/* BasicAuth */}}
		user, pass, ok := r.BasicAuth()
		if !ok {
			return p, nil
		}
		payload.{{ .UsernameField }} = {{ if .UsernamePointer }}&{{ end }}user
		payload.{{ .PasswordField }} = {{ if .PasswordPointer }}&{{ end }}pass

	{{- else if eq .Scheme.Kind 3 }}{{/* APIKey */}}
		{{- if eq .Scheme.In "query" }}
		key := r.URL.Query().Get({{ printf "%q" .Scheme.Name }})
		if key == "" {
			return p, nil
		}
		payload.{{ .CredField }} = {{ if .CredPointer }}&{{ end }}key
		{{- else }}
		key := r.Header.Get({{ printf "%q" .Scheme.Name }})
		if key == "" {
			return p, nil
		}
		payload.{{ .CredField }} = {{ if .CredPointer }}&{{ end }}key
		{{- end }}

	{{- else }}{{/* OAuth2 and JWT */}}
		{{- if eq .Scheme.In "query" }}
		token{{ .Scheme.Type }} := r.URL.Query().Get({{ printf "%q" .Scheme.Name }})
		if token{{ .Scheme.Type }} == "" {
			return p, nil
		}
		payload.{{ .CredField }} = {{ if .CredPointer }}&{{ end }}token{{ .Scheme.Type }}
		{{- else }}
		h{{ .Scheme.Type }} := r.Header.Get({{ printf "%q" .Scheme.Name }})
		if h{{ .Scheme.Type }} == "" {
			return p, nil
		}
		token{{ .Scheme.Type }} := strings.TrimPrefix(h{{ .Scheme.Type }}, "Bearer ")
		payload.{{ .CredField }} = {{ if .CredPointer }}&{{ end }}token{{ .Scheme.Type }}
		{{- end }}

	{{- end }}

{{- end }}
		return payload, nil
	}
}
`

// input: SecureEncoderData
const authEncoderT = `{{ printf "%s returns an encoder for requests sent to the %s %s endpoint that is security scheme aware." .SecureRequestEncoder .ServiceName .MethodName | comment }}
func {{ .SecureRequestEncoder }}(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, interface{}) error {
	rawEncoder := {{ .RequestEncoder }}(encoder)
	return func(req *http.Request, v interface{}) error {
		if err := rawEncoder(req, v); err != nil {
			return err
		}
		payload := v.({{ .PayloadType }})
		{{- if querySchemes .Schemes }}
		values := req.URL.Query()
		{{- end }}
{{- range .Schemes }}

	{{- if eq .Scheme.Kind 2 }}{{/* BasicAuth */}}
		req.SetBasicAuth({{ if .UsernamePointer }}*{{ end }}payload.{{ .UsernameField }}, {{ if .PasswordPointer }}*{{ end }}payload.{{ .PasswordField }})

	{{- else if eq .Scheme.Kind 3 }}{{/* APIKey */}}
		{{- if eq .Scheme.In "query" }}
		values.Add({{ printf "%q" .Scheme.Name }}, {{ if .CredPointer }}*{{ end }}payload.{{ .CredField }})
		{{- else }}
		req.Header.Add({{ printf "%q" .Scheme.Name }}, {{ if .CredPointer }}*{{ end }}payload.{{ .CredField }})
		{{- end }}

	{{- else }}{{/* OAuth2 and JWT */}}
		{{- if eq .Scheme.In "query" }}
		values.Add({{ printf "%q" .Scheme.Name }}, {{ if .CredPointer }}*{{ end }}payload.{{ .CredField }})
		{{- else }}
		req.Header.Add({{ printf "%q" .Scheme.Name }}, fmt.Sprintf("Bearer %s", {{ if .CredPointer }}*{{ end }}payload.{{ .CredField }}))
		{{- end }}

	{{- end }}

{{- end }}
		{{- if querySchemes .Schemes }}
		req.URL.RawQuery = values.Encode()
		{{- end }}
		return nil
	}
}
`
