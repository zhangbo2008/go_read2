// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package reflect

// VisibleFields returns all the visible fields in t, which must be a
// struct type. A field is defined as visible if it's accessible
// directly with a FieldByName call. The returned fields include fields
// inside anonymous struct members and unexported fields. They follow
// the same order found in the struct, with anonymous fields followed
// immediately by their promoted fields.
//
// For each element e of the returned slice, the corresponding field
// can be retrieved from a value v of type t by calling v.FieldByIndex(e.Index).
func VisibleFields(t Type) []StructField { //返回一个结构体里面的可以访问的字段.
	if t == nil {
		panic("reflect: VisibleFields(nil)")
	}
	if t.Kind() != Struct {
		panic("reflect.VisibleFields of non-struct type")
	}
	w := &visibleFieldsWalker{
		byName:   make(map[string]int),
		visiting: make(map[Type]bool),
		fields:   make([]StructField, 0, t.NumField()),
		index:    make([]int, 0, 2),
	}
	w.walk(t)
	// Remove all the fields that have been hidden.
	// Use an in-place removal that avoids copying in
	// the common case that there are no hidden fields.
	j := 0
	for i := range w.fields { //上面的walk方法我们分析了, 结果保存在w.fields里面.下面统计一下然后切片.
		f := &w.fields[i]
		if f.Name == "" { //有一些名字是空字符串的.我们跳过他.通过42行来修改fields去掉这些空内容.
			continue
		}
		if i != j { //36行会触发下一次循环39行.让i比j大.这是真实索引应该是j,所以把f放到真正索引j里面.
			// A field has been removed. We need to shuffle
			// all the subsequent elements up.
			w.fields[j] = *f
		}
		j++
	}
	return w.fields[:j] //计数完毕切片即可.
}

type visibleFieldsWalker struct {
	byName   map[string]int
	visiting map[Type]bool
	fields   []StructField
	index    []int
}

// walk walks all the fields in the struct type t, visiting
// fields in index preorder and appending them to w.fields
// (this maintains the required ordering).
// Fields that have been overridden have their
// Name field cleared.
func (w *visibleFieldsWalker) walk(t Type) {
	if w.visiting[t] {
		return
	}
	w.visiting[t] = true
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i) //f是第i个字段
		w.index = append(w.index, i)
		add := true
		if oldIndex, ok := w.byName[f.Name]; ok { //如果在byName字典里面.那么byName里面能获取到这个f.Name上次遍历到的时候的索引.因为我们字典是嵌套的,所以可以存在同名的key//这时候我们就需要保存index小的那个,因为那个是更外层的包含更多的信息.最终结果我们就保存到w.fields这个数组里面.
			old := &w.fields[oldIndex]
			if len(w.index) == len(old.Index) {
				// Fields with the same name at the same depth
				// cancel one another out. Set the field name
				// to empty to signify that has happened, and
				// there's no need to add this field.
				old.Name = ""
				add = false
			} else if len(w.index) < len(old.Index) {
				// The old field loses because it's deeper than the new one.
				old.Name = ""
			} else {
				// The old field wins because it's shallower than the new one.
				add = false
			}
		}
		if add {
			// Copy the index so that it's not overwritten
			// by the other appends.
			f.Index = append([]int(nil), w.index...)
			w.byName[f.Name] = len(w.fields)
			w.fields = append(w.fields, f)
		}
		if f.Anonymous { //如果这个字段也是一个结构体.
			if f.Type.Kind() == Pointer {
				f.Type = f.Type.Elem()
			}
			if f.Type.Kind() == Struct {
				w.walk(f.Type)
			}
		}
		w.index = w.index[:len(w.index)-1]
	}
	delete(w.visiting, t)
}
