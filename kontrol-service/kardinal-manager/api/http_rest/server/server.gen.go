// Package kardinal_manager_server_rest_server provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen/v2 version v2.2.1-0.20240604070534-2f0ff757704b DO NOT EDIT.
package kardinal_manager_server_rest_server

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
	strictecho "github.com/oapi-codegen/runtime/strictmiddleware/echo"
	. "kardinal.kontrol/kardinal-manager/api/http_rest/types"
)

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// Delete virtual service
	// (DELETE /virtual-services)
	DeleteVirtualServices(ctx echo.Context) error
	// List virtual services
	// (GET /virtual-services)
	GetVirtualServices(ctx echo.Context) error
	// Create virtual service
	// (POST /virtual-services)
	PostVirtualServices(ctx echo.Context) error
}

// ServerInterfaceWrapper converts echo contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler ServerInterface
}

// DeleteVirtualServices converts echo context to params.
func (w *ServerInterfaceWrapper) DeleteVirtualServices(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshaled arguments
	err = w.Handler.DeleteVirtualServices(ctx)
	return err
}

// GetVirtualServices converts echo context to params.
func (w *ServerInterfaceWrapper) GetVirtualServices(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshaled arguments
	err = w.Handler.GetVirtualServices(ctx)
	return err
}

// PostVirtualServices converts echo context to params.
func (w *ServerInterfaceWrapper) PostVirtualServices(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshaled arguments
	err = w.Handler.PostVirtualServices(ctx)
	return err
}

// This is a simple interface which specifies echo.Route addition functions which
// are present on both echo.Echo and echo.Group, since we want to allow using
// either of them for path registration
type EchoRouter interface {
	CONNECT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	HEAD(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	OPTIONS(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PATCH(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	TRACE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
}

// RegisterHandlers adds each server route to the EchoRouter.
func RegisterHandlers(router EchoRouter, si ServerInterface) {
	RegisterHandlersWithBaseURL(router, si, "")
}

// Registers handlers, and prepends BaseURL to the paths, so that the paths
// can be served under a prefix.
func RegisterHandlersWithBaseURL(router EchoRouter, si ServerInterface, baseURL string) {

	wrapper := ServerInterfaceWrapper{
		Handler: si,
	}

	router.DELETE(baseURL+"/virtual-services", wrapper.DeleteVirtualServices)
	router.GET(baseURL+"/virtual-services", wrapper.GetVirtualServices)
	router.POST(baseURL+"/virtual-services", wrapper.PostVirtualServices)

}

type NotOkJSONResponse ResponseInfo

type DeleteVirtualServicesRequestObject struct {
}

type DeleteVirtualServicesResponseObject interface {
	VisitDeleteVirtualServicesResponse(w http.ResponseWriter) error
}

type DeleteVirtualServices200JSONResponse VirtualService

func (response DeleteVirtualServices200JSONResponse) VisitDeleteVirtualServicesResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)

	return json.NewEncoder(w).Encode(response)
}

type DeleteVirtualServicesdefaultJSONResponse struct {
	Body       ResponseInfo
	StatusCode int
}

func (response DeleteVirtualServicesdefaultJSONResponse) VisitDeleteVirtualServicesResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(response.StatusCode)

	return json.NewEncoder(w).Encode(response.Body)
}

type GetVirtualServicesRequestObject struct {
}

type GetVirtualServicesResponseObject interface {
	VisitGetVirtualServicesResponse(w http.ResponseWriter) error
}

type GetVirtualServices200JSONResponse map[string]VirtualService

func (response GetVirtualServices200JSONResponse) VisitGetVirtualServicesResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)

	return json.NewEncoder(w).Encode(response)
}

type GetVirtualServicesdefaultJSONResponse struct {
	Body       ResponseInfo
	StatusCode int
}

func (response GetVirtualServicesdefaultJSONResponse) VisitGetVirtualServicesResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(response.StatusCode)

	return json.NewEncoder(w).Encode(response.Body)
}

