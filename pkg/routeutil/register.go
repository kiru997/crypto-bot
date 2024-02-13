package routeutil

import (
	"net/http"
	"path"
	"sync"

	"github.com/gin-gonic/gin"

	"example.com/greetings/pkg/swag"
	"example.com/greetings/pkg/swag/endpoint"
	"example.com/greetings/pkg/swag/swagger"
	"example.com/greetings/pkg/transportutil"
	// "otanics.gitlab.com/otanicsserver/backend/pkg/apperror"
)

var API = swag.New(
	swag.Title("Diff backend"),
	// swag.Version(version.Version),
	// swag.ContactEmail("support@otanics.com"),
	swag.BasePath("/"),
	swag.Schemes("http", "https"),
	swag.SecurityScheme("Authorization", swagger.APIKeySecurity("Authorization", "header")),
	swag.SecurityScheme("DeviceID", swagger.APIKeySecurity("DeviceID", "header")),
	// swag.Description(fmt.Sprintf("GoVersion: %s <br/> GitHash: %s", version.GoVersion, version.GitHash)+apperror.ExposeDocs()),
)

var once sync.Once

func ServingDocs(c *gin.Context) {
	API.Host = c.Request.Host
	API.Tags = []swagger.Tag{}

	once.Do(func() {
		// if version.Version != "v-.-.-" {
		// 	API.Schemes = []string{"https", "http"}
		// }
		API.Schemes = []string{"https", "http"}
		// for k, d := range API.Definitions {
		// 	for k2, p := range d.Properties {
		// 		pv := reflect.New(p.GoType)
		// 		switch v := pv.Interface().(type) {
		// 		case enum.Enum:
		// 			for _, e := range v.EnumDescriptions() {
		// 				p.Enum = append(p.Enum, e)
		// 			}
		// 			p.Format = "string"

		// 			API.Definitions[k].Properties[k2] = p
		// 		}
		// 	}
		// }
	})

	c.JSON(http.StatusOK, API)
}

type RegisterOption int

const (
	RegisterOptionSkipAuth RegisterOption = iota + 1
	RegisterOptionDeprecated
)

var l sync.RWMutex

func AddEndpoint(
	g *gin.RouterGroup,
	relativePath string,
	handler gin.HandlerFunc,
	req, resp interface{},
	description string,
	opts ...RegisterOption,
) {
	l.Lock()
	defer l.Unlock()

	path := joinPaths(g.BasePath(), relativePath)

	swaggerOpts := []endpoint.Option{
		endpoint.Tags(g.BasePath()),
		endpoint.Security("Authorization", []string{}...),
		endpoint.Body(req, "Request Body", true),
		endpoint.Response(http.StatusOK, resp, "Response"),
	}

	for _, o := range opts {
		switch o {
		case RegisterOptionSkipAuth:
			transportutil.RegisterPublicEndpoint(path)
		case RegisterOptionDeprecated:
			swaggerOpts = append(swaggerOpts, endpoint.Deprecated(true))
		}
	}

	API.AddEndpoint(endpoint.New(
		http.MethodPost,
		path,
		description,
		swaggerOpts...,
	))

	g.POST(relativePath, handler)
}

func AddCustomEndpoint(
	g *gin.RouterGroup,
	relativePath string,
	handler gin.HandlerFunc,
	description string,
	rOpt RegisterOption,
	opts ...endpoint.Option,
) {
	l.Lock()
	defer l.Unlock()

	opts = append(opts, endpoint.Security("Authorization", []string{}...))
	path := joinPaths(g.BasePath(), relativePath)
	if rOpt == RegisterOptionSkipAuth {
		transportutil.RegisterPublicEndpoint(path)
	}

	API.AddEndpoint(endpoint.New(
		http.MethodPost,
		path,
		description,
		opts...,
	))

	g.POST(relativePath, handler)
}

func joinPaths(absolutePath, relativePath string) string {
	if relativePath == "" {
		return absolutePath
	}

	finalPath := path.Join(absolutePath, relativePath)
	appendSlash := lastChar(relativePath) == '/' && lastChar(finalPath) != '/'
	if appendSlash {
		return finalPath + "/"
	}
	return finalPath
}

func lastChar(str string) uint8 {
	if str == "" {
		panic("The length of the string can't be 0")
	}
	return str[len(str)-1]
}

func Parameter(p swagger.Parameter) endpoint.Option {
	return func(b *endpoint.Builder) {
		if b.Endpoint.Parameters == nil {
			b.Endpoint.Parameters = []swagger.Parameter{}
		}

		b.Endpoint.Parameters = append(b.Endpoint.Parameters, p)
	}
}
