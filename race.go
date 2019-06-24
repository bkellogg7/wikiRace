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

var prefixes = [...]string{"/wiki/Main_Page", "/wiki/Special","/wiki/File","/wiki/Help","/wiki/Wikipedia:"}

func checkListOfPrefixes(href string)bool{
  for _, prefix := range prefixes{
    if strings.HasPrefix(href,prefix){
      return true
    }
  }
  return false
}

func processElement(index int, element *goquery.Selection)(string) {
  // See if the href attribute exists on the element
  href, exists := element.Attr("href")
  if exists && strings.HasPrefix(href,"/wiki/") && ! checkListOfPrefixes(href){
    return "https://en.m.wikipedia.org" + string(href)
  }
  return ""
}

func getNewLinks(a node, out chan<- []node,path_num int){
  q := make([]node,0)
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
  links := document.Find("a").Map(processElement)
  for _, link := range links{
      if link != ""{
        q = append(q, node{url:link,path:path_num,parent:&a})
      }
  }
  out <- q
}


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

func FindShortestWikiPath(article1 string, article2 string)(string, string){
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

   m := make(map[string]node)
   next_q1 := make([]node, 1)
   next_q1[0] = start
   next_q2 := make([]node,1)
   next_q2[0] = end

   var a1 node
   var a2 node
   q1 := make([]node, 0)
   q2 := make([]node, 0)

   for len(next_q1) > 0 || len(next_q2) > 0{
      q1 = next_q1
      q2 = next_q2
      next_q1 = make([]node, 0)
      next_q2 = make([]node, 0)
      paths := make([]string, 0)

      out1 := make(chan []node)
      out1Count := 0
      out2 := make(chan []node)
      out2Count := 0

      //fmt.Println(len(q1))
      //fmt.Println(len(q2))

      for len(q1) > 0 || len(q2) > 0{
        var val node
        var found bool

        if len(q1) > 0{
           a1, q1 = q1[0], q1[1:]
           if val, found = m[a1.url]; found && val.path != 1 {
             paths = append(paths, getPath(&a1, &val))
            }
            if !found {
              m[a1.url] = a1
              out1Count += 1
              go getNewLinks(a1, out1, 1)
            }
        }

        if len(q2) > 0{
          a2, q2 = q2[0], q2[1:]
          if val, found = m[a2.url]; found && val.path != 2 {
             paths = append(paths, getPath(&val, &a2))
           }
          if !found{
           m[a2.url] = a2
           out2Count +=1
           go getNewLinks(a2, out2, 2)
          }
        }
       }
       // fmt.Println("/////////////////////////////////////////////////////////////////////////////")
       if len(paths) != 0 {
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
