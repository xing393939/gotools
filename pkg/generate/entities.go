package generate

import (
	"bytes"
	"encoding/json"
	"google.golang.org/protobuf/types/descriptorpb"
	"reflect"
)

var swaggerMapTypes = map[string]reflect.Kind{
	"string":   reflect.String,
	"*string":  reflect.String,
	"int":      reflect.Int,
	"*int":     reflect.Int,
	"uint":     reflect.Uint,
	"*uint":    reflect.Uint,
	"int8":     reflect.Int8,
	"*int8":    reflect.Int8,
	"uint8":    reflect.Uint8,
	"*uint8":   reflect.Uint8,
	"int16":    reflect.Int16,
	"*int16":   reflect.Int16,
	"uint16":   reflect.Uint16,
	"*uint16":  reflect.Uint16,
	"int32":    reflect.Int,
	"*int32":   reflect.Int,
	"uint32":   reflect.Int,
	"*uint32":  reflect.Int,
	"uint64":   reflect.Int64,
	"*uint64":  reflect.Int64,
	"int64":    reflect.Int64,
	"*int64":   reflect.Int64,
	"[]string": reflect.Slice,
	"[]int":    reflect.Slice,
	"[]int64":  reflect.Slice,
	"[]int32":  reflect.Slice,
	"[]uint32": reflect.Slice,
	"[]uint64": reflect.Slice,
	"bool":     reflect.Bool,
	"*bool":    reflect.Bool,
	"struct":   reflect.Struct,
	"*struct":  reflect.Struct,
	"float32":  reflect.Float32,
	"*float32": reflect.Float32,
	"float64":  reflect.Float64,
	"*float64": reflect.Float64,
}

// http://swagger.io/specification/#infoObject
type swaggerInfoObject struct {
	Title          string `json:"title"`
	Description    string `json:"description,omitempty"`
	TermsOfService string `json:"termsOfService,omitempty"`
	Version        string `json:"version"`

	Contact *swaggerContactObject `json:"contact,omitempty"`
	License *swaggerLicenseObject `json:"license,omitempty"`
}

// http://swagger.io/specification/#contactObject
type swaggerContactObject struct {
	Name  string `json:"name,omitempty"`
	URL   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
}

// http://swagger.io/specification/#licenseObject
type swaggerLicenseObject struct {
	Name string `json:"name,omitempty"`
	URL  string `json:"url,omitempty"`
}

// http://swagger.io/specification/#externalDocumentationObject
type swaggerExternalDocumentationObject struct {
	Description string `json:"description,omitempty"`
	URL         string `json:"url,omitempty"`
}

// http://swagger.io/specification/#swaggerObject
type swaggerObject struct {
	Swagger             string                              `json:"swagger"`
	Info                swaggerInfoObject                   `json:"info"`
	Host                string                              `json:"host,omitempty"`
	BasePath            string                              `json:"basePath,omitempty"`
	Schemes             []string                            `json:"schemes"`
	Consumes            []string                            `json:"consumes"`
	Produces            []string                            `json:"produces"`
	Paths               swaggerPathsObject                  `json:"paths"`
	Definitions         swaggerDefinitionsObject            `json:"definitions"`
	StreamDefinitions   swaggerDefinitionsObject            `json:"x-stream-definitions,omitempty"`
	SecurityDefinitions swaggerSecurityDefinitionsObject    `json:"securityDefinitions,omitempty"`
	Security            []swaggerSecurityRequirementObject  `json:"security,omitempty"`
	ExternalDocs        *swaggerExternalDocumentationObject `json:"externalDocs,omitempty"`
}

// http://swagger.io/specification/#securityDefinitionsObject
type swaggerSecurityDefinitionsObject map[string]swaggerSecuritySchemeObject

// http://swagger.io/specification/#securitySchemeObject
type swaggerSecuritySchemeObject struct {
	Type             string              `json:"type"`
	Description      string              `json:"description,omitempty"`
	Name             string              `json:"name,omitempty"`
	In               string              `json:"in,omitempty"`
	Flow             string              `json:"flow,omitempty"`
	AuthorizationURL string              `json:"authorizationUrl,omitempty"`
	TokenURL         string              `json:"tokenUrl,omitempty"`
	Scopes           swaggerScopesObject `json:"scopes,omitempty"`
}

