package userdb

import (
	"encoding/binary"
)

//Set
func (udb *Userdb) SetFileDownloadSize(value uint64) error {
	k := []byte("dbtotalfilesdownloadeddata")
	v := make([]byte, 8)
	binary.PutUvarint(v, value)
	return udb.ldb.Put(k, v)
}

func (udb *Userdb) SetFileUploadSize(value uint64) error {
	k := []byte("dbtotalfilesuploadeddata")
	v := make([]byte, 8)
	binary.PutUvarint(v, value)
	return udb.ldb.Put(k, v)
}
func (udb *Userdb) SetTMDownloadSize(value uint64) error {
	k := []byte("dbtotaltmdownloadeddata")
	v := make([]byte, 8)
	binary.PutUvarint(v, value)
	return udb.ldb.Put(k, v)
}

//Get
func (udb *Userdb) GetFileDownloadSize() (uint64, error) {
	k := []byte("dbtotalfilesdownloadeddata")
	v, err := udb.ldb.Get(k)
	if(err != nil){
		return 0, err
	}
	value, _ := binary.Uvarint(v)
	return  value, nil
}

func (udb *Userdb) GetFileUploadSize() (uint64, error) {
	k := []byte("dbtotalfilesuploadeddata")
	v, err := udb.ldb.Get(k)
	if(err != nil){
		return 0, err
	}
	value, _ := binary.Uvarint(v)
	return  value, nil
}

func (udb *Userdb) GetTMDownloadSize() (uint64, error) {
	k := []byte("dbtotaltmdownloadeddata")
	v, err := udb.ldb.Get(k)
	if(err != nil){
		return 0, err
	}
	value, _ := binary.Uvarint(v)
	return  value, nil
}
