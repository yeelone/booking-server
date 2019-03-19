package service

import (
	"booking/models"
	"fmt"
	"strings"

	"github.com/juju/utils/set"
	"github.com/lexkong/log"
	"gopkg.in/jdkato/prose.v2"
)

func ngrams(s []string, n int) set.Strings {
	result := make(set.Strings)
	for i := 0; i < len(s)-n+1; i++ {
		result.Add(strings.Join(s[i:i+n], " "))
	}
	return result
}

func Ngrams(content string) (dicts []models.Dictionary,err error) {
	log.Info("Ngrams function called.")


	doc, err := prose.NewDocument(content)
	if err != nil {
		log.Fatal("prose new document occue ", err)
	}

	phrases := []string{}

	// Iterate over the doc's sentences:
	for _, sent := range doc.Sentences() {
		subSent := strings.Split(sent.Text, " ")
		for _, n := range []int{1,2, 3, 4, 5, 6, 7, 8} {
			for key := range ngrams(subSent, n) {
				if strings.Contains(key, ",") {
					continue
				}

				if strings.Contains(key, "\"") {
					continue
				}

				if strings.Contains(key, "'") {
					key = strings.Replace(key, "'", "''", -1)
				}

				if strings.Contains(key, ".") {
					key = strings.Replace(key, ".", "", -1)
				}
				phrases = append(phrases, key)
			}
		}

	}

	limit := 5000
	count := 0
	for i := range phrases {
		if count == limit || i == len(phrases)-1 {
			count = 0
			dict , err := models.QueryPhrases(phrases[i-limit : i])

			if err != nil {
				fmt.Println("err", err)
				continue
			}

			for _, d := range dict {
				singleWord := strings.Replace(d.Word, " ", "", -1) //为了把类似于 be \ an \ ok 之类的两个字符的词汇过滤掉
				if d.Tag == "zk" || d.Tag == "gk" || d.Tag == "zk gk" || len(singleWord) == 2 {
					//忽略中考和高考的词汇,词长为2或者3的单词也排除，太简单的没什么必要
					continue
				}
				dicts = append(dicts, d)
			}
		}
		count++
	}

	return dicts,nil
}
