package main

import (
	"crypto/sha1"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strconv"

	"github.com/imroc/req"
	"sync"

	_ "net/http/pprof"

	"net/http"
)

type resp struct {
	message string
}

var partSize = 4 * 1024 * 1024
var maxSplit = 4
var server string

func postContent(tid string, pid int, content []byte, c chan int) {
	header := req.Header{
		"Accept": "application/json",
	}
	// param := req.Param{
	// 	"partContent": string(content),
	// }
	_, err := req.Post(server+"/"+tid+"/part/"+strconv.Itoa(pid), header, content)
	if err == nil {
		// fmt.Println(r)
	} else {
		log.Fatalf("post data to %s/%s/part/%s | %v", server, tid, strconv.Itoa(pid), err)
	}
	c <- pid
}

func done(tid string) {
	header := req.Header{
		"Accept": "application/json",
	}
	_, err := req.Post(server+"/"+tid+"/done", header)
	if err == nil {
		// fmt.Println(r)
	} else {
		log.Fatalf("end %s error", tid)
	}
}

func getSize(path string) int64 {
	fileInfo, err := os.Stat(path)
	if err != nil {
		log.Fatalf("get file size error %v", err)
	}
	fileSize := fileInfo.Size()
	return fileSize
}

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	if len(os.Args) == 3 {
		server = os.Args[1]
		uploadFileName := os.Args[2]
		f, e := os.Open(uploadFileName)
		defer f.Close()

		h := sha1.New()
		if e == nil {
			if _, err := io.Copy(h, f); err != nil {
				log.Fatalf("sha calc error | %v", err)
			} else {
				fileSHA1 := fmt.Sprintf("%x", h.Sum(nil))
				log.Printf(fileSHA1)
				header := req.Header{
					"Accept": "application/json",
				}
				fileSize := getSize(uploadFileName)
				param := req.Param{
					"fileName":    uploadFileName,
					"totalLength": fileSize,
				}
				r, err := req.Put(server+"/"+fileSHA1, header, param)
				if err != nil {
					log.Fatalf("Put error | %v", err)
				} else {
					fmt.Println(r)
					partLen := int(math.Ceil(float64(fileSize) / float64(partSize)))
					fmt.Printf("total %d block\n", partLen)
					c := make(chan int, 10)
					limit := make(chan int, 4)
					go func() {
						for pid := range c {
							fmt.Printf("process %d done\n", pid)
						}
						fmt.Printf("all done\n")
					}()

					w := &sync.WaitGroup{}
					for i := 0; i < partLen; i++ {
						w.Add(1)
						limit <- 1
						go func(i int) {
							defer w.Done()
							// buffer := make([]byte, partSize)
							buffer := BufferPool.Get().([]byte)
							len, err := f.ReadAt(buffer, int64(i*partSize))
							if err != nil {
								if len == 0 {
									log.Fatalf("read file part %d error | %v", i, err)
								} else {
									postContent(fileSHA1, i, (buffer[:len]), c)
								}
							} else {
								postContent(fileSHA1, i, (buffer[:len]), c)
							}

							BufferPool.Put(buffer)
							fmt.Printf("process %d/%d\n", i, partLen)
							<-limit
						}(i)
					}

					w.Wait()
					close(c)

					done(fileSHA1)
					fmt.Printf("upload done")
				}
			}
		} else {
			log.Fatalf("open file error | %v", e)
		}
	} else {
		log.Fatal("need file name")
	}
}

var (
	BufferPool sync.Pool
)

func init() {
	BufferPool.New = func() interface{} {
		return make([]byte, partSize)
	}
}
