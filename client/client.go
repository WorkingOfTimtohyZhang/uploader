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
)

type resp struct {
	message string
}

var partSize = 4 * 1024 * 1024
var server string

func postContent(tid string, pid int, content string) {
	header := req.Header{
		"Accept": "application/json",
	}
	param := req.Param{
		"partContent": content,
	}
	_, err := req.Post(server+"/"+tid+"/part/"+strconv.Itoa(pid), header, param)
	if err == nil {
		// fmt.Println(r)
	} else {
		log.Fatalf("post data to %s/%s/part/%s | %v", server, tid, strconv.Itoa(pid), err)
	}
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
					buffer := make([]byte, partSize)
					for i := 0; i < partLen; i++ {
						len, err := f.ReadAt(buffer, int64(i*partSize))
						if err != nil {
							if len == 0 {
								log.Fatalf("read file part %d error | %v", i, err)
							} else {
								postContent(fileSHA1, i, string(buffer[:len]))
							}
						} else {
							postContent(fileSHA1, i, string(buffer[:len]))
						}
						fmt.Printf("process %d/%d", i, partLen)
					}
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
