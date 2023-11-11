package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"log"

	stat "gonum.org/v1/gonum/stat"

	"github.com/gorilla/websocket"
)

var (
	urlFlag     = flag.String("url", "ws://127.0.0.1:8080/ws", "File containing data to POST. Remember also to set -T")
	requests    = flag.Int("n", 1000000, "Number of requests to perform")
	concurrency = flag.Int("c", 1000, "Number of multiple requests to make at a time")
	// timeout      = flag.Int("s", 30, "Seconds to max. wait for each response")
	// postfile     = flag.String("p", "", "File containing data to POST. Remember also to set -T")
	// putfile      = flag.String("u", "", "File containing data to PUT. Remember also to set -T")
	// content_type = flag.String("T", "text/plain", "Content-type header to use for POST/PUT data, eg. "+
	// 	"\n'application/x-www-form-urlencoded'")
	// verbosity  = flag.Bool("v", false, "How much troubleshooting info to print")
	// print_info = flag.Bool("w", false, "Print out results in HTML tables")
	// cookie     = flag.String("C", "", "Add cookie, eg. 'Apache=1234'. (repeatable)")
	// header     = flag.String("H", "", "Add Arbitrary header line, eg. 'Accept-Encoding: gzip' "+
	// 	"\nInserted after all normal header lines. (repeatable)")
	// version        = flag.Bool("V", false, "Print version number and exit")
	keep_alive = flag.Bool("k", true, "Use HTTP KeepAlive feature")
	// collected_file = flag.String("g", "", "Output collected data to gnuplot format file.")
	// csv_file       = flag.String("e", "", "Output CSV file with percentages served")
	// receive_error  = flag.Bool("r", false, "Don't exit on socket receive errors.")
	binary = flag.Bool("b", false, "Send data use websocket binary mode")
	usage  = flag.Bool("h", false, "Display usage information (this message)")
)

var (
	addr              string
	wgWorker          sync.WaitGroup
	wgReceiver        sync.WaitGroup
	connectionChan    chan *websocket.Conn
	requestResultChan chan *RequestResults
)

// RequestResults 请求结果
type RequestResults struct {
	Time          int64 // 请求时间 纳秒
	IsSucceed     bool  // 是否请求成功
	ErrCode       int   // 错误码
	SendedBytes   int
	ReceivedBytes int

	ConnectTime float64
	SentTime    float64
	RecvTime    float64
	TotalTime   float64
}

// result
var (
	spendedTime time.Duration

	completeRequests int
	failedRequests   int
	totalSecessTime  int64

	totalDataSent int64
	totalDataRecv int64

	ConnectTimes []float64
	SentTimes    []float64
	RecvTimes    []float64
	TotalTimes   []float64
)

func main() {
	flag.Usage = func() {
		fmt.Println("Usage: " + os.Args[0] + " [options] -url [http://]hostname[:port]/path")
		fmt.Println("Options are:")
		flag.PrintDefaults()
	}
	flag.Parse()

	printErrorIfTrue := func(b bool, tips string) {
		if b {
			fmt.Println(tips)
		}
	}
	if *usage || *requests == 0 || len(*urlFlag) == 0 {
		printErrorIfTrue(*concurrency > *requests,
			"wb: Cannot use concurrency level greater than total number of requests")
		printErrorIfTrue(len(*urlFlag) == 0,
			"wb: invalid URL")
		flag.Usage()
		return
	}

	Dispose()
	printResult()
}

