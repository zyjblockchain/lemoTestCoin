// 存储LemoAddress对应的交易的时间戳
package store

import (
	"github.com/boltdb/bolt"
	"log"
	"strconv"
)

// 存储 key = LemoAddress,value = tx.Expiration
func Putdb(LemoAddress string, expiration uint64) error {
	db, err := bolt.Open("lemo.db", 0600, nil)
	if err != nil {
		log.Println("open db file error:", err)
		return err
	}
	defer db.Close()
	// 创建/打开一个表
	err = db.Update(func(tx *bolt.Tx) error {
		// 打开一个表，如果不存在则创建
		b, err := tx.CreateBucketIfNotExists([]byte("bucket"))
		if err != nil {
			log.Println("open bucket error:", err)
			return err
		}
		// uint64转string
		str := strconv.FormatUint(expiration, 10)
		// 对表进行操作,这里是put操作
		err = b.Put([]byte(LemoAddress), []byte(str))
		return err
	})

	return err
}

// 获得db中LemoAddress对应的时间戳
func Getdb(LemoAddress string) (uint64, error) {
	var time uint64
	db, err := bolt.Open("lemo.db", 0600, nil)
	if err != nil {
		log.Println("open db file error:", err)
		return 0, err
	}
	defer db.Close()
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("bucket"))
		byteTime := b.Get([]byte(LemoAddress))
		// 如果db没有这个key,则返回nil
		if byteTime == nil {
			time = 0
			return nil
		} else {
			// []byte 转 uint64
			time, err = strconv.ParseUint(string(byteTime), 10, 64)
			if err != nil {
				log.Println("parse uint64 error:", err)
				return err
			}
		}
		return nil
	})

	return time, err
}
