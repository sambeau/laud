package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	_ "github.com/joho/godotenv/autoload"
	"github.com/supabase-community/supabase-go"
)

const version = 0.6

const baseBookUrl = "https://www.audible.co.uk/pd/"
const baseSearchUrl = "https://www.audible.co.uk/search?"

// exponentially dole out fewer points from spot 1–100
// 0.9794 spreads 250 points over the top 300 places
// chosen because it gives a nice curve at the top
const popularityFactor = 0.9794
const popularityTopScore = 250

// these 2 numbers multiplied shouldn't be bigger than 500
const pageSize = 50     // can be: 20, 30, 40, 50
const pagesToFetch = 10 // mostly for debugging, 1–50

type Category string

const (
	categorySciFiFantasy      Category = "19378442031"
	categoryFantasy           Category = "19378443031"
	categoryFantasyEpic       Category = "19378451031"
	categoryFantasyAdventure  Category = "19378444031"
	categoryFantasyCreatures  Category = "19378449031"
	categoryFantasyHumour     Category = "19378455031"
	categorySciFi             Category = "19378464031"
	categorySciFiHard         Category = "19378474031"
	categorySciFiHumor        Category = "19378475031"
	categorySciFiSpaceExplore Category = "19378479031"
	categorySciFiSpaceOpera   Category = "19378480031"
	categoryYASciFiFantasy    Category = "19377879031"
	categoryKidsSciFiFantasy  Category = "19377132031"
)

var categories []Category = []Category{
	categorySciFiFantasy,
	categoryFantasy,
	categorySciFi,
	categoryYASciFiFantasy,
	categoryKidsSciFiFantasy,
	categoryFantasyEpic,
	categoryFantasyAdventure,
	categoryFantasyCreatures,
	categoryFantasyHumour,
	categorySciFiHard,
	categorySciFiHumor,
	categorySciFiSpaceExplore,
	categorySciFiSpaceOpera,
}

func (c Category) Friendly() string {
	switch c {
	case categorySciFiFantasy:
		return "SciFi Fantasy"
	case categoryFantasy:
		return "Fantasy"
	case categorySciFi:
		return "SciFi"
	case categoryKidsSciFiFantasy:
		return "Children's SciFi Fantasy"
	case categoryYASciFiFantasy:
		return "YA SciFi Fantasy"
	case categoryFantasyEpic:
		return "Fantasy Epic"
	case categoryFantasyAdventure:
		return "Fantasy Adventure"
	case categoryFantasyCreatures:
		return "Fantasy Creatures"
	case categoryFantasyHumour:
		return "Fantasy Humour"
	case categorySciFiHard:
		return "SciFi Hard"
	case categorySciFiHumor:
		return "SciFi Humor"
	case categorySciFiSpaceExplore:
		return "SciFi Space Exploration"
	case categorySciFiSpaceOpera:
		return "SciFi Space Opera"
	}

	return "Unknown Category"
}

func (c Category) Tags() []string {
	switch c {
	case categorySciFiFantasy:
		return []string{"Science Fiction & Fantasy", "Science Fiction", "Fantasy"}
	case categoryFantasy:
		return []string{"Science Fiction & Fantasy", "Fantasy"}
	case categorySciFi:
		return []string{"Science Fiction & Fantasy", "Science Fiction"}
	case categoryKidsSciFiFantasy:
		return []string{"Children's", "Science Fiction & Fantasy", "Science Fiction", "Fantasy"}
	case categoryYASciFiFantasy:
		return []string{"Teen & Young Adult", "Science Fiction & Fantasy", "Science Fiction", "Fantasy"}
	case categoryFantasyEpic:
		return []string{"Science Fiction & Fantasy", "Fantasy", "Epic"}
	case categoryFantasyAdventure:
		return []string{"Science Fiction & Fantasy", "Fantasy", "Adventure"}
	case categoryFantasyCreatures:
		return []string{"Science Fiction & Fantasy", "Fantasy", "Creatures"}
	case categoryFantasyHumour:
		return []string{"Science Fiction & Fantasy", "Fantasy", "Humour"}
	case categorySciFiHard:
		return []string{"Science Fiction & Fantasy", "Science Fiction", "Hard"}
	case categorySciFiHumor:
		return []string{"Science Fiction & Fantasy", "Science Fiction", "Humor"}
	case categorySciFiSpaceExplore:
		return []string{"Science Fiction & Fantasy", "Science Fiction", "Space", "Space Exploration"}
	case categorySciFiSpaceOpera:
		return []string{"Science Fiction & Fantasy", "Science Fiction", "Space", "Space Opera"}
	}

	return []string{}
}

type Sort string

const (
	sortFeatured Sort = "" // seems to produce guff!
	sortPop      Sort = "popularity-rank"
	sortReview   Sort = "review-rank"
)

