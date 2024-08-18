package lib

import (
	"time"

	"github.com/go-shiori/go-readability"
)

func GetArticleContent(url string) (string, error) {
	article, err := readability.FromURL(url, 30*time.Second)
	if err != nil {
		return "", err
	}
	return article.TextContent, nil
}
