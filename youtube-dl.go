package ytdlgo

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/tidwall/gjson"
)

type ytdl struct {
	exec                    string
	max_songs_from_playlist int
}

func YoutubeDL() ytdl {
	return ytdl{
		exec:                    "youtube-dl",
		max_songs_from_playlist: 20,
	}
}

type Song struct {
	Title    string
	Duration int
	Url      string
	Stream   string
	Thumb    string
}

func (song Song) String() string {
	return fmt.Sprintf("Song{\n\tTitle: %v\n\tDuration: %v\n\tUrl: %v\n\tStream: %v\n\tThumb: %v\n}\n",
		song.Title, song.Duration, song.Url, song.Stream, song.Thumb)
}

// run youtube-dl with args
func (ytdl *ytdl) runYoutubeDL(args ...string) (string, error) {
	cmd := exec.Command(ytdl.exec, args...)
	out, err := cmd.Output()
	return string(out), err
}

// search youtube for name and return number of songs
func (ytdl *ytdl) YoutubeSearch(name string, number int) (songs []Song) {
	out, err := ytdl.runYoutubeDL(
		fmt.Sprintf("ytsearch%v:%v", number, name),
		"-x", "-j", "-f bestaudio")
	if err != nil {
		return nil
	}

	jsons := strings.Split(out, "\n")
	for _, json := range jsons {
		song := ytdl.SongFromJson(json)
		if (Song{}) == song {
			continue
		}

		songs = append(songs, song)
	}
	return
}

func (ytdl *ytdl) SongFromJson(json string) (song Song) {
	if json == "" {
		return Song{}
	}

	title := gjson.Get(json, "title")
	if !title.Exists() {
		return Song{}
	}

	duration_str := gjson.Get(json, "duration")
	if !duration_str.Exists() {
		return Song{}
	}
	duration, _ := strconv.Atoi(fmt.Sprint(duration_str))

	id := gjson.Get(json, "id")
	if !id.Exists() {
		return Song{}
	}

	stream := gjson.Get(json, "url")
	if !stream.Exists() {
		return Song{}
	}

	thumb := gjson.Get(json, "thumbnail")
	if !thumb.Exists() {
		return Song{}
	}

	song = Song{
		Title:    fmt.Sprint(title),
		Duration: duration,
		Url:      fmt.Sprintf("https://www.youtube.com/watch?v=%v", id),
		Stream:   fmt.Sprint(stream),
		Thumb:    fmt.Sprint(thumb),
	}
	return
}

// return json from url/name
func (ytdl *ytdl) json(search string, url bool) (json string) {
	var out string
	var err error

	if url {
		out, err = ytdl.runYoutubeDL(search, "-j", "-x", "-f bestaudio")
	} else {
		out, err = ytdl.runYoutubeDL(
			fmt.Sprintf("ytsearch1:%v", search), "-j", "-x", "-f bestaudio")
	}

	if err != nil {
		return ""
	}

	return out
}

func (ytdl *ytdl) SongsFromURL(url string) (songs []Song) {
	json := ytdl.json(url, true)
	if json == "" {
		return []Song{}
	}

	jsons := strings.Split(json, "\n")
	if len(jsons) == 1 {
		return []Song{ytdl.SongFromJson(json)}
	}

	count := len(jsons)
	if count > ytdl.max_songs_from_playlist {
		count = ytdl.max_songs_from_playlist
	}

	for _, json := range jsons[:count] {
		if json == "" {
			continue
		}
		songs = append(songs, ytdl.SongFromJson(json))
	}

	return
}

func (ytdl *ytdl) SongFromName(name string) (song Song) {
	json := ytdl.json(name, false)
	if json == "" {
		return Song{}
	}

	return ytdl.SongFromJson(json)
}
