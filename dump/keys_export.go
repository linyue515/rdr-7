package dump

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/919927181/rdr/decoder"
	"github.com/dustin/go-humanize"
	"github.com/urfave/cli"
)

/**
 * @file about: 将统计结果输出到STDOUT 或 File
 */

// keys is function for command `keys`
// output all keys in rdb file(s) get from args
func Export_All_keys(cli *cli.Context) {
	if cli.NArg() < 1 {
		fmt.Fprintln(cli.App.ErrWriter, " requires at least 1 argument")
		return
	}
	cst := time.FixedZone("CST", 8*60*60)
	currentTime := time.Now().In(cst)
	exportFileName := fmt.Sprintf("rdb-all-keys-%s.txt", currentTime.Format("20060102-150405"))

	// 创建或打开文件，追加模式; os.O_APPEND|os.O_CREATE|os.O_WRONLY表示以追加模式打开文件，如果文件不存在则创建，并且只允许写入0644
	file, err := os.OpenFile(exportFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("create report file %s failed!", exportFileName)
		return
	}
	defer file.Close()


	//使用 bufio.Writer 可减少系统调用，提升性能,适合大量数据
	fileWriter := bufio.NewWriter(file)
	defer fileWriter.Flush() // WriteString 后内容暂存在缓冲区，必须调用 Flush() 才会真正写入磁盘文件


	fmt.Fprintln(cli.App.Writer, "export start ...")
	start_milliseconds := time.Now().UnixMilli()

	//{"db":0,"key":"student:1:name","expiration":"2025-12-28T06:48:05.844+08:00","size":104,"type":"string","encoding":"string","value":"wangxing"},
	//db,key,type,encoding,size,humanizeSize,numOfElem,expiration
    // 遍历传参 rdb file
	nArgs := cli.NArg()
	for i := 0; i < nArgs; i++ {
		rdb_file := cli.Args().Get(i)
		fmt.Fprintln(cli.App.Writer, "rdb parse start ...")
		rdbDecoder := decoder.NewDecoder()
		go Decode(cli, rdbDecoder, rdb_file)

		fmt.Fprintln(cli.App.Writer, "write to file start ...")
		fileWriter.WriteString("key,type,encoding,size,humanizeSize,numOfElem,expiration,lruIdle,lfuFreq,db\n")
		for e := range rdbDecoder.Entries {
			expiryStr := ""  //key的过期时间
			if e.Expiration > 0 {
				expiryStr = time.Unix(0, e.Expiration*int64(time.Millisecond)).Format("2006-01-02 15:04:05")
			}
			lruIdleStr:=""  //key的最后一次访问时间，maxmemory-policy配置的淘汰策略是volatile-lru或allkeys-lru，它记录的是Key的最后一次访问时间
			if e.LruIdle > 0 {
				lruIdleStr = time.Unix(0, int64(e.LruIdle)*int64(time.Millisecond)).Format("2006-01-02 15:04:05")
			}

			rowWords := []string{
				e.Key,
				e.Type,
				e.Encoding,
				strconv.FormatUint(e.Bytes, 10),
				humanize.Bytes(e.Bytes),
				strconv.FormatUint(e.NumOfElem, 10),
				expiryStr,
				lruIdleStr,
				strconv.Itoa(e.LfuFreq),
				strconv.Itoa(e.Db),
			}

			fileWriter.WriteString(strings.Join(rowWords, ", ")+"\n")
		}

	}

	end_milliseconds := time.Now().UnixMilli()
	fmt.Printf("export finished, all keys have write to file ./%s.\n", exportFileName)
	fmt.Printf("time use %d ms.\n", (end_milliseconds - start_milliseconds))

}

