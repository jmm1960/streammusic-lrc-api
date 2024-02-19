package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

var neteaseCloudMusicAPIHost = flag.String("ncmapi", "http://localhost:3000", "NeteaseCloudMusicAPI Host")

func main() {
	flag.Parse()
	nc := &NeteaseCloudMusicAPIClient{Host: *neteaseCloudMusicAPIHost}
	http.HandleFunc("/lyrics", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		title := query.Get("title")
		artist := query.Get("artist")
		album := query.Get("album")
		path := query.Get("path")
		log.Println("request lyric", "title:", title, "artist:", artist, "album:", album, "path:", path)
		if path != "" {
			if localLyric := readLocalLyric(path); localLyric != nil {
				log.Println("local lyric found", "path:", path)
				w.Write(localLyric)
				return
			}
			log.Println("local lyric not found", "path:", path)
		}
		fmt.Fprintf(w, "%s - %s - %s", title, artist, album)
		keywords := fmt.Sprintf("%s %s %s", title, artist, album)
		sr, err := nc.Search(keywords)
		if err != nil {
			log.Println("request search fail", "title:", title, "err:", err)
			w.WriteHeader(http.StatusInternalServerError)
		} else if sr.SongCount == 0 {
			w.Write([]byte{})
		} else {
			song := sr.Songs[0]
			songId := song.Id
			log.Println("request lyric", "title:", song.Name, "album:", song.Album.Name, "songId:", songId)
			lr, err := nc.Lyric(songId)
			if err != nil {
				log.Println("request lyric fail", "title:", song.Name, "album:", song.Album.Name, "songId:", songId, "err:", err)
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				w.Write([]byte(lr.Lrc.Lyric))
			}
		}
	})
	http.ListenAndServe(":8092", nil)
}

func readLocalLyric(musicPath string) []byte {
	bashMusicPath := strings.TrimSuffix(musicPath, filepath.Ext(musicPath))
	lrcPath := bashMusicPath + ".lrc"
	_, err := os.Stat(lrcPath)
	if err == nil {
		file, err := os.ReadFile(lrcPath)
		if err == nil {
			return file
		}
	}
	txtLyricPath := bashMusicPath + ".txt"
	_, err = os.Stat(txtLyricPath)
	if err == nil {
		file, err := os.ReadFile(txtLyricPath)
		if err == nil {
			return file
		}
	}
	return nil
}

type NeteaseCloudMusicAPIClient struct {
	Host string
}

func (c *NeteaseCloudMusicAPIClient) Lyric(songId int64) (lr LyricResponse, err error) {
	values := url.Values{}
	values.Add("id", fmt.Sprintf("%d", songId))
	endpoint := "/lyric"
	resp, err := http.Get(fmt.Sprintf("%s%s?%s", c.Host, endpoint, values.Encode()))
	if err != nil {
		return lr, err
	}
	if resp.StatusCode != http.StatusOK {
		return lr, fmt.Errorf("status not ok: %d", resp.StatusCode)
	}
	all, err := io.ReadAll(resp.Body)
	if err != nil {
		return lr, err
	}
	var rlt RawResult
	err = json.Unmarshal(all, &rlt)
	if err != nil {
		return lr, err
	}
	if rlt.Code != 200 {
		return lr, fmt.Errorf("rlt not ok: %d", rlt.Code)
	}
	err = json.Unmarshal(all, &lr)
	return lr, err
}

