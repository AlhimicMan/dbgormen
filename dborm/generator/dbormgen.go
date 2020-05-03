package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"reflect"
	"strings"
)

var fileName = flag.String("name", "", "Set file name to generate code for")

type DbeParam struct {
	TableName string `json:"table"`
}

type StructInfo struct {
	Name     string
	GenParam *DbeParam
	Target   *ast.GenDecl
}

type ColInfo struct {
	ColName   string
	FieldName string
	NotNull   bool
	ColType   string
	FieldType string
}

type TableInfo struct {
	StructName string
	TableName  string
	PrimaryKey *ColInfo
	Columns    []*ColInfo
}

/*
id int(11) NOT NULL AUTO_INCREMENT,
  title varchar(255) NOT NULL,
  description text NOT NULL,
  updated varchar(255) DEFAULT NULL,
  PRIMARY KEY (id)
*/
func (tableD *TableInfo) generateCreateTable(out *os.File) error {
	fmt.Fprint(out, "func (in *"+tableD.StructName+") createTable(db *sql.DB) (error) {\n")
	var resSQLq = fmt.Sprintf("\tsqlQ := `CREATE TABLE %s (\n", tableD.TableName)
	for _, col := range tableD.Columns {
		colSQL := col.ColName + " " + col.ColType
		if col.NotNull {
			colSQL += " NOT NULL"
		}
		if col == tableD.PrimaryKey {
			colSQL += " AUTO_INCREMENT"
		}
		colSQL += ",\n"
		resSQLq += colSQL
	}
	if tableD.PrimaryKey != nil {
		resSQLq += fmt.Sprintf("PRIMARY KEY (%s)\n", tableD.PrimaryKey.ColName)
	}
	resSQLq += ")`\n"
	fmt.Fprint(out, resSQLq)
	fmt.Fprint(out, "\t_, err := db.Exec(sqlQ)\n\t\tif err != nil {\n\t\t\treturn err\n\t\t}\n")
	fmt.Fprint(out, "\t return nil\n}\n\n")
	return nil
}

func (tableD *TableInfo) generateCreate(out *os.File) error {
	fmt.Fprint(out, "func (in *"+tableD.StructName+") Create(db *sql.DB) (error) {\n")
	var columns, valuePlaces, valuesListParams string
	for _, col := range tableD.Columns {
		if col == tableD.PrimaryKey {
			continue
		}
		columns += "`" + col.ColName + "`,"
		valuePlaces += "?,"
		valuesListParams += "in." + col.FieldName + ","
	}
	columns = columns[:len(columns)-1]
	valuePlaces = valuePlaces[:len(valuePlaces)-1]
	valuesListParams = valuesListParams[:len(valuesListParams)-1]

	resSQLq := fmt.Sprintf("\tsqlQ := \"INSERT INTO %s (%s) VALUES (%s);\"\n",
		tableD.TableName,
		columns,
		valuePlaces)
	fmt.Fprintln(out, resSQLq)
	fmt.Fprintf(out, "result, err := db.Exec(sqlQ, %s)\n", valuesListParams)
	fmt.Fprintln(out, `if err != nil {
		return err
	}`)
	//Setting id if we have primary key
	if tableD.PrimaryKey != nil {
		fmt.Fprintf(out, `lastId, err := result.LastInsertId()
		if err != nil {
			return nil
		}`)
		fmt.Fprintf(out, "\nin.%s = %s(lastId)\n", tableD.PrimaryKey.FieldName, tableD.PrimaryKey.FieldType)
	}
	fmt.Fprintln(out, "return nil\n}\n\n")
	//in., _ := result.LastInsertId()`)
	return nil
}

