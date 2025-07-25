package main

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func generateMessage(g *protogen.GeneratedFile, message *protogen.Message) {
	structName := message.GoIdent.GoName

	// 生成私有字段结构体
	generatePrivateStruct(g, message, structName)

	// 生成脏标记结构体
	generateDirtyStruct(g, message, structName)

	// 生成构造函数
	generateConstructor(g, message, structName)

	// 生成ensureDirty私有方法
	generateEnsureDirtyMethod(g, message, structName)

	// 生成notifyParentDirty私有方法
	generateNotifyParentDirtyMethod(g, message, structName)

	// 生成Getter/Setter方法
	generateAccessors(g, message, structName)

	// 生成脏标记管理方法
	generateDirtyMethods(g, message, structName)
}

func generatePrivateStruct(g *protogen.GeneratedFile, message *protogen.Message, structName string) {
	g.P("// ", structName, " 结构体，所有字段为私有")
	g.P("type ", structName, " struct {")
	g.P("\t// 私有字段")

	for _, field := range message.Fields {
		fieldName := strings.ToLower(field.GoName[:1]) + field.GoName[1:]
		fieldType := getGoType(field)
		jsonTag := fmt.Sprintf("`json:\"%s\"`", getFieldName(field))
		g.P("\t", fieldName, " ", fieldType, " ", jsonTag)
	}

	g.P()
	g.P("\t// 脏标记跟踪（公开字段以支持反射）")
	g.P("\tDirty *", structName, "Dirty")
	g.P("\t// 父对象通知回调，用于嵌套脏标记同步")
	g.P("\tparentNotifier ParentNotifier")
	g.P("\tparentFieldIndex int // 在父对象中的字段索引")
	g.P("}")
	g.P()
}

func generateDirtyStruct(g *protogen.GeneratedFile, message *protogen.Message, structName string) {
	g.P("// ", structName, "Dirty 脏标记结构体")
	g.P("type ", structName, "Dirty struct {")

	// 计算需要多少个uint64来存储所有字段的脏标记
	fieldCount := len(message.Fields)
	bitmapSize := (fieldCount + 63) / 64 // 向上取整

	g.P("\t// 使用位图存储字段脏标记，每个位代表一个字段")
	if bitmapSize == 1 {
		g.P("\tFieldsBitmap uint64")
	} else {
		g.P("\tFieldsBitmap [", bitmapSize, "]uint64")
	}

	// 为数组和字典字段生成额外的跟踪
	for i, field := range message.Fields {
		if isArrayOrMap(field) {
			publicFieldName := strings.Title(strings.ToLower(field.GoName[:1]) + field.GoName[1:])
			g.P("\t", publicFieldName, "Elements map[interface{}]bool // 跟踪具体元素的变更")
		}
		// 生成字段索引常量注释
		g.P("\t// ", field.GoName, " field index: ", i)
	}

	g.P("\tTotalChanges int // 总变更数量")
	g.P("\tTotalFields  int // 总字段数量")
	g.P("}")
	g.P()

	// 生成字段索引常量
	g.P("// ", structName, " 字段索引常量")
	g.P("const (")
	for i, field := range message.Fields {
		constName := fmt.Sprintf("%s%sFieldIndex", structName, field.GoName)
		g.P("\t", constName, " = ", i)
	}
	g.P(")")
	g.P()
}

func generateConstructor(g *protogen.GeneratedFile, message *protogen.Message, structName string) {
	g.P("// New", structName, " 创建新的", structName, "实例")
	g.P("func New", structName, "() *", structName, " {")
	g.P("\treturn &", structName, "{")
	g.P("\t\tDirty: &", structName, "Dirty{")
	g.P("\t\t\tTotalFields: ", len(message.Fields), ",")

	// 初始化数组和字典的元素跟踪
	for _, field := range message.Fields {
		if isArrayOrMap(field) {
			publicFieldName := strings.Title(strings.ToLower(field.GoName[:1]) + field.GoName[1:])
			g.P("\t\t\t", publicFieldName, "Elements: make(map[interface{}]bool),")
		}
	}

	g.P("\t\t},")
	g.P("\t}")
	g.P("}")
	g.P()

	// 生成SetParentNotifier方法
	g.P("// SetParentNotifier 设置父对象通知器")
	g.P("func (x *", structName, ") SetParentNotifier(notifier ParentNotifier, fieldIndex int) {")
	g.P("\tif x == nil {")
	g.P("\t\treturn")
	g.P("\t}")
	g.P("\tx.parentNotifier = notifier")
	g.P("\tx.parentFieldIndex = fieldIndex")
	g.P("}")
	g.P()
}

