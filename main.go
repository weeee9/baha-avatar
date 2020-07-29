package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
)

var (
	lastReqTime = time.Now()
	avatarBytes []byte
	kirito404   []byte
	err         error
)

const (
	postURL = `https://forum.gamer.com.tw/C.php?page=81000&bsn=60076&snA=5037743`
	kirito  = `https://imgur.com/6hBAhdf`
)

func main() {
	router := gin.Default()
	router.Static("/img", "./img")

	router.GET("/avatar.png", renderAvatar)

	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}

	router.Run(":" + port)
}

func renderAvatar(c *gin.Context) {
	now := time.Now()

	if now.Sub(lastReqTime) < 5*time.Second && len(avatarBytes) != 0 {
		log.Println("request time < 3s")
		size := len(avatarBytes)
		c.Header("Content-Length", strconv.Itoa(size))
		c.Data(http.StatusOK, "image/png", avatarBytes)
	} else {
		lastReqTime = now
		resp, err := http.Get(postURL)
		if err != nil {
			log.Println(err)
			c.Redirect(http.StatusTemporaryRedirect, kirito)
			return
		}
		defer resp.Body.Close()

		dom, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			log.Println(err)
			c.Redirect(http.StatusTemporaryRedirect, kirito)
			return
		}
		userIDs := dom.Find(".userid")
		// get last user id from userIDs
		lastUserID := userIDs.Eq(userIDs.Length() - 1).Text()
		avatarURL := getAvatarURLByUserID(lastUserID)

		avatarBytes, err = downloadAvatar(avatarURL)
		if err != nil {
			log.Println(err)
			c.Redirect(http.StatusTemporaryRedirect, kirito)
			return
		}
		size := len(avatarBytes)

		c.Header("Content-Length", strconv.Itoa(size))
		c.Data(http.StatusOK, "image/png", avatarBytes)
	}
}

func getAvatarURLByUserID(userID string) string {
	userID = strings.ToLower(userID)
	firstLetter := string(userID[0])
	secondLetter := string(userID[1])
	url := fmt.Sprintf("https://avatar2.bahamut.com.tw/avataruserpic/%s/%s/%s/%s.png", firstLetter, secondLetter, userID, userID)
	return url
}

func downloadAvatar(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return b, nil
}
