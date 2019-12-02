package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
)

//
var myFilePathHERE = "PAST_YOUR_FILE_PATH_CONTAINING_THE_INTERFACE"
var myInterfaceNameHere = "PAST_YOUR_INTERFACE_NAME"
var whichPackageReturnHere = "PAST_THE_RETURN_STRUCT_PACKAGE"

var r = strings.NewReplacer("*", "*")

func main() {
	readLine(myFilePathHERE, myInterfaceNameHere, whichPackageReturnHere)
}

func readLine(path string, interfaceName string, packageName string) {
	inFile, _ := os.Open(path)
	defer inFile.Close()
	scanner := bufio.NewScanner(inFile)
	scanner.Split(bufio.ScanLines)
	pattern := fmt.Sprintf("type %s interface {", interfaceName)
	var funcFromInterface []string
	found := false
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(strings.Trim(line, "\t"), "//") {
			continue
		}
		if found {
			if strings.Contains(line, "}") {
				generateTheOUtput(&funcFromInterface, packageName, interfaceName)
				return
			}
			if line != "" {
				funcFromInterface = append(funcFromInterface, strings.Trim(line, "\t"))
			}
		}

		if line == pattern {
			found = true
		}
	}
}

func generateTheOUtput(funcFromInterface *[]string, packageName string, intefaceName string) {
	var buffer bytes.Buffer
	generateTypeMockOptions(funcFromInterface, &buffer, intefaceName)
	generateTypeMockOptionsDefault(funcFromInterface, &buffer, intefaceName)
	buffer.WriteString(fmt.Sprintf("type mock%sOption func(*mock%sOptions)\n", intefaceName, intefaceName))
	buffer.WriteString("\n")

	generateWithOption(funcFromInterface, &buffer, intefaceName)
	generateCoreMockMethod(funcFromInterface, &buffer, intefaceName)
	generateStruct(&buffer, packageName, intefaceName)
	println(buffer.String())
}

func generateWithOption(funcFromInterface *[]string, buffer *bytes.Buffer, intefaceName string) {
	for _, v := range *funcFromInterface {
		i := strings.IndexAny(v, "(")
		v = r.Replace(v)
		methodName := v[:i]
		funcParam := v[i:]
		buffer.WriteString(fmt.Sprintf("func withFunc%s(f func%s) mock%sOption {\n", methodName, funcParam, intefaceName))
		buffer.WriteString(fmt.Sprintf("\treturn func(o *mock%sOptions) {\n", intefaceName))
		buffer.WriteString(fmt.Sprintf("\t\to.func%s = f \n", methodName))
		buffer.WriteString("\t}\n")
		buffer.WriteString("}\n")
	}
	buffer.WriteString("\n")
	buffer.WriteString("\n")
}

func generateTypeMockOptionsDefault(funcFromInterface *[]string, buffer *bytes.Buffer, intefaceName string) {
	rO := strings.NewReplacer("(", "")
	rC := strings.NewReplacer(")", "")
	buffer.WriteString(fmt.Sprintf("var defaultMock%sOptions = mock%sOptions{\n", intefaceName, intefaceName))
	for _, v := range *funcFromInterface {
		i := strings.IndexAny(v, "(")
		v = r.Replace(v)
		methodName := v[:i]
		funcParam := v[i:]
		j := strings.Index(v, ")")
		returnParam := v[j:]
		returnParam = rO.Replace(returnParam)
		returnParam = rC.Replace(returnParam)

		params := strings.Split(returnParam, ",")
		var paramBuf bytes.Buffer
		for _, s := range params {
			if strings.Contains(s, "string") {
				paramBuf.WriteString("\"\"")
			} else if strings.Contains(s, "int") {
				paramBuf.WriteString("0")
			} else if strings.Contains(s, "float64") {
				paramBuf.WriteString("0.0")
			} else if strings.Contains(s, "bool") {
				paramBuf.WriteString("false")
			} else {
				paramBuf.WriteString("nil")
			}
			paramBuf.WriteString(",")
		}
		returnParamString := paramBuf.String()
		returnParamString = returnParamString[:len(returnParamString)-1]

		buffer.WriteString(fmt.Sprintf("\tfunc%s :  func %s {\n", methodName, funcParam))

		buffer.WriteString(fmt.Sprintf("\t\treturn %s \n", returnParamString))
		buffer.WriteString("\t}, \n")
	}
	buffer.WriteString("} \n")
	buffer.WriteString("\n")
	buffer.WriteString("\n")
}

func generateTypeMockOptions(funcFromInterface *[]string, buffer *bytes.Buffer, intefaceName string) {
	buffer.WriteString(fmt.Sprintf("type mock%sOptions struct { \n", intefaceName))
	for _, v := range *funcFromInterface {
		i := strings.IndexAny(v, "(")
		v = r.Replace(v)
		methodName := v[:i]
		funcParam := v[i:]
		buffer.WriteString(fmt.Sprintf("\tfunc%s   func %s \n", methodName, funcParam))
	}
	buffer.WriteString("} \n")
	buffer.WriteString("\n")
	buffer.WriteString("\n")
}

func generateCoreMockMethod(funcFromInterface *[]string, buffer *bytes.Buffer, intefaceName string) {
	for _, v := range *funcFromInterface {
		i := strings.IndexAny(v, "(")
		j := strings.Index(v, ")")
		methodName := v[:i]
		funcParam := v[i+1 : j]
		params := strings.Split(funcParam, " ")
		var paramBuf bytes.Buffer
		for i, s := range params {
			if (i % 2) == 0 {
				paramBuf.WriteString(s)
				paramBuf.WriteString(",")
			}
		}
		paramString := paramBuf.String()
		paramString = paramString[0:(len(paramString) - 1)]
		v = r.Replace(v)
		v = fmt.Sprintf("func (s *mock%s) %s {\n", intefaceName, v)
		buffer.WriteString(v)
		buffer.WriteString(fmt.Sprintf("\treturn s.options.func%s(%s)\n", methodName, paramString))
		buffer.WriteString("}\n")
	}
}

func generateStruct(buffer *bytes.Buffer, packageName string, intefaceName string) {
	buffer.WriteString(fmt.Sprintf("type mock%s struct { \n", intefaceName))
	buffer.WriteString(fmt.Sprintf("options mock%sOptions\n", intefaceName))
	buffer.WriteString("}\n")
	if packageName == "" {
		buffer.WriteString(fmt.Sprintf("func newMock%s (opt ...mock%sOption) %s { ", intefaceName, intefaceName, intefaceName))

	} else {
		buffer.WriteString(fmt.Sprintf("func newMock%s (opt ...mock%sOption) %s.%s { ", intefaceName, intefaceName, packageName, intefaceName))

	}
	buffer.WriteString(fmt.Sprintf("opts := defaultMock%sOptions\n", intefaceName))
	buffer.WriteString(fmt.Sprintf("for _, o := range opt {\n"))
	buffer.WriteString(fmt.Sprintf("o(&opts)\n"))
	buffer.WriteString(fmt.Sprintf("}\n"))
	buffer.WriteString(fmt.Sprintf("return &mock%s{\n", intefaceName))
	buffer.WriteString(fmt.Sprintf("options: opts,\n"))
	buffer.WriteString(fmt.Sprintf("}\n"))
	buffer.WriteString("}\n")

}