func generateAccessors(g *protogen.GeneratedFile, message *protogen.Message, structName string) {
	for i, field := range message.Fields {
		fieldName := strings.ToLower(field.GoName[:1]) + field.GoName[1:]
		publicName := field.GoName
		fieldType := getGoType(field)
		fieldIndex := i

		// Getter方法
		g.P("// Get", publicName, " 获取", fieldName, "字段的值")
		g.P("func (x *", structName, ") Get", publicName, "() ", fieldType, " {")
		g.P("\tif x == nil {")
		g.P("\t\treturn ", getZeroValue(field))
		g.P("\t}")

		// 如果是message类型，设置父对象通知器
		if field.Desc.Kind() == protoreflect.MessageKind {
			g.P("\tif x.", fieldName, " != nil {")
			constName := fmt.Sprintf("%s%sFieldIndex", structName, field.GoName)
			g.P("\t\tx.", fieldName, ".SetParentNotifier(x, ", constName, ")")
			g.P("\t}")
		}

		g.P("\treturn x.", fieldName)
		g.P("}")
		g.P()

		// Setter方法
		g.P("// Set", publicName, " 设置", fieldName, "字段的值")
		g.P("func (x *", structName, ") Set", publicName, "(v ", fieldType, ") {")
		g.P("\tif x == nil {")
		g.P("\t\treturn")
		g.P("\t}")
		g.P("\tx.EnsureDirty()")

		// 检查值是否真的改变了
		g.P("\tif !reflect.DeepEqual(x.", fieldName, ", v) {")
		g.P("\t\tif !x.isFieldDirty(", fieldIndex, ") {")
		g.P("\t\t\tx.Dirty.TotalChanges++")
		g.P("\t\t}")
		g.P("\t\tx.setFieldDirty(", fieldIndex, ")")
		g.P("\t\tx.", fieldName, " = v")

		// 如果是message类型，设置父对象通知器
		if field.Desc.Kind() == protoreflect.MessageKind {
			g.P("\t\tif x.", fieldName, " != nil {")
			constName := fmt.Sprintf("%s%sFieldIndex", structName, field.GoName)
			g.P("\t\t\tx.", fieldName, ".SetParentNotifier(x, ", constName, ")")
			g.P("\t\t}")
		}

		// 通知父对象脏标记更新
		g.P("\t\tx.notifyParentDirty()")
		g.P("\t}")
		g.P("}")
		g.P()

		// 如果是数组或字典，生成额外的操作方法
		if isArrayOrMap(field) {
			generateCollectionMethods(g, message, field, structName, fieldName, publicName, fieldType)
		}
	}
}

func generateCollectionMethods(g *protogen.GeneratedFile, message *protogen.Message, field *protogen.Field, structName, fieldName, publicName, fieldType string) {
	publicFieldName := strings.Title(fieldName)
	fieldIndex := -1

	// 找到字段索引
	for i, f := range message.Fields {
		if f == field {
			fieldIndex = i
			break
		}
	}

	if field.Desc.IsList() {
		// 数组操作方法
		elementType := getElementType(field)

		g.P("// Add", publicName, "Element 向", fieldName, "添加元素")
		g.P("func (x *", structName, ") Add", publicName, "Element(v ", elementType, ") {")
		g.P("\tif x == nil {")
		g.P("\t\treturn")
		g.P("\t}")
		g.P("\tx.EnsureDirty()")
		g.P("\tx.", fieldName, " = append(x.", fieldName, ", v)")
		g.P("\tindex := len(x.", fieldName, ") - 1")
		g.P("\tx.Dirty.", publicFieldName, "Elements[index] = true")
		g.P("\tif !x.isFieldDirty(", fieldIndex, ") {")
		g.P("\t\tx.Dirty.TotalChanges++")
		g.P("\t\tx.setFieldDirty(", fieldIndex, ")")
		g.P("\t}")
		g.P("}")
		g.P()

		g.P("// Set", publicName, "Element 设置", fieldName, "指定位置的元素")
		g.P("func (x *", structName, ") Set", publicName, "Element(index int, v ", elementType, ") {")
		g.P("\tif x == nil || index < 0 || index >= len(x.", fieldName, ") {")
		g.P("\t\treturn")
		g.P("\t}")
		g.P("\tx.EnsureDirty()")
		g.P("\tif !reflect.DeepEqual(x.", fieldName, "[index], v) {")
		g.P("\t\tx.", fieldName, "[index] = v")
		g.P("\t\tx.Dirty.", publicFieldName, "Elements[index] = true")
		g.P("\t\tif !x.isFieldDirty(", fieldIndex, ") {")
		g.P("\t\t\tx.Dirty.TotalChanges++")
		g.P("\t\t\tx.setFieldDirty(", fieldIndex, ")")
		g.P("\t\t}")
		g.P("\t}")
		g.P("}")
		g.P()

	} else if field.Desc.IsMap() {
		// 字典操作方法
		keyType, valueType := getMapTypes(field)

		g.P("// Set", publicName, "Value 设置", fieldName, "中指定键的值")
		g.P("func (x *", structName, ") Set", publicName, "Value(key ", keyType, ", value ", valueType, ") {")
		g.P("\tif x == nil {")
		g.P("\t\treturn")
		g.P("\t}")
		g.P("\tx.EnsureDirty()")
		g.P("\tif x.", fieldName, " == nil {")
		g.P("\t\tx.", fieldName, " = make(", fieldType, ")")
		g.P("\t}")
		g.P("\toldValue, exists := x.", fieldName, "[key]")
		g.P("\tif !exists || !reflect.DeepEqual(oldValue, value) {")
		g.P("\t\tx.", fieldName, "[key] = value")
		g.P("\t\tx.Dirty.", publicFieldName, "Elements[key] = true")
		g.P("\t\tif !x.isFieldDirty(", fieldIndex, ") {")
		g.P("\t\t\tx.Dirty.TotalChanges++")
		g.P("\t\t\tx.setFieldDirty(", fieldIndex, ")")
		g.P("\t\t}")
		g.P("\t}")
		g.P("}")
		g.P()
	}
}

