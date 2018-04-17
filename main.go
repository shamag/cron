package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/logger"
	"github.com/kataras/iris/middleware/recover"
	"github.com/robfig/cron"
)

func main() {

	app := iris.New()
	app.Use(recover.New())
	app.Use(logger.New())
	var netClient = &http.Client{
		Timeout: time.Second * 30,
	}
	c := cron.New()
	c.AddFunc("1 * * * * *", func() {
		fmt.Println("Every minute")
		netClient.Get("http://localhost:8080/api/v1/run")
	})
	c.Start()
	type Order struct {
		OrderID string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
	}

	v1 := app.Party("/api/v1")
	v1.Get("/run", func(ctx iris.Context) {
		var buf bytes.Buffer
		var response, _ = netClient.Get("http://localhost:8080/api/v1/csv")
		io.Copy(&buf, response.Body)
		reader := csv.NewReader(bufio.NewReader(&buf))
		defer response.Body.Close()
		var orders []Order
		for {
			line, error := reader.Read()
			if error == io.EOF {
				break
			} else if error != nil {
				log.Fatal(error)
			}
			orders = append(orders, Order{
				OrderID: line[0],
				Email:   line[1],
				Name:    line[2],
			})
		}
		orderJSON, _ := json.Marshal(orders)
		netClient.Post("http://localhost:8080/api/v1/orders", "application/json", bytes.NewBuffer(orderJSON))
		ctx.Write(orderJSON)
	})
	v1.Get("/csv", func(ctx iris.Context) {
		file := "./example.csv"
		ctx.SendFile(file, "file.csv")
	})
	v1.Post("/orders", func(ctx iris.Context) {
		var orders []Order
		err := ctx.ReadJSON(&orders)
		if err != nil {
			ctx.Writef("wrong format")
			return
		}
		fmt.Println("Received: %#+v\n", orders)
		ctx.Writef("Received: %#+v\n", orders)
	})

	app.Run(iris.Addr("localhost:8080"))

}
