package main

import (
    "fmt"
    "os"
    "net/http"
    "strings"
    "github.com/PuerkitoBio/goquery"
    "log"
    "sync"
  )

type node struct {
  url string
  path int
  parent *node
}

var prefixes = [...]string{"/wiki/Main_Page", "/wiki/Special","/wiki/File","/wiki/Help"}

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

func getNewLinks(a node, q []node,path_num int)[]node {
  resp, err := http.Get(a.url)
  if err != nil{
    log.Fatal("Error getting page", err)
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
  return q
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
              next_q1 = getNewLinks(a1,next_q1,1)
            }
        }

        if len(q2) > 0{
          a2, q2 = q2[0], q2[1:]
          if val, found = m[a2.url]; found && val.path != 2 {
             paths = append(paths, getPath(&val, &a2))
           }
          if !found{
           m[a2.url] = a2
           next_q2 = getNewLinks(a2,next_q2,2)
          }
        }
       }
       fmt.Println(paths)
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
  fmt.Println(path)


}
