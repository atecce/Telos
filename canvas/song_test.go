package canvas

import (
	"io/ioutil"
	"log"
	"strings"
	"sync"
	"testing"
)

var files = map[string]string{
	"Present Tense":       "song_tests/present_tense.txt",
	"Hey Mami":            "song_tests/hey_mami.txt",
	"Dance, Dance, Dance": "song_tests/dance_dance_dance.txt",
}

func readFile(name string) string {
	data, err := ioutil.ReadFile(name)
	if err != nil {
		log.Fatal("failed to read file", name)
	}
	return strings.Trim(string(data), "\n")
}

func TestParse(t *testing.T) {

	songs := []Song{
		{
			Album: &Album{
				Name: "A Moon Shaped Pool",
			},
			Url:  "https://www.lyrics.com/lyric/32911894/Present+Tense",
			Name: "Present Tense",
		},
		{
			Album: &Album{
				Name: "Sylvan Esso",
			},
			Url:  "https://www.lyrics.com/lyric/30800248/Sylvan+Esso/Hey+Mami",
			Name: "Hey Mami",
		},
		{
			Album: &Album{
				Name: "Lykke Li",
			},
			Url:  "https://www.lyrics.com/lyric/13943003/Lykke+Li/Dance%2C+Dance%2C+Dance",
			Name: "Dance, Dance, Dance",
		},
	}

	wg := new(sync.WaitGroup)
	for _, song := range songs {
		wg.Add(1)
		song.Parse(wg)

		lyrics := readFile(files[song.Name])
		if lyrics != song.Lyrics {
			t.Error("failed to parse", song.Name, "expected:\n", []byte(lyrics), "\ngot:\n", []byte(song.Lyrics))
		}
	}
}
