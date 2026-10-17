// Package retrieval picks the instruction passages that best match a decision
// (BM25 over paragraph chunks), for a model given rules per decision instead of
// all of them up front.
package retrieval

import (
	"math"
	"sort"
	"strings"
	"unicode"
)

// TopK is how many passages one decision gets.
const TopK = 4

// BM25 parameters: term-frequency saturation and length normalisation.
const (
	bm25K1 = 1.5
	bm25B  = 0.75
)

// Chunking marks: a paragraph boundary, a heading and a separator line.
const (
	paragraphSep = "\n\n"
	headingMark  = "#"
	separatorRow = "---"
)

// Index is a BM25 index over the chunks of one text.
type Index struct {
	chunks []chunk
	df     map[string]int
	avgLen float64
}

type chunk struct {
	text   string
	tf     map[string]int
	length int
}

// New splits text into chunks of about chunkWords words. A chunk starts at a
// heading or at a paragraph boundary, and never mixes two headings.
func New(text string, chunkWords int) *Index {
	ix := &Index{df: map[string]int{}}
	var cur []string
	words := 0
	flush := func() {
		if len(cur) > 0 {
			ix.add(strings.Join(cur, paragraphSep))
		}
		cur, words = nil, 0
	}
	for _, p := range strings.Split(text, paragraphSep) {
		p = strings.TrimSpace(p)
		n := len(strings.Fields(p))
		switch {
		case p == "" || p == separatorRow:
			continue
		case strings.HasPrefix(p, headingMark):
			flush()
		case words+n > chunkWords:
			flush()
		}
		cur = append(cur, p)
		words += n
	}
	flush()
	ix.finish()
	return ix
}

func (ix *Index) add(text string) {
	tf := map[string]int{}
	terms := tokens(text)
	for _, t := range terms {
		tf[t]++
	}
	for t := range tf {
		ix.df[t]++
	}
	ix.chunks = append(ix.chunks, chunk{text: text, tf: tf, length: len(terms)})
}

func (ix *Index) finish() {
	total := 0
	for _, c := range ix.chunks {
		total += c.length
	}
	if len(ix.chunks) > 0 {
		ix.avgLen = float64(total) / float64(len(ix.chunks))
	}
}

// Top returns the k chunks that score best against query, best first. Chunks
// that share no term with the query are not returned.
func (ix *Index) Top(query string, k int) []string {
	type scored struct {
		i     int
		score float64
	}
	var hits []scored
	terms := unique(tokens(query))
	for i, c := range ix.chunks {
		score := 0.0
		for _, t := range terms {
			f := float64(c.tf[t])
			if f == 0 {
				continue
			}
			idf := math.Log(1 + (float64(len(ix.chunks))-float64(ix.df[t])+0.5)/(float64(ix.df[t])+0.5))
			norm := bm25K1 * (1 - bm25B + bm25B*float64(c.length)/ix.avgLen)
			score += idf * f * (bm25K1 + 1) / (f + norm)
		}
		if score > 0 {
			hits = append(hits, scored{i, score})
		}
	}
	sort.SliceStable(hits, func(a, b int) bool { return hits[a].score > hits[b].score })
	var out []string
	for _, h := range hits[:min(k, len(hits))] {
		out = append(out, ix.chunks[h.i].text)
	}
	return out
}

// tokens lowercases text and splits it on anything that is not a letter or digit.
func tokens(text string) []string {
	return strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
}

func unique(ts []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, t := range ts {
		if !seen[t] {
			seen[t] = true
			out = append(out, t)
		}
	}
	return out
}
