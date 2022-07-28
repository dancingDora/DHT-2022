package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

type launchPackage struct {
	hashes [20]byte
	index  int
}

type downloadPackage struct {
	data  []byte
	index int
}

func Upload(fileName string, targetPath string, node *dhtNode) error {
	var blockNum int
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Println("[LAUNCH] Fail to Open the File")
		return err
	}
	var fileByteSize = len(content)
	if fileByteSize%PieceSize == 0 {
		blockNum = fileByteSize / PieceSize
	} else {
		blockNum = fileByteSize/PieceSize + 1
	}
	var pieces []byte = make([]byte, 20*blockNum)
	ch1 := make(chan int, blockNum+20)
	ch2 := make(chan launchPackage, blockNum+20)
	for i := 1; i <= blockNum; i++ {
		ch1 <- i
	}
	var flag1, flag2 bool = true, true
	for flag1 {
		select {
		case index := <-ch1:
			l := (index - 1) * PieceSize
			r := index + PieceSize
			if r > fileByteSize {
				r = fileByteSize
			}
			go uploadToNetwork(blockNum, node, index, content[l:r], ch1, ch2)
		case <-time.After(TimeWait):
			flag1 = false
		}
		time.Sleep(100 * time.Millisecond)
	}
	ch3 := make(chan string)
	ch4 := make(chan BencodeInfo)
	ch5 := make(chan string)
	go MakeTorrentFile(fileName, targetPath, ch3, ch4, ch5)
	for flag2 {
		select {
		case pack := <-ch2:
			index := pack.index
			copy(pieces[(index-1)*20:index*20], pack.hashes[:])
		default:
			flag2 = false
		}
	}
	ch3 <- string(pieces)
	torrentInfo := <-ch4
	torrentConten := <-ch5
	temp, err := torrentInfo.hash()
	infoHash := fmt.Sprintf("%x", temp)
	if err != nil {
		return err
	}
	ok := myself.Put(infoHash, torrentConten)
	if !ok {
		return errors.New("Put failed")
	}
	fmt.Println("Magnet Link Generates Successfully : ")
	fmt.Println("magnet: ?xt = urn:sha1" + infoHash + "&dn=" + torrentInfo.Name)
	return nil
}

func uploadToNetwork(totalSize int, node *dhtNode, index int, data []byte, ch1 chan int, ch2 chan launchPackage) {
	info := PieceInfo{index, data}
	hashKey, err := info.Hash()
	if err != nil {
		ch1 <- index
		return
	}
	var sx16 string = fmt.Sprintf("%x", hashKey)
	ok := (*node).Put(sx16, string(data))
	if !ok {
		fmt.Println("Fail to upload")
		ch1 <- index
		return
	}
	ch2 <- launchPackage{hashKey, index}
	fmt.Println("Uploading", float64(totalSize-len(ch1))/float64(totalSize)*100, "%...")
	return
}

func download(torrentName string, targetPath string, node *dhtNode) error {
	torrentFile, err := os.Open(torrentName)
	if err != nil {
		fmt.Println("[DOWNLOAD] Fail to Open the File")
		return err
	}
	BT, err := Open(torrentFile)
	if err != nil {
		fmt.Println("[DOWNLOAD] Fail to Unmarshal the File")
		return err
	}
	allInfo, err := BT.toTorrentFile()
	if err != nil {
		fmt.Println("[DOWNLOAD] Fail to Transform the BT")
		return err
	}
	fmt.Println("[DOWNLOAD] Start Downloading ", allInfo.Name)
	var content []byte = make([]byte, allInfo.Length)
	var blockSize int = len(allInfo.PieceHashes)
	ch1 := make(chan int, blockSize+5)
	ch2 := make(chan downloadPackage, blockSize+5)
	for i := 1; i <= blockSize; i++ {
		ch1 <- i
	}
	flag1, flag2 := true, true
	for flag1 {
		select {
		case index := <-ch1:
			go DownloadFromNetwork(blockSize, node, allInfo.PieceHashes[index-1], index, ch1, ch2)
		case <-time.After(TimeWait):
			fmt.Println("[DOWNLOAD]Download and Verify Finished")
			flag1 = false
		}
		time.Sleep(100 * time.Millisecond)
	}
	for flag2 {
		select {
		case pack := <-ch2:
			l := allInfo.PieceLength * (pack.index - 1)
			r := allInfo.PieceLength * (pack.index)
			if r > allInfo.Length {
				r = allInfo.Length
			}
			copy(content[l:r], pack.data[:])
		default:
			var dfile string
			if targetPath == "" {
				dfile = allInfo.Name
			} else {
				dfile = targetPath + "/" + allInfo.Name
			}
			err2 := ioutil.WriteFile(dfile, content, 0644)
			if err2 != nil {
				fmt.Println("[DOWNLOAD] Fail to Write in File with Error : ", err2)
				return err2
			}
			flag2 = false
		}
	}
	return nil
}

func DownloadFromNetwork(totalSize int, node *dhtNode, hashSet [20]byte, index int, ch1 chan int, ch2 chan downloadPackage) {
	var sx16 string = fmt.Sprintf("%x", hashSet)
	ok, rawData := (*node).Get(sx16)
	if !ok {
		fmt.Println("Fail to Download Try again :(")
		ch1 <- index
		return
	}
	info := PieceInfo{index, []byte(rawData)}
	verifyHash, err := info.Hash()
	if err != nil {
		fmt.Println("Fail to hash when verify due to ", err, " Try again :(")
		ch1 <- index // repeat to wait queue
		return
	}
	if verifyHash != hashSet {
		fmt.Println("Fail to Verify Try again :(")
		ch1 <- index
		return
	}
	ch2 <- downloadPackage{[]byte(rawData), index}
	fmt.Println("Downloading ", float64(totalSize-len(ch1))/float64(totalSize)*100, "% ....")
	return
}