func printResult() {

	u, _ := url.Parse(addr)
	fmt.Printf("Server Scheme:          %v\n", u.Scheme)
	fmt.Printf("Server Hostname:        %v\n", u.Hostname())
	fmt.Printf("Server Port:            %v\n\n", u.Port())

	fmt.Printf("Document Path:          %v\n\n", u.Path)
	// fmt.Printf("Document Length:        73 bytes\n")

	fmt.Printf("Concurrency Level:      %v\n", *concurrency)
	fmt.Printf("Time taken for tests:   %v\n", spendedTime)
	fmt.Printf("Complete requests:      %v\n", completeRequests)
	fmt.Printf("Failed requests:        %v\n", failedRequests)
	fmt.Printf("Total data sent:        %v bytes\n", totalDataSent)
	fmt.Printf("Total data recv:        %v bytes\n", totalDataRecv)

	// fmt.Printf("HTML transferred:       438000 bytes\n")
	fmt.Printf("Requests per second:    %.2f [#/sec] (mean)\n",
		float32(completeRequests)/float32(float32(spendedTime)/float32(time.Second)))
	fmt.Printf("Time per request:       %.2f [ms] (mean)\n",
		float32(totalSecessTime)/float32(completeRequests-failedRequests)/float32(time.Millisecond))
	// fmt.Printf("Time per request:       0.628 [ms] (mean, across all concurrent requests)\n")
	fmt.Printf("Transfer rate:          %v [Kbytes/sec] received\n"+
		"                        %v kb/s sent\n\n",
		float32(totalDataRecv)/(float32(spendedTime)/float32(time.Second))/1000,
		float32(totalDataSent)/(float32(spendedTime)/float32(time.Second))/1000)

	fmt.Printf("Connection Times (ms)\n" +
		"\t\t\t\t\tmin  mean[+/-sd] median   max\n")
	sort.Float64s(ConnectTimes)
	mean, std := stat.MeanStdDev(ConnectTimes, nil)
	median := stat.Quantile(0.5, stat.Empirical, ConnectTimes, nil)
	fmt.Printf("Connect:\t %.0f \t\t %.0f \t %.0f \t %.0f \t %0.f\n",
		ConnectTimes[0], mean, std,
		median, ConnectTimes[len(ConnectTimes)-1])
	sort.Float64s(SentTimes)
	mean, std = stat.MeanStdDev(SentTimes, nil)
	median = stat.Quantile(0.5, stat.Empirical, SentTimes, nil)
	fmt.Printf("Sent:\t\t\t %.0f \t\t %.0f \t %.0f \t %.0f \t %0.f\n",
		SentTimes[0], mean, std,
		median, SentTimes[len(SentTimes)-1])
	sort.Float64s(RecvTimes)
	mean, std = stat.MeanStdDev(RecvTimes, nil)
	median = stat.Quantile(0.5, stat.Empirical, RecvTimes, nil)
	fmt.Printf("Recv:\t\t\t %.0f \t\t %.0f \t %.0f \t %.0f \t %0.f\n",
		RecvTimes[0], mean, std,
		median, RecvTimes[len(RecvTimes)-1])
	sort.Float64s(TotalTimes)
	mean, std = stat.MeanStdDev(TotalTimes, nil)
	median = stat.Quantile(0.5, stat.Empirical, TotalTimes, nil)
	fmt.Printf("Total:\t\t %.0f \t\t %.0f \t %.0f \t %.0f \t %0.f\n\n",
		TotalTimes[0], mean, std,
		median, TotalTimes[len(TotalTimes)-1])
	// fmt.Printf("ERROR: The median and mean for the initial connection time are more than twice the standard\n" +
	// 	"deviation apart. These results are NOT reliable.\n")

	leng := len(TotalTimes)
	fmt.Printf("Percentage of the requests served within a certain time (ms)\n")
	fmt.Printf("  50%%    %.0f\n", TotalTimes[int(leng*50/100)-1])
	fmt.Printf("  66%%    %.0f\n", TotalTimes[int(leng*66/100)-1])
	fmt.Printf("  75%%    %.0f\n", TotalTimes[int(leng*75/100)-1])
	fmt.Printf("  80%%    %.0f\n", TotalTimes[int(leng*80/100)-1])
	fmt.Printf("  90%%    %.0f\n", TotalTimes[int(leng*90/100)-1])
	fmt.Printf("  95%%    %.0f\n", TotalTimes[int(leng*95/100)-1])
	fmt.Printf("  98%%    %.0f\n", TotalTimes[int(leng*98/100)-1])
	fmt.Printf("  99%%    %.0f\n", TotalTimes[int(leng*99/100)-1])
	fmt.Printf(" 100%%    %.0f (longest request)\n", TotalTimes[int(leng*100/100)-1])
}