var sorts []Sort = []Sort{
	// sortFeatured, // removed as seems to produce guff!
	sortPop,
	sortReview,
}

func (s Sort) Friendly() string {
	switch s {
	case sortFeatured:
		return "Featured"
	case sortPop:
		return "Popularity"
	case sortReview:
		return "Review Score"
	}
	return "Unknown Sort"
}

func makeSearchUrl(n Category, s Sort, page int) string {
	url := baseSearchUrl
	if n != "" {
		url += "&node=" + string(n)
	}
	url += "&sort=" + string(s)
	if page > 0 {
		url += fmt.Sprintf("&pageSize=%d&page=%d", pageSize, page)
	}
	return url
}

func stringsToInts(ss []string) []int {
	ns := len(ss)
	ints := make([]int, ns)
	for i := 0; i < ns; i++ {
		n, err := strconv.Atoi(strings.ReplaceAll(ss[i], ",", ""))
		if err != nil {
			log.Fatal("ERR!: stringsToInts:", err)
		}
		ints[i] = n
	}
	return ints
}

type Book struct {
	Title              string    `json:"title" selector:"h1"`
	SubTitle           string    `json:"subtitle" selector:".bc-col-5 span ul li:nth-child(2)"`
	Author             string    `json:"author" selector:".authorLabel > a"`
	AuthorLink         string    `json:"authorlink" selector:".authorLabel > a" attr:"href"`
	Series             string    `json:"series" selector:".seriesLabel > a"`
	SeriesLink         string    `json:"serieslink" selector:".seriesLabel > a" attr:"href"`
	Format             string    `json:"format" selector:".format"`
	ReleaseDate        time.Time `json:"releasedate"`
	Image              string    `json:"image" selector:"#center-1 .bc-col-3 > div > div:nth-child(1) > img" attr:"src"`
	Sample             string    `json:"sample" selector:"[id*=sample-player] > button" attr:"data-mp3"`
	Id                 string    `json:"asin" selector:"[id*=sample-player] > button" attr:"sample-asin"`
	Link               string    `json:"link"`
	Summary            string    `json:"summary"`
	Copyright          string    `json:"copyright" selector:"#center-9 > div > div > div:nth-child(3) > span"`
	Tags               []string  `json:"tags" selector:"#center-10 > div > div > div > div > span > span > a > span > span.bc-chip-text"`
	RatingsOverall     []string  `json:"ratingsoverall" selector:"#center-16 > div.bc-container > div.bc-row-responsive.bc-spacing-s6 > div:nth-child(1) > span > ul > li.histogram-rating span:nth-child(5)"`
	RatingsPerformance []string  `json:"ratingsperformance" selector:"#center-16 > div.bc-container > div.bc-row-responsive.bc-spacing-s6 > div:nth-child(2) > span > ul > li.histogram-rating span:nth-child(5)"`
	RatingsStory       []string  `json:"ratingsstory" selector:"#center-16 > div.bc-container > div.bc-row-responsive.bc-spacing-s6 > div:nth-child(3) > span > ul > li.histogram-rating span:nth-child(5)"`
	Rating             float64   `json:"rating"`
	RatingPerformance  float64   `json:"ratingperformance"`
	RatingStory        float64   `json:"ratingstory"`
	DurationInMins     int       `json:"durationInMins"`
	PopularityScore    float64   `json:"popularity"`
}

type tag struct {
	asin string
	tag  string
}

type BookCollector struct {
	books           map[string]bool
	bannedTags      map[string]bool
	bannedWords     []string
	db              *supabase.Client
	listCollector   *colly.Collector
	detailCollector *colly.Collector
	currentCategory Category
	currentSort     Sort
	popularityScore float64
}

var inEnglishRx = regexp.MustCompile(`(?i)Language:\s+English`)
var findPreOrderRx = regexp.MustCompile(`(?i)pre-?order`)
var findNotRatedRx = regexp.MustCompile(`(?i)Not\srated\syet`)
var fixFormatRx = regexp.MustCompile(`\s+`)
var findMinRx = regexp.MustCompile(`(?i)(\d+)M`)
var findHourRx = regexp.MustCompile(`(?i)(\d+)H`)

