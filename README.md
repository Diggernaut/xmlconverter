# xmlconverter

This repo provides fast way to convert your mongo collection or json file to xml.

Install `go get github.com/Diggernaut/xmlconverter`

Usage:

`go build -o xmlconverter main.go`

`./xmlconverter -db="somedb" -col="somecol" -dbaddr="127.0.0.1" > out.xml`

`./xmlconverter -file="test.json" > out.xml`


or with go run

`go run main.go -db="somedb" -col="somecol" -dbaddr="127.0.0.1" > out.xml`

`go run main.go -file="test.json" > out.xml`