// 辅助函数

func getGoType(field *protogen.Field) string {
	// 先检查是否是数组或映射
	if field.Desc.IsList() {
		elementType := getElementType(field)
		return "[]" + elementType
	}

	if field.Desc.IsMap() {
		keyType, valueType := getMapTypes(field)
		return fmt.Sprintf("map[%s]%s", keyType, valueType)
	}

	// 再检查基本类型
	switch field.Desc.Kind() {
	case protoreflect.StringKind:
		return "string"
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return "int32"
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return "int64"
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return "uint32"
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return "uint64"
	case protoreflect.BoolKind:
		return "bool"
	case protoreflect.FloatKind:
		return "float32"
	case protoreflect.DoubleKind:
		return "float64"
	case protoreflect.BytesKind:
		return "[]byte"
	case protoreflect.MessageKind:
		return "*" + string(field.Message.GoIdent.GoName)
	}

	return "interface{}"
}

func getElementType(field *protogen.Field) string {
	switch field.Desc.Kind() {
	case protoreflect.StringKind:
		return "string"
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return "int32"
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return "int64"
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return "uint32"
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return "uint64"
	case protoreflect.BoolKind:
		return "bool"
	case protoreflect.FloatKind:
		return "float32"
	case protoreflect.DoubleKind:
		return "float64"
	case protoreflect.BytesKind:
		return "[]byte"
	case protoreflect.MessageKind:
		return "*" + string(field.Message.GoIdent.GoName)
	default:
		return "interface{}"
	}
}

func getMapTypes(field *protogen.Field) (string, string) {
	if !field.Desc.IsMap() {
		return "interface{}", "interface{}"
	}

	// 获取map的key和value类型
	mapEntry := field.Message
	var keyType, valueType string

	for _, f := range mapEntry.Fields {
		switch f.Desc.Name() {
		case "key":
			keyType = getGoTypeForMapField(f)
		case "value":
			valueType = getGoTypeForMapField(f)
		}
	}

	if keyType == "" {
		keyType = "interface{}"
	}
	if valueType == "" {
		valueType = "interface{}"
	}

	return keyType, valueType
}

func getGoTypeForMapField(field *protogen.Field) string {
	switch field.Desc.Kind() {
	case protoreflect.StringKind:
		return "string"
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return "int32"
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return "int64"
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return "uint32"
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return "uint64"
	case protoreflect.BoolKind:
		return "bool"
	case protoreflect.FloatKind:
		return "float32"
	case protoreflect.DoubleKind:
		return "float64"
	case protoreflect.BytesKind:
		return "[]byte"
	case protoreflect.MessageKind:
		return "*" + string(field.Message.GoIdent.GoName)
	default:
		return "interface{}"
	}
}

func getZeroValue(field *protogen.Field) string {
	switch field.Desc.Kind() {
	case protoreflect.StringKind:
		return "\"\""
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind,
		protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind,
		protoreflect.Uint32Kind, protoreflect.Fixed32Kind,
		protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return "0"
	case protoreflect.BoolKind:
		return "false"
	case protoreflect.FloatKind, protoreflect.DoubleKind:
		return "0.0"
	case protoreflect.BytesKind:
		return "nil"
	default:
		return "nil"
	}
}

