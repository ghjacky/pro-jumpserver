package wsserver

import (
	"bufio"
	"encoding/json"
	socketio "github.com/googollee/go-socket.io"
	"os"
	"sync"
	"time"
	"zeus/models"
)

const (
	keyCtrlC = 3
	keyCtrlD = 4
	keySpace = 32
	keyLeft  = 68
	keyRight = 67
	keyUp    = 65
	keyDown  = 66
)

var signal = make(chan int, 0)

func play(filepath string, conn socketio.Conn) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	var donec = make(chan bool)
	f, err := os.OpenFile(filepath, os.O_RDONLY, 0644)
	if err != nil {
		//log.Fatalf("couldn't open file %s", filepath)
	}
	defer func() {
		if err := f.Close(); err != nil {
			//log.Printf("couldn't close file %s", filepath)
		}
	}()
	var allRecoredBytesData = []models.Event{}
	scanner := bufio.NewScanner(f)

	// 此处有必要优化（当文件过大时，可能造成内存占用量陡增）
	for scanner.Scan() {
		recordItem := models.Event{}
		line := scanner.Text()
		//log.Printf(fmt.Sprintf("%d", sleeping))
		lineBytes := []byte(line)
		if err != nil {
			//log.Fatalf("couldn't parse line string from file %s", filepath)
		}
		//log.Printf("got line string: %s", line)
		if err := json.Unmarshal(lineBytes, &recordItem); err != nil {
			//log.Fatalf("couldn't parse line bytes: %s", err.Error())
		}
		allRecoredBytesData = append(allRecoredBytesData, recordItem)
	}

	//log.Printf("start play")
	os.Stdout.Write([]byte("\x1bc"))
	go func() {
		once := sync.Once{}
		defer wg.Done()
		//log.Printf("start scan file")
		lastTime := int64(0)
		sleeping := 0 * time.Nanosecond
		var sig = make(chan int, 0)
		conn.Emit("play_begin", 1)
		allRecoredBytesData = append(allRecoredBytesData, models.Event{})
		for i, recordItem := range allRecoredBytesData {
			if len(allRecoredBytesData) == i+1 {
				conn.Emit("playing", "^END^")
				//conn.Emit("play_end", 1)
				break
			}
			select {
			case s := <-signal:
				switch s {
				case 1:
					wait(donec)
				case 2:
					donec <- true
				case 0:
					//conn.Emit("playing", "^END^")
					return
				}
			default:
				currentTime := recordItem.Timestamp
				once.Do(func() {
					lastTime = recordItem.Timestamp
				})
				sleeping = time.Duration(currentTime-lastTime) * time.Nanosecond
				// 播放速度控制
				go func() {
					time.Sleep(sleeping / 1)
					<-sig
				}()
				sig <- 1
				// 此处将记录的数据写入 stdout 或 其他writer
				//_, err := os.Stdout.Write(recordItem.Data)
				conn.Emit("playing", string(recordItem.Data))
				lastTime = recordItem.Timestamp
				if err != nil {
					continue
				}
			}
		}
	}()
	wg.Wait()
}

func wait(c chan bool) {
	<-c
}

func playStop(s int) {
	signal <- s
}
