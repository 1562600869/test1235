package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"
)

const dataFile = "data.json"
const lockFile = "data.lock"

type Store struct {
	dir string
}

func NewStore() *Store {
	return &Store{dir: "."}
}

func (s *Store) dataPath() string {
	return filepath.Join(s.dir, dataFile)
}

func (s *Store) lockPath() string {
	return filepath.Join(s.dir, lockFile)
}

func (s *Store) lock() (*os.File, error) {
	f, err := os.OpenFile(s.lockPath(), os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("打开锁文件失败: %v", err)
	}
	err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX)
	if err != nil {
		f.Close()
		return nil, fmt.Errorf("获取锁失败: %v", err)
	}
	return f, nil
}

func (s *Store) unlock(f *os.File) {
	syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	f.Close()
}

func (s *Store) Load() (*Database, error) {
	path := s.dataPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &Database{
			Materials:      []Material{},
			InboundRecords: []InboundRecord{},
			Orders:         []Order{},
		}, nil
	}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取数据文件失败: %v", err)
	}
	var db Database
	if len(data) > 0 {
		if err := json.Unmarshal(data, &db); err != nil {
			return nil, fmt.Errorf("解析数据文件失败: %v", err)
		}
	}
	if db.Materials == nil {
		db.Materials = []Material{}
	}
	if db.InboundRecords == nil {
		db.InboundRecords = []InboundRecord{}
	}
	if db.Orders == nil {
		db.Orders = []Order{}
	}
	return &db, nil
}

func (s *Store) Save(db *Database) error {
	data, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化数据失败: %v", err)
	}
	tmpPath := s.dataPath() + ".tmp"
	if err := ioutil.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("写入临时文件失败: %v", err)
	}
	if err := os.Rename(tmpPath, s.dataPath()); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("原子替换文件失败: %v", err)
	}
	return nil
}

type TxFunc func(db *Database) error

func (s *Store) Transaction(fn TxFunc) (*Database, error) {
	lockF, err := s.lock()
	if err != nil {
		return nil, err
	}
	defer s.unlock(lockF)

	db, err := s.Load()
	if err != nil {
		return nil, err
	}

	if err := fn(db); err != nil {
		return nil, err
	}

	if err := s.Save(db); err != nil {
		return nil, err
	}

	return db, nil
}

func FindMaterial(db *Database, id string) *Material {
	for i := range db.Materials {
		if db.Materials[i].ID == id {
			return &db.Materials[i]
		}
	}
	return nil
}

func FindOrder(db *Database, id string) *Order {
	for i := range db.Orders {
		if db.Orders[i].ID == id {
			return &db.Orders[i]
		}
	}
	return nil
}

func ValidateDate(date string) error {
	if len(date) != 10 {
		return errors.New("日期格式应为 YYYY-MM-DD")
	}
	if date[4] != '-' || date[7] != '-' {
		return errors.New("日期格式应为 YYYY-MM-DD")
	}
	for i, c := range date {
		if i == 4 || i == 7 {
			continue
		}
		if c < '0' || c > '9' {
			return errors.New("日期格式应为 YYYY-MM-DD")
		}
	}
	year := (int(date[0]-'0') * 1000) + (int(date[1]-'0') * 100) + (int(date[2]-'0') * 10) + int(date[3]-'0')
	month := (int(date[5]-'0') * 10) + int(date[6]-'0')
	day := (int(date[8]-'0') * 10) + int(date[9]-'0')
	if month < 1 || month > 12 {
		return errors.New("月份必须在 1-12 之间")
	}
	if day < 1 || day > 31 {
		return errors.New("日期必须在 1-31 之间")
	}
	_ = year
	return nil
}

func ValidateMonth(month string) error {
	if len(month) != 7 {
		return errors.New("月份格式应为 YYYY-MM")
	}
	if month[4] != '-' {
		return errors.New("月份格式应为 YYYY-MM")
	}
	for i, c := range month {
		if i == 4 {
			continue
		}
		if c < '0' || c > '9' {
			return errors.New("月份格式应为 YYYY-MM")
		}
	}
	m := (int(month[5]-'0') * 10) + int(month[6]-'0')
	if m < 1 || m > 12 {
		return errors.New("月份必须在 1-12 之间")
	}
	return nil
}
