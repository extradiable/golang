package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"aqwari.net/xml/xmltree"

	"aqwari.net/xml/xsd"
)

type Column struct {
	Name     string
	Type     string
	Comment  string
	Required bool
}

func (c Column) String() string {
	return fmt.Sprintf("Name: %s\nType: %s\nRequired: %s\nComment: %s", c.Name, c.Type, strconv.FormatBool(c.Required), c.Comment)
}

type Table struct {
	Name    string
	Comment string
	Parent  *Table
	Columns []Column
}

func (t Table) String() string {
	return fmt.Sprintf("Name: %s\nComment: %s", t.Name, t.Comment)
}

func (t Table) ToSQL() {
	fmt.Printf("create table %s (\n", t.Name)
	fmt.Println(");")
}

var tableMap map[string]Table
var simpleTypeMap map[string]string

func init() {
	tableMap = make(map[string]Table)
	simpleTypeMap = make(map[string]string)
}

func processElement(element xmltree.Element) {
	t := Table{}
	t.Name = element.Attr("", "name")
	for _, annotation := range element.Search("http://www.w3.org/2001/XMLSchema", "annotation") {
		for _, documentation := range annotation.Search("http://www.w3.org/2001/XMLSchema", "documentation") {
			t.Comment = string(documentation.Content)
		}
	}
	tableMap[t.Name] = t
}

func processComplexType(element xmltree.Element) {
	name := element.Attr("", "name")
	table, found := tableMap[name]
	if !found {
		//panic(fmt.Sprintf("\ntable: %s has not been processed", name))
		return
	}
	for _, attribute := range element.Search("http://www.w3.org/2001/XMLSchema", "attribute") {
		column := Column{
			Name:     attribute.Attr("", "name"),
			Required: attribute.Attr("", "use") == "required",
			Type:     attribute.Attr("", "type"),
		}
		docs := attribute.Search("http://www.w3.org/2001/XMLSchema", "documentation")
		if len(docs) > 1 {
			fmt.Printf("warning: Multiple document nodes found at complexType %s\n", name)
		}
		for _, doc := range docs {
			column.Comment = string(doc.Content)
		}
		table.Columns = append(table.Columns, column)
	}
}

func main() {
	var file = flag.String("f", "", "XSD file")
	flag.Parse()
	bytes, _ := os.ReadFile(*file)
	var elements []*xmltree.Element
	var err error
	elements, err = xsd.Normalize(bytes)
	if err != nil {
		panic(err)
	}
	element := elements[0]
	// On the first pass, tables and columns are initialized
	for i := 0; i < len(element.Children); i++ {
		child := element.Children[i]
		if child.Name.Local == "element" {
			processElement(child)
		} else if child.Name.Local == "complexType" {
			processComplexType(child)
		}
	}
	// On the second pass, simple types are resolved
	// Finally, tables are printed:
	for k := range tableMap {
		tableMap[k].ToSQL()
	}
}
