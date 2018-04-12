package nicenshtein

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"testing"
)

func TestContainsWord(t *testing.T) {
	nice := NewNicenshtein()

	nice.AddWord("Prinzhorn")

	if !nice.ContainsWord("Prinzhorn") {
		t.Error("Should contain Prinzhorn")
	}

	//Prefix
	if nice.ContainsWord("Prinz") {
		t.Error("Should not contain Prinz")
	}

	//Suffix
	if nice.ContainsWord("horn") {
		t.Error("Should not contain horn")
	}

	//Case sensitive
	if nice.ContainsWord("prinzhorn") {
		t.Error("Should not contain prinzhorn")
	}

	//Diacrit (runes / utf-8)
	if nice.ContainsWord("Prinzhôrn") {
		t.Error("Should not contain Prinzhôrn")
	}
}

func TestCollectWords(t *testing.T) {
	nice := NewNicenshtein()

	nice.AddWord("Prinzhorn")
	nice.AddWord("prinzhorn")
	nice.AddWord("Crème fraîche")
	nice.AddWord("👻💩💩👻")

	closestWords := make(map[string]int)

	nice.CollectWords(&closestWords, "Prinzhorn", 2)

	if !reflect.DeepEqual(closestWords, map[string]int{"Prinzhorn": 0, "prinzhorn": 1}) {
		t.Error("Prinzhorn + prinzhorn not found")
	}

	closestWords = make(map[string]int)

	nice.CollectWords(&closestWords, "Creme fraîche", 2)

	if !reflect.DeepEqual(closestWords, map[string]int{"Crème fraîche": 1}) {
		t.Error("Crème fraîche not found")
	}

	closestWords = make(map[string]int)

	nice.CollectWords(&closestWords, "👻💩💩💩👻", 2)

	if !reflect.DeepEqual(closestWords, map[string]int{"👻💩💩👻": 1}) {
		t.Error("👻💩💩👻 not found")
	}
}

func TestIndexFile(t *testing.T) {
	nice := NewNicenshtein()

	nice.IndexFile("./password.1.txt")

	closestWords := make(map[string]int)

	nice.CollectWords(&closestWords, "password", 1)

	if !nice.ContainsWord("password1") {
		t.Error("Should contain password1")
	}

	//81 is the number of passwords in password.1.txt.
	//They are all within a distance of 1 to "password".
	if len(closestWords) != 81 {
		t.Error(fmt.Sprintf("Should return 81 results for \"password\", got %d", len(closestWords)))
	}
}

func randString() string {
	return base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(rand.Int())))
}

func prepareIndex(nice *Nicenshtein) {
	//For benchmarking we want deterministic values.
	rand.Seed(1)

	for i := 0; i < 10000; i++ {
		nice.AddWord(randString())
	}
}

func BenchmarkAddWord(b *testing.B) {
	nice := NewNicenshtein()

	prepareIndex(&nice)

	word := randString()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		nice.AddWord(word)
	}
}

func BenchmarkContainsWord(b *testing.B) {
	nice := NewNicenshtein()

	prepareIndex(&nice)

	word := randString()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		nice.ContainsWord(word)
	}
}

func BenchmarkCollectWords(b *testing.B) {
	nice := NewNicenshtein()

	prepareIndex(&nice)

	word := randString()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		out := make(map[string]int)
		nice.CollectWords(&out, word, 3)
	}
}
