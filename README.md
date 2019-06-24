# wikiRace
Small project to find the shortest path between any two wikipedia pages.

## Description
This project finds the shortest path of wikipedia page links between any two articles. A Breadth first search (BFS) is utilized
to accomplish this along with ansynchronous threads to gather the new links of new articles. Links to non article pages are ignored
for the purpose of the race

## Dependencies
This project is written in golang with use of the third party package github.com/PuerkitoBio/goquery.

## Usage
``` $ ./wikiRace "article 1" "article 2" ```

