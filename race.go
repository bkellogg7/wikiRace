package main

import (
    "fmt"
    "os"
    "net/http"
    "strings"
    "github.com/PuerkitoBio/goquery"
    "log"
  )

type node struct {
  url string
  path int
  parent *node
}

//Slice of prefixes of wikipedia pages to ignore in order to only get pages of actual articles
var prefixes = [...]string{"/wiki/Main_Page", "/wiki/Special","/wiki/File","/wiki/Help","/wiki/Wikipedia:"}

/* checkListOfPrefixes: Checks url against slice of prefixes to ensure the linke is to an article.*/
func checkListOfPrefixes(href string)bool{
  for _, prefix := range prefixes{
    if strings.HasPrefix(href,prefix){
      return true
    }
  }
  return false
}

/* processElement: Checks if url is valid for the race and formats it to the full url
                   to prepare it for a get request.*/
func processElement(index int, element *goquery.Selection)(string) {
  // See if the href attribute exists on the element
  href, exists := element.Attr("href")
  if exists && strings.HasPrefix(href,"/wiki/") && ! checkListOfPrefixes(href){
    return "https://en.m.wikipedia.org" + string(href)
  }
  return ""
}

/* getNewLinks: Ansynchronous function that performs a get request to a url to
                ensure it is valid, creates a slice of all the valid urls from
                the page for the race, returns the slice in the provided channel. */
func getNewLinks(a node, out chan<- []node,path_num int){
  //slic to add new links to
  q := make([]node,0)

  //Get request for the url
  resp, err := http.Get(a.url)
  if err != nil{
    log.Print("Error getting page", err)
    out <- q
    return
  }

  document, queryError := goquery.NewDocumentFromReader(resp.Body)
  if queryError != nil {
     log.Fatal("Error loading HTTP response body. ", queryError)
  }

  //find all links within the body of the page
  links := document.Find("a").Map(processElement)
  for _, link := range links{
      if link != ""{
        q = append(q, node{url:link,path:path_num,parent:&a})
      }
  }

  //send slice to the channel
  out <- q
}

/*getPath: concatenate the path from node1 to the starting node with the path of
           node2 to the ending node.*/
func getPath(node1 *node, node2 *node)string{
  path := ""
  for node1 != nil{
    path = node1.url + " -> " + path
    node1 = node1.parent
  }
  path = strings.TrimRight(path," -> ")
  node2 = node2.parent
  for node2 != nil{
    path = path + " -> " + node2.url
    node2 = node2.parent
  }
  return path
}


func VisitNode(m map[string]node, q []node, threadCount int,  out chan<- []node, path int, paths []string)(map[string]node, []node, int, []string){
  if len(q) > 0{
    var a node
    var val node
    var found bool

    //pop a url from the queue
    a, q = q[0], q[1:]
    //if the url is already in the map and it was not found in this half
    //of the search, then path has been found from start to finish
    if val, found = m[a.url]; found && val.path != path {
     paths = append(paths, getPath(&a, &val))
    }
    if !found {
      //add url to map
      m[a.url] = a
      threadCount += 1
      //asynch thread to get new links on page
      go getNewLinks(a, out, path)
    }
  }
  return m, q, threadCount, paths
}

/* FindShortestWikiPath: Returns a string indicating the shortest path between
                         two wikipeida articles. Utilizes a BFS that starts from
                         both the start and finish articles.*/
func FindShortestWikiPath(article1 string, article2 string)(string, string){

  //convert the proviced article names to wikipedia url's
  start := node {
                        url : "https://en.m.wikipedia.org/wiki/" + strings.Replace(article1, " ", "_", -1),
                        path: 1,
                        parent : nil,
                      }
  end := node {
                        url : "https://en.m.wikipedia.org/wiki/" + strings.Replace(article2, " ", "_", -1),
                        path: 2,
                        parent : nil,
                      }

   //map from a url string to node representing a webpage that has been visited
   m := make(map[string]node)

   //preliminary queues for the next round of BFS
   next_q1 := make([]node, 1)
   next_q1[0] = start
   next_q2 := make([]node,1)
   next_q2[0] = end

   //current queue of url's to be visited in the BFS
   q1 := make([]node, 0)
   q2 := make([]node, 0)

   //for each round of bfs
   for len(next_q1) > 0 || len(next_q2) > 0{
      q1 = next_q1
      q2 = next_q2

      next_q1 = make([]node, 0)
      next_q2 = make([]node, 0)

      //contains all valid paths from start to finish found in a round
      paths := make([]string, 0)

      //chnnels for each queue to return new links found on each page
      out1 := make(chan []node)
      out2 := make(chan []node)

      //number of threads for each channel
      out1Count := 0
      out2Count := 0

      //for each element in the queues in this round of BFS
      for len(q1) > 0 || len(q2) > 0{

        m, q1, out1Count, paths = VisitNode(m, q1, out1Count, out1, 1, paths)

        m, q2, out2Count, paths = VisitNode(m, q2, out2Count, out2, 2, paths)

       }

       //If a valid path from starting node to finishing node is found
       if len(paths) != 0 {
         //find the shortest path in the list of paths
         shortestLength := len(paths[0])
         index := 0
         for i, path := range paths {
           if length := strings.Count(path, "->"); length < shortestLength{
             shortestLength = length
             index = i
           }
         }
         return paths[index], ""
       }

       //For each thread in ech channel, get the result from the thread
       //and add it to next queue
       for i := 0; i < out1Count; i++ {
         next_q1 = append(next_q1, <-out1...)
       }
       for i := 0; i < out2Count; i++ {
         next_q2 = append(next_q2, <-out2...)
       }

   }
   return "No Path Found.",""

}

func FindArticleAdress(article string)(bool, string){
  var wikiURL = "https://en.wikipedia.org/wiki/"
  wikiURL = wikiURL + strings.Replace(article, " ", "_", -1)
  resp, err := http.Get(wikiURL)
  status := resp.StatusCode
  if err != nil || status == 404{
    return false, "The article " + article + " does not exist."
  }
  return true, ""

}


func main() {
  if len(os.Args) < 3{
    fmt.Println("Please provide two wikipedia article names as arguments.")
    return
  }
	var article1 = os.Args[1]
  var article2 = os.Args[2]

  exists1, err := FindArticleAdress(article1)
  if !exists1 {
    fmt.Println(err)
  }
  exists2, err := FindArticleAdress(article2)
  if !exists2 {
    fmt.Println(err)
  }

  path, _ := FindShortestWikiPath(article1,article2)
  fmt.Println("Path: ",path)


}
