package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"sync"

	"github.com/Diggernaut/mxj"
	"gopkg.in/mgo.v2"
)

func init() {
	mxj.JsonUseNumber = true
	mxj.XmlGoEmptyElemSyntax()
}

// XMLOBJ is wrapper on xml object
type XMLOBJ struct {
	data []byte
	id   int
}

// SL is container for xml objects
type SL []XMLOBJ

func main() {
	// read flags, and if something wrong = exit
	var col string
	var db string
	var filename string
	var dbaddr string

	flag.StringVar(&db, "db", "", "mongo database name")
	flag.StringVar(&col, "col", "", "mongo collection name")
	flag.StringVar(&dbaddr, "dbaddr", "", "mongo host addr")
	flag.StringVar(&filename, "file", "", "json filename")
	flag.Parse()

	if db == "" || col == "" || dbaddr == "" {
	} else if filename == "" {
		fmt.Fprintln(os.Stderr, "Use flag Luke")
		os.Exit(1)
	}
	if db != "" && col != "" && dbaddr != "" {
		mongoToXml(db, col, dbaddr)
		return
	}
	if filename != "" {
		fileToXml(filename)
		return
	}

}
// MapToXML is function thats map mongo objects to xml
func MapToXML(obj map[string]interface{}, id int, wg *sync.WaitGroup, ch chan XMLOBJ) {
	// make sure we call done anyway
	defer wg.Done()
	data, err := mxj.AnyXmlIndentByte(obj, "", " ")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	ch <- XMLOBJ{data: data, id: id}
	// for lower memory usage
	//data = nil
	//obj = nil
}
func mongoToXml(db, col, dbaddr string) {
	// connect to database
	Session, err := mgo.Dial(dbaddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can`t connect to database, reason:  %v\n", err)
		os.Exit(1)
	}
	defer Session.Close()
	Session.SetMode(mgo.Monotonic, true)
	c := Session.DB(db).C(col)
	// get len of objects in database, if err exit
	l, err := c.Count()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can`t get count of items in database, reason: %v\n", err)
		os.Exit(1)
	}
	// do job
	// waitgroup
	var wg sync.WaitGroup
	counter := 0
	// chan for data
	cha := make(chan (XMLOBJ), l)
	// chan for break loop
	done := make(chan bool, 1)
	// container
	sl := make([]XMLOBJ, l)
	// goroutines scheduler
	runtime.Gosched()
	// loop over items in db, for each make new goroutine
	// add each to waitgroup
	for i := 0; i < l; i++ {
		var result map[string]interface{}
		_ = c.Find(nil).Limit(i + 1).Skip(i - 1).One(&result)
		// delete id field
		delete(result, "_id")
		wg.Add(1)
		go MapToXML(result, counter, &wg, cha)
		counter++
	}
	// this func wait for all goroutines done
	go func(group *sync.WaitGroup) {
		group.Wait()
		done <- true
	}(&wg)
	// loop over channels
loop:
	for {
		select {
		// save each item in slice
		case item := <-cha:
			sl[item.id] = item
			// break loop when all goroutines finish
		default:
			select {
			case stop := <-done:
				if stop {
					break loop
				}
			}

		}
	}

	fmt.Fprintln(os.Stdout, "<?xml version='1.0' encoding='UTF-8'?>")
	fmt.Fprintln(os.Stdout, "<items>")
	// iterate over container and print data in Stdout
	for i := range sl {
		fmt.Fprintln(os.Stdout, string(sl[i].data))
	}
	fmt.Fprintln(os.Stdout, "</items>")
}
func fileToXml(filename string) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	m := make(map[string]interface{})
	err = json.Unmarshal(data, &m)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	fmt.Fprintln(os.Stdout, "<?xml version='1.0' encoding='UTF-8'?>")
	fmt.Fprintln(os.Stdout, "<items>")
	// iterate over container and print data in Stdout
	fmt.Fprintln(os.Stdout, string(FileMapToXML(m)))

	fmt.Fprintln(os.Stdout, "</items>")

}
func FileMapToXML(m map[string]interface{}) []byte {
	data, err := mxj.AnyXmlIndentByte(m, "", " ", "")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	return data

}