func (bc *BookCollector) setupCollectors() {
	bc.listCollector.OnHTML(".productListItem", func(e *colly.HTMLElement) {
		id := ""
		e.ForEachWithBreak("[id*=sample-player] > button", func(_ int, h *colly.HTMLElement) bool {
			id = h.Attr("sample-asin")
			return false
		})
		title := ""
		e.ForEachWithBreak("h3 > a", func(_ int, h *colly.HTMLElement) bool {
			title = h.Text
			return false
		})
		productText := e.DOM.Text()
		// is this book in English?
		if !inEnglishRx.MatchString(productText) {
			log.Println("- - SKIP: NOT ENGLISH:", title)
			return
		}
		// is this book pre-order only?
		if findPreOrderRx.MatchString(productText) {
			log.Println("- - SKIP: PRE-ORDER:", title)
			return
		}
		// has this book been rated yet?
		if findNotRatedRx.MatchString(productText) {
			log.Println("- - SKIP: NOT RATED:", title)
			return
		}
		// does this book contain banned words?
		for _, bannedWord := range bc.bannedWords {
			if strings.Contains(productText, bannedWord) {
				log.Printf("- - SKIP: word '%s' in %s", bannedWord, title)
				return
			}
		}
		// tag it as we've now seen it in this category, even if we've already seen it in another
		for _, tag := range bc.currentCategory.Tags() {
			bc.addBookTagToDB(id, tag)
		}

		// have we fetched this book before?
		if _, ok := bc.books[id]; ok {
			// we still need to score it's popularity
			if bc.currentSort == sortPop && bc.popularityScore > 0.0 {
				bc.updatePopularityScoreInDB(id, bc.getNextPopularityScore(bc.currentSort))
			}
			log.Println("- - SKIP:", id, "SEEN BEFORE")
			return
		}
		// not in the map so go and fetch it
		link := baseBookUrl + id
		bc.detailCollector.Visit(e.Request.AbsoluteURL(link))
	})

	bc.detailCollector.OnHTML("body > div.adbl-page.desktop", func(e *colly.HTMLElement) {

		// boom!
		b := &Book{}
		e.Unmarshal(b)

		// does this book contain banned tags?
		for _, tag := range b.Tags {
			if _, ok := bc.bannedTags[tag]; ok {
				log.Printf("- - SKIP: tag '%s' in %s", tag, b.Title)
				return
			}
		}

		// fix a few things

		b.Format = fixFormatRx.ReplaceAllString(b.Format, " ")
		b.Link = baseBookUrl + b.Id

		e.ForEachWithBreak("#center-9 > div > div > div:nth-child(2) > span", func(_ int, h *colly.HTMLElement) bool {
			html, err := h.DOM.Html()
			if err != nil {
				log.Fatal("ERR!: ForEachWithBreak:", err)
			}
			b.Summary = html
			return false
		})
		if len(b.RatingsOverall) > 0 {
			b.Rating = starSort(stringsToInts(b.RatingsOverall))
		}
		if len(b.RatingsPerformance) > 0 {
			b.RatingPerformance = starSort(stringsToInts(b.RatingsPerformance))
		}
		if len(b.RatingsStory) > 0 {
			b.RatingStory = starSort(stringsToInts(b.RatingsStory))
		}
		// pull data from javascript json
		jsonData := ""
		e.ForEachWithBreak("#bottom-0 > script", func(_ int, h *colly.HTMLElement) bool {
			jsonData += h.Text
			return false
		})
		data := []map[string]interface{}{}
		err := json.Unmarshal([]byte(jsonData), &data)
		if err != nil {
			log.Println("ERR!: skipping json", err)
		} else {
			// find release date
			if data != nil && data[0] != nil && data[0]["datePublished"] != nil {
				datePublishedString := data[0]["datePublished"].(string)
				datePublished, err := time.Parse("2006-01-02", datePublishedString)
				if err != nil {
					log.Println("- - ERR!: Could not parse time:", err)
				}
				b.ReleaseDate = datePublished
			} else {
				log.Println("ERR!: skipping datePublished")
			}

			// Find duration in minutes
			if data != nil && data[0] != nil && data[0]["duration"] != nil {
				durationString := data[0]["duration"].(string)
				durationHours := 0
				dhs := findHourRx.FindStringSubmatch(durationString)
				if len(dhs) != 0 {
					durationHours, err = strconv.Atoi(dhs[1])
					if err != nil {
						log.Println("- - ERR!: couldn't convert duration", err)
					}
				}
				durationMins := 0
				dms := findMinRx.FindStringSubmatch(durationString)
				if len(dms) != 0 {
					durationMins, err = strconv.Atoi(dms[1])
					if err != nil {
						log.Println("- - ERR!: couldn't convert duration", err)
					}
				}
				durationInMins := durationHours*60 + durationMins
				b.DurationInMins = durationInMins
			} else {
				log.Println("ERR!: skipping duration")
			}
		}

		// as this is the first time we've seen this book
		// we can calculate it's base popularity
		b.PopularityScore = bc.getNextPopularityScore(bc.currentSort)

		// add to books
		bc.books[b.Id] = true
		//

		// check database
		_, count, err := bc.db.From("books").Select("asin", "exact", false).Eq("asin", b.Id).Execute()
		if err != nil {
			log.Fatal("ERR!: DATABASE:", err)
		}
		if count == 0 {
			// add to database
			log.Printf("- • BOOK: %s (%1.2f★) %s, by %s\n", b.Id, b.Rating, b.Title, b.Author)
			_, _, err = bc.db.From("books").Insert(b, false, "", "", "").Execute()
			if err != nil {
				log.Println("ERR!: DATABASE: id:%s", err, b.Id)
				log.Println("%#v", b)
			}
		} else {
			// we still need to update popularity score
			if bc.currentSort == sortPop && bc.popularityScore > 0.0 {
				bc.updatePopularityScoreInDB(b.Id, b.PopularityScore)
			}

			log.Printf("- - DUP!: %s (%d)", b.Id, count)
		}
	})
}

