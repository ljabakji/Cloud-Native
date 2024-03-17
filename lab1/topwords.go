// Find the top K most common words in a text document.
// Input path: location of the document, K top words
// Output: Slice of top K words
// For this excercise, word is defined as characters separated by a whitespace

// Note: You should use `checkError` to handle potential errors.

package textproc

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
)

func topWords(path string, K int) []WordCount {
	// Your code here.....

	file, err := os.Open(path)
	checkError(err)
	defer file.Close()

	//Create a map to store word occurrences
	wordCount := make(map[string]int)

	// Read the file line by line
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		words := strings.Fields(line) // Split the line into words

		// Count occurences of each word
		for _, word := range words {
			word = strings.ToLower(word)
			wordCount[word]++
		}
	}

	//Check for errors during scanning
	checkError(scanner.Err())

	//Create a slice to store word-count pairs
	var wordCountList []WordCount

	//Populate the slice with word-count pairs
	for word, count := range wordCount {
		wordCountList = append(wordCountList, WordCount{Word: word, Count: count})
	}

	//Sort the slice based on count and word
	sortWordCounts(wordCountList)

	// Return the top K occurrences
	if K > len(wordCountList) {
		K = len(wordCountList)
	}

	return wordCountList[:K]

}

//--------------- DO NOT MODIFY----------------!

// A struct that represents how many times a word is observed in a document
type WordCount struct {
	Word  string
	Count int
}

// Method to convert struct to string format
func (wc WordCount) String() string {
	return fmt.Sprintf("%v: %v", wc.Word, wc.Count)
}

// Helper function to sort a list of word counts in place.
// This sorts by the count in decreasing order, breaking ties using the word.

func sortWordCounts(wordCounts []WordCount) {
	sort.Slice(wordCounts, func(i, j int) bool {
		wc1 := wordCounts[i]
		wc2 := wordCounts[j]
		if wc1.Count == wc2.Count {
			return wc1.Word < wc2.Word
		}
		return wc1.Count > wc2.Count
	})
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
