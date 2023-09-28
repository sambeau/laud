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
- Scraped 12 categories, with 2 sort orders each

So, I now have a database with >7,000 audiobooks in it. Unfortunately, this includes titles not in English which are of no use to me, so I will probably have to run it again and remove them.

So that just leaves the simple task of building an app (and working out a few niggles).

## Basic Details

The Go files (laud.go and starsort.go) make up the Audible scraper. It slurps up categories (nodes, as Audible calls them) and fills a Superbase database. Supabase is just Postgres + Postgrest, so you should be able to make it work with any SQL database easily enough (there's only a few DB calls).

I'm currently listening to a lot of Fantasy & Sci-Fi, so I only scraped those categories. It's a little tricky to find out what the categories are, so I ended up going to a category on the Audible website and examining the URL string for a `node` parameter, which tend to look like `node=19378442031`.

These are the (updated) list of `nodes` I scrape, which I've renamed `categories`:

	categorySciFiFantasy            Category = "19378442031"
	categoryFantasy                 Category = "19378443031"
	categoryFantasyEpic             Category = "19378451031"
	categoryFantasyAdventure        Category = "19378444031"
	categoryFantasyCreatures        Category = "19378449031"
	categoryFantasyHumour           Category = "19378455031"
	categorySciFi                   Category = "19378464031"
	categorySciFiHard               Category = "19378474031"
	categorySciFiHumor              Category = "19378475031"
	categorySciFiSpaceExplore       Category = "19378479031"
	categorySciFiSpaceOpera         Category = "19378480031"
	categoryYASciFiFantasy          Category = "19377879031"
	categoryKidsSciFiFantasy        Category = "19377132031"
	categoryActionAdventure         Category = "19378254031"
	categoryKidsActionAdventure     Category = "19376663031"
	categoryMysteryThrillerSuspense Category = "19378257031"

I also scrape more than one sort-order, as you are limited to 500 books in each search. So I search:

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

## Popularity Scores

I've added an experimental popularity score. As Audible don't expose download figures on the site I have had to make up my own based on where, and how often, a book appears in each category when sorted by popularity. I've used a magic formula where the top book in each list gets 500 points added to its popularity score and I exponentially shrink this number for all the other placings. By around 300 the score is zero. The more lists a title appears in, the more points it gets.

I'll no doubt have to make adjustments to it.

## Tags

I import all of Audible's tags and add ones for each category, as Audible doesn't include those in the tags for some reason.


## Banned Tags and Words

I now keep a list of tags and words I don't want. At one point I was getting self-help books and other guff, so I filtered them out (this turned out to be a bug). However, I can still get rid of books I don't want: Erotica, LitRPG, BDSM etc.

## Results
I now have 6,717 title in the database. It's smaller than previous runs as I'm now ignoring over 1,000 Warhammer and LitRPG titles.

I thought it would be interesting to compare review score with popularity score. As you can see there three titles that appear popular but have very low review scores. That seems suspicious to me and it smells of rigging/gaming the system, though it may be that Audible uses US sales as part of the popularity score while omitting them from the ratings score (I'm scraping the .co.uk store, not the .com).

| author               -  | title                                                                               | rating  | popularity |
| ----------------------- | ----------------------------------------------------------------------------------- | ------- | ---------- |
| M.E. Robinson           | Loremaster                                                                          | 2.40366 | 500        |
| Damien Hanson           | BuyMort: Rise of the Window Puncher: How I Became the Accidental Warlord of Arizona | 2.69738 | 489.7      |
| Benjamin Wallace        | Dads vs. Zombies: Year 2                                                            | 2.69738 | 414.582    |


Highest rated books:

| author                | title                                                            | rating  | popularity |
| --------------------- | ---------------------------------------------------------------- | ------- | ---------- |
| J.K. Rowling          | Harry Potter and the Deathly Hallows, Book 7                     | 4.9472  | 423.302    |
| J.K. Rowling          | Harry Potter and the Half-Blood Prince, Book 6                   | 4.92378 | 450.579    |
| J. R. R. Tolkien      | The Two Towers                                                   | 4.92102 | 358.37     |
| J.K. Rowling          | Harry Potter and the Prisoner of Azkaban, Book 3                 | 4.91749 | 381.462    |
| J.K. Rowling          | Harry Potter and the Goblet of Fire, Book 4                      | 4.91441 | 389.485    |
| J. R. R. Tolkien      | The Return of the King                                           | 4.90369 | 329.74     |
| Bernard Cornwell      | Sharpe's Enemy: The Defence of Portugal, Christmas 1812          | 4.90257 | 0          |
| J.K. Rowling          | Harry Potter and the Philosopher's Stone, Book 1                 | 4.89957 | 195.963    |
| J.K. Rowling          | Harry Potter and the Order of the Phoenix, Book 5                | 4.89876 | 441.297    |
| J. R. R. Tolkien      | The Hobbit                                                       | 4.89626 | 96.5646    |
| J.K. Rowling          | Harry Potter and the Chamber of Secrets, Book 2                  | 4.89611 | 479.612    |
| Jodi Taylor           | The Good, the Bad and the History                                | 4.89041 | 0          |
| J. R. R. Tolkien      | The Fellowship of the Ring                                       | 4.88994 | 176.594    |
| Robert Jordan         | A Memory of Light                                                | 4.88002 | 0          |
| Bernard Cornwell      | War Lord                                                         | 4.87922 | 0          |
| Dan Abnett            | Saturnine                                                        | 4.87582 | 0          |
| J. R.R. Tolkien       | The Return of the King                                           | 4.87483 | 0          |
| David Gemmell         | The First Chronicles of Druss the Legend                         | 4.8716  | 0          |
| Mark Tufo             | Hiraeth                                                          | 4.8712  | 0          |
| Bernard Cornwell      | Sharpe's Eagle: The Talavera Campaign, July 1809                 | 4.87074 | 0          |
| Craig Alanson         | Fallout                                                          | 4.8707  | 0          |
| Craig Alanson         | Critical Mass                                                    | 4.86808 | 0          |
| Robert Jordan         | Towers of Midnight                                               | 4.86502 | 0          |
| Jim Butcher           | Changes: The Dresden Files, Book 12                              | 4.86369 | 0          |
| Terry Mancour         | Hedgewitch                                                       | 4.86172 | 0          |
| J. R.R. Tolkien       | The Two Towers                                                   | 4.861   | 6.45106    |
| Joe Abercrombie       | The Trouble with Peace                                           | 4.85908 | 2.86467    |
| Bernard Cornwell      | Sharpe's Regiment: The Invasion of France, June to November 1813 | 4.8588  | 0          |
| Brandon Sanderson     | Words of Radiance                                                | 4.8577  | 397.678    |
| Robert Jordan         | The Gathering Storm                                              | 4.85652 | 0          |
| Bernard Cornwell      | Sharpe's Sword: The Salamanca Campaign, June and July 1812       | 4.85541 | 0          |
| Michael Stephen Fuchs | Endgame                                                          | 4.85483 | 0          |
| Terry Pratchett       | Lords and Ladies                                                 | 4.8539  | 0          |
| Suzanne Collins       | Catching Fire                                                    | 4.85355 | 450.579    |
| Bernard Cornwell      | Sharpe's Honour: The Vitoria Campaign, February to June 1813     | 4.85259 | 0          |
| Jessica Townsend      | Hollowpox                                                        | 4.85184 | 83.4716    |
| Sarah J. Stone        | Nathanial (Dragon Shifter Romance)                               | 4.85131 | 0          |
| Bernard Cornwell      | Sharpe's Company: The Siege of Badajoz, January to April 1812    | 4.85099 | 0          |
| Jodi Taylor           | Saving Time                                                      | 4.85058 | 0          |
| Terry Mancour         | Necromancer                                                      | 4.85014 | 0          |
| Stephen King          | The Green Mile                                                   | 4.84901 | 0          |
| Terry Pratchett       | Carpe Jugulum                                                    | 4.84877 | 0          |
| Suzanne Collins       | The Hunger Games: Special Edition                                | 4.84849 | 373.604    |
| Stephen Fry           | Heroes                                                           | 4.84415 | 373.604    |
| Jason Anspach         | Galaxy's Edge, Part V                                            | 4.84384 | 0          |
| Jessica Townsend      | Wundersmith                                                      | 4.84131 | 51.7155    |
| Chris Smith           | Kid Normal and the Final Five                                    | 4.84118 | 4.43519    |
| Terry Mancour         | Arcanist                                                         | 4.84079 | 0          |
| James S. A. Corey     | Tiamat's Wrath                                                   | 4.84039 | 0          |
| TurtleMe              | Reckoning                                                        | 4.8401  | 0          |

Most popular books:

| author                  | title                                                                               | rating  | popularity |
| ----------------------- | ----------------------------------------------------------------------------------- | ------- | ---------- |
| M.E. Robinson           | Loremaster                                                                          | 2.40366 | 500        |
| Sarah J. Maas           | A Court of Mist and Fury                                                            | 4.78393 | 500        |
| Rick Riordan            | Percy Jackson and the Olympians: The Chalice of the Gods                            | 3.82815 | 500        |
| Damien Hanson           | BuyMort: Rise of the Window Puncher: How I Became the Accidental Warlord of Arizona | 2.69738 | 489.7      |
| Dina Gregory            | Gingerella                                                                          | 4.06145 | 489.7      |
| Lucinda Whiteley        | Horrid Henry: Hugely Horrid                                                         | 4.03231 | 489.7      |
| Sarah J. Maas           | A Court of Wings and Ruin                                                           | 4.69171 | 489.7      |
| Peter Benchley          | Jaws                                                                                | 4.36502 | 479.612    |
| Graham McNeill          | Blood of the Emperor                                                                | 4.06129 | 479.612    |
| J.K. Rowling            | Harry Potter and the Chamber of Secrets, Book 2                                     | 4.89611 | 479.612    |
| Lucinda Whiteley        | Horrid Henry: How to Be Horrid                                                      | 4.17858 | 469.732    |
| Chelsea Sedoti          | As You Wish                                                                         | 2.93654 | 469.732    |
| Paul Lynch              | Prophet Song                                                                        | 2.93654 | 469.732    |
| John Wiltshire          | The Paths Less Travelled                                                            | 2.40366 | 469.732    |
| Annamaria Murphy        | Curious Under the Stars                                                             | 4.60948 | 460.056    |
| Lucinda Whiteley        | Horrid Henry: Big Story Bonanza                                                     | 4.34821 | 460.056    |
| J.K. Rowling            | Harry Potter and the Half-Blood Prince, Book 6                                      | 4.92378 | 450.579    |
| Suzanne Collins         | Catching Fire                                                                       | 4.85355 | 450.579    |
| Astrid Lindgren         | Pippi Longstocking                                                                  | 4.19699 | 450.579    |
| Douglas Adams           | Dirk Gently's Holistic Detective Agency                                             | 4.50663 | 450.579    |
| George R.R. Martin      | A Clash of Kings                                                                    | 4.74214 | 450.579    |
| J.K. Rowling            | Harry Potter and the Order of the Phoenix, Book 5                                   | 4.89876 | 441.297    |
| John Buchan             | The Thirty-Nine Steps                                                               | 4.31383 | 441.297    |
| Sarah J. Maas           | Empire of Storms                                                                    | 4.70841 | 441.297    |
| Nick Jones              | And Then She Vanished                                                               | 4.48123 | 441.297    |
| Douglas Adams           | The Long Dark Tea-Time of the Soul                                                  | 4.42196 | 441.297    |
| Lucinda Whiteley        | Horrid Henry: Awesome Adventures                                                    | 4.38373 | 441.297    |
| Rob Grant               | The Quanderhorn Collexion                                                           | 3.54133 | 432.206    |
| Rin Chupeco             | The Bone Witch                                                                      | 3.99198 | 432.206    |
| Erskine Childers        | The Riddle Of The Sands                                                             | 4.28135 | 432.206    |
| George R.R. Martin      | A Storm of Swords                                                                   | 4.79073 | 432.206    |
| J.K. Rowling            | Harry Potter and the Deathly Hallows, Book 7                                        | 4.9472  | 423.302    |
| Douglas Adams           | Dirk Gently: Two BBC Radio Full-Cast Dramas                                         | 4.32044 | 423.302    |
| Sarah J. Maas           | A Court of Thorns and Roses                                                         | 4.45476 | 423.302    |
| Jeaniene Frost          | Wicked All Night                                                                    | 4.65947 | 423.302    |
| Arthur Ransome          | Swallowdale                                                                         | 4.69577 | 414.582    |
| Sarah J. Maas           | Queen of Shadows                                                                    | 4.72169 | 414.582    |
| Benjamin Wallace        | Dads vs. Zombies: Year 2                                                            | 2.69738 | 414.582    |
| Lucy Courtenay          | Mermaid School                                                                      | 4.62252 | 414.582    |
| Alastair Reynolds       | Redemption Ark                                                                      | 4.44161 | 414.582    |
| George R.R. Martin      | A Feast for Crows                                                                   | 4.46649 | 406.042    |
| Frances Hodgson Burnett | The Secret Garden                                                                   | 4.67702 | 406.042    |
| Dan Wells               | Zero G                                                                              | 4.666   | 406.042    |
| John Scalzi             | The Ghost Brigades                                                                  | 4.37908 | 406.042    |
| Dan Gutman              | Mission Unstoppable                                                                 | 4.53704 | 406.042    |
| Robert Galbraith        | The Ink Black Heart                                                                 | 4.3067  | 406.042    |
| Brandon Sanderson       | Words of Radiance                                                                   | 4.8577  | 397.678    |
| John Scalzi             | The Android's Dream                                                                 | 4.39613 | 397.678    |
| Sarah J. Maas           | A Court of Mist and Fury (Part 2 of 2) (Dramatized Adaptation)                      | 4.75819 | 397.678    |
| J.K. Rowling            | Harry Potter and the Goblet of Fire, Book 4                                         | 4.91441 | 389.485    |

## TODO

- Design an iOS App
- Make an iOS App
- Add a simple way to restart a scraping job
- Add a way to update in parts, perhaps an option to do one category at once
- Add a column to mark a title that has an error on import, so we can do a run of just errors

