package parser

import (
	"fmt"
	"reflect"
	"testing"

	"go/ast"
)

type NoMatchField struct {
	ast.Expr
}

func TestGetFieldType(t *testing.T) {
	tt := []struct {
		Name                     string
		ExpectedResult           string
		ExpectedFundamentalTypes []string
		InputField               ast.Expr
	}{
		{
			Name:           "Test *ast.Ident as primitive",
			ExpectedResult: "int",
			InputField: &ast.Ident{
				Name: "int",
			},
			ExpectedFundamentalTypes: []string{},
		},
		{
			Name:           "Test *ast.Ident as not primitive",
			ExpectedResult: fmt.Sprintf("%s%s", packageConstant, "TestClass"),
			InputField: &ast.Ident{
				Name: "TestClass",
			},
			ExpectedFundamentalTypes: []string{fmt.Sprintf("%s%s", packageConstant, "TestClass")},
		},
		{
			Name:           "Test *ast.ArrayType",
			ExpectedResult: "[]int",
			InputField: &ast.ArrayType{
				Elt: &ast.Ident{
					Name: "int",
				},
			},
			ExpectedFundamentalTypes: []string{},
		},
		{
			Name:           "Test *ast.SelectorExpr",
			ExpectedResult: "goplantuml.TestClass",
			InputField: &ast.SelectorExpr{
				X: &ast.Ident{
					Name: "puml",
				},
				Sel: &ast.Ident{
					Name: "TestClass",
				},
			},
			ExpectedFundamentalTypes: []string{"goplantuml.TestClass"},
		},
		{
			Name:           "Test *ast.MapType",
			ExpectedResult: "<font color=blue>map</font>[string]int",
			InputField: &ast.MapType{
				Key: &ast.Ident{
					Name: "string",
				},
				Value: &ast.Ident{
					Name: "int",
				},
			},
			ExpectedFundamentalTypes: []string{},
		},
		{
			Name:           "Test *ast.StarExpr",
			ExpectedResult: "*int",
			InputField: &ast.StarExpr{
				X: &ast.Ident{
					Name: "int",
				},
			},
			ExpectedFundamentalTypes: []string{},
		},
		{
			Name:           "Test *ast.ChanType",
			ExpectedResult: "<font color=blue>chan</font> int",
			InputField: &ast.ChanType{
				Value: &ast.Ident{
					Name: "int",
				},
			},
			ExpectedFundamentalTypes: []string{},
		},
		{
			Name:           "Test *ast.StructType",
			ExpectedResult: "<font color=blue>struct</font>{int, string}",
			InputField: &ast.StructType{
				Fields: &ast.FieldList{
					List: []*ast.Field{
						{
							Type: &ast.Ident{
								Name: "int",
							},
						},
						{
							Type: &ast.Ident{
								Name: "string",
							},
						},
					},
				},
			},
			ExpectedFundamentalTypes: []string{},
		},
		{
			Name:           "Test *ast.InterfaceType",
			ExpectedResult: "<font color=blue>interface</font>{Foo <font color=blue>func</font>(*FooComposed) *FooComposed}",
			InputField: &ast.InterfaceType{
				Methods: &ast.FieldList{
					List: []*ast.Field{
						{
							Names: []*ast.Ident{
								{
									Name: "Foo",
								},
							},
							Type: &ast.FuncType{
								Params: &ast.FieldList{
									List: []*ast.Field{
										{
											Names: []*ast.Ident{
												{
													Name: "var1",
												},
											},
											Type: &ast.StarExpr{
												X: &ast.Ident{
													Name: "FooComposed",
												},
											},
										},
									},
								},
								Results: &ast.FieldList{
									List: []*ast.Field{
										{
											Names: nil,
											Type: &ast.StarExpr{
												X: &ast.Ident{
													Name: "FooComposed",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			ExpectedFundamentalTypes: []string{},
		},
		{
			Name:                     "Test *ast.FuncType with one result",
			ExpectedResult:           "<font color=blue>func</font>(*FooComposed) *FooComposed",
			ExpectedFundamentalTypes: []string{},
			InputField: &ast.FuncType{
				Params: &ast.FieldList{
					List: []*ast.Field{
						{
							Names: []*ast.Ident{
								{
									Name: "var1",
								},
							},
							Type: &ast.StarExpr{
								X: &ast.Ident{
									Name: "FooComposed",
								},
							},
						},
					},
				},
				Results: &ast.FieldList{
					List: []*ast.Field{
						{
							Names: nil,
							Type: &ast.StarExpr{
								X: &ast.Ident{
									Name: "FooComposed",
								},
							},
						},
					},
				},
			},
		},
		{
			Name:                     "Test *ast.FuncType with two results",
			ExpectedResult:           "<font color=blue>func</font>(*FooComposed) (*FooComposed, *string)",
			ExpectedFundamentalTypes: []string{},
			InputField: &ast.FuncType{
				Params: &ast.FieldList{
					List: []*ast.Field{
						{
							Names: []*ast.Ident{
								{
									Name: "var1",
								},
							},
							Type: &ast.StarExpr{
								X: &ast.Ident{
									Name: "FooComposed",
								},
							},
						},
					},
				},
				Results: &ast.FieldList{
					List: []*ast.Field{
						{
							Names: nil,
							Type: &ast.StarExpr{
								X: &ast.Ident{
									Name: "FooComposed",
								},
							},
						},
						{
							Names: nil,
							Type: &ast.StarExpr{
								X: &ast.Ident{
									Name: "string",
								},
							},
						},
					},
				},
			},
		},
		{
			Name:                     "Test not match field type",
			ExpectedResult:           "",
			ExpectedFundamentalTypes: []string{},
			InputField:               &NoMatchField{},
		},
		{
			Name:                     "Test *ast.Ellipsis",
			ExpectedResult:           "...int",
			ExpectedFundamentalTypes: []string{},
			InputField: &ast.Ellipsis{
				Elt: &ast.Ident{
					Name: "int",
				},
			},
		},
	}
	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			inputAliasMap := map[string]string{
				"puml": "goplantuml",
			}
			result, fundamentalTypes := getFieldType(tc.InputField, inputAliasMap)
			if result != tc.ExpectedResult {
				t.Errorf("Expected result to be %s, got %s", tc.ExpectedResult, result)
			}

			if !reflect.DeepEqual(fundamentalTypes, tc.ExpectedFundamentalTypes) {
				t.Errorf("Expected result to be %v, got %v", tc.ExpectedFundamentalTypes, fundamentalTypes)
			}
		})
	}
}

func TestIsPrimitiveStringPointer(t *testing.T) {
	if !isPrimitiveString("*int") {
		t.Errorf("TestIsPrimitiveStringPointer: expecting true, got false")
	}
}
