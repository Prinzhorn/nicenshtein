package nicenshtein

import (
	"bufio"
	"os"
	"strings"
	"unicode/utf8"
)

//A Trie structure that maps runes to a list of following (child-) runes.
//`word` serves two purposes:
//1. If it is not an empty string, it marks the end of a word like a flag
//2. It stores the word that the path to it spells
type RuneNode struct {
	children map[rune]*RuneNode
	word     string
}

type Nicenshtein struct {
	root *RuneNode
}

func NewNicenshtein() Nicenshtein {
	var nice Nicenshtein
	nice.root = &RuneNode{make(map[rune]*RuneNode), ""}

	return nice
}

func (nice *Nicenshtein) IndexFile(filePath string) error {
	file, err := os.Open(filePath)

	if err != nil {
		return err
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		nextWord := strings.TrimSpace(scanner.Text())
		nice.AddWord(nextWord)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func (nice *Nicenshtein) AddWord(word string) {
	if len(word) == 0 {
		return
	}

	var currentNode *RuneNode = nice.root

	for index, runeValue := range word {
		childNode, ok := currentNode.children[runeValue]

		//We have not indexed this rune yet, create a new entry.
		if !ok {
			childNode = &RuneNode{make(map[rune]*RuneNode), ""}
			currentNode.children[runeValue] = childNode
		}

		//The node at the end of a word stores the full word, which also marks the end.
		//This makes the index less memory efficient, but vastly improves query performance.
		//Otherwise each query would need to collect the runes along the path and concat the word.
		if index == len(word)-len(string(runeValue)) {
			childNode.word = word
		}

		currentNode = childNode
	}
}

func (nice *Nicenshtein) ContainsWord(word string) bool {
	var currentNode *RuneNode = nice.root

	for _, runeValue := range word {
		childNode, ok := currentNode.children[runeValue]

		//Current rune not indexed, dead end.
		if !ok {
			return false
		}

		currentNode = childNode
	}

	//Does this node terminate with the given word?
	return currentNode.word == word
}

func (nice *Nicenshtein) collectClosestWords(out *map[string]byte, currentNode *RuneNode, word string, distance byte, maxDistance byte) {
	//We have eated all runes, let's see if we have reached a node with a valid word.
	if len(word) == 0 {
		if currentNode.word != "" {
			knownDistance, ok := (*out)[currentNode.word]

			//We have not seen this word or we have found a smaller distance.
			if !ok || distance < knownDistance {
				(*out)[currentNode.word] = distance
			}
		}

		return
	}

	runeValue, size := utf8.DecodeRuneInString(word)
	wordWithoutFirstRune := word[size:]
	nextNode := currentNode.children[runeValue]

	if nextNode != nil {
		//Move forward by one rune without incrementing the distance.
		//This is just regular Trie walking sans Levenshtein.
		nice.collectClosestWords(out, nextNode, wordWithoutFirstRune, distance, maxDistance)
	}

	//Here we keep walking the Trie, but with a twist.
	//We do each of the Levenshtein edits at the current position
	//and walk the Trie as if nothing cool has happened.
	if distance < maxDistance {
		distance++

		//For substitution and insertion we need to apply them
		//for every rune at the current node.
		for runeValue, _ := range currentNode.children {
			//Substitution (replace the first rune with the current one).
			nice.collectClosestWords(out, currentNode, string(runeValue)+wordWithoutFirstRune, distance, maxDistance)

			//Insertion (add the current rune as prefix).
			nice.collectClosestWords(out, currentNode, string(runeValue)+word, distance, maxDistance)
		}

		//Deletion (skip first rune).
		nice.collectClosestWords(out, currentNode, wordWithoutFirstRune, distance, maxDistance)
	}
}

func (nice *Nicenshtein) CollectClosestWords(out *map[string]byte, word string, maxDistance byte) {
	nice.collectClosestWords(out, nice.root, word, 0, maxDistance)
}
