package utils

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

func parseFile(fileName string) (*ast.File, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, fileName, nil, parser.AllErrors)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func hasTag(field *ast.Field, tag string) bool {
	if field.Tag == nil {
		return false
	}
	return strings.Contains(field.Tag.Value, tag)
}

func getSnakeCaseFromTag(field *ast.Field) string {
	if field.Tag != nil {
		tag := field.Tag.Value
		tag = strings.Trim(tag, "`")
		tagParts := strings.Split(tag, " ")
		for _, part := range tagParts {
			if strings.HasPrefix(part, "boil:") {
				return strings.Trim(strings.TrimPrefix(part, "boil:"), `"`)
			}
		}
	}
	return ""
}

func getTypeFromFieldType(fieldType ast.Expr) SQLBoilerType {
	var tp SQLBoilerType
	switch t := fieldType.(type) {
	case *ast.Ident:
		tp = SQLBoilerType{
			OriginalName:  t.Name,
			FormattedName: t.Name,
		}
	case *ast.StarExpr:
		ft := getTypeFromFieldType(t.X)

		tp = SQLBoilerType{
			OriginalName:  ft.OriginalName,
			FormattedName: "*" + ft.OriginalName,
		}
	case *ast.SelectorExpr:
		ft := getTypeFromFieldType(t.X)

		tp = SQLBoilerType{
			OriginalName:  ft.OriginalName + "." + t.Sel.Name,
			FormattedName: ft.OriginalName + "." + t.Sel.Name,
		}
	default:
		fmt.Println("unknown", t)
	}

	tp.FormattedName = sqlboilerTypeToType(tp.FormattedName)

	if _, ok := enumCacheMap[tp.OriginalName]; ok {
		tp.IsEnum = true
	}

	return tp
}

var sqlboilerTypes = map[string]string{
	"time":    "time.Time",
	"json":    "any",
	"decimal": "float64",
}

func sqlboilerTypeToType(s string) string {
	var formattedString = s

	if strings.Contains(formattedString, "time") {
		modelImports = append(modelImports, "time")
	}

	if strings.Contains(s, ".") {
		splitted := strings.Split(s, ".")
		formattedString = strings.ToLower(splitted[1])
	}

	if val, ok := sqlboilerTypes[formattedString]; ok {
		formattedString = val
	}

	if strings.HasSuffix(formattedString, "array") {
		formattedString = "[]" + strings.TrimSuffix(formattedString, "array")
	}

	if strings.HasPrefix(s, "null.") {
		formattedString = "*" + formattedString
	}

	return formattedString
}

func convertGoTypeToTypescript(t SQLBoilerType) string {
	var formattedString = t.FormattedName

	formattedString = strings.TrimPrefix(formattedString, "*")

	if strings.Contains(formattedString, "int") || strings.Contains(formattedString, "float") {
		formattedString = "number"
	}

	formattedString = strings.ReplaceAll(formattedString, "bool", "boolean")
	formattedString = strings.ReplaceAll(formattedString, "time.Time", "Date")

	if strings.HasSuffix(formattedString, "Slice") {
		return fmt.Sprintf("%v[]", strings.TrimSuffix(formattedString, "Slice"))
	}

	if strings.HasPrefix(formattedString, "[]") {
		return fmt.Sprintf("%v[]", strings.TrimPrefix(formattedString, "[]"))
	}

	return formattedString

}