type PostVirtualServicesRequestObject struct {
	Body *PostVirtualServicesJSONRequestBody
}

type PostVirtualServicesResponseObject interface {
	VisitPostVirtualServicesResponse(w http.ResponseWriter) error
}

type PostVirtualServices200JSONResponse VirtualService

func (response PostVirtualServices200JSONResponse) VisitPostVirtualServicesResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)

	return json.NewEncoder(w).Encode(response)
}

type PostVirtualServicesdefaultJSONResponse struct {
	Body       ResponseInfo
	StatusCode int
}

func (response PostVirtualServicesdefaultJSONResponse) VisitPostVirtualServicesResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(response.StatusCode)

	return json.NewEncoder(w).Encode(response.Body)
}

// StrictServerInterface represents all server handlers.
type StrictServerInterface interface {
	// Delete virtual service
	// (DELETE /virtual-services)
	DeleteVirtualServices(ctx context.Context, request DeleteVirtualServicesRequestObject) (DeleteVirtualServicesResponseObject, error)
	// List virtual services
	// (GET /virtual-services)
	GetVirtualServices(ctx context.Context, request GetVirtualServicesRequestObject) (GetVirtualServicesResponseObject, error)
	// Create virtual service
	// (POST /virtual-services)
	PostVirtualServices(ctx context.Context, request PostVirtualServicesRequestObject) (PostVirtualServicesResponseObject, error)
}

type StrictHandlerFunc = strictecho.StrictEchoHandlerFunc
type StrictMiddlewareFunc = strictecho.StrictEchoMiddlewareFunc

func NewStrictHandler(ssi StrictServerInterface, middlewares []StrictMiddlewareFunc) ServerInterface {
	return &strictHandler{ssi: ssi, middlewares: middlewares}
}

type strictHandler struct {
	ssi         StrictServerInterface
	middlewares []StrictMiddlewareFunc
}