func getFieldName(field *protogen.Field) string {
	return string(field.Desc.Name())
}

func isArrayOrMap(field *protogen.Field) bool {
	return field.Desc.IsList() || field.Desc.IsMap()
}

// generateDirtyInitialization 生成dirty初始化代码
func generateDirtyInitialization(g *protogen.GeneratedFile, message *protogen.Message, structName string) {
	g.P("\tif x.Dirty == nil {")
	g.P("\t\tx.Dirty = &", structName, "Dirty{TotalFields: ", len(message.Fields), "}")

	// 初始化数组和字典的元素跟踪
	for _, f := range message.Fields {
		if isArrayOrMap(f) {
			publicFieldName := strings.Title(strings.ToLower(f.GoName[:1]) + f.GoName[1:])
			g.P("\t\tx.Dirty.", publicFieldName, "Elements = make(map[interface{}]bool)")
		}
	}

	g.P("\t}")
}

// generateEnsureDirtyMethod 生成ensureDirty私有方法
func generateEnsureDirtyMethod(g *protogen.GeneratedFile, message *protogen.Message, structName string) {
	g.P("// EnsureDirty 确保dirty结构体已初始化")
	g.P("func (x *", structName, ") EnsureDirty() {")
	generateDirtyInitialization(g, message, structName)
	g.P("}")
	g.P()
}

// generateNotifyParentDirtyMethod 生成notifyParentDirty私有方法
func generateNotifyParentDirtyMethod(g *protogen.GeneratedFile, message *protogen.Message, structName string) {
	g.P("// notifyParentDirty 通知父对象更新脏标记")
	g.P("func (x *", structName, ") notifyParentDirty() {")
	g.P("\tif x == nil || x.parentNotifier == nil {")
	g.P("\t\treturn")
	g.P("\t}")
	g.P("\t// 直接调用父对象的NotifyFieldChanged方法，避免反射")
	g.P("\tx.parentNotifier.NotifyFieldChanged(x.parentFieldIndex)")
	g.P("}")
	g.P()

	// 生成NotifyFieldChanged方法实现
	g.P("// NotifyFieldChanged 实现ParentNotifier接口")
	g.P("func (x *", structName, ") NotifyFieldChanged(fieldIndex int) {")
	g.P("\tif x == nil {")
	g.P("\t\treturn")
	g.P("\t}")
	g.P("\tx.EnsureDirty()")
	g.P("\tif !x.isFieldDirty(fieldIndex) {")
	g.P("\t\tx.Dirty.TotalChanges++")
	g.P("\t\tx.setFieldDirty(fieldIndex)")
	g.P("\t}")
	g.P("\tx.notifyParentDirty() // 递归通知父对象")
	g.P("}")
	g.P()

	// 生成位图操作辅助方法
	g.P("// isFieldDirty 检查指定字段是否脏")
	g.P("func (x *", structName, ") isFieldDirty(fieldIndex int) bool {")
	g.P("\tif x == nil || x.Dirty == nil || fieldIndex < 0 {")
	g.P("\t\treturn false")
	g.P("\t}")

	fieldCount := len(message.Fields)
	bitmapSize := (fieldCount + 63) / 64

	if bitmapSize == 1 {
		g.P("\treturn (x.Dirty.FieldsBitmap & (1 << uint(fieldIndex))) != 0")
	} else {
		g.P("\tbitmapIndex := fieldIndex / 64")
		g.P("\tbitIndex := fieldIndex % 64")
		g.P("\treturn (x.Dirty.FieldsBitmap[bitmapIndex] & (1 << uint(bitIndex))) != 0")
	}
	g.P("}")
	g.P()

	g.P("// setFieldDirty 设置指定字段为脏")
	g.P("func (x *", structName, ") setFieldDirty(fieldIndex int) {")
	g.P("\tif x == nil || x.Dirty == nil || fieldIndex < 0 {")
	g.P("\t\treturn")
	g.P("\t}")

	if bitmapSize == 1 {
		g.P("\tx.Dirty.FieldsBitmap |= (1 << uint(fieldIndex))")
	} else {
		g.P("\tbitmapIndex := fieldIndex / 64")
		g.P("\tbitIndex := fieldIndex % 64")
		g.P("\tx.Dirty.FieldsBitmap[bitmapIndex] |= (1 << uint(bitIndex))")
	}
	g.P("}")
	g.P()
}
