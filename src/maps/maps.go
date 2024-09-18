// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package maps defines various functions useful with maps of any type.
package maps

// Equal reports whether two maps contain the same key/value pairs.
// Values are compared using ==.
func Equal[M1, M2 ~map[K]V, K, V comparable](m1 M1, m2 M2) bool { //M1,M2是map, 底层是map即可.~表示不限制必须是map, 底层是map即可. KV是可比较类型. 然后这里定义Equal方法,就是挨个遍历map元素进行元素比较即可.
	if len(m1) != len(m2) {
		return false
	}
	for k, v1 := range m1 {
		if v2, ok := m2[k]; !ok || v1 != v2 {
			return false
		}
	}
	return true
}

// EqualFunc is like Equal, but compares values using eq.
// Keys are still compared with ==.
func EqualFunc[M1 ~map[K]V1, M2 ~map[K]V2, K comparable, V1, V2 any](m1 M1, m2 M2, eq func(V1, V2) bool) bool { // 这里面就是比较函数不是等号了,而是eq函数.
	if len(m1) != len(m2) {
		return false
	}
	for k, v1 := range m1 {
		if v2, ok := m2[k]; !ok || !eq(v1, v2) {
			return false
		}
	}
	return true
}

// clone is implemented in the runtime package.
func clone(m any) any

// Clone returns a copy of m.  This is a shallow clone:
// the new keys and values are set using ordinary assignment.
func Clone[M ~map[K]V, K comparable, V any](m M) M { //输入M类型, 输出一个M类型的拷贝,这里是浅拷贝,拷贝m对象,m对象里面的各个key value,还是m拷贝之前的那个对象的里面的key,value. 浅拷贝节省空间.提高效率, 但是不能完全分离之前的底层数据.
	// Preserve nil in case it matters.
	if m == nil {
		return nil
	}
	return clone(m).(M)
}

// Copy copies all key/value pairs in src adding them to dst.
// When a key in src is already present in dst,
// the value in dst will be overwritten by the value associated
// with the key in src.
func Copy[M1 ~map[K]V, M2 ~map[K]V, K comparable, V any](dst M1, src M2) {
	for k, v := range src {
		dst[k] = v
	}
}

// DeleteFunc deletes any key/value pairs from m for which del returns true.
func DeleteFunc[M ~map[K]V, K comparable, V any](m M, del func(K, V) bool) {
	for k, v := range m {
		if del(k, v) { // 删除m里面的全部key, value对. 删完就是一个空字典了.
			delete(m, k)
		}
	}
}
