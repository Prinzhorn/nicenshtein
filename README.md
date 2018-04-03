# Nicenshtein

Efficiently index and search a dictionary by Levenshtein distance. This is done by creating a trie (prefix tree) as an index and then walking the trie for collecting all words within a given distance. We keep track of the number of edits that have been made and walk multiple paths at the same time until all edits are consumed.

It is safe to use with utf-8 strings as it uses runes internally.

Check out [nicenshtein-server](https://github.com/Prinzhorn/nicenshtein-server) as well, it has a demo live at [https://nicenshtein.now.sh](https://nicenshtein.now.sh).

# API

## NewNicenshtein()

Returns a new instance of a Nicenshtein index with the following methods:

## IndexFile(filePath string): error

Indexes every single line in the given file using `AddWord`.

## AddWord(word string)

Adds a `word` to the index.

## ContainsWord(word string): bool

Returns whether or not the index contains the given `word`.

## CollectWords(out \*map[string]byte, word string, maxDistance byte)

Will fill `out` (maps words to distances) with all words that are within `maxDistance` of `word`.