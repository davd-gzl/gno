// PKGPATH: gno.land/r/demo/crossinpfinal
package crossinpfinal

// Stress the finalizer. The caller holds an object that is not yet real
// when it enters the helper, and stores it afterwards. If returning from
// the helper writes the caller's realm back early, the later store meets
// a half-finalized realm.

import "filetests/extern/crossinp_adv"

type Node struct {
	Name string
	Next *Node
}

var root *Node

func main(cur realm) {
	tmp := &Node{Name: "a"}
	tmp.Next = &Node{Name: "b"}
	println(crossinp_adv.Direct(cur))
	root = tmp
	println(root.Name, root.Next.Name)
}

// Output:
// gno.land/r/demo/crossinpfinal
// a b
