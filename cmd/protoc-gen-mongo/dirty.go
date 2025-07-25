package main

import (
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
)

func generateDirtyMethods(g *protogen.GeneratedFile, message *protogen.Message, structName string) {
	fieldCount := len(message.Fields)
	bitmapSize := (fieldCount + 63) / 64

	// 重置脏标记
	g.P("// ResetDirty 重置所有脏标记")
	g.P("func (x *", structName, ") ResetDirty() {")
	g.P("\tif x == nil || x.Dirty == nil {")
	g.P("\t\treturn")
	g.P("\t}")
	g.P("\tx.Dirty.TotalChanges = 0")

	// 重置位图
	if bitmapSize == 1 {
		g.P("\tx.Dirty.FieldsBitmap = 0")
	} else {
		g.P("\tfor i := range x.Dirty.FieldsBitmap {")
		g.P("\t\tx.Dirty.FieldsBitmap[i] = 0")
		g.P("\t}")
	}

	// 重置数组和字典的元素跟踪
	for _, field := range message.Fields {
		if isArrayOrMap(field) {
			publicFieldName := strings.Title(strings.ToLower(field.GoName[:1]) + field.GoName[1:])
			g.P("\tx.Dirty.", publicFieldName, "Elements = make(map[interface{}]bool)")
		}
	}
	g.P("}")
	g.P()

	// 检查是否有脏数据
	g.P("// IsDirty 检查是否有脏数据")
	g.P("func (x *", structName, ") IsDirty() bool {")
	g.P("\tif x == nil || x.Dirty == nil {")
	g.P("\t\treturn false")
	g.P("\t}")
	g.P("\treturn x.Dirty.TotalChanges > 0")
	g.P("}")
	g.P()

	// 检查特定字段是否脏
	for i, field := range message.Fields {
		publicName := field.GoName

		g.P("// Is", publicName, "Dirty 检查", publicName, "字段是否有变更")
		g.P("func (x *", structName, ") Is", publicName, "Dirty() bool {")
		g.P("\treturn x.isFieldDirty(", i, ")")
		g.P("}")
		g.P()
	}

	// 生成获取脏字段数量的方法
	g.P("// GetDirtyFieldCount 获取脏字段数量")
	g.P("func (x *", structName, ") GetDirtyFieldCount() int {")
	g.P("\tif x == nil || x.Dirty == nil {")
	g.P("\t\treturn 0")
	g.P("\t}")
	g.P("\treturn x.Dirty.TotalChanges")
	g.P("}")
	g.P()

	// 生成获取所有脏字段索引的方法
	g.P("// GetDirtyFieldIndexes 获取所有脏字段的索引")
	g.P("func (x *", structName, ") GetDirtyFieldIndexes() []int {")
	g.P("\tif x == nil || x.Dirty == nil {")
	g.P("\t\treturn nil")
	g.P("\t}")
	g.P("\tvar indexes []int")
	g.P("\tfor i := 0; i < ", fieldCount, "; i++ {")
	g.P("\t\tif x.isFieldDirty(i) {")
	g.P("\t\t\tindexes = append(indexes, i)")
	g.P("\t\t}")
	g.P("\t}")
	g.P("\treturn indexes")
	g.P("}")
	g.P()
}
