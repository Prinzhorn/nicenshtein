package nicenshtein

import (
	"bufio"
	"os"
	"strings"
	"unicode/utf8"
)

//A trie structure that maps runes to a list of following (child-) runes.
//`word` serves two purposes:
//1. If it is not an empty string, it marks the end of a word like a flag
//2. It stores the word that the path to it spells
type RuneNode struct {
	children map[rune]*RuneNode
	length   int
	word     string
}

type Nicenshtein struct {
	root *RuneNode
}

func NewNicenshtein() Nicenshtein {
	var nice Nicenshtein
	nice.root = &RuneNode{make(map[rune]*RuneNode), 0, ""}

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
	numRunes := utf8.RuneCountInString(word)

	if numRunes == 0 {
		return
	}

	currentNode := nice.root
	runeIndex := 0

	for _, runeValue := range word {
		//Store the length of the largest word below this node.
		if numRunes-runeIndex > currentNode.length {
			currentNode.length = numRunes - runeIndex
		}

		childNode, ok := currentNode.children[runeValue]

		//We have not indexed this rune yet, create a new entry.
		if !ok {
			childNode = &RuneNode{make(map[rune]*RuneNode), 0, ""}
			currentNode.children[runeValue] = childNode
		}

		currentNode = childNode
		runeIndex++
	}

	//The node at the end of a word stores the full word, which also marks the end.
	//This makes the index less memory efficient, but vastly improves query performance.
	//Otherwise each query would need to collect the runes along the path and concat the word.
	currentNode.word = word
}

func (nice *Nicenshtein) ContainsWord(word string) bool {
	currentNode := nice.root
	numRunes := utf8.RuneCountInString(word)
	runeIndex := 0

	for _, runeValue := range word {
		if numRunes-runeIndex > currentNode.length {
			return false
		}

		childNode, ok := currentNode.children[runeValue]

		//Current rune not indexed, dead end.
		if !ok {
			return false
		}

		currentNode = childNode
		runeIndex++
	}

	//Does this node terminate with the given word?
	return currentNode.word == word
}

func (nice *Nicenshtein) CollectWords(out *map[string]int, word string, maxDistance int) {
	nice.collectWords(out, nice.root, word, 0, 0, maxDistance)
}

func (nice *Nicenshtein) collectWords(out *map[string]int, currentNode *RuneNode, word string, depth int, distance, maxDistance int) {
	numRunes := utf8.RuneCountInString(word)
	remainingEdits := maxDistance - distance

	//We have eated all runes, let's see if we have reached a node with a valid word.
	if numRunes == 0 {
		//A word does indeed terminate at this node.
		if currentNode.word != "" {
			knownDistance, ok := (*out)[currentNode.word]

			//We have not seen this word or we have found a smaller distance.
			if !ok || distance < knownDistance {
				(*out)[currentNode.word] = distance
			}
		}

		//This catches the case where we have eaten all runes but we can still insert new ones at the end.
		if remainingEdits > 0 {
			for runeValue, _ := range currentNode.children {
				nice.collectWords(out, currentNode, string(runeValue), depth+1, distance+1, maxDistance)
			}
		}

		return
	}

	//There are no words below this node long enough.
	if currentNode.length < numRunes-depth-remainingEdits {
		return
	}

	runeValue, size := utf8.DecodeRuneInString(word)
	wordWithoutFirstRune := word[size:]
	nextNode := currentNode.children[runeValue]

	if nextNode != nil {
		//Move forward by one rune without incrementing the distance.
		//This is just regular trie walking sans Levenshtein.
		nice.collectWords(out, nextNode, wordWithoutFirstRune, depth+1, distance, maxDistance)
	}

	//Here we keep walking the trie, but with a twist.
	//We do each of the Levenshtein edits at the current position
	//and walk the trie as if nothing cool has happened.
	if remainingEdits > 0 {
		distance++

		//For substitution and insertion we need to apply them
		//for every rune at the current node.
		for runeValue, _ := range currentNode.children {
			//Substitution (replace the first rune with the current one).
			nice.collectWords(out, currentNode, string(runeValue)+wordWithoutFirstRune, depth+1, distance, maxDistance)

			//Insertion (add the current rune as prefix).
			nice.collectWords(out, currentNode, string(runeValue)+word, depth+1, distance, maxDistance)
		}

		//Deletion (skip first rune).
		nice.collectWords(out, currentNode, wordWithoutFirstRune, depth+1, distance, maxDistance)
	}
}