// Dispose 处理函数
func Dispose() {

	connectionChan = make(chan *websocket.Conn, 1000)
	go dialer()

	requestResultChan = make(chan *RequestResults, 1000)
	wgReceiver.Add(1)
	go receiver()

	addr = *urlFlag
	if !strings.HasSuffix(addr, "ws") {
		addr = "ws://" + addr
	}

	start := time.Now()
	u, _ := url.Parse(addr)
	fmt.Printf("Benchmarking %v (be patient)\n", u.Hostname())

	for i := 0; i < *concurrency; i++ {
		wgWorker.Add(1)

		times := (int)(*requests / (*concurrency))
		if i == 0 && (*requests)%(*concurrency) > 0 {
			times += (*requests) % (*concurrency)
		}
		data := []byte(`{"service": "echo", "method": "Echo", "EchoRequest": {"name":"tester"}}`)
		go worker(i, addr, times, data)

		// // 注意:时间间隔太短会出现连接失败的报错 默认连接时长:20毫秒(公网连接)
		// time.Sleep(5 * time.Millisecond)
	}

	// 等待所有的数据都发送完成
	wgWorker.Wait()
	spendedTime = time.Since(start)

	// 数据全部处理完成了
	wgReceiver.Wait()
}

func receiver() {
	for i := 1; i < *requests+1; i++ {
		rr := <-requestResultChan
		completeRequests++
		if !rr.IsSucceed {
			failedRequests++
		} else {
			totalSecessTime += rr.Time
			totalDataSent += int64(rr.SendedBytes)
			totalDataRecv += int64(rr.ReceivedBytes)

			ConnectTimes = append(ConnectTimes, rr.ConnectTime/float64(time.Millisecond))
			RecvTimes = append(RecvTimes, rr.RecvTime/float64(time.Millisecond))
			SentTimes = append(SentTimes, rr.SentTime/float64(time.Millisecond))
			TotalTimes = append(TotalTimes, rr.TotalTime/float64(time.Millisecond))
		}

		patch := int(*requests) / 10
		if i != 0 && (i%patch == 0 || i == *requests) {
			fmt.Printf("Completed %v requests\n", i)
		}
	}
	fmt.Println("")
	wgReceiver.Done()
}

func dialer() {
	// count := *requests
	// if *keep_alive {
	// 	count = *concurrency
	// }
	// for i := 0; i < count; i++ {
	// 	c, _, err := websocket.DefaultDialer.Dial(addr, nil)
	// 	if err != nil {
	// 		log.Fatalf("dial error fff: %v", err)
	// 	}
	// 	connectionChan <- c
	// 	time.Sleep(time.Millisecond)
	// }
}

func worker(id int, addr string, times int, data []byte) {
	var c *websocket.Conn
	var err error
	if *keep_alive {
		// c = <-connectionChan
		c, _, err = websocket.DefaultDialer.Dial(addr, nil)
		if err != nil {
			log.Fatalf("{%v-*} dial error: %v", id, err)
		}
		defer c.Close()
	}
	for i := 0; i < times; i++ {
		// var c *websocket.Conn
		// var err error

		start := time.Now()

		var message []byte
		var tc time.Duration
		var timeBegin time.Time

		messageType := websocket.TextMessage
		if *binary {
			messageType = websocket.BinaryMessage
		}

		rr := &RequestResults{
			IsSucceed: false,
		}

		if c == nil {
			// c = <-connectionChan
			c, _, err = websocket.DefaultDialer.Dial(addr, nil)
			if err != nil {
				log.Fatalf("{%v-%v} dial error: %v", id, i, err)
				// panic(err.Error() + "   " + fmt.Sprint(id))
			}
			defer c.Close()
		}
		rr.ConnectTime = float64(time.Since(start))

		timeBegin = time.Now()
		err = c.WriteMessage(messageType, data)
		if err != nil {
			log.Fatalf("write error: %v", err)
			goto SEND_TO_CHAN
		}
		rr.SentTime = float64(time.Since(timeBegin))

		timeBegin = time.Now()
		_, message, err = c.ReadMessage()
		if err != nil {
			log.Fatalf("read error: %v", err)
			goto SEND_TO_CHAN
		}
		rr.RecvTime = float64(time.Since(timeBegin))
		rr.TotalTime = float64(time.Since(start))

		// log.Printf("recv data: %s", string(message))

		if !(*keep_alive) {
			c.Close()
			c = nil
		}

		tc = time.Since(start)
		// log.Printf("time cost = %v\n", tc)

		rr.Time = int64(tc)
		rr.IsSucceed = true
		rr.SendedBytes = len(data)
		rr.ReceivedBytes = len(message)
	SEND_TO_CHAN:
		requestResultChan <- rr

	}
	wgWorker.Done()
}
