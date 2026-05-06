package engine

type Rule struct {
	Name string
}

type Hit struct {
	Rule      Rule
	Offset    int // byte offset of first token in source
	EndOffset int // byte offset just past last token
	Len       int // number of tokens consumed
}

type trieNode struct {
	children map[string]*trieNode
	rule     *Rule
}

type Trie struct {
	root trieNode
}

func (t *Trie) Add(tokens []string, rule Rule) {
	node := &t.root
	for _, tok := range tokens {
		if node.children == nil {
			node.children = make(map[string]*trieNode)
		}
		child, ok := node.children[tok]
		if !ok {
			child = &trieNode{}
			node.children[tok] = child
		}
		node = child
	}
	node.rule = &rule
}

// Match walks the trie from tokens[start] and returns the longest match.
func (t *Trie) Match(tokens []Token, start int) (Rule, int, bool) {
	node := &t.root
	last := Rule{}
	lastLen := 0
	found := false
	for i := start; i < len(tokens); i++ {
		child, ok := node.children[tokens[i].Text]
		if !ok {
			break
		}
		node = child
		if node.rule != nil {
			last = *node.rule
			lastLen = i - start + 1
			found = true
		}
	}
	return last, lastLen, found
}

// Scan finds all non-overlapping longest matches in tokens, in order.
func (t *Trie) Scan(tokens []Token) []Hit {
	var hits []Hit
	i := 0
	for i < len(tokens) {
		rule, n, ok := t.Match(tokens, i)
		if ok {
			last := tokens[i+n-1]
			hits = append(hits, Hit{
				Rule:      rule,
				Offset:    tokens[i].Offset,
				EndOffset: last.Offset + last.Len,
				Len:       n,
			})
			i += n
		} else {
			i++
		}
	}
	return hits
}