// http://swagger.io/specification/#scopesObject
type swaggerScopesObject map[string]string

// http://swagger.io/specification/#securityRequirementObject
type swaggerSecurityRequirementObject map[string][]string

// http://swagger.io/specification/#pathsObject
type swaggerPathsObject map[string]swaggerPathItemObject

// http://swagger.io/specification/#pathItemObject
type swaggerPathItemObject struct {
	Get    *swaggerOperationObject `json:"get,omitempty"`
	Delete *swaggerOperationObject `json:"delete,omitempty"`
	Post   *swaggerOperationObject `json:"post,omitempty"`
	Put    *swaggerOperationObject `json:"put,omitempty"`
	Patch  *swaggerOperationObject `json:"patch,omitempty"`
}

// http://swagger.io/specification/#operationObject
type swaggerOperationObject struct {
	Summary     string                  `json:"summary,omitempty"`
	Description string                  `json:"description,omitempty"`
	OperationID string                  `json:"operationId"`
	Responses   swaggerResponsesObject  `json:"responses"`
	Parameters  swaggerParametersObject `json:"parameters,omitempty"`
	RequestBody *struct {
		Content swaggerContentObject `json:"content,omitempty"`
	} `json:"requestBody,omitempty"`
	Tags         []string                            `json:"tags,omitempty"`
	Deprecated   bool                                `json:"deprecated,omitempty"`
	Consumes     []string                            `json:"consumes,omitempty"`
	Security     *[]swaggerSecurityRequirementObject `json:"security,omitempty"`
	ExternalDocs *swaggerExternalDocumentationObject `json:"externalDocs,omitempty"`
}

type (
	swaggerParametersObject []swaggerParameterObject
	swaggerContentObject    map[string]swaggerParametersObject
)

// http://swagger.io/specification/#parameterObject
type swaggerParameterObject struct {
	Name             string              `json:"name"`
	Description      string              `json:"description,omitempty"`
	In               string              `json:"in,omitempty"`
	Required         bool                `json:"required"`
	Type             string              `json:"type,omitempty"`
	Format           string              `json:"format,omitempty"`
	Items            *swaggerItemsObject `json:"items,omitempty"`
	Enum             []string            `json:"enum,omitempty"`
	CollectionFormat string              `json:"collectionFormat,omitempty"`
	Default          string              `json:"default,omitempty"`
	MinItems         *int                `json:"minItems,omitempty"`
	Example          string              `json:"example,omitempty"`

	// Or you can explicitly refer to another type. If this is defined all
	// other fields should be empty
	Schema *swaggerSchemaObject `json:"schema,omitempty"`
}

// core part of schema, which is common to itemsObject and schemaObject.
// http://swagger.io/specification/#itemsObject
type schemaCore struct {
	Type    string `json:"type,omitempty"`
	Format  string `json:"format,omitempty"`
	Ref     string `json:"$ref,omitempty"`
	Example string `json:"example,omitempty"`

	Items *swaggerItemsObject `json:"items,omitempty"`
	// If the item is an enumeration include a list of all the *NAMES* of the
	// enum values.  I'm not sure how well this will work but assuming all enums
	// start from 0 index it will be great. I don't think that is a good assumption.
	Enum    []string `json:"enum,omitempty"`
	Default string   `json:"default,omitempty"`
}

type swaggerItemsObject schemaCore

// http://swagger.io/specification/#responsesObject
type swaggerResponsesObject map[string]swaggerResponseObject

// http://swagger.io/specification/#responseObject
type swaggerResponseObject struct {
	Description string              `json:"description"`
	Schema      swaggerSchemaObject `json:"schema"`
}

type keyVal struct {
	Key   string
	Value interface{}
}

type swaggerSchemaObjectProperties []keyVal

func (op swaggerSchemaObjectProperties) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString("{")
	for i, kv := range op {
		if i != 0 {
			buf.WriteString(",")
		}
		key, err := json.Marshal(kv.Key)
		if err != nil {
			return nil, err
		}
		buf.Write(key)
		buf.WriteString(":")
		val, err := json.Marshal(kv.Value)
		if err != nil {
			return nil, err
		}
		buf.Write(val)
	}

	buf.WriteString("}")
	return buf.Bytes(), nil
}