func (tableD *TableInfo) generateQuery(out *os.File) error {
	fmt.Fprint(out, "func (in *"+tableD.StructName+") Query(db *sql.DB) ([]*"+tableD.StructName+", error) {\n")

	fmt.Fprintf(out, "\tsqlQ := \"SELECT * FROM %s;\"\n", tableD.TableName)
	fmt.Fprintf(out, "rows, err := db.Query(sqlQ)\n")
	fmt.Fprintf(out, "results := make([]*%s, 0)\n", tableD.StructName)
	fmt.Fprintf(out, `for rows.Next() {`)
	fmt.Fprintf(out, "\t tempR := &%s{}\n", tableD.StructName)
	var valuesListParams string
	for _, col := range tableD.Columns {
		valuesListParams += "&tempR." + col.FieldName + ","
	}
	valuesListParams = valuesListParams[:len(valuesListParams)-1]

	fmt.Fprintf(out, "\terr = rows.Scan(%s)\n", valuesListParams)
	fmt.Fprintf(out, `if err != nil {
		return nil, err
		}`)
	fmt.Fprintf(out, "\n\tresults = append(results, tempR)")
	fmt.Fprintf(out, `}
		return results, nil
	}`)
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "")
	return nil
}

func (tableD *TableInfo) generateUpdate(out *os.File) error {
	fmt.Fprint(out, "func (in *"+tableD.StructName+") Update(db *sql.DB) (error) {\n")
	var updVals, valuesListParams string
	for _, col := range tableD.Columns {
		if col == tableD.PrimaryKey {
			continue
		}
		updVals += "`" + col.ColName + "`=?,"
		valuesListParams += "in." + col.FieldName + ","
	}
	updVals = updVals[:len(updVals)-1]
	valuesListParams += "in." + tableD.PrimaryKey.FieldName

	resSQLq := fmt.Sprintf("\tsqlQ := \"UPDATE %s SET %s WHERE %s = ?;\"\n",
		tableD.TableName,
		updVals,
		tableD.PrimaryKey.ColName)
	fmt.Fprintln(out, resSQLq)
	fmt.Fprintf(out, "_, err := db.Exec(sqlQ, %s)\n", valuesListParams)
	fmt.Fprintln(out, `if err != nil {
		return err
	}`)

	fmt.Fprintln(out, "return nil\n}\n\n")
	//in., _ := result.LastInsertId()`)
	return nil
}

func (tableD *TableInfo) generateDelete(out *os.File) error {
	fmt.Fprint(out, "func (in *"+tableD.StructName+") Delete(db *sql.DB) (error) {\n")
	fmt.Fprintf(out, "sqlQ := \"DELETE FROM %s WHERE id = ?\"\n", tableD.TableName)

	fmt.Fprintf(out, "_, err := db.Exec(sqlQ, in.%s)\n", tableD.PrimaryKey.FieldName)

	fmt.Fprintln(out, `if err != nil {
		return err
	}
	return nil
}`)
	fmt.Fprintln(out)
	return nil
}

