package main

import (
	"DBV_Script/dbv"
	"log/slog"
)

func main() {
	if err := dbv.Parser(); err != nil {
		slog.Error(err.Error())
		return
	}
	dbv.Run()
	// TestAddBarToWriteDataToFileFromIO()
}

// func TestAddBarToWriteDataToFileFromIO() {
// 	url := "http://172.17.0.1:8080/1.py"
// 	data, err := dbv.SenderGetAllRaw(url)
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
// 	size := int64(len(data))
// 	reader, err := dbv.SenderGetReader(url)
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
// 	path := "test/TestWriteRawDataToFile.txt"
// 	title := "1.py"
// 	err = dbv.AddBarToWriteDataToFileFromIO(path, reader, size, title)
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
// 	dbv.WaitBarFinish()
// }
