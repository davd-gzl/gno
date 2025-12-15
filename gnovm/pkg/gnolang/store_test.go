package gnolang

import (
	"fmt"
	"io"
	"path"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	storetypes "github.com/gnolang/gno/tm2/pkg/store/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransactionStore(t *testing.T) {
	db := memdb.NewMemDB()
	tm2Store := dbadapter.StoreConstructor(db, storetypes.StoreOptions{})

	st := NewStore(nil, tm2Store, tm2Store)
	wrappedTm2Store := tm2Store.CacheWrap()
	txSt := st.BeginTransaction(wrappedTm2Store, wrappedTm2Store, nil)
	m := NewMachineWithOptions(MachineOptions{
		PkgPath: "gno.vm/t/hello",
		Store:   txSt,
		Output:  io.Discard,
	})
	_, pv := m.RunMemPackage(&std.MemPackage{
		Type: MPUserProd,
		Name: "hello",
		Path: "gno.vm/t/hello",
		Files: []*std.MemFile{
			{Name: "hello.gno", Body: "package hello; func main() { println(A(11)); }; type A int"},
		},
	}, true)
	m.SetActivePackage(pv)
	m.RunMain()

	// mem package should only exist in txSt
	// (check both memPackage and types - one is stored directly in the db,
	// the other uses txlog)
	assert.Nil(t, st.GetMemPackage("gno.vm/t/hello"))
	assert.NotNil(t, txSt.GetMemPackage("gno.vm/t/hello"))
	assert.PanicsWithValue(t, "unexpected type with id gno.vm/t/hello.A", func() { st.GetType("gno.vm/t/hello.A") })
	assert.NotNil(t, txSt.GetType("gno.vm/t/hello.A"))

	// use write on the stores
	txSt.Write()
	wrappedTm2Store.Write()

	// mem package should exist and be ==.
	res := st.GetMemPackage("gno.vm/t/hello")
	assert.NotNil(t, res)
	assert.Equal(t, txSt.GetMemPackage("gno.vm/t/hello"), res)
	helloA := st.GetType("gno.vm/t/hello.A")
	assert.NotNil(t, helloA)
	assert.Equal(t, txSt.GetType("gno.vm/t/hello.A"), helloA)
}

func TestTransactionStore_blockedMethods(t *testing.T) {
	// These methods should panic as they modify store settings, which should
	// only be changed in the root store.
	assert.Panics(t, func() { transactionStore{}.SetPackageGetter(nil) })
	assert.Panics(t, func() { transactionStore{}.SetNativeResolver(nil) })
}

func TestCopyFromCachedStore(t *testing.T) {
	// Create cached store, with a type and a mempackage.
	c1 := memdb.NewMemDB()
	c1s := dbadapter.StoreConstructor(c1, storetypes.StoreOptions{})
	c2 := memdb.NewMemDB()
	c2s := dbadapter.StoreConstructor(c2, storetypes.StoreOptions{})
	cachedStore := NewStore(nil, c1s, c2s)
	cachedStore.SetType(&DeclaredType{
		PkgPath: "io",
		Name:    "Reader",
		Base:    BoolType,
	})
	cachedStore.AddMemPackage(&std.MemPackage{
		Type: MPStdlibAll,
		Name: "math",
		Path: "math",
		Files: []*std.MemFile{
			{Name: "math.gno", Body: "package math"},
		},
	}, MPAnyAll)

	// Create dest store and copy.
	d1, d2 := memdb.NewMemDB(), memdb.NewMemDB()
	d1s := dbadapter.StoreConstructor(d1, storetypes.StoreOptions{})
	d2s := dbadapter.StoreConstructor(d2, storetypes.StoreOptions{})
	destStore := NewStore(nil, d1s, d2s)
	destStoreTx := destStore.BeginTransaction(nil, nil, nil) // CopyFromCachedStore requires a tx store.
	CopyFromCachedStore(destStoreTx, cachedStore, c1s, c2s)
	destStoreTx.Write()

	assert.Equal(t, c1, d1, "cached baseStore and dest baseStore should match")
	assert.Equal(t, c2, d2, "cached iavlStore and dest iavlStore should match")
	assert.Equal(t, cachedStore.cacheTypes, destStore.cacheTypes, "cacheTypes should match")
}

