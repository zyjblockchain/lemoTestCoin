package store

import (
	"fmt"
	"github.com/boltdb/bolt"
	"log"
)

func main() {
	// 创建名为my.db的数据库
	db, err := bolt.Open("my.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	// 创建表
	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("myBucket"))
		if err != nil {
			return err
		}
		err = b.Put([]byte("name"), []byte("sandy"))
		return err
	})

	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("myBucket"))
		v := b.Get([]byte("name"))
		fmt.Println("name=", string(v))
		return nil
	})

}