// http://swagger.io/specification/#schemaObject
type swaggerSchemaObject struct {
	schemaCore
	// Properties can be recursively defined
	Properties           *swaggerSchemaObjectProperties `json:"properties,omitempty"`
	AdditionalProperties *swaggerSchemaObject           `json:"additionalProperties,omitempty"`

	Description string `json:"description,omitempty"`
	Title       string `json:"title,omitempty"`

	ExternalDocs *swaggerExternalDocumentationObject `json:"externalDocs,omitempty"`

	ReadOnly         bool     `json:"readOnly,omitempty"`
	MultipleOf       float64  `json:"multipleOf,omitempty"`
	Maximum          float64  `json:"maximum,omitempty"`
	ExclusiveMaximum bool     `json:"exclusiveMaximum,omitempty"`
	Minimum          float64  `json:"minimum,omitempty"`
	ExclusiveMinimum bool     `json:"exclusiveMinimum,omitempty"`
	MaxLength        uint64   `json:"maxLength,omitempty"`
	MinLength        uint64   `json:"minLength,omitempty"`
	Pattern          string   `json:"pattern,omitempty"`
	MaxItems         uint64   `json:"maxItems,omitempty"`
	MinItems         uint64   `json:"minItems,omitempty"`
	UniqueItems      bool     `json:"uniqueItems,omitempty"`
	MaxProperties    uint64   `json:"maxProperties,omitempty"`
	MinProperties    uint64   `json:"minProperties,omitempty"`
	Required         []string `json:"required,omitempty"`
}

// http://swagger.io/specification/#definitionsObject
type swaggerDefinitionsObject map[string]swaggerSchemaObject

// Internal type mapping from FQMN to descriptor.Message. Used as a set by the
// findServiceMessages function.
type messageMap map[string]*Message

// Internal type mapping from FQEN to descriptor.Enum. Used as a set by the
// findServiceMessages function.
type enumMap map[string]*Enum

type Message struct {
	*descriptorpb.DescriptorProto
	// File is the file where the message is defined.
	File *File
	// Outers is a list of outer messages if this message is a nested type.
	Outers []string
	// Fields is a list of message fields.
	Fields []*Field
	// Index is proto path index of this message in File.
	Index int
	// ForcePrefixedName when set to true, prefixes a type with a package prefix.
	ForcePrefixedName bool
}

type Enum struct {
	*descriptorpb.EnumDescriptorProto
	// File is the file where the enum is defined
	File *File
	// Outers is a list of outer messages if this enum is a nested type.
	Outers []string
	// Index is a enum index value.
	Index int
	// ForcePrefixedName when set to true, prefixes a type with a package prefix.
	ForcePrefixedName bool
}

type File struct {
	*descriptorpb.FileDescriptorProto
	// GoPkg is the go package of the go file generated from this file.
	GoPkg GoPackage
	// GeneratedFilenamePrefix is used to construct filenames for generated
	// files associated with this source file.
	//
	// For example, the source file "dir/foo.proto" might have a filename prefix
	// of "dir/foo". Appending ".pb.go" produces an output file of "dir/foo.pb.go".
	GeneratedFilenamePrefix string
	// Messages is the list of messages defined in this file.
	Messages []*Message
	// Enums is the list of enums defined in this file.
	Enums []*Enum
	// Services is the list of services defined in this file.
	Services []interface{}
}

type Field struct {
	*descriptorpb.FieldDescriptorProto
	// Message is the message type which this field belongs to.
	Message *Message
	// FieldMessage is the message type of the field.
	FieldMessage *Message
	// ForcePrefixedName when set to true, prefixes a type with a package prefix.
	ForcePrefixedName bool
}

type GoPackage struct {
	// Path is the package path to the package.
	Path string
	// Name is the package name of the package
	Name string
	// Alias is an alias of the package unique within the current invocation of gRPC-Gateway generator.
	Alias string
}

// Internal type to store used references.
type refMap map[string]struct{}
