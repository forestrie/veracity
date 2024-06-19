package veracity

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"

	"github.com/datatrails/go-datatrails-merklelog/massifs"
	"github.com/urfave/cli/v2"
)

type Node struct {
	prefix      []byte
	terminating bool
	Left        *Node
	Right       *Node
	Value       string
	Hash        string
}

func slice2Uint64(slice []byte) uint64 {
	return binary.BigEndian.Uint64(slice)
}

func uint642slice(n uint64) []byte {
	b := make([]byte, 32)
	binary.BigEndian.PutUint64(b, n)
	return b
}

func (n *Node) isInternal() bool {
	return n.Value == "" && !n.terminating
}

func (n *Node) isTerminating() bool {
	return n.terminating
}

type BinaryUrkelTrie struct {
	Root *Node
}

func NewBinaryUrkelTrie() *BinaryUrkelTrie {
	return &BinaryUrkelTrie{}
}

func newLeafFromLeaf(leaf *Node) *Node {
	return &Node{
		Value:  leaf.Value,
		prefix: leaf.prefix,
		Hash:   leaf.Hash,
	}
}

func hash(data string) string {
	h := sha256.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func (t *BinaryUrkelTrie) Insert(bkey, value string) {
	//binaryKey := toBinaryString(key)

	// if len(bkey) != 32 {
	// 	panic("bad key length")
	// }

	nn := &Node{
		prefix: []byte(bkey),
		Value:  value,
		Hash:   hash(value),
	}
	t.Root = t.insert(t.Root, nn, 0)
}

func toBinaryString(key string) string {
	var binaryStr string
	for i := 0; i < len(key); i++ {
		binaryStr += fmt.Sprintf("%08b", key[i])
	}
	return binaryStr
}

func commonPrefixS(a, b string) int {
	prefix := 0
	for i := 0; i < len(a); i++ {
		if i >= len(b) {
			return prefix
		}
		if b[i] != a[i] {
			return prefix
		}
		prefix++
	}
	return prefix
}

func prefixForDepth(depth int, prefix []byte, lastSignificant *int) []byte {
	b := make([]byte, len(prefix))
	for i, bb := range prefix {
		if i == depth/8 {
			break
		}
		b[i] = bb
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

func (t *BinaryUrkelTrie) insert(node *Node, new *Node, depth int) *Node {
	if node == nil {
		//node = new
		return new
	}

	//consume teminating node
	if !new.isInternal() && !new.isTerminating() && node.isTerminating() {
		return new
	}
	// if len(key) == 0 {
	// 	node.Value = value
	// 	node.Hash = hash(value)
	// 	return node
	// }

	if bytes.Equal(new.prefix, node.prefix) && depth == len(node.prefix)*8 {
		return new
	}
	//targetDepth := commonPrefixS(node.prefix, new.prefix)
	tdepth := commonPrefix(node.prefix, new.prefix)
	//tidx := (tdepth / 8)
	// mask := 0x1 << (7 - (tdepth - tidx*8))
	// r := new.prefix[tidx] & byte(mask)
	// fmt.Printf("1 %s\n", toBinaryString(string(rune(node.prefix))))
	// fmt.Printf("2 %s\n", toBinaryString(string(rune(new.prefix))))
	// fmt.Printf("%s\n", toBinaryString(string(uint642slice(new.prefix)[tidx])))
	// fmt.Printf("%s\n", toBinaryString(string([]byte{byte(mask)})))
	// fmt.Printf("---- %d: %d shifting by: %d %t\n", tdepth, tidx, 7-(tdepth-tidx*8), r == 0x0)

	prefixNext := tdepth + 1
	if node.isTerminating() || node.isInternal() || new.isInternal() {
		prefixNext = min(tdepth+1, depth+1)
	}

	left := !bit(prefixNext, new.prefix)

	one := 1
	zero := 0

	if left { //new.prefix[targetDepth] == '0' {
		if node.isInternal() {
			// we traverse deeper
			node.Left = t.insert(node.Left, new, depth+1)
		} else if node.isTerminating() {
			// replace terminating node
			node.terminating = false
			node.Left = t.insert(node.Left, new, depth+1)
			node.Right = &Node{
				prefix:      prefixForDepth(depth+1, new.prefix, &one), //(node.prefix << 1) + 1, // "1",
				terminating: true,
			}
		} else {
			// we move the curent one to the side
			newNode := &Node{

				// shift max int and then and it - cut off targetDepth most significant
				prefix: prefixForDepth(tdepth, new.prefix, nil), // node.prefix[:tdepth],
				Left:   new,
				Right:  newLeafFromLeaf(node),
				Hash:   hash(new.Hash + node.Hash),
			}
			// as above
			node.prefix = prefixForDepth(depth, new.prefix, nil)
			node.Value = ""
			node.Hash = ""
			node.Left = &Node{
				prefix:      prefixForDepth(depth+1, new.prefix, &zero),
				terminating: true,
			}
			node.Right = &Node{
				prefix:      prefixForDepth(depth+1, new.prefix, &one),
				terminating: true,
			}

			if depth != tdepth {
				return t.insert(node, newNode, depth)
			}

			return newNode
		}
	} else {
		if node.isInternal() {
			node.Right = t.insert(node.Right, new, depth+1)
		} else if node.isTerminating() {
			node.terminating = false
			node.Right = t.insert(node.Right, new, depth+1)
			node.Left = &Node{
				prefix:      prefixForDepth(depth+1, new.prefix, &zero), //node.prefix + "0",
				terminating: true,
			}
		} else {
			// we move the curent one to the side

			newNode := &Node{
				prefix: prefixForDepth(tdepth, new.prefix, nil), //node.prefix[:targetDepth],
				Left:   newLeafFromLeaf(node),
				Right:  new,
				Hash:   hash(node.Hash + new.Hash),
			}
			node.prefix = prefixForDepth(depth, new.prefix, nil) //node.prefix[:depth]
			node.Value = ""
			node.Hash = ""
			node.Left = &Node{
				prefix:      prefixForDepth(depth+1, new.prefix, &zero), //node.prefix + "0",
				terminating: true,
			}
			node.Right = &Node{
				prefix:      prefixForDepth(depth+1, new.prefix, &one), //node.prefix + "1",
				terminating: true,
			}

			if depth != tdepth { //len(newNode.prefix) {
				return t.insert(node, newNode, depth)
			}

			return newNode

			// node.Left = &Node{
			// 	Value:  node.Value,
			// 	Hash:   node.Hash,
			// 	prefix: node.prefix,
			// }
			// node.Value = ""
			// node.prefix = new.Value[:depth]
			// node.Right = t.insert(node.Right, new, depth+1)
		}
	}

	nodeHash := node.Value
	if node.Left != nil {
		nodeHash += node.Left.Hash
	}
	if node.Right != nil {
		nodeHash += node.Right.Hash
	}
	node.Hash = hash(nodeHash)

	return node
}

func (t *BinaryUrkelTrie) Get(key string) (string, bool) {
	binaryKey := toBinaryString(key)
	node := t.Root
	for i := 0; i < len(binaryKey); i++ {
		if node == nil {
			return "", false
		}
		if binaryKey[i] == '0' {
			node = node.Left
		} else {
			node = node.Right
		}
	}
	if node == nil || node.Value == "" {
		return "", false
	}
	return node.Value, true
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
	b := value[(pos-1)/8]
	r := (pos) % 8
	m := (1 << r)
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
				urkel.Insert(hex.EncodeToString(key), hex.EncodeToString(val))
				//urkel = AddToUrkle(CreateLeafithKeyValue(key, val), urkel, 0, 0)

				fmt.Printf("%d: %s %d %d\n", i+cmd.massif.Start.FirstIndex, hex.EncodeToString(key), len(key), math.Ilogb(float64(len(key))*8))

			}
			//fmt.Printf("---- : %s: %s\n", hex.EncodeToString(urkel.min), hex.EncodeToString(urkel.max))
			return nil // fmt.Errorf("'%s' not found", cCtx.String("value"))
		},
	}
}