func generateMethods(reqStruct *StructInfo, out *os.File) {
	for _, spec := range reqStruct.Target.Specs {
		fmt.Fprintln(out, "")
		currType, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}
		currStruct, ok := currType.Type.(*ast.StructType)
		if !ok {
			continue
		}

		fmt.Printf("\tgenerating createTable methods for %s\n", currType.Name.Name)

		curTable := &TableInfo{
			TableName: reqStruct.GenParam.TableName,
			Columns:   make([]*ColInfo, 0, len(currStruct.Fields.List)),
		}

		for _, field := range currStruct.Fields.List {
			if len(field.Names) == 0 {
				continue
			}
			tableCol := &ColInfo{FieldName: field.Names[0].Name}
			var fieldIsPrimKey bool
			var preventThisField bool
			if field.Tag != nil {
				tag := reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1])
				tagVal := tag.Get("dbe")
				fmt.Println("dbe:", tagVal)
				tagParams := strings.Split(tagVal, ",")
			PARAMSLOOP:
				for _, param := range tagParams {
					switch param {
					case "primary_key":
						if curTable.PrimaryKey == nil {
							fieldIsPrimKey = true
							tableCol.NotNull = true
						} else {
							log.Panicf("Table %s cannot have more then 1 primary key!", currType.Name.Name)
						}
					case "not_null":
						tableCol.NotNull = true
					case "-":
						preventThisField = true
						break PARAMSLOOP
					default:
						tableCol.ColName = param
					}

				}
				if preventThisField {
					continue
				}
			}
			if tableCol.ColName == "" {
				tableCol.ColName = tableCol.FieldName
			}
			if fieldIsPrimKey {
				curTable.PrimaryKey = tableCol
			}
			//Determine field type
			var fieldType string
			switch field.Type.(type) {
			case *ast.Ident:
				fieldType = field.Type.(*ast.Ident).Name
			case *ast.SelectorExpr:
				fieldType = field.Type.(*ast.SelectorExpr).Sel.Name
			}
			//fieldType := field.Type.(*ast.Ident).Name
			fmt.Printf("%s- %s\n", tableCol.FieldName, fieldType)
			//Check for integers
			if strings.Contains(fieldType, "int") {
				tableCol.ColType = "integer"
			} else {
				//Check for other types
				switch fieldType {
				case "string":
					tableCol.ColType = "text"
				case "bool":
					tableCol.ColType = "boolean"
				case "Time":
					tableCol.ColType = "TIMESTAMP"
				default:
					log.Panicf("Field type %s not supported", fieldType)
				}
			}
			tableCol.FieldType = fieldType
			curTable.Columns = append(curTable.Columns, tableCol)
			curTable.StructName = currType.Name.Name

		}
		curTable.generateCreateTable(out)

		fmt.Printf("\tgenerating CRUD methods for %s\n", currType.Name.Name)
		curTable.generateCreate(out)
		curTable.generateQuery(out)
		curTable.generateUpdate(out)
		curTable.generateDelete(out)
	}
}

func main() {
	flag.Parse()
	fmt.Println("DB ORM code generator")
	if *fileName == "" {
		fmt.Println("Filename not set. Exiting...")
		return
	}
	fName := *fileName
	fNameOut := fName[:len(fName)-3] + "_dbe.go"
	fmt.Println(fNameOut)

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, *fileName, nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	out, _ := os.Create(fNameOut)
	defer out.Close()

	fmt.Fprintln(out, `package `+node.Name.Name)
	fmt.Fprintln(out) // empty line
	fmt.Fprintln(out, `import "database/sql"`)
	fmt.Fprintln(out) // empty line

	for _, f := range node.Decls {
		genD, ok := f.(*ast.GenDecl)
		if !ok {
			fmt.Printf("SKIP %T is not *ast.GenDecl\n", f)
			continue
		}
		targetStruct := &StructInfo{}
		var thisIsStruct bool
		for _, spec := range genD.Specs {
			currType, ok := spec.(*ast.TypeSpec)
			if !ok {
				fmt.Printf("SKIP %T is not ast.TypeSpec\n", spec)
				continue
			}

			currStruct, ok := currType.Type.(*ast.StructType)
			if !ok {
				fmt.Printf("SKIP %T is not ast.StructType\n", currStruct)
				continue
			}
			targetStruct.Name = currType.Name.Name
			thisIsStruct = true
		}
		//Getting comments
		var needCodegen bool
		var dbeParams string
		if thisIsStruct {
			for _, comment := range genD.Doc.List {
				needCodegen = needCodegen || strings.HasPrefix(comment.Text, "// dbe")
				if len(comment.Text) < 7 {
					dbeParams = ""
				} else {
					dbeParams = strings.Replace(comment.Text, "// dbe:", "", 1)
				}
			}
		}
		if needCodegen {
			targetStruct.Target = genD
			genParams := &DbeParam{}
			if len(dbeParams) != 0 {
				err := json.Unmarshal([]byte(dbeParams), genParams)
				if err != nil {
					fmt.Printf("Error encoding DBE params for structure %s\n", targetStruct.Name)
					continue
				}
			} else {
				genParams.TableName = targetStruct.Name
			}

			targetStruct.GenParam = genParams
			generateMethods(targetStruct, out)
		}
	}
}
