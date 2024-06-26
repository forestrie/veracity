package veracity

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"

	"github.com/datatrails/go-datatrails-merklelog/massifs"
	"github.com/urfave/cli/v2"
)

type Node struct {
	prefix      []byte
	pdepth      int
	terminating bool
	Left        *Node
	Right       *Node
	Value       string
	Hash        string
}

func (n *Node) isInternal() bool {
	return n.Value == "" && !n.terminating
}

func (n *Node) isTerminating() bool {
	return n.terminating
}

func (n *Node) isLeaf() bool {
	return !n.isTerminating() && !n.isInternal()
}

type BinaryUrkelTrie struct {
	Root *Node
}

func NewBinaryUrkelTrie() *BinaryUrkelTrie {
	return &BinaryUrkelTrie{}
}

func newLeafFromLeaf(leaf *Node) *Node {
	return &Node{
		Value:       leaf.Value,
		prefix:      leaf.prefix,
		pdepth:      leaf.pdepth,
		Hash:        leaf.Hash,
		terminating: leaf.terminating,
	}
}

func hash(data string) string {
	h := sha256.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func (t *BinaryUrkelTrie) Insert(key []byte, value string) {
	//binaryKey := toBinaryString(key)

	// if len(bkey) != 32 {
	// 	panic("bad key length")
	// }

	nn := &Node{
		prefix: key,
		pdepth: len(key) * 8,
		Value:  value,
		Hash:   hash(value),
	}
	t.Root = t.insert(t.Root, nn, 0)
}

func prefixForDepth(depth int, prefix []byte, lastSignificant *int) []byte {
	b := make([]byte, len(prefix))
	for i, bb := range prefix {
		if i == depth/8 {
			break
		}
		b[i] = bb
	}

	if depth/8 == len(prefix) {
		return b
	}

	b[depth/8] = (prefix[depth/8] >> (depth % 8)) << (depth % 8)

	if lastSignificant != nil && *lastSignificant == 1 {
		mask := (byte(0xff) >> (depth % 8))
		b[depth/8] = b[depth/8] | mask
	} else if lastSignificant != nil {
		mask := (byte(0xff) >> (depth % 8)) << (depth % 8)
		b[depth/8] = b[depth/8] & mask
	}

	return b
}

// func (t *BinaryUrkelTrie) insert(node *Node, new *Node, depth int) *Node {
// 	// if we have empty tree just return the leaf
// 	if node == nil {
// 		return new
// 	}

// 	// if we hit a leaf find out appropriate depth and create a merge node

// 	if node.isLeaf() && new.isLeaf() {
// 		tdepth := commonPrefix(node.prefix, new.prefix)
// 		left := !bit(tdepth, new.prefix)
// 		nn := &Node{
// 			prefix:      node.prefix,
// 			pdepth:      tdepth,
// 			terminating: false,
// 		}
// 		if left {
// 			nn.Left = new
// 			nn.Right = node
// 		} else {
// 			nn.Left = node
// 			nn.Right = new
// 		}
// 		return nn
// 	}

// 	// if it's internal node we go deeper until we reach appropriate depth
// 	if node.isInternal() {
// 		for depth != new.pdepth {
// 			left := !bit(depth, new.prefix)
// 			if left {
// 				node.Left = t.insert(node.Left, new, depth+1)
// 				if node.Left.isLeaf() {
// 					break
// 				}
// 			} else {
// 				node.Right = t.insert(node.Right, new, depth+1)
// 				if node.Right.isLeaf() {
// 					break
// 				}
// 			}
// 		}
// 	}

// 	if node.isTerminating() {
// 		return node
// 	}

// 	return node
// }

func (t *BinaryUrkelTrie) insert(node *Node, new *Node, depth int) *Node {
	if node == nil {
		return new
	}

	//consume teminating node
	if !new.isInternal() && !new.isTerminating() && node.isTerminating() {
		return new
	}

	if bytes.Equal(new.prefix, node.prefix) && depth == len(node.prefix)*8 {
		return new
	}
	//targetDepth := commonPrefixS(node.prefix, new.prefix)
	tdepth := min(commonPrefix(node.prefix, new.prefix), node.pdepth)

	prefixNext := tdepth
	if node.isTerminating() || node.isInternal() || new.isInternal() || (node == t.Root && node.isInternal()) {
		prefixNext = min(tdepth, depth)
	}

	left := !bit(prefixNext, new.prefix)

	one := 1
	zero := 0

	// if node.isInternal() {
	// 	if left {
	// 		node.Left = t.insert(node.Left, new, depth+1)
	// 	} else {
	// 		node.Right = t.insert(node.Right, new, depth+1)
	// 	}
	// } else if node.isTerminating() && new.isInternal() && depth == new.pdepth {
	// 	return new
	// } else if node.isTerminating() {

	// 	if new.isInternal() {
	// 		n = terminatingNode(prefixForDepth(depth+1, new.prefix, nn), depth+1)
	// 	}
	// 	node.terminating = false
	// 	n = t.insert(n, new, depth+1)
	// 	o = terminatingNode(prefixForDepth(depth+1, new.prefix, oo), depth+1)
	// } else {
	// 	var newNode *Node
	// 	if left {
	// 		newNode = &Node{

	// 			// shift max int and then and it - cut off targetDepth most significant
	// 			prefix: prefixForDepth(tdepth, new.prefix, nil),
	// 			pdepth: tdepth,
	// 			Left:   new,
	// 			Right:  newLeafFromLeaf(node),
	// 			Hash:   hash(new.Hash + node.Hash),
	// 		}
	// 	} else {
	// 		newNode = &Node{
	// 			prefix: prefixForDepth(tdepth, new.prefix, nil), //node.prefix[:targetDepth],
	// 			pdepth: tdepth,
	// 			Left:   newLeafFromLeaf(node),
	// 			Right:  new,
	// 			Hash:   hash(node.Hash + new.Hash),
	// 		}
	// 	}

	// 	nodeToDoubleTerminating(node, new.prefix, depth)
	// 	if depth != tdepth {
	// 		return t.insert(node, newNode, depth)
	// 	}
	// }

	if left { //new.prefix[targetDepth] == '0' {
		if node.isInternal() {
			// we traverse deeper
			node.Left = t.insert(node.Left, new, depth+1)
		} else if node.isTerminating() {
			// replace terminating node
			if new.isInternal() && depth == new.pdepth {
				return new
			}

			if new.isInternal() {
				node.Left = terminatingNode(prefixForDepth(depth+1, new.prefix, &zero), depth+1)
			}

			node.terminating = false
			node.Left = t.insert(node.Left, new, depth+1)
			node.Right = terminatingNode(prefixForDepth(depth+1, new.prefix, &one), depth+1)
		} else {
			// we move the curent one to the side
			newNode := &Node{

				// shift max int and then and it - cut off targetDepth most significant
				prefix: prefixForDepth(tdepth, new.prefix, nil),
				pdepth: tdepth,
				Left:   new,
				Right:  newLeafFromLeaf(node),
				Hash:   hash(new.Hash + node.Hash),
			}
			// as above
			nodeToDoubleTerminating(node, new.prefix, depth)
			// node.terminating = false
			// node.prefix = prefixForDepth(depth, new.prefix, nil)
			// node.pdepth = depth
			// node.Value = ""
			// node.Hash = ""
			// node.Left = terminatingNode(prefixForDepth(depth+1, new.prefix, &zero), depth+1)
			// node.Right = terminatingNode(prefixForDepth(depth+1, new.prefix, &one), depth+1)

			if depth != tdepth {
				return t.insert(node, newNode, depth)
			}

			return newNode
		}
	} else {
		if node.isInternal() {
			node.Right = t.insert(node.Right, new, depth+1)
		} else if node.isTerminating() {
			// replace terminating node
			if new.isInternal() && depth == new.pdepth {
				return new
			}
			if new.isInternal() {
				node.Right = terminatingNode(prefixForDepth(depth+1, new.prefix, &one), depth+1)
			}
			node.terminating = false
			node.Right = t.insert(node.Right, new, depth+1)
			node.Left = terminatingNode(prefixForDepth(depth+1, new.prefix, &zero), depth+1)
		} else {
			// we move the curent one to the side

			newNode := &Node{
				prefix: prefixForDepth(tdepth, new.prefix, nil), //node.prefix[:targetDepth],
				pdepth: tdepth,
				Left:   newLeafFromLeaf(node),
				Right:  new,
				Hash:   hash(node.Hash + new.Hash),
			}

			nodeToDoubleTerminating(node, new.prefix, depth)
			// node.terminating = false
			// node.prefix = prefixForDepth(depth, new.prefix, nil) //node.prefix[:depth]
			// node.pdepth = depth
			// node.Value = ""
			// node.Hash = ""
			// node.Left = terminatingNode(prefixForDepth(depth+1, new.prefix, &zero), depth+1)
			// node.Right = terminatingNode(prefixForDepth(depth+1, new.prefix, &one), depth+1)
			if depth != tdepth { //len(newNode.prefix) {
				return t.insert(node, newNode, depth)
			}

			return newNode
		}
	}

	if node.isTerminating() {
		return node
	}

	nodeHash := node.Value
	if node.Left != nil {
		nodeHash += node.Left.Hash
	}
	if node.Right != nil {
		nodeHash += node.Right.Hash
	}
	node.Hash = hash(nodeHash)

	node.pdepth = depth

	return node
}

func nodeToDoubleTerminating(node *Node, prefix []byte, depth int) {
	one := 1
	zero := 0
	node.terminating = false
	node.prefix = prefixForDepth(depth, prefix, nil)
	node.pdepth = depth
	node.Value = ""
	node.Hash = ""
	node.Left = terminatingNode(prefixForDepth(depth+1, prefix, &zero), depth+1)
	node.Right = terminatingNode(prefixForDepth(depth+1, prefix, &one), depth+1)
}

func terminatingNode(prefix []byte, depth int) *Node {
	return &Node{
		prefix:      prefix,
		pdepth:      depth,
		terminating: true,
		Hash:        "0000000000000000000000000000000000000000000000000000000000000000",
	}
}

func (t *BinaryUrkelTrie) GetPath(key []byte) ([]string, string, bool) {
	path := []string{}
	node := t.Root
	for i := 0; i <= len(key)*8; i++ {

		if node == nil {
			return path, "", false
		}

		path = append(path, node.Hash)
		// this is leaf value we have to break out
		if node.Value != "" {
			break
		}

		if node.isTerminating() {
			return path, "", false
		}

		if bit(i, key) {
			node = node.Right
		} else {
			node = node.Left
		}
	}

	//path = append(path, node.Hash)
	if !bytes.Equal(node.prefix, key) {
		return path, node.Value, false
	}

	return path, node.Value, true
}

// const (
// 	terminatingHash = 0x0
// )

// type NodeType int

// const (
// 	INTERNAL NodeType = iota
// 	LEAF
// 	NULL
// )

// type node interface {
// 	Hash() []byte
// 	GetPrefix() []byte
// 	GetType() NodeType
// }

// type internal struct {
// 	left, right node
// 	prefix      []byte
// 	hash        []byte
// }

// func (n *internal) Hash() []byte {
// 	return n.hash
// }

// func (n *internal) GetPrefix() []byte {
// 	return n.prefix
// }

// func (n *)

// type leaf struct {
// 	prefix []byte
// 	hash   []byte
// 	key    []byte
// }

// type nullNode struct {
// 	prefix []byte
// 	hash   []byte
// }

// func insert(n node, key, value []byte, depth int) node {
// 	switch t := n.(type) {
// 	case nullNode:
// 		fmt.Printf("got null")
// 	case internal:
// 		fmt.Printf("got internal")
// 	case leaf:
// 		fmt.Printf("got leaf")
// 	default:
// 		fmt.Printf("unknown")
// 	}
// }

// var fhash = []byte{0xff}

// type unode struct {
// 	value []byte
// 	l, h  []byte
// 	key   []byte
// }

// func commonDepth(prefix int, kl int) int {
// 	return prefix
// 	// if prefix == 0 {
// 	// 	return 1
// 	// }
// 	// return kl - math.Ilogb(float64(prefix))
// }

func bit(pos int, value []byte) bool {
	b := value[(pos)/8]
	r := (pos) % 8
	m := 1 << (7 - r)
	return b&byte(m) != 0
}

func commonPrefix(node, value []byte) int {
	commonBytes := 0
	for i, b := range node {
		if len(value) <= i {
			break
		}
		bb := b ^ value[i]
		if bb == 0x0 {
			commonBytes += 1
			continue
		}

		break
	}

	if commonBytes == len(node) || commonBytes == len(value) {
		fmt.Printf("Common prefix %d\n", commonBytes*8)
		return commonBytes * 8
	}

	nb := node[commonBytes]
	vb := value[commonBytes]

	commonBits := 8
	for nb != vb {
		nb = nb >> 1
		vb = vb >> 1
		commonBits -= 1
	}

	fmt.Printf("Common prefix %d\n", commonBytes*8+commonBits)
	return commonBytes*8 + commonBits
}

// type node struct {
// 	value       []byte
// 	key         []byte
// 	min, max    []byte
// 	left, right *node
// }

// func (n *node) isLeaf() bool {
// 	return n.left == nil && n.right == nil
// }

// func (n *node) fallsToTheLeft(value []byte) bool {
// 	// if we're lower than lower band
// 	// we go to the left for sure
// 	if bytes.Compare(value, n.min) == -1 {
// 		return true
// 	}

// 	// if we're higher than the higher band we go to the right
// 	if bytes.Compare(value, n.max) == 1 {
// 		return false
// 	}

// 	// if we're lower then left high we go to the left
// 	if n.left != nil && bytes.Compare(value, n.left.max) == -1 {
// 		return true
// 	}

// 	return false
// }

// func CreateLeafithKeyValue(key, value []byte) *node {
// 	fmt.Printf("adding leaf: %s with value: %s\n", hex.EncodeToString(key), hex.EncodeToString(value))
// 	return &node{
// 		min:   key,
// 		max:   key,
// 		key:   key,
// 		value: value,
// 		left:  nil,
// 		right: nil,
// 	}
// }

// func AddToUrkle(leaf, urkle *node, depth int, prefix int) *node {
// 	if urkle == nil || (urkle.key == nil && urkle.left == nil && urkle.right == nil) {
// 		fmt.Printf("got empty urkle\n")
// 		return leaf
// 	}

// 	if bytes.Equal(urkle.key, fhash) {
// 		urkle.key = leaf.key
// 		urkle.value = leaf.value
// 		urkle.min = leaf.key
// 		urkle.max = leaf.key
// 		return urkle
// 	}

// 	if depth >= prefix && !leaf.isLeaf() {
// 		if urkle.left != nil {
// 			urkle.right = AddToUrkle(leaf, urkle, depth+1, prefix)
// 			return urkle
// 		}

// 		if urkle.right != nil {
// 			urkle.left = AddToUrkle(leaf, urkle, depth+1, prefix)
// 			return urkle
// 		}
// 	}
// 	// if prefix >= depth {
// 	// 	// add null nodes
// 	// 	return leaf
// 	// }

// 	fmt.Printf("adding to urkle: %s %d\n", hex.EncodeToString(urkle.key), depth)
// 	lprefix := commonPrefix(urkle.min, leaf.key)
// 	rprefix := commonPrefix(urkle.max, leaf.key)
// 	maxPrefix := int(math.Max(float64(lprefix), float64(rprefix)))
// 	fmt.Printf("prefixes: %d %d (max: %d)\n", lprefix, rprefix, maxPrefix)
// 	left := false
// 	if lprefix == rprefix {
// 		cmin := bytes.Compare(urkle.min, leaf.key)
// 		//cmax := bytes.Compare(urkle.max, leaf.key)
// 		if cmin == 1 {
// 			left = true
// 		}

// 		// if cmax == -1 {
// 		// 	right = true
// 		// }

// 		// right = true
// 	}

// 	if lprefix > rprefix {
// 		left = true

// 	} //else {
// 	// 	right = true
// 	// }

// 	newNode := &node{
// 		left:  urkle.left,
// 		right: urkle.right,
// 		min:   urkle.min,
// 		max:   urkle.max,
// 		key:   urkle.key,
// 		value: urkle.value,
// 	}

// 	if left {
// 		fmt.Printf("going left\n")
// 		// if it's a leaf we just do a new node l:r and return it if depth has been reached
// 		if urkle.isLeaf() {
// 			leaf = &node{
// 				left:  leaf,
// 				right: urkle,
// 				min:   leaf.key,
// 				max:   urkle.key,
// 			}
// 		}

// 		if depth >= maxPrefix {
// 			return leaf
// 		}

// 		if urkle.isLeaf() {
// 			newNode.left = &node{
// 				left: &node{key: fhash},
// 			}
// 			newNode.right = &node{
// 				key: fhash,
// 			}
// 			newNode.min = leaf.min
// 			newNode.max = leaf.max
// 			newNode.key = nil
// 			newNode.value = nil
// 		}

// 		// otherwise we add null node and go deeper
// 		newNode.left = AddToUrkle(leaf, newNode.left, depth+1, maxPrefix)
// 		newNode.min = newNode.left.min
// 	} else {
// 		fmt.Printf("going right\n")
// 		// if we hit a leaf we create a node from it and either return it
// 		// or go deeper depending on maxPrefix
// 		if urkle.isLeaf() {
// 			leaf = &node{
// 				left:  urkle,
// 				right: leaf,
// 				min:   urkle.key,
// 				max:   leaf.key,
// 			}
// 		}

// 		if depth >= maxPrefix && depth >= prefix {
// 			return leaf
// 		}

// 		// we're not deep enough and it was a leaf so need to add
// 		// terminating hash
// 		if urkle.isLeaf() {
// 			newNode.left = &node{
// 				key: fhash,
// 			}
// 			newNode.right = &node{
// 				right: &node{key: fhash},
// 			}
// 			newNode.min = leaf.min
// 			newNode.max = leaf.max
// 			newNode.key = nil
// 			newNode.value = nil
// 		}

// 		// urkle was a leaf
// 		// but we made it here
// 		// so we need to go deeper
// 		newNode.right = AddToUrkle(leaf, newNode.right, depth+1, maxPrefix)
// 		newNode.max = newNode.right.max
// 	}

// 	return newNode

// 	// // this is leaf, so we need to spawn nodes
// 	// if urkle.left == nil && urkle.right == nil {
// 	// 	fmt.Printf("got leaf %s\n", hex.EncodeToString(urkle.key))
// 	// 	if bytes.Equal(urkle.key, []byte{terminatingHash}) {
// 	// 		// we need to consume this one maybe
// 	// 	}

// 	// 	c := bytes.Compare(urkle.key, leaf.key)
// 	// 	// may want to add X nodes if more than one off
// 	// 	switch c {
// 	// 	case 0:
// 	// 		fmt.Printf("equal val\n")
// 	// 		return urkle
// 	// 	case 1:
// 	// 		meetingPoint := commonDepth(commonPrefix(urkle.key, leaf.key))
// 	// 		fmt.Printf("lower val - will split @%d currently @%d\n", meetingPoint, depth)
// 	// 		newNode := &node{}
// 	// 		newNode.left = leaf
// 	// 		newNode.right = urkle
// 	// 		newNode.l = leaf.key
// 	// 		newNode.h = urkle.key
// 	// 		newNode.key = nil // TODO make it hash key + key
// 	// 		newNode.value = nil
// 	// 		return newNode
// 	// 	case -1:
// 	// 		meetingPoint := commonDepth(commonPrefix(urkle.key, leaf.key))
// 	// 		fmt.Printf("higher val - will split @%d currently @%d\n", meetingPoint, depth)
// 	// 		newNode := &node{}
// 	// 		newNode.left = urkle
// 	// 		newNode.right = leaf
// 	// 		newNode.l = urkle.key
// 	// 		newNode.h = leaf.key
// 	// 		newNode.key = nil // TODO make it hash key + key
// 	// 		newNode.value = nil
// 	// 		return newNode
// 	// 	}
// 	// }

// 	// it's not a leaf - we need to go deeper until we find a leaf again

// 	// we need to add a leaf to the left
// 	// if urkle.fallsToTheLeft(leaf.key) {
// 	// 	fmt.Printf("goint to the left\n")
// 	// 	if bytes.Compare(leaf.key, urkle.l) == -1 {
// 	// 		urkle.l = leaf.key
// 	// 	}
// 	// 	return AddToUrkle(leaf, urkle.left, depth+1)
// 	// }

// 	// fmt.Printf("goint to the right\n")
// 	// if bytes.Compare(leaf.key, urkle.h) == 1 {
// 	// 	urkle.h = leaf.key
// 	// }
// 	// return AddToUrkle(leaf, urkle.right, depth+1)
// }

// NewUrkelCommand implements a sub command which construcst an urkel trie
func NewUrkelCommand() *cli.Command {
	return &cli.Command{Name: "urkel",
		Usage: "construct an urkel from massif",
		Flags: []cli.Flag{
			&cli.Int64Flag{
				Name: "massif", Aliases: []string{"m"},
			},
			&cli.StringFlag{
				Name: "value", Aliases: []string{"v"},
			},
			&cli.BoolFlag{Name: "massif-relative", Aliases: []string{"r"}},
		},
		Action: func(cCtx *cli.Context) error {
			var err error
			cmd := &CmdCtx{}

			if err = cfgMassif(cmd, cCtx); err != nil {
				return err
			}

			//targetValue, err := hex.DecodeString(cCtx.String("value"))
			if err != nil {
				return err
			}
			start := cmd.massif.LogStart()
			count := cmd.massif.Count()
			urkel := NewBinaryUrkelTrie()
			for i := uint64(0); i < count; i++ {
				val := cmd.massif.Data[start+i*massifs.ValueBytes : start+i*massifs.ValueBytes+massifs.ValueBytes]
				key := cmd.massif.Data[start+i*massifs.TrieKeyBytes : start+i*massifs.TrieKeyBytes+massifs.TrieKeyBytes]
				urkel.Insert(key, hex.EncodeToString(val))
				//urkel = AddToUrkle(CreateLeafithKeyValue(key, val), urkel, 0, 0)

				fmt.Printf("%d: %s %d %d\n", i+cmd.massif.Start.FirstIndex, hex.EncodeToString(key), len(key), math.Ilogb(float64(len(key))*8))

			}
			//fmt.Printf("---- : %s: %s\n", hex.EncodeToString(urkel.min), hex.EncodeToString(urkel.max))
			return nil // fmt.Errorf("'%s' not found", cCtx.String("value"))
		},
	}
}