func TestFindByPrefix(t *testing.T) {
	stdlibs := []string{"abricot", "balloon", "call", "dingdong", "gnocchi"}
	pkgs := []string{
		"fruits.org/t/abricot",
		"fruits.org/t/abricot/fraise",
		"fruits.org/t/fraise",
	}

	cases := []struct {
		Prefix   string
		Limit    int
		Expected []string
	}{
		{"", 100, append(stdlibs, pkgs...)}, // no prefix == everything
		{"fruits.org", 100, pkgs},
		{"fruits.org/t/abricot", 100, []string{
			"fruits.org/t/abricot", "fruits.org/t/abricot/fraise",
		}},
		{"fruits.org/t/abricot/", 100, []string{
			"fruits.org/t/abricot/fraise",
		}},
		{"fruits", 100, pkgs}, // no stdlibs (prefixed by "_" keys)
		{"_", 100, stdlibs},
		{"_/a", 100, []string{"abricot"}},
		// special case
		{string([]byte{255}), 100, []string{}}, // using 255 as prefix, should not panic
		{string([]byte{0}), 100, []string{}},   // using 0 as prefix, should not panic
		// testing iter seq
		{"_", 0, []string{}},
		{"_", 2, stdlibs[:2]},
	}

	// Create cached store, with a type and a mempackage.
	d1, d2 := memdb.NewMemDB(), memdb.NewMemDB()
	d1s := dbadapter.StoreConstructor(d1, storetypes.StoreOptions{})
	d2s := dbadapter.StoreConstructor(d2, storetypes.StoreOptions{})
	store := NewStore(nil, d1s, d2s)

	// Add stdlibs
	for _, lib := range stdlibs {
		store.AddMemPackage(&std.MemPackage{
			Type: MPStdlibAll,
			Name: lib,
			Path: lib,
			Files: []*std.MemFile{
				{Name: lib + ".gno", Body: "package " + lib},
			},
		}, MPStdlibAll)
	}

	// Add pkgs
	for _, pkg := range pkgs {
		name := path.Base(pkg)
		store.AddMemPackage(&std.MemPackage{
			Type: MPUserProd,
			Name: name,
			Path: pkg,
			Files: []*std.MemFile{
				{Name: name + ".gno", Body: "package " + name},
			},
		}, MPUserProd)
	}

	for _, tc := range cases {
		name := fmt.Sprintf("%s:limit(%d)", tc.Prefix, tc.Limit)
		t.Run(name, func(t *testing.T) {
			t.Logf("lookup prefix: %q, limit: %d", tc.Prefix, tc.Limit)

			paths := []string{}

			var count int
			yield := func(path string) bool {
				if count >= tc.Limit {
					return false
				}

				paths = append(paths, path)
				count++
				return true // continue
			}

			store.FindPathsByPrefix(tc.Prefix)(yield) // find stdlibs
			require.Equal(t, tc.Expected, paths)
		})
	}
}

func TestDeepCloneObject_Block(t *testing.T) {
	alloc := NewAllocator(10000)

	// Create a block with some values
	intVal := TypedValue{T: IntType}
	intVal.SetInt(42)
	strVal := TypedValue{T: StringType, V: StringValue("test")}

	block := &Block{
		ObjectInfo: ObjectInfo{},
		Values:     []TypedValue{intVal, strVal},
	}

	// Clone the block
	cloned := deepCloneObject(block, alloc).(*Block)

	// Verify the clone is not the same object
	require.NotSame(t, block, cloned)

	// Verify the values are equal
	require.Equal(t, len(block.Values), len(cloned.Values))
	require.Equal(t, block.Values[0].T, cloned.Values[0].T)
	require.Equal(t, block.Values[0].GetInt(), cloned.Values[0].GetInt())
	require.Equal(t, int64(42), cloned.Values[0].GetInt())

	// Modify original - should not affect clone
	block.Values[0].SetInt(100)
	require.Equal(t, int64(42), cloned.Values[0].GetInt())
}