func (bc *BookCollector) getAllPages(category Category, sort Sort) {
	bc.currentCategory = category
	bc.currentSort = sort
	bc.popularityScore = popularityTopScore // give points to the top 250 or so
	for pageNumber := 1; pageNumber <= pagesToFetch; pageNumber++ {
		log.Printf("- PAGE: %d of %d (books: %d to %d) (%s by %s)\n", pageNumber, pagesToFetch, ((pageNumber-1)*pageSize)+1, pageNumber*pageSize, category.Friendly(), sort.Friendly())

		url := makeSearchUrl(category, sort, pageNumber)
		log.Println("- - LOAD:", url)
		bc.listCollector.Visit(url)
	}
}
func (bc *BookCollector) getDebugPage(url string) {
	log.Println("- - LOAD:", url)
	bc.listCollector.Visit(url)
}

func (bc *BookCollector) addBookTagToDB(id, tag string) {
	log.Printf("Tag %s with %s\n", id, tag)
	bc.db.Rpc("insert_tag", "", map[string]string{"asin": id, "tag": tag})
}

func (bc *BookCollector) getNextPopularityScore(sort Sort) float64 {
	if sort != sortPop {
		return 0.0
	}
	ps := bc.popularityScore
	bc.popularityScore = bc.popularityScore * popularityFactor
	if bc.popularityScore < 1.0 {
		bc.popularityScore = 0.0
	}
	return ps * 2.0 // make it out of 500
}

func (bc *BookCollector) updatePopularityScoreInDB(id string, score float64) {
	bc.db.Rpc("add_to_popularity_score", "", map[string]interface{}{"asin_param": id, "score_param": score})
}

func main() {

	log.Printf("Laudible v%f\n", version)

	API_URL := os.Getenv("API_URL")
	API_KEY := os.Getenv("API_KEY")

	SbClient, err := supabase.NewClient(API_URL, API_KEY, nil)
	if err != nil {
		fmt.Println("cannot initalize client", err)
	}

	listCollector := colly.NewCollector(
		colly.AllowedDomains("www.audible.co.uk"),
		colly.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.6 Safari/605.1.15"),
	)
	detailCollector := listCollector.Clone()

	bookCollector := BookCollector{
		books:           map[string]bool{},
		bannedTags:      map[string]bool{},
		bannedWords:     []string{},
		db:              SbClient,
		listCollector:   listCollector,
		detailCollector: detailCollector,
	}
	bookCollector.setupCollectors()

	// pre-seed books with database contents

	allKnownIds := []map[string]string{}
	count, err := SbClient.From("books").Select("asin", "exact", false).ExecuteTo(&allKnownIds)
	if err != nil {
		log.Fatal("ERR!: DATABASE:", err)
	}
	log.Printf("INFO: %d books in database\n", count)

	// load the banned tags

	bannedTags := []map[string]string{}
	count, err = SbClient.From("banned_tags").Select("tag", "exact", false).ExecuteTo(&bannedTags)
	if err != nil {
		log.Fatal("ERR!: DATABASE:", err)
	}
	log.Printf("INFO: %d banned tags in database\n", count)

	// load the banned words

	bannedWords := []map[string]string{}
	count, err = SbClient.From("banned_words").Select("word", "exact", false).ExecuteTo(&bannedWords)
	if err != nil {
		log.Fatal("ERR!: DATABASE:", err)
	}
	log.Printf("INFO: %d banned words in database\n", count)

	// convert books to a fast asin lookup
	for _, book := range allKnownIds {
		bookCollector.books[book["asin"]] = true
	}
	// convert banned_tags to a fast tag lookup
	for _, tag := range bannedTags {
		bookCollector.bannedTags[tag["tag"]] = true
	}
	// convert banned_tags to a fast tag lookup
	for _, word := range bannedWords {
		bookCollector.bannedWords = append(bookCollector.bannedWords, word["word"])
	}

	for _, category := range categories {
		for _, sort := range sorts {
			log.Printf("CATEGORY: %s sorted by %s", category.Friendly(), sort.Friendly())
			bookCollector.getAllPages(category, sort)

			// tell the database to update all the tags
			log.Println("TAGS: update:", SbClient.Rpc("update_all_tags", "", nil))
		}
	}
}
