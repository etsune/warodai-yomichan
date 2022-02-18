package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	yomi "github.com/etsune/bkrs2yomi/pkg/yomi"
)

type WarodaiHeader struct {
	Kana          string
	Kanji         string
	Transcription string
	Tag           string
	Id            string
}

type WarodaiEntry struct {
	Header   WarodaiHeader
	Meanings []string
}

type lineCLeaner struct {
	bbCodeRx      *regexp.Regexp
	newLineTrimRx *regexp.Regexp
}

var EntryCount, GEntryCount, TermFileIndex = 0, 0, 1
var LineCLeaner lineCLeaner

func CreateLineCLeaner() lineCLeaner {
	return lineCLeaner{
		bbCodeRx:      regexp.MustCompile(`</?(i|a(\shref.+?)?)>`),
		newLineTrimRx: regexp.MustCompile(`[\.;]($|[\n\r]+)`),
	}
}

func StartConverting(dir string) {
	yomi.CreateTempDir()
	LineCLeaner = CreateLineCLeaner()

	title := "Warodai"
	url := "https://www.warodai.ru/"
	description := "Словарь Warodai, compiled with warodai-yomichan"
	rev := time.Now().Format("2006-01-02")
	yomi.CreateIndexFile(rev, title, url, description)

	ProcessAllFiles(dir)
}

func ProcessAllFiles(dir string) {
	var yomiList yomi.YomiTermList

	var walk = func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			fmt.Println(err)
		}

		if !info.IsDir() && filepath.Ext(path) == ".txt" {
			bodyBuf, err := os.ReadFile(path)
			if err != nil {
				log.Fatal(err)
			}
			ProcessRawText(bodyBuf, &yomiList)
			if EntryCount >= 10000 {
				WriteYomiTermFile(&yomiList)
			}
		}
		return nil
	}

	var err = filepath.WalkDir(dir, walk)
	if err != nil {
		fmt.Println(err)
	}

	if EntryCount > 0 {
		WriteYomiTermFile(&yomiList)
	}
}

func WriteYomiTermFile(yomiList *yomi.YomiTermList) {
	yomi.WriteYomiFile(*yomiList, TermFileIndex)
	*yomiList = yomi.YomiTermList{}
	GEntryCount += EntryCount
	TermFileIndex++
	EntryCount = 0
	fmt.Println(GEntryCount)
}

func ProcessRawText(bodyRaw []byte, yomiList *yomi.YomiTermList) {
	var body = string(bodyRaw)
	var entry = ParseEntry(&body)
	var yomiEntries = WarodaiToYomiEntries(&entry)
	*yomiList = append(*yomiList, yomiEntries...)
	EntryCount += len(yomiEntries)
}

func CleanEntryMeanings(lines []string, cleaner *lineCLeaner) []string {
	var result []string
	for i := range lines {
		var curLine = lines[i]
		curLine = strings.Trim(curLine, ";.")
		curLine = strings.TrimSpace(curLine)
		curLine = cleaner.bbCodeRx.ReplaceAllString(curLine, "")
		curLine = cleaner.newLineTrimRx.ReplaceAllString(curLine, "\n")
		result = append(result, curLine)
	}

	return result
}

func WarodaiToYomiEntries(entry *WarodaiEntry) yomi.YomiTermList {
	var readings = strings.Split(entry.Header.Kana, ",")
	var kanjis = strings.Split(entry.Header.Kanji, ",")
	var result yomi.YomiTermList

	if entry.Header.Kanji != "" && len(readings) != len(kanjis) && len(kanjis) != 1 {
		fmt.Printf("Wrong entry header %s, %s\n", entry.Header.Kana, entry.Header.Kanji)
		return result
	}

	var meanings = CleanEntryMeanings(entry.Meanings, &LineCLeaner)

	for i := range readings {
		var curKana = strings.TrimSpace(readings[i])

		if entry.Header.Kanji != "" && len(readings) == len(kanjis) {
			// каждому чтению соответствует написание
			// きやり, きやりおんど【木遣り, 木遣り音頭】
			var curKanjis = strings.Split(kanjis[i], "･")
			for _, j := range curKanjis {
				result = append(result, yomi.YomiTerm{
					Expression: strings.TrimSpace(j),
					Reading:    curKana,
					Glossary:   meanings,
				})
			}
		} else if entry.Header.Kanji != "" && len(kanjis) == 1 {
			// два+ чтения, одно написание
			// ろくろくび, ろくろっくび【轆轤首】
			result = append(result, yomi.YomiTerm{
				Expression: entry.Header.Kanji,
				Reading:    curKana,
				Glossary:   meanings,
			})

		} else {
			// нет написаний
			// ばちゃん, ばちゃんと
			result = append(result, yomi.YomiTerm{
				Expression: curKana,
				Reading:    curKana,
				Glossary:   meanings,
			})
		}

	}

	return result
}

func ParseEntry(entry *string) WarodaiEntry {
	var split = strings.SplitN(*entry, "\n", 2)
	if len(split) != 2 {
		fmt.Println("Wrong entry, " + *entry)
	}

	var header, body = split[0], split[1]

	var warodaiEntry = WarodaiEntry{
		Header:   ParseWarodaiHeader(&header),
		Meanings: ParseWarodaiMeanings(&body),
	}

	return warodaiEntry
}

func ParseWarodaiMeanings(body *string) []string {
	entrySplitRx := regexp.MustCompile(`(^|[\r\n]+)\d+\)`)
	var meanings = entrySplitRx.Split(*body, -1)
	var result []string
	for i := range meanings {
		var line = meanings[i]
		line = strings.TrimSpace(line)

		if len(line) == 0 {
			continue
		}
		result = append(result, line)
	}
	return result
}

type HeaderFragmentType int

const (
	_ HeaderFragmentType = iota
	Kana
	Kanji
	Transcription
	Tag
	Id
)

func ParseWarodaiHeader(header *string) WarodaiHeader {
	var result = WarodaiHeader{}
	var currentFragType HeaderFragmentType = Kana

	for _, ch := range *header {
		switch ch {
		case '【':
			currentFragType = Kanji
		case '(':
			currentFragType = Transcription
		case '[':
			currentFragType = Tag
		case '〔':
			currentFragType = Id
		default:
			if strings.ContainsRune("】)〕]", ch) {
				continue
			}
			switch currentFragType {
			case Kana:
				result.Kana += string(ch)
			case Kanji:
				result.Kanji += string(ch)
			case Transcription:
				result.Transcription += string(ch)
			case Tag:
				result.Tag += string(ch)
			case Id:
				result.Id += string(ch)
			}
		}
	}

	result.Kana = strings.TrimSpace(result.Kana)
	result.Kanji = strings.TrimSpace(result.Kanji)
	result.Tag = strings.TrimSpace(result.Tag)
	result.Transcription = strings.TrimSpace(result.Transcription)
	result.Id = strings.TrimSpace(result.Id)

	return result
}
