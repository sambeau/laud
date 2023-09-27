# Laudible

Search Audible books using a properly weighted star rating.

## Intro

I needed something new to listen to while doing the dishes and Audible's search is, frankly, broken. So I did the thing that programmery people do — I said, "I'll just scrape all of Audible and write my own search engine. How hard can it be?…"

Like so many programmery things — it's not 'hard' but it isn't simple or quick either.

## The Plan

The plan was to scrape all the Audible categories I care about and fill a database with them. Next, I wanted to apply a better rating algorithm to them, in the hope of seeing a more representative list of good titles. Finally, I want to build a better search app, so I can find my next book without having to use the terrible juddery lists in the Audible iOS app. I could build a web app, but I mostly have my phone in my hand when I'm looking for a new book to listen to.

So far, I have done the following:

- Built the scraper
- Created the online database
- Written a better rating algorithm
- Scraped 12 categories, with 3 sort orders each

So, I now have a database with 7,230 audiobooks in it. Unfortunately, this includes titles not in English which are of no use to me, so I will probably have to run it again and remove them.

So that just leaves the simple task of building an app (and working out a few niggles).

## Basic Details

The Go files (laud.go and starsort.go) make up the Audible scraper. It slurps up categories (nodes, as Audible calls them) and fills a Superbase database. Supabase is just Postgres + Postgrest, so you should be able to make it work with any SQL database easily enough (there's only a few DB calls).

I'm currently listening to a lot of Fantasy & Sci-Fi, so I only scraped those categories. It's a little tricky to find out what the categories are, so I ended up going to a category on the Audible website and examining the URL string for a `node` parameter, which tend to look like `node=19378442031`.

These are the `nodes` I scrape, which I've renamed `categories`:

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

I also scrape more than one sort-order, as you are limited to 500 books in each search. So I search:

	sortFeatured Sort = ""
	sortPop      Sort = "popularity-rank"
	sortReview   Sort = "review-rank"

Again, these are a parameter in the search URL, with the more sensible name of `sort`.

## The Scraping

As with all scraping, this code relies heavily on Audible keeping the same design. It's extra-complicated by there being two Audibles — the one you see if signed in and the one when signed out. It took me quite a while to realise that I was looking at the wrong HTML. Furthermore, Audible uses javascript to populate some fields, especially the ones that need translation (like date formats). Again, it took me a while to work out that I needed to scrape some JSON data in a script tag to find the published date.

The scraper itself is reasonably reliable (though occasionally, it messes up, possibly due to network issues) and currently you have to run it again from scratch (or edit the code to comment out categories already scraped). It won't ask for something that is already in the cache or the database (and it fills the cache from the database before starting). However, it will still request all the category list pages on every run (which is only 10 pages, but it adds up, once you multiply by three for each sort).

One annoying thing I've found is that, when you fins a title, it isn't tagged by the category in which you found it, so a Fantasy title won't be tagged as "Sci-Fi", just "Time Travel". Which probably means having to scan every page for every category If I want to use metadata tags.

## The Better Rating algorithm

The issue I have with star rating is that a title with two 5★ ratings will rank the same as a title with 1,000 5★ ratings. It's infuriating. So I went looking for a better approach that would address this issue. I'd like to say I delved deep into the world of "Dirichlet prior distribution on the probability vectors", but in reality I looked at a StackOverflow page and converted some bonkers Python code into a working Go function.

The algorithm I use is a Go version of Evan Miller's "Ranking Items With Star Ratings" which uses a Bayesian approximation to provide a better average rating.

- [Evan Miller, Ranking Items With Star Ratings](http://www.evanmiller.org/ranking-items-with-star-ratings.html)

I scrape all the raw data about how many people scored a title 5★, .., 1★ and use those to recalculate a better average.

Tolkien's "The Two Towers" read by Andy Serkis (4,154 ratings), is rated:

	["3,918", "196", "26", "4", "6"]

Amazon gives it a rating of `4.9` which is pretty close to my score of `4.9209`. However, it's top of my list (with the highest score), but comes in 14th on Amazon's list—behind "The Eye of the Bedlam Bride" (170 ratings, top spot) and "Mythology: Indian Mythology, Gods, Goddesses, and Stories" (100 ratings, 3rd spot).

The ratings for "Mythology: Indian Mythology, Gods, Goddesses, and Stories" are, the somewhat suspicious-looking:

	["98", "1", "0", "0", "1"]

Amazon gives it `5.0`. I give it `4.75239`, 474th on my list, which isn't bad. It's still, surprisingly, ahead of "Ready Player One" (4.75208) and "Mistborn" (4.69725).

You can find the star rating code in `starsort.go`. And yes, it looks very mathsy because the Python code was mathsy.

## TODO

- Design an iOS App
- Make an iOS App
- Add a simple way to restart a scraping job
- Add an option to only scrape for one language
- Break tags up into new tags table
- Add a way to update in parts, perhaps an option to do one category at once
- Add a column to mark a title that has an error on import, so we can do a run of just errors
- Reject based on tag e.g. ["Biographies & Memoirs","Mental Health","Psychology","Gender Studies","Performing Arts"]
- add "Mystery, Thriller, Suspense" (and for YA)
- remove featured
~~Add category 19377132031 children's sci-fi fantasy~~
