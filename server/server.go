package main

import (
	"crypto/sha1"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"bufio"
	"github.com/gin-gonic/gin"
)

type pTask struct {
	id          string
	fileName    string
	totalLength int
	pFile       *os.File
}

var taskMap = make(map[string]pTask)
var partSize = 4 * 1024 * 1024

func main() {
	r := gin.Default()
	r.PUT("/:tid", func(c *gin.Context) {
		tid := c.Param("tid")
		fileName := c.PostForm("fileName")
		sTotalLength := c.PostForm("totalLength")
		iTotalLength, iErr := strconv.Atoi(sTotalLength)
		if iErr == nil {
			outputFile, outputError := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0666)
			if outputError == nil {
				ptask := pTask{tid, fileName, iTotalLength, outputFile}
				taskMap[tid] = ptask
				c.JSON(200, gin.H{
					"message": "success",
				})
			} else {
				log.Fatal("create file")
			}
		} else {
			log.Fatal("[totalLength]strconv error")
		}
	})
	r.POST("/:tid/part/:pid", func(c *gin.Context) {
		tid := c.Param("tid")
		sPid := c.Param("pid")

		closer := c.Request.Body
		iPid, err := strconv.Atoi(sPid)
		if err != nil {
			log.Fatal(err)
		}

		reader := bufio.NewReaderSize(closer, 4094)
		targetFile := taskMap[tid].pFile

		start := int64(iPid * partSize)
		p := make([]byte, 1024)
		for {
			l, err := reader.Read(p)
			if l > 0 {
				n, err := targetFile.WriteAt(p[:l], start)
				start = start + int64(n)
				if err != nil {
					log.Fatal(err)
				}
			}
			if err != nil {
				if err == io.EOF {
					break
				}
				log.Fatal(err)
			}
		}

		c.JSON(200, gin.H{
			"message": "write success",
		})

	})
	r.POST("/:tid/done", func(c *gin.Context) {
		tid := c.Param("tid")
		// 客户端控制，假定这里传输已经结束了
		h := sha1.New()
		taskMap[tid].pFile.Sync()
		taskMap[tid].pFile.Seek(0, 0)
		if _, err := io.Copy(h, taskMap[tid].pFile); err != nil {
			log.Fatalf("calc sha1 | %v", err)
		} else {
			fileSHA1 := fmt.Sprintf("%x", h.Sum(nil))
			if fileSHA1 == tid {
				outputFile, outputError := os.OpenFile(taskMap[tid].fileName+".sha1", os.O_WRONLY|os.O_CREATE, 0666)
				if outputError == nil {
					defer outputFile.Close()
					outputFile.WriteString(fileSHA1)
					closeErr := taskMap[tid].pFile.Close()
					if closeErr == nil {
						delete(taskMap, tid)
						c.JSON(200, gin.H{
							"message": "sha1 checked",
						})
					} else {
						log.Fatalf("close file | %v", closeErr)
					}
				} else {
					log.Fatalf("create sha1 file | %v", outputError)
				}

			}

		}

	})
	r.Run() // listen and serve on 0.0.0.0:8080
}