type LyricResponse struct {
	Uncollected bool `json:"uncollected"`
	Sgc         bool `json:"sgc"`
	Sfy         bool `json:"sfy"`
	Qfy         bool `json:"qfy"`
	LyricUser   struct {
		Id       int    `json:"id"`
		Status   int    `json:"status"`
		Demand   int    `json:"demand"`
		Userid   int    `json:"userid"`
		Nickname string `json:"nickname"`
		Uptime   int64  `json:"uptime"`
	} `json:"lyricUser"`
	Lrc struct {
		Version int    `json:"version"`
		Lyric   string `json:"lyric"`
	} `json:"lrc"`
	Klyric struct {
		Version int    `json:"version"`
		Lyric   string `json:"lyric"`
	} `json:"klyric"`
	Tlyric struct {
		Version int    `json:"version"`
		Lyric   string `json:"lyric"`
	} `json:"tlyric"`
	Romalrc struct {
		Version int    `json:"version"`
		Lyric   string `json:"lyric"`
	} `json:"romalrc"`
}

func (c *NeteaseCloudMusicAPIClient) Search(keywords string) (sr SearchResponse, err error) {
	values := url.Values{}
	values.Add("keywords", keywords)
	endpoint := "/search"
	resp, err := http.Get(fmt.Sprintf("%s%s?%s", c.Host, endpoint, values.Encode()))
	if err != nil {
		return sr, err
	}
	if resp.StatusCode != http.StatusOK {
		return sr, fmt.Errorf("status not ok: %d", resp.StatusCode)
	}
	all, err := io.ReadAll(resp.Body)
	if err != nil {
		return sr, err
	}
	var rlt RawResultPage
	err = json.Unmarshal(all, &rlt)
	if err != nil {
		return sr, err
	}
	if rlt.Code != 200 {
		return sr, fmt.Errorf("rlt not ok: %d", rlt.Code)
	}
	err = json.Unmarshal(rlt.Result, &sr)
	return sr, err
}

type SearchResponse struct {
	Songs []struct {
		Id      int64  `json:"id"`
		Name    string `json:"name"`
		Artists []struct {
			Id        int           `json:"id"`
			Name      string        `json:"name"`
			PicUrl    interface{}   `json:"picUrl"`
			Alias     []interface{} `json:"alias"`
			AlbumSize int           `json:"albumSize"`
			PicId     int           `json:"picId"`
			FansGroup interface{}   `json:"fansGroup"`
			Img1V1Url string        `json:"img1v1Url"`
			Img1V1    int           `json:"img1v1"`
			Trans     interface{}   `json:"trans"`
		} `json:"artists"`
		Album struct {
			Id     int    `json:"id"`
			Name   string `json:"name"`
			Artist struct {
				Id        int           `json:"id"`
				Name      string        `json:"name"`
				PicUrl    interface{}   `json:"picUrl"`
				Alias     []interface{} `json:"alias"`
				AlbumSize int           `json:"albumSize"`
				PicId     int           `json:"picId"`
				FansGroup interface{}   `json:"fansGroup"`
				Img1V1Url string        `json:"img1v1Url"`
				Img1V1    int           `json:"img1v1"`
				Trans     interface{}   `json:"trans"`
			} `json:"artist"`
			PublishTime int64    `json:"publishTime"`
			Size        int      `json:"size"`
			CopyrightId int      `json:"copyrightId"`
			Status      int      `json:"status"`
			PicId       int64    `json:"picId"`
			Mark        int      `json:"mark"`
			Alia        []string `json:"alia,omitempty"`
			TransNames  []string `json:"transNames,omitempty"`
		} `json:"album"`
		Duration    int         `json:"duration"`
		CopyrightId int         `json:"copyrightId"`
		Status      int         `json:"status"`
		Alias       []string    `json:"alias"`
		Rtype       int         `json:"rtype"`
		Ftype       int         `json:"ftype"`
		Mvid        int         `json:"mvid"`
		Fee         int         `json:"fee"`
		RUrl        interface{} `json:"rUrl"`
		Mark        int         `json:"mark"`
		TransNames  []string    `json:"transNames,omitempty"`
	} `json:"songs"`
	HasMore   bool `json:"hasMore"`
	SongCount int  `json:"songCount"`
}

type RawResultPage struct {
	Result json.RawMessage `json:"result"`
	Code   int             `json:"code"`
}

type RawResult struct {
	Code int `json:"code"`
}
