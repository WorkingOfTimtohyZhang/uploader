package main

import "github.com/gin-gonic/gin"

type Task struct {
	id          string
	fileName    string
	totalLength int	
}

var taskMap map[string]Task{}

func main() {
	r := gin.Default()
	r.PUT("/:tid", func(c *gin.Context) {
		tid := c.Param("tid")
		fileName := c.PostForm("fileName")
		totalLength := c.PostForm("totalLength")
		taskMap[tid] = {
			tid: tid,
			fileName: fileName,
			totalLength: totalLength
		}
		c.JSON(200, gin.H{
			"message": "success"
		})
	})
	r.POST("/:tid/part/:pid", func(c *gin.Context) {
		tid := c.Param("tid")
		pid := c.Param("pid")
		partContent := c.PostForm("partContent")
		fileName := taskMap[tid].fileName
		c.JSON(200, gin.H{
			"message": "success"
		})
	})
	r.POST("/:tid/done", func(c *gin.Context) {
		tid := c.Param("tid")
		c.JSON(200, gin.H{
			"message": "success"
		})
	})
	r.Run() // listen and serve on 0.0.0.0:8080
}