func TestDeepCloneObject_PackageValue(t *testing.T) {
	alloc := NewAllocator(10000)

	// Create a TypedValue for the block
	intVal := TypedValue{T: IntType}
	intVal.SetInt(10)

	// Create a package with a block
	pkg := &PackageValue{
		ObjectInfo: ObjectInfo{},
		PkgName:    "test",
		PkgPath:    "gno.land/p/test",
		FNames:     []string{"func1", "func2"},
		Block: &Block{
			ObjectInfo: ObjectInfo{},
			Values:     []TypedValue{intVal},
		},
	}

	// Clone the package
	cloned := deepCloneObject(pkg, alloc).(*PackageValue)

	// Verify the clone is not the same object
	require.NotSame(t, pkg, cloned)
	require.Equal(t, pkg.PkgName, cloned.PkgName)
	require.Equal(t, pkg.PkgPath, cloned.PkgPath)

	// Verify the block is also cloned
	origBlock := pkg.Block.(*Block)
	clonedBlock := cloned.Block.(*Block)
	require.NotSame(t, origBlock, clonedBlock)
	require.Equal(t, origBlock.Values[0].GetInt(), clonedBlock.Values[0].GetInt())

	// Modify original - should not affect clone
	origBlock.Values[0].SetInt(99)
	require.Equal(t, int64(10), clonedBlock.Values[0].GetInt())
}

func TestDeepCloneObject_ArrayValue(t *testing.T) {
	alloc := NewAllocator(10000)

	// Create TypedValues for the array
	val1 := TypedValue{T: IntType}
	val1.SetInt(1)
	val2 := TypedValue{T: IntType}
	val2.SetInt(2)
	val3 := TypedValue{T: IntType}
	val3.SetInt(3)

	// Create an array
	arr := &ArrayValue{
		ObjectInfo: ObjectInfo{},
		List:       []TypedValue{val1, val2, val3},
	}

	// Clone the array
	cloned := deepCloneObject(arr, alloc).(*ArrayValue)

	// Verify the clone is not the same object
	require.NotSame(t, arr, cloned)
	require.Equal(t, len(arr.List), len(cloned.List))

	// Verify values are equal
	for i := range arr.List {
		require.Equal(t, arr.List[i].GetInt(), cloned.List[i].GetInt())
	}

	// Modify original - should not affect clone
	arr.List[0].SetInt(999)
	require.Equal(t, int64(1), cloned.List[0].GetInt())
}

func TestDeepCloneObject_StructValue(t *testing.T) {
	alloc := NewAllocator(10000)

	// Create field values
	field1 := TypedValue{T: IntType}
	field1.SetInt(42)
	field2 := TypedValue{T: StringType, V: StringValue("hello")}

	// Create a struct
	s := &StructValue{
		ObjectInfo: ObjectInfo{},
		Fields:     []TypedValue{field1, field2},
	}

	// Clone the struct
	cloned := deepCloneObject(s, alloc).(*StructValue)

	// Verify the clone is not the same object
	require.NotSame(t, s, cloned)
	require.Equal(t, len(s.Fields), len(cloned.Fields))

	// Verify values are equal
	require.Equal(t, s.Fields[0].GetInt(), cloned.Fields[0].GetInt())
	require.Equal(t, s.Fields[1].V.(StringValue), cloned.Fields[1].V.(StringValue))

	// Modify original - should not affect clone
	s.Fields[0].SetInt(100)
	require.Equal(t, int64(42), cloned.Fields[0].GetInt())
}

func TestDeepCloneObject_MapValue(t *testing.T) {
	alloc := NewAllocator(10000)

	// Create a map with some entries
	m := &MapValue{
		ObjectInfo: ObjectInfo{},
		List: &MapList{
			Head: nil,
			Tail: nil,
			Size: 0,
		},
		vmap: make(map[MapKey]*MapListItem),
	}

	// Add some items
	key1 := TypedValue{T: StringType, V: StringValue("key1")}
	val1 := TypedValue{T: IntType}
	val1.SetInt(10)
	item1 := &MapListItem{Key: key1, Value: val1}
	m.List.Head = item1
	m.List.Tail = item1
	m.List.Size = 1
	keyStr1, _ := key1.ComputeMapKey(nil, false)
	m.vmap[keyStr1] = item1

	// Clone the map
	cloned := deepCloneObject(m, alloc).(*MapValue)

	// Verify the clone is not the same object
	require.NotSame(t, m, cloned)
	require.Equal(t, m.List.Size, cloned.List.Size)

	// Verify the list items are different objects
	require.NotSame(t, m.List.Head, cloned.List.Head)
	require.Equal(t, m.List.Head.Value.GetInt(), cloned.List.Head.Value.GetInt())

	// Modify original - should not affect clone
	m.List.Head.Value.SetInt(999)
	require.Equal(t, int64(10), cloned.List.Head.Value.GetInt())
}

func TestDeepCloneObject_Nil(t *testing.T) {
	alloc := NewAllocator(10000)

	// Clone nil object
	cloned := deepCloneObject(nil, alloc)

	// Verify it returns nil
	require.Nil(t, cloned)
}
