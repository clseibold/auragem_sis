package islam

// Most used languages on GeminiSpace:
// english, german, russian, french, finnish, spanish, italian
// Languages to add based on this: french, finish

import (
	// "context"
	// "database/sql"
	// "time"
	"fmt"
	"strings"

	// "net/url"
	// "strconv"
	// "unicode/utf8"

	// "github.com/krixano/ponixserver/src/db"

	sis "gitlab.com/clseibold/smallnetinformationservices"
	// "runtime"
	// "golang.org/x/net/html"
	// "github.com/microcosm-cc/bluemonday"
)

type QuranVersionListItem struct {
	Name       string
	Identifier string
}

func HandleIslamicTexts(g sis.ServerHandle) {
	quranSurahs := GetSurahs()

	// en.yusufali, en.pickthall, en.sahih, en.asad, en.sarwar (Shi'a)

	englishQuranVersions := []QuranVersionListItem{
		QuranVersionListItem{"Yusuf Ali", "en.yusufali"},
		QuranVersionListItem{"Pickthall", "en.pickthall"},
		QuranVersionListItem{"Saheeh International", "en.sahih"},
		QuranVersionListItem{"Asad", "en.asad"},
		QuranVersionListItem{"Sarwar", "en.sarwar"},
		QuranVersionListItem{"Shakir", "en.shakir"},
	}

	// Spanish: es.cortes Cortes, es.asad Asad
	spanishQuranVersions := []QuranVersionListItem{
		QuranVersionListItem{"Cortes", "es.cortes"},
		QuranVersionListItem{"Asad", "es.asad"},
	}

	// German: de.aburida Abu Rida, de.bubenheim Bubenheim & Elyas, de.khoury Khoury, de.zaidan Zaidan
	germanQuranVersions := []QuranVersionListItem{
		QuranVersionListItem{"Abu Rida", "de.aburida"},
		QuranVersionListItem{"Bubenheim & Elyas", "de.bubenheim"},
		QuranVersionListItem{"Khoury", "de.khoury"},
		QuranVersionListItem{"Zaidan", "de.zaidan"},
	}

	russianQuranVersions := []QuranVersionListItem{
		QuranVersionListItem{"Elmir Kuliev", "ru.kuliev"},
		QuranVersionListItem{"V. Porokhova", "ru.porokhova"},
	}

	frenchQuranVersions := []QuranVersionListItem{
		QuranVersionListItem{"Hamidullah", "fr.hamidullah"},
	}

	farsiQuranVersions := []QuranVersionListItem{
		QuranVersionListItem{"AbdolMohammad Ayati", "fa.ayati"},
		QuranVersionListItem{"Mohammad Mahdi Fooladvand", "fa.fooladvand"},
		QuranVersionListItem{"Mahdi Elahi Ghomshei", "fa.ghomshei"},
		QuranVersionListItem{"Naser Makarem Shirazi", "fa.makarem"},
		QuranVersionListItem{"Hussain Ansarian", "fa.ansarian"},
		QuranVersionListItem{"Abolfazl Bahrampour", "fa.bahrampour"},
		QuranVersionListItem{"Baha'oddin Khorramshahi", "fa.khorramshahi"},
		QuranVersionListItem{"Sayyed Jalaloddin Mojtabavi", "fa.mojtabavi"},
		QuranVersionListItem{"Mostafa Khorramdel", "fa.khorramdel"},
		QuranVersionListItem{"Mohammad Kazem Moezzi", "fa.moezzi"},
	}

	/*arabicQuranVersions := []QuranVersionListItem {
		QuranVersionListItem { "Abu Rida", "de.aburida" },
		QuranVersionListItem { "Bubenheim & Elyas", "de.bubenheim" },
		QuranVersionListItem { "Khoury", "de.khoury" },
		QuranVersionListItem { "Zaidan", "de.zaidan" },
	}*/

	versionNames := make(map[string]string)
	for _, version := range englishQuranVersions {
		versionNames[version.Identifier] = version.Name
	}
	for _, version := range spanishQuranVersions {
		versionNames[version.Identifier] = version.Name
	}
	for _, version := range germanQuranVersions {
		versionNames[version.Identifier] = version.Name
	}
	for _, version := range russianQuranVersions {
		versionNames[version.Identifier] = version.Name
	}
	for _, version := range frenchQuranVersions {
		versionNames[version.Identifier] = version.Name
	}
	for _, version := range farsiQuranVersions {
		versionNames[version.Identifier] = version.Name
	}

	versionNames["arabic"] = "Qur'an"

	g.AddRoute("/scriptures/islam", func(request sis.Request) {
		var builder strings.Builder
		fmt.Fprintf(&builder, "## Qur'an Versions\n\n=> /scriptures/islam/quran/arabic/ Arabic\n")
		fmt.Fprintf(&builder, "### English\n")
		for _, version := range englishQuranVersions {
			if version.Identifier != "" {
				fmt.Fprintf(&builder, "=> /scriptures/islam/quran/%s/ %s\n", version.Identifier, version.Name)
			} else {
				fmt.Fprintf(&builder, "\n")
			}
		}
		fmt.Fprintf(&builder, "\n### Spanish\n")
		for _, version := range spanishQuranVersions {
			if version.Identifier != "" {
				fmt.Fprintf(&builder, "=> /scriptures/islam/quran/%s/ %s\n", version.Identifier, version.Name)
			} else {
				fmt.Fprintf(&builder, "\n")
			}
		}
		fmt.Fprintf(&builder, "\n### German\n")
		for _, version := range germanQuranVersions {
			if version.Identifier != "" {
				fmt.Fprintf(&builder, "=> /scriptures/islam/quran/%s/ %s\n", version.Identifier, version.Name)
			} else {
				fmt.Fprintf(&builder, "\n")
			}
		}
		fmt.Fprintf(&builder, "\n### Russian\n")
		for _, version := range russianQuranVersions {
			if version.Identifier != "" {
				fmt.Fprintf(&builder, "=> /scriptures/islam/quran/%s/ %s\n", version.Identifier, version.Name)
			} else {
				fmt.Fprintf(&builder, "\n")
			}
		}
		fmt.Fprintf(&builder, "\n### French\n")
		for _, version := range frenchQuranVersions {
			if version.Identifier != "" {
				fmt.Fprintf(&builder, "=> /scriptures/islam/quran/%s/ %s\n", version.Identifier, version.Name)
			} else {
				fmt.Fprintf(&builder, "\n")
			}
		}
		fmt.Fprintf(&builder, "\n### Farsi\n")
		for _, version := range farsiQuranVersions {
			if version.Identifier != "" {
				fmt.Fprintf(&builder, "=> /scriptures/islam/quran/%s/ %s\n", version.Identifier, version.Name)
			} else {
				fmt.Fprintf(&builder, "\n")
			}
		}

		request.Gemini(fmt.Sprintf(`# Islamic Texts

=> https://alquran.cloud/ Powered by Al Quran Cloud

%s

Tags: #quran #qur'an #koran #القرآن 
`, builder.String()))
	})

	g.AddRoute("/scriptures/islam/quran/:version", func(request sis.Request) {
		versionId := request.GetParam("version")
		var builder strings.Builder
		for _, surah := range quranSurahs {
			fmt.Fprintf(&builder, "=> /scriptures/islam/quran/%s/%d/ Surah %d: %s (%s)\n", versionId, surah.Number, surah.Number, surah.EnglishNameTranslation, surah.EnglishName)
		}

		request.Gemini(fmt.Sprintf(`# %s

=> /scriptures/islam Qur'an Versions

%s

=> https://alquran.cloud/ Powered by Al Quran Cloud
`, versionNames[versionId], builder.String()))
	})

	g.AddRoute("/scriptures/islam/quran/:version/:surah", func(request sis.Request) {
		versionId := request.GetParam("version")
		surahNumber := request.GetParam("surah")

		surah := GetSurah(versionId, surahNumber)

		var builder strings.Builder
		for _, ayah := range surah.Ayahs {
			fmt.Fprintf(&builder, "[%d] %s\n", ayah.NumberInSurah, ayah.Text)
		}

		if surah.Number > 1 || surah.Number < len(quranSurahs) {
			fmt.Fprintf(&builder, "\n\n")
		}
		if surah.Number > 1 {
			fmt.Fprintf(&builder, "=> /scriptures/islam/quran/%s/%d/ Previous\n", versionId, surah.Number-1)
		}
		if surah.Number < len(quranSurahs) {
			fmt.Fprintf(&builder, "=> /scriptures/islam/quran/%s/%d/ Next\n", versionId, surah.Number+1)
		}

		request.Gemini(fmt.Sprintf(`# %s, Surah %d: %s (%s)

=> /scriptures/islam/quran/%s/ Surahs

%s

=> https://alquran.cloud/ Powered by Al Quran Cloud
`, versionNames[versionId], surah.Number, surah.EnglishNameTranslation, surah.EnglishName, versionId, builder.String()))
	})
}
