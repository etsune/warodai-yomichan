package main

import (
	"strings"
	"testing"
)

func TestWarodaiToYomiEntries(t *testing.T) {
	var tests = []struct {
		entry              WarodaiEntry
		expectedExpression string
		expectedKana       string
	}{
		{WarodaiEntry{Meanings: []string{}, Header: WarodaiHeader{Kana: "うちあげる", Kanji: "打ち揚げる･打ち上げる"}}, "打ち揚げる", "うちあげる"},
		{WarodaiEntry{Meanings: []string{}, Header: WarodaiHeader{Kana: "ばちゃん, ばちゃんと"}}, "ばちゃん", "ばちゃん"},
		{WarodaiEntry{Meanings: []string{}, Header: WarodaiHeader{Kana: "ろくろくび, ろくろっくび", Kanji: "轆轤首"}}, "轆轤首", "ろくろくび"},
		{WarodaiEntry{Meanings: []string{}, Header: WarodaiHeader{Kana: "あおつづら, あおつづらふじ", Kanji: "青葛･防己･防已･木防已, 青葛藤･木防已"}}, "青葛", "あおつづら"},
	}

	for _, testCase := range tests {
		var result = WarodaiToYomiEntries(&testCase.entry)

		if result[0].Expression != testCase.expectedExpression {
			t.Errorf("Incorrect Expression. Expect %s, got %s", testCase.expectedExpression, result[0].Expression)
		}
		if result[0].Reading != testCase.expectedKana {
			t.Errorf("Incorrect Reading. Expect %s, got %s", testCase.expectedKana, result[0].Reading)
		}
	}
}

func TestCleanEntryMeanings(t *testing.T) {
	var tests = []struct {
		line     []string
		expected []string
	}{
		{[]string{
			"половина;.",
			`рыбий клей; <i>ср.</i> <a href="#005-13-65">にべもない</a>.`,
			"сосуд с маслом <i>(для волос и т. п.)</i>;\r\n<i>тех.</i> масляный резервуар.",
		}, []string{
			"половина",
			"рыбий клей; ср. にべもない",
			"сосуд с маслом (для волос и т. п.)\nтех. масляный резервуар",
		}},
	}

	var cleaner = CreateLineCLeaner()

	for _, testCase := range tests {
		var result = CleanEntryMeanings(testCase.line, &cleaner)

		for i := range result {
			if testCase.expected[i] != result[i] {
				t.Errorf("Incorrect result. Expect %s, got %s", testCase.expected[i], result[i])
			}
		}
	}
}

func TestParseEntry(t *testing.T) {
	var tests = []struct {
		body                 string
		expectedFirstMeaning string
		expectedKanji        string
	}{
		{"てんげん【天元】(тэнгэн)〔006-56-61〕\n1) центр вселенной;\n2) центр доски <i>(для игры в го)</i>.", "центр вселенной;", "天元"},
	}

	for _, testCase := range tests {
		var result = ParseEntry(&testCase.body)

		if result.Meanings[0] != testCase.expectedFirstMeaning {
			t.Errorf("Incorrect Meanings. Expect %s, got %s", testCase.expectedFirstMeaning, result.Meanings[0])
		}

		if result.Header.Kanji != testCase.expectedKanji {
			t.Errorf("Incorrect Kanji. Expect %s, got %s", testCase.expectedKanji, result.Header.Kanji)
		}
	}
}

func TestParseWarodaiMeanings(t *testing.T) {
	var tests = []struct {
		body     string
		expected string
	}{
		{"1) приобретать, покупать;\n2) вербовать за награду.", "приобретать, покупать;|вербовать за награду."},
		{"компенсация, возмещение;", "компенсация, возмещение;"},
		{"закладываемая вещь, заклад, залог;\n～がない закладывать нечего.", "закладываемая вещь, заклад, залог;\n～がない закладывать нечего."},
		{"1) в городе; на территории города\n市中は火の消えたようです;\n2) открытый (вольный) рынок.", "в городе; на территории города\n市中は火の消えたようです;|открытый (вольный) рынок."},
	}

	for _, testCase := range tests {
		var resultArr = ParseWarodaiMeanings(&testCase.body)
		var result = strings.Join(resultArr, "|")

		if result != testCase.expected {
			t.Errorf("Incorrect result. Expect %s, got %s", testCase.expected, result)
		}
	}
}

func TestParseWarodaiHeader(t *testing.T) {
	var tests = []struct {
		header                string
		expectedKana          string
		expectedKanji         string
		expectedTranscription string
		expectedTag           string
		expectedId            string
	}{
		{"さんしゅう【三舟】(сансю:)〔008-87-74〕", "さんしゅう", "三舟", "сансю:", "", "008-87-74"},
		{"きやり, きやりおんど【木遣り, 木遣り音頭】(кияри, кияриондо)〔000-28-75〕", "きやり, きやりおんど", "木遣り, 木遣り音頭", "кияри, кияриондо", "", "000-28-75"},
		{"ヴォーカル・フォア(во:кару-фоа)〔004-53-33〕", "ヴォーカル・フォア", "", "во:кару-фоа", "", "004-53-33"},
		{"ルワンダ(Руванда) [геогр.]〔008-71-47〕", "ルワンダ", "", "Руванда", "геогр.", "008-71-47"},
	}

	for _, testCase := range tests {
		var result = ParseWarodaiHeader(&testCase.header)

		if result.Kana != testCase.expectedKana {
			t.Errorf("Incorrect Kana. Expect %s, got %s", testCase.expectedKana, result.Kana)
		}
		if result.Kanji != testCase.expectedKanji {
			t.Errorf("Incorrect Kanji. Expect %s, got %s", testCase.expectedKanji, result.Kanji)
		}
		if result.Transcription != testCase.expectedTranscription {
			t.Errorf("Incorrect Transcription. Expect %s, got %s", testCase.expectedTranscription, result.Transcription)
		}
		if result.Tag != testCase.expectedTag {
			t.Errorf("Incorrect Tag. Expect %s, got %s", testCase.expectedTag, result.Tag)
		}
		if result.Id != testCase.expectedId {
			t.Errorf("Incorrect Id. Expect %s, got %s", testCase.expectedId, result.Id)
		}
	}
}
