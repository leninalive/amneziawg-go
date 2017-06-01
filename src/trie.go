package main

/* Binary trie
 *
 * Syncronization done seperatly
 * See: routing.go
 *
 * Todo: Better commenting
 */

type Trie struct {
	cidr  uint
	child [2]*Trie
	bits  []byte
	peer  *Peer

	// Index of "branching" bit
	bit_at_byte  uint
	bit_at_shift uint
}

/* Finds length of matching prefix
 * Maybe there is a faster way
 *
 * Assumption: len(s1) == len(s2)
 */
func commonBits(s1 []byte, s2 []byte) uint {
	var i uint
	size := uint(len(s1))
	for i = 0; i < size; i += 1 {
		v := s1[i] ^ s2[i]
		if v != 0 {
			v >>= 1
			if v == 0 {
				return i*8 + 7
			}

			v >>= 1
			if v == 0 {
				return i*8 + 6
			}

			v >>= 1
			if v == 0 {
				return i*8 + 5
			}

			v >>= 1
			if v == 0 {
				return i*8 + 4
			}

			v >>= 1
			if v == 0 {
				return i*8 + 3
			}

			v >>= 1
			if v == 0 {
				return i*8 + 2
			}

			v >>= 1
			if v == 0 {
				return i*8 + 1
			}
			return i * 8
		}
	}
	return i * 8
}

func (node *Trie) RemovePeer(p *Peer) *Trie {
	if node == nil {
		return node
	}

	// Walk recursivly

	node.child[0] = node.child[0].RemovePeer(p)
	node.child[1] = node.child[1].RemovePeer(p)

	if node.peer != p {
		return node
	}

	// Remove peer & merge

	node.peer = nil
	if node.child[0] == nil {
		return node.child[1]
	}
	return node.child[0]
}

func (node *Trie) choose(key []byte) byte {
	return (key[node.bit_at_byte] >> node.bit_at_shift) & 1
}

func (node *Trie) Insert(key []byte, cidr uint, peer *Peer) *Trie {

	// At leaf

	if node == nil {
		return &Trie{
			bits:         key,
			peer:         peer,
			cidr:         cidr,
			bit_at_byte:  cidr / 8,
			bit_at_shift: 7 - (cidr % 8),
		}
	}

	// Traverse deeper

	common := commonBits(node.bits, key)
	if node.cidr <= cidr && common >= node.cidr {
		if node.cidr == cidr {
			node.peer = peer
			return node
		}
		bit := node.choose(key)
		node.child[bit] = node.child[bit].Insert(key, cidr, peer)
		return node
	}

	// Split node

	newNode := &Trie{
		bits:         key,
		peer:         peer,
		cidr:         cidr,
		bit_at_byte:  cidr / 8,
		bit_at_shift: 7 - (cidr % 8),
	}

	cidr = min(cidr, common)

	// Check for shorter prefix

	if newNode.cidr == cidr {
		bit := newNode.choose(node.bits)
		newNode.child[bit] = node
		return newNode
	}

	// Create new parent for node & newNode

	parent := &Trie{
		bits:         key,
		peer:         nil,
		cidr:         cidr,
		bit_at_byte:  cidr / 8,
		bit_at_shift: 7 - (cidr % 8),
	}

	bit := parent.choose(key)
	parent.child[bit] = newNode
	parent.child[bit^1] = node

	return parent
}

func (node *Trie) Lookup(key []byte) *Peer {
	var found *Peer
	size := uint(len(key))
	for node != nil && commonBits(node.bits, key) >= node.cidr {
		if node.peer != nil {
			found = node.peer
		}
		if node.bit_at_byte == size {
			break
		}
		bit := node.choose(key)
		node = node.child[bit]
	}
	return found
}

func (node *Trie) Count() uint {
	if node == nil {
		return 0
	}
	l := node.child[0].Count()
	r := node.child[1].Count()
	return l + r
}