// DeleteVirtualServices operation middleware
func (sh *strictHandler) DeleteVirtualServices(ctx echo.Context) error {
	var request DeleteVirtualServicesRequestObject

	handler := func(ctx echo.Context, request interface{}) (interface{}, error) {
		return sh.ssi.DeleteVirtualServices(ctx.Request().Context(), request.(DeleteVirtualServicesRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "DeleteVirtualServices")
	}

	response, err := handler(ctx, request)

	if err != nil {
		return err
	} else if validResponse, ok := response.(DeleteVirtualServicesResponseObject); ok {
		return validResponse.VisitDeleteVirtualServicesResponse(ctx.Response())
	} else if response != nil {
		return fmt.Errorf("unexpected response type: %T", response)
	}
	return nil
}

// GetVirtualServices operation middleware
func (sh *strictHandler) GetVirtualServices(ctx echo.Context) error {
	var request GetVirtualServicesRequestObject

	handler := func(ctx echo.Context, request interface{}) (interface{}, error) {
		return sh.ssi.GetVirtualServices(ctx.Request().Context(), request.(GetVirtualServicesRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "GetVirtualServices")
	}

	response, err := handler(ctx, request)

	if err != nil {
		return err
	} else if validResponse, ok := response.(GetVirtualServicesResponseObject); ok {
		return validResponse.VisitGetVirtualServicesResponse(ctx.Response())
	} else if response != nil {
		return fmt.Errorf("unexpected response type: %T", response)
	}
	return nil
}

// PostVirtualServices operation middleware
func (sh *strictHandler) PostVirtualServices(ctx echo.Context) error {
	var request PostVirtualServicesRequestObject

	var body PostVirtualServicesJSONRequestBody
	if err := ctx.Bind(&body); err != nil {
		return err
	}
	request.Body = &body

	handler := func(ctx echo.Context, request interface{}) (interface{}, error) {
		return sh.ssi.PostVirtualServices(ctx.Request().Context(), request.(PostVirtualServicesRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "PostVirtualServices")
	}

	response, err := handler(ctx, request)

	if err != nil {
		return err
	} else if validResponse, ok := response.(PostVirtualServicesResponseObject); ok {
		return validResponse.VisitPostVirtualServicesResponse(ctx.Response())
	} else if response != nil {
		return fmt.Errorf("unexpected response type: %T", response)
	}
	return nil
}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"H4sIAAAAAAAC/9yUz07bQBDGX2U1rcTFiVO4+UYLRRFtQAmlh4rD1h47C/buMjMORZHfvVrbJORP1SLB",
	"padsdsfffvPbmVlC6irvLFphSJZAyN5ZxvbPxMnFXVikzgpaCUvtfWlSLcbZ+JadDXuczrHSYfWeMIcE",
	"3sVr1bg75XjaS49t7qBpmggy5JSMD1qQwDeLvzymgplCIkcQQvqPg/bG98kSPDmPJKbzmroMw2/uqNIC",
	"CdTGytEhRCCPHiEBYwULJGgiqJBZF214f8hCxhbhrNv4t0yuQmxwSXhfG8IMkh+dwPqOqHN2s/Lhft5i",
	"KuGqDZlkCWjrKiicTqcXU4hgPPl8ARF8P55OxpOzZxJrt9eGpNblDGlhUtylYnW1L80ty23UrsMQZnrY",
	"m081PZ1d5XWpji/Hij2mJu9LQuWOlMxRnWvKjNWl+qqtLpDUs7qJlJEDVjVjpsQpXYsbFGiRtKBKS4NW",
	"1Ozk/ICVtplipAXSgE2GqkUZgRgpg9GdS575gggWSNz5HQ0/DEeBl/NotTeQwNFwNDyCCLyWeYsqXnQs",
	"B9zB5C7tEqUlGKi27scZJHDS7m/SZ4g22+dwNHq15tl66D3tM6vTFJlD+k8uoA3KdV3Kn/RXhuOu2due",
	"q6tK0+MqTdWTUT2Z8AK64FA5O8xumggKlF1iZyivjEtnmQlHurzcqPmXYNxb82/P9Yth2abKf8PqHe/h",
	"eul4L9j7Glk+uuzxTUtwPUWEamz+uwb4RKhf2gCtQju0QsT26FzNrKqfWV2oMnY9M09OryGCmkpIYC7i",
	"OYnjh4eH4V0fMMxwET/9GfRCcZhrzU3zOwAA//+rRL0K0gcAAA==",
}

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %w", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}

	return buf.Bytes(), nil
}

var rawSpec = decodeSpecCached()

// a naive cached of a decoded swagger spec
func decodeSpecCached() func() ([]byte, error) {
	data, err := decodeSpec()
	return func() ([]byte, error) {
		return data, err
	}
}

// Constructs a synthetic filesystem for resolving external references when loading openapi specifications.
func PathToRawSpec(pathToFile string) map[string]func() ([]byte, error) {
	res := make(map[string]func() ([]byte, error))
	if len(pathToFile) > 0 {
		res[pathToFile] = rawSpec
	}

	return res
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file. The external references of Swagger specification are resolved.
// The logic of resolving external references is tightly connected to "import-mapping" feature.
// Externally referenced files must be embedded in the corresponding golang packages.
// Urls can be supported but this task was out of the scope.
func GetSwagger() (swagger *openapi3.T, err error) {
	resolvePath := PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		pathToFile := url.String()
		pathToFile = path.Clean(pathToFile)
		getSpec, ok := resolvePath[pathToFile]
		if !ok {
			err1 := fmt.Errorf("path not found: %s", pathToFile)
			return nil, err1
		}
		return getSpec()
	}
	var specData []byte
	specData, err = rawSpec()
	if err != nil {
		return
	}
	swagger, err = loader.LoadFromData(specData)
	if err != nil {
		return
	}
	return
}
