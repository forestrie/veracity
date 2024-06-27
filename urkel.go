//go:build urkel

package veracity

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
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

// isInternal returns true if a node is not a leaf
func (n *Node) isInternal() bool {
	return n.Value == "" && !n.terminating
}

// isTerminating returns true if a node is terminating node
// with a zero hash
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
	nn := &Node{
		prefix: key,
		// this is new leaf - so set the significant bits to all of them
		pdepth: len(key) * 8,
		Value:  value,
		Hash:   hash(value),
	}
	t.Root = t.insert(t.Root, nn, 0)
}

// prefixForDepth returns a new prefix with depth bits from the prefix
// with last significant bit set to lastSignigicant if one is defined
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

// insert inserts new node into a tree
// two leafs that share a common key prefix will end up
// on the depth equal to lengt of the commmon part of key - essentially
// ending up where they would normally meet in the tree
// all the missing branches will be filled in by terminating nodes
// with hash set to all zero
// this implementation follows the design notes in: https://handshake.org/files/handshake.txt
func (t *BinaryUrkelTrie) insert(node *Node, new *Node, depth int) *Node {

	// if we have an nil noce we just replace
	// with new node (this will only be a case for first leaf)
	if node == nil {
		return new
	}

	// if we're inserting a leaf and our node is a terminating node
	// we just replace it
	if !new.isInternal() && !new.isTerminating() && node.isTerminating() {
		return new
	}

	// find out on how many bits we match the current node
	// up to the depth of the node
	tdepth := min(commonPrefix(node.prefix, new.prefix), node.pdepth)

	// which bit to check for left?right
	prefixNext := tdepth
	if node.isTerminating() || node.isInternal() || new.isInternal() || (node == t.Root && node.isInternal()) {
		// if we're not dealing with leafs we want to clamp to current depth
		prefixNext = min(tdepth, depth)
	}

	// check the bit after the common part
	// zeroes go to the left
	left := !bit(prefixNext, new.prefix)

	one := 1
	zero := 0

	if left { //new.prefix[targetDepth] == '0' {
		if node.isInternal() {
			// we hit internal node - we must go deeper
			node.Left = t.insert(node.Left, new, depth+1)
		} else if node.isTerminating() {
			// if we're inserting an internal node
			// add the depth if aligned with our target depth
			// we just replace the terminating node
			if new.isInternal() && depth == new.pdepth {
				return new
			}

			if new.isInternal() {
				// if the deptht is different and we're insetgin internal node
				// we create another terminating node to the left
				node.Left = terminatingNode(prefixForDepth(depth+1, new.prefix, &zero), depth+1)
			}

			// make current node not terminating any more
			// and insert our new node to the left then add terminating one on the right
			node.terminating = false
			node.Left = t.insert(node.Left, new, depth+1)
			node.Right = terminatingNode(prefixForDepth(depth+1, new.prefix, &one), depth+1)
		} else {
			// we have two leafs
			// we need to create an internal node with a target depth
			// of common prefix between our leafs
			newNode := &Node{
				prefix: prefixForDepth(tdepth, new.prefix, nil),
				pdepth: tdepth,
				Left:   new,
				Right:  newLeafFromLeaf(node),
				Hash:   hash(new.Hash + node.Hash),
			}
			// make current node into an internal node with two terminating leafs
			nodeToDoubleTerminating(node, new.prefix, depth)
			// and if the depth does not match the target depth insert our new internal node
			if depth != tdepth {
				return t.insert(node, newNode, depth)
			}

			// otherwise just return new node
			return newNode
		}
	} else {
		// repeat for right side scenario
		if node.isInternal() {
			node.Right = t.insert(node.Right, new, depth+1)
		} else if node.isTerminating() {
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
			newNode := &Node{
				prefix: prefixForDepth(tdepth, new.prefix, nil), //node.prefix[:targetDepth],
				pdepth: tdepth,
				Left:   newLeafFromLeaf(node),
				Right:  new,
				Hash:   hash(node.Hash + new.Hash),
			}

			nodeToDoubleTerminating(node, new.prefix, depth)
			if depth != tdepth {
				return t.insert(node, newNode, depth)
			}

			return newNode
		}
	}

	// if it's terminating node just return as we don't have anything below it
	if node.isTerminating() {
		return node
	}

	// on the way back up the tree
	// we need to update the hashes
	// we combine left + right
	nodeHash := node.Value
	if node.Left != nil {
		nodeHash += node.Left.Hash
	}
	if node.Right != nil {
		nodeHash += node.Right.Hash
	}
	node.Hash = hash(nodeHash)

	// set the depth to current
	node.pdepth = depth
	return node
}

// turn node into a double terminated node
//
//	n     ->     n
//	            / \
//	           x   x
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

// return new terminating node with a prefix and depth target
// and hash of all 0
func terminatingNode(prefix []byte, depth int) *Node {
	return &Node{
		prefix:      prefix,
		pdepth:      depth,
		terminating: true,
		Hash:        "0000000000000000000000000000000000000000000000000000000000000000",
	}
}

// GetPath gets a path for a key
// return list of hex strings for all nodes leading up
// to leaf with specified key - and return value of that leaf
// as well as true if the key was found or false if we hit a terminating node
// or other leaf early
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

	if !bytes.Equal(node.prefix, key) {
		return path, node.Value, false
	}

	return path, node.Value, true
}

// bit returns true if bit at positing is 1, false otherwise
func bit(pos int, value []byte) bool {
	b := value[(pos)/8]
	r := (pos) % 8
	m := 1 << (7 - r)
	return b&byte(m) != 0
}

// returns lengts of common prefix of two byte arrays
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

	return commonBytes*8 + commonBits
}
