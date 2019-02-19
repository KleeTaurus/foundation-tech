package main

import (
	"crypto/sha256"
	"fmt"
	"strconv"
	"time"

	cedar "github.com/iohub/Ahocorasick"
)

const count = 5 * 1000 * 1000 // 500万

func findByKey(key []byte, trie *cedar.Cedar) {
	now := time.Now()
	val, err := trie.Get(key)
	elapsed := time.Since(now)
	if err != nil {
		fmt.Printf("!!!Can not found the key %s\n", string(key))
	} else {
		fmt.Printf("Found the key %s, value %s, took %s\n", string(key), val.(string), elapsed)
	}
}

func hash(preimage []byte) string {
	sum := sha256.Sum256(preimage)
	return fmt.Sprintf("%x", sum)
}

func createTrie() *cedar.Cedar {
	trie := cedar.NewCedar()
	for i := 1; i <= count; i++ {
		if i%500000 == 0 {
			fmt.Printf("There are %d keys have been inserted.\n", i)
		}
		// key: index with string type, value: the hash value of key
		key := strconv.Itoa(i)
		val := hash([]byte(key))
		trie.Insert([]byte(key), val)
	}
	return trie
}

func main() {
	fmt.Printf("Creating a new trie with %d keys.\n", count)
	trie := createTrie()
	fmt.Printf("The trie has been created.\n")

	testingKeys := []string{
		"100101", "127817", "128742", "181343",
		"287333", "xyz297", "applex", "google",
		"i97343", "908282", "827344", "11123123",
		"1024", "2048", "4096", "50238", "123343",
		"860987", "499123",
	}
	for _, k := range testingKeys {
		findByKey([]byte(k), trie)
	}
}
